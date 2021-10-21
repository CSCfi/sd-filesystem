# SDA-Filesystem

This desktop software makes use of the [SD-Connect Proxy API](docs/API.md) to build a FUSE (Filesystem in Userspace) layer.


### Requirements

Set environment variables `FS_SD_CONNECT_METADATA_API`, `FS_SD_CONNECT_DATA_API` and `FS_SD_CONNECT_CERTS` before running program.

For test environment use:

```
export FS_SD_CONNECT_METADATA_API=https://connect-metadata-api-test.sd.csc.fi
export FS_SD_CONNECT_DATA_API=https://connect-data-api-test.sd.csc.fi

# Connection requires a certificate
# SD_CONNECT_CERTS should be the file that contains the necessary certificates
export FS_SD_CONNECT_CERTS=cert.pem	#FS_SD_CONNECT_CERTS should be the file that contains the necessary certificates
```

## Graphical User Interface

###  Dependencies
Go version 1.16

cgofuse and its [dependencies on different operating systems](https://github.com/billziss-gh/cgofuse#how-to-build).

Install [Qt for Go](https://github.com/therecipe/qt/wiki/Installation). Regardless of the operating system, there are [multiple ways](https://github.com/therecipe/qt/wiki/Available-Tools) of installing this package. Required that `GO111MODULE=on`.


### Setup

On linux install required packages and vendor dependencies
```
./dev_utils/setup-linux.sh
```

Note: for some vendor modules there might be warnings such as:
```
INFO[0427] installing full qt/bluetooth                 
go install: no install location for directory /home/<user>/sd-connect-fuse-master/vendor/github.com/therecipe/qt/bluetooth outside GOPATH
	For more details see: 'go help gopath'
```
These are ok, and are caused as of go 1.14+ 
```
go command now verifies that the main module's vendor/modules.txt file is consistent with its go.mod file.
```

### Run

```
qtdeploy build desktop cmd/qt/main.go
./cmd/qt/deploy/darwin/qt_project.app/Contents/MacOS/qt_project  // Path slightly different for other OS`
```

### Deploy

To deploy binary to Virtual Machine (VM):
```
qtdeploy build desktop cmd/qt/main.go
tar -czf deploy.tar.gz -C cmd/qt/deploy linux
```

Copy the archive of the deployment environment.

## Command Line Interface

The CLI binary will require a username and password for accessing the SD-Connect Proxy API.

### Build and Run
```
go build -o ./go-fuse ./cmd/cli/main.go
./go-fuse
```
