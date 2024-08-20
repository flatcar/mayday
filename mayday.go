package main

import (
	"archive/tar"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/flatcar/mayday/mayday"
	"github.com/flatcar/mayday/mayday/plugins/command"
	"github.com/flatcar/mayday/mayday/plugins/docker"
	"github.com/flatcar/mayday/mayday/plugins/file"
	"github.com/flatcar/mayday/mayday/plugins/journal"
	"github.com/flatcar/mayday/mayday/plugins/symlink"
	mtar "github.com/flatcar/mayday/mayday/tar"
	"github.com/flatcar/mayday/mayday/tarable"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"strings"
)

const (
	dirPrefix     = "/mayday"
	configDefault = "/etc/mayday/default.json"
)

type Config struct {
	Files       []File      `mapstructure:"files"`
	Directories []Directory `mapstructure:"directories"`
	Commands    []Command   `mapstructure:"commands"`
}

type File struct {
	Name string `mapstructure:"name"`
	Link string `mapstructure:"link"`
}

type Directory struct {
	Name string `mapstructure:"name"`
	Link string `mapstructure:"link"`
}

type Command struct {
	Args []string `mapstructure:"args"`
	Link string   `mapstructure:"link"`
}

func openFile(f File) (*file.MaydayFile, error) {
	content, err := os.Open(f.Name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(f.Name)
	if err != nil {
		return nil, err
	}

	header, err := tar.FileInfoHeader(fi, f.Name)
	header.Name = f.Name
	if err != nil {
		return nil, err
	}

	opened := file.New(content, header, f.Name, f.Link)
	return opened, nil
}

func main() {
	pflag.StringP("config", "c", configDefault, "path configuration file (in place of profile)")
	pflag.BoolP("danger", "d", false, "collect potentially sensitive information (ex, container logs)")
	pflag.StringP("profile", "p", "", "set of data to be collected (default: everything)")
	pflag.StringP("output", "o", "", "output file (default: /tmp/mayday-{hostname}-{current time}.tar.gz)")

	// binds cli flag "danger" to viper config danger, etc.
	viper.BindPFlag("danger", pflag.Lookup("danger"))
	viper.BindPFlag("config", pflag.Lookup("config"))
	viper.BindPFlag("output", pflag.Lookup("output"))
	viper.BindPFlag("profile", pflag.Lookup("profile"))
	// cli arg takes precendence over anything in config files
	pflag.Parse()

	// can't define both config and profile at the same time
	if viper.GetString("config") != configDefault && viper.GetString("profile") != "" {
		log.Fatal("--profile option cannot be used with --config option. (Point --config to full path of file.)")
	}

	if viper.GetString("config") == configDefault {
		// Flatcar config location
		viper.AddConfigPath("/usr/share/mayday")
	}

	path_with_filename := strings.Split(viper.GetString("config"), "/")
	path := path_with_filename[:len(path_with_filename)-1]

	if len(path) == 0 {
		viper.AddConfigPath(".")
	} else {
		viper.AddConfigPath(strings.Join(path, "/"))
	}

	if viper.GetString("profile") != "" {
		viper.SetConfigName(viper.GetString("profile"))
	} else {
		filename := path_with_filename[len(path_with_filename)-1]
		filename_no_ext := strings.Split(filename, ".json")[0]
		viper.SetConfigName(filename_no_ext)
	}

	err := viper.ReadInConfig()
	if err != nil {
		// viper returns an `unsupported config type ""` error if it can't find a file
		// https://github.com/spf13/viper/issues/210
		if strings.HasSuffix(err.Error(), `Type ""`) {
			log.Fatalf("Could not find configuration file in %s",
				viper.GetString("config"))
		}
		log.Printf("Error reading configuration file.")
		log.Fatalf("Fatal error reading config: %s\n", err)
	}

	log.Printf("loading config from %s", viper.ConfigFileUsed())

	var tarables []tarable.Tarable

	var C Config

	// fill C with configuration data
	viper.Unmarshal(&C)

	journals, err := journal.List()
	if err != nil {
		log.Fatal(err)
	}

	containers, dockerLogs, err := docker.GetContainers()
	if err != nil {
		log.Println("Could not connect to docker. Verify mayday has permissions to read /var/lib/docker.")
		log.Printf("Connection error: %s", err)
	}

	df := make([]File, 0) // list of files in directories to include
	for _, d := range C.Directories {
		log.Printf("reading directory %s\n", d)
		err := filepath.Walk(d.Name, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil {
				log.Printf("error opening path %q: %v\n", path, err)
				return err
			}
			if info.Mode().IsRegular() {
				mf := File{Name: path}
				df = append(df, mf)
			}
			return nil
		})

		if err == nil {
			sl := symlink.New(d.Name, d.Link)
			tarables = append(tarables, sl)
		}
	}

	for _, f := range df {
		mf, err := openFile(f)
		if err != nil {
			log.Printf("error opening %s: %s\n", f.Name, err)
		} else {
			defer mf.Close()
			tarables = append(tarables, mf)
		}
	}

	for _, f := range C.Files {
		mf, err := openFile(f)
		if err != nil {
			log.Printf("error opening %s: %s\n", f.Name, err)
		} else {
			defer mf.Close()
			tarables = append(tarables, mf)
		}
	}

	for _, c := range C.Commands {
		tarables = append(tarables, command.New(c.Args, c.Link))
	}

	for _, j := range journals {
		tarables = append(tarables, j)
	}

	for _, c := range containers {
		tarables = append(tarables, c)
	}

	for _, l := range dockerLogs {
		tarables = append(tarables, l)
	}

	now := time.Now().Format("200601021504.999999999")

	outputFile := viper.GetString("output")
	if outputFile == "" {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknownhost"
		}
		ws := os.TempDir() + dirPrefix + "-" + hostname + "-" + now
		outputFile = ws + ".tar.gz"
	}

	var t mtar.Tar

	tarfile, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer tarfile.Close()
	t.Init(tarfile, now)

	mayday.Run(t, tarables)
	t.Close()

	log.Printf("Output saved in %v\n", outputFile)
	log.Printf("All done!")

	return
}
