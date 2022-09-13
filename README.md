# mayday

...man overboard!

```


     ___ ___   ____  __ __  ___     ____  __ __
    |   |   | /    ||  |  ||   \   /    ||  |  |
    | _   _ ||  o  ||  |  ||    \ |  o  ||  |  |
    |  \_/  ||     ||  ~  ||  D  ||     ||  ~  |
    |   |   ||  _  ||___, ||     ||  _  ||___, |
    |   |   ||  |  ||     ||     ||  |  ||     |
    |___|___||__|__||____/ |_____||__|__||____/




~~~~~~~~~~~~~~~~~~~~~~~~~\o/~~~~~~~~~~~~~~~~~~~~~~~~

                                   ><>
               <><  
```

## overview
Mayday is a tool to simplify gathering support information.  It is built in the
spirit of sysreport, son of sysreport (sosreport), and similar support tools.  
Mayday gathers information about the configuration, hardware, and running state
of a system.


## goals
The goals of mayday are:
  * simplify gathering information about a running system into a single command
  * collect information into one single file to be transferred to support staff
  * when possible the file should be small enough to be sent via email (<10MB)
  * *not* collect sensitive information like crypto keys, password hashes, etc
  * extensible through plugin system

## usage
In it's most simplistic form, all a user needs to do is run `mayday`:

```
$ mayday
```

This will collect a basic set of data and emit it in a tar archive for
transmission to a systems administrator, site reliability engineer, or support
technician for further troubleshooting.

In addition, more data can be collected by running as the superuser:

```
$ sudo mayday
```

If permissions are missing or if a dependency is not installed, the report files
for that check will be empty. Running as superuser will ensure that all permissions
are given to access the kernel logs, pstore content, coredump info, and so on.

Even more data can be collected by adding the `--danger` flag:

```
$ sudo mayday --danger
```

## what's collected

By default, mayday operates in a "safe" mode. No sensitive information is
collected -- only information that support might need. This includes things like:

* network connections, firewall rules, and hostname
* information about currently running processes, including open files and ports
* the system journal with the kernel log via `journalctl`, depending on the user this needs `sudo`
* pstore dmesg logs, normally needs `sudo`
* log files from systemd services
* the list of logged coredumps (depends on `systemd-coredump`), works best with `sudo`
* filesystem and memory usage information
* information about docker and rkt containers, including network and state but NOT logs

The following is collected only if the `--danger` flag is activated:

* logs and environment variables of docker and rkt containers

The following information is **never** collected:

* private keys
* passwords

## integration

### about
Mayday can be integrated into other projects by defining a default configuration
file at either the location `/etc/mayday/default.json` or
`/usr/share/mayday/default.json`. Through the use of [viper](https://github.com/spf13/viper)
YAML and TOML are now supported as well, though [Flatcar Container Linux](https://flatcar.org)
will continue to use JSON as the mechanism of choice.  If multiple products are
to be supported specialized configurations can be provided as "profiles" located
in the above directories (e.g. `/etc/mayday/quay.json`) and the referenced via:

```
$ mayday -p quay
```

### configuration syntax

The configuration file is comprised of objects (As of 1.0.0 valid objects are
"files" and "commands").  A example of the syntax can be seen in the file
[default.json](https://github.com/flatcar-linux/mayday/blob/master/default.json).
Each top level object contains an array of the relevant items to collect.
Optionally items can be annotated with a "link" which will provide an easy to
locate pointer for commonly accessed data.

### collection
Files are directly retrieved. Commands are executed and the results of standard
output (`stdout`) are collected. Assets are placed into a Go "tarable"
interface and then gzipped and serialized out to a file on disk.
