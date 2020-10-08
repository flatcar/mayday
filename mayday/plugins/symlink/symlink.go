package symlink

import (
	"archive/tar"
	"bytes"

	"github.com/flatcar-linux/mayday/mayday/tarable"
)

type MaydaySymlink struct {
	name string // the name of the file/directory on the filesystem
	link string // a link to make in the root of the tarball
}

func New(n string, l string) *MaydaySymlink {
	f := new(MaydaySymlink)
	f.name = n
	f.link = l

	return f
}

func (f *MaydaySymlink) Content() *bytes.Buffer {
	return nil
}

func (f *MaydaySymlink) Header() *tar.Header {
	var b bytes.Buffer
	return tarable.Header(&b, f.Name())
}

func (f *MaydaySymlink) Name() string {
	return f.name
}

func (f *MaydaySymlink) Link() string {
	return f.link
}
