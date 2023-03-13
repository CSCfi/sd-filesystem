# SDA-Filesystem / Data Gateway

[![Linting go code](https://github.com/CSCfi/sda-filesystem/actions/workflows/linting.yml/badge.svg)](https://github.com/CSCfi/sda-filesystem/actions/workflows/linting.yml)
[![Unit Tests](https://github.com/CSCfi/sda-filesystem/actions/workflows/unittest.yml/badge.svg)](https://github.com/CSCfi/sda-filesystem/actions/workflows/unittest.yml)
[![Coverage Status](https://coveralls.io/repos/github/CSCfi/sda-filesystem/badge.svg?branch=master)](https://coveralls.io/github/CSCfi/sda-filesystem?branch=master)

**This project has been rebranded as Data Gateway**

Data Gateway makes use of the:

- [SD Connect Proxy API](docs/SD-Connect-API.md) 
- [SD Apply/SD Submit Download API](docs/SD-Submit-API.md) 

It builds a FUSE (Filesystem in Userspace) layer and uses Airlock to export files to SD Connect. Software currently supports Linux, macOS and Windows for:
- [Graphical User Interface](#graphical-user-interface)
- [Command Line Interface](#command-line-interface)

Binaries are built on each release for all supported Operating Systems.

### Requirements

Go version 1.20

Set these environment variables before running the application:
- `FS_SD_CONNECT_API` - API for SD-Connect
- `FS_SD_SUBMIT_API` – a comma-separated list of APIs for SD Apply/SD Submit
- `SDS_ACCESS_TOKEN` - a JWT for authenticating to the SD APIs
- `FS_CERTS` - path to a file that contains certificates required by SD Connect, SD Apply/SD Submit, and SDS AAI 

Optional envronment variables:

- `CSC_USERNAME` - username for SDA-Filesystem
- `CSC_PASSWORD` - password for SDA-Filesystem and Airlock CLI

For test environment follow instructions at https://gitlab.ci.csc.fi/sds-dev/local-proxy

## Graphical User Interface

###  Dependencies

`cgofuse` and its [dependencies on different operating systems](https://github.com/billziss-gh/cgofuse#how-to-build).

Install [Wails](https://wails.io/docs/gettingstarted/installation) and its dependencies.

### Build and Run

Before running/building the repository for the first time, generate the frontend assests by running:
```
npm install --prefix frontend
npm run build --prefix frontend
```

To run in development mode:
```
cd cmd/gui
wails dev
``` 

To build for production:
```
cd cmd/gui

# For Linux and macOS
wails build -upx -trimpath -clean -s

# For Windows
wails build -upx -trimpath -clean -s -webview2=embed
```

### Deploy

See [Linux setup](docs/linux-setup.md).

## Command Line Interface

Two command line binaries are released, one for SDA-Filesystem and one for Airlock. 

### SDA-Fileystem

The CLI binary will require a username and password for accessing the SD-Connect Proxy API. Username is given as input. Password is either given as input or in an environmental variable.

#### Build and Run
```
go build -o ./go-fuse ./cmd/fuse/main.go
```
Test install.
```
./go-fuse -help                        
Usage of ./go-fuse:
  -alsologtostderr
    	log to standard error as well as files
  -http_timeout int
    	Number of seconds to wait before timing out an HTTP request (default 20)
  -log_backtrace_at value
    	when logging hits line file:N, emit a stack trace
  -log_dir string
    	If non-empty, write log files in this directory
  -loglevel string
    	Logging level. Possible values: {debug,info,warning,error} (default "info")
  -logtostderr
    	log to standard error instead of files
  -mount string
    	Path to Data Gateway mount point
  -project string
    	SD Connect project if it differs from that in the VM
  -sdapply
      Connect only to SD Apply
  -stderrthreshold value
    	logs at or above this threshold go to stderr
  -v value
    	log level for V logs
  -vmodule value
    	comma-separated list of pattern=N settings for file-filtered logging

```
Example run: `./go-fuse -mount=$HOME/ExampleMount` will create the FUSE layer in the directory `$HOME/ExampleMount` for both 'SD Connect' and 'SD Apply'.

### Airlock

The CLI binary will require a username, a bucket and a filename. Password is either given as input or in an environmental variable.

#### Build and Run
```
go build -o ./airlock ./cmd/airlock/main.go
```
Test install.
```
./airlock -help
Usage of ./airlock:
  -alsologtostderr
    	log to standard error as well as files
  -debug
    	Enable debug prints
  -journal-number string
    	Journal Number/Name specific for Findata uploads
  -log_backtrace_at value
    	when logging hits line file:N, emit a stack trace
  -log_dir string
    	If non-empty, write log files in this directory
  -logtostderr
    	log to standard error instead of files
  -original-file string
    	Filename of original unecrypted file when uploading pre-encrypted file from Findata vm
  -project string
    	SD Connect project if it differs from that in the VM
  -quiet
    	Print only errors
  -segment-size int
    	Maximum size of segments in Mb used to upload data. Valid range is 10-4000. (default 4000)
  -stderrthreshold value
    	logs at or above this threshold go to stderr
  -v value
    	log level for V logs
  -vmodule value
    	comma-separated list of pattern=N settings for file-filtered logging
``` 

Example run: `./airlock username ExampleBucket ExampleFile` will export file `ExampleFile` to bucket `ExampleBucket`.

## Troubleshooting
See [troubleshooting](docs/troubleshooting.md) for fixes to known issues.

## License

Data Gateway is released under `MIT`, see [LICENSE](LICENSE).

[Wails](https://wails.io) is released under [MIT](https://github.com/wailsapp/wails/blob/master/LICENSE)

[CgoFuse](https://github.com/billziss-gh/cgofuse) is released under [MIT](https://github.com/billziss-gh/cgofuse/blob/master/LICENSE.txt)
