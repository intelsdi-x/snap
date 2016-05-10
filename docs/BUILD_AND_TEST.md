<!--
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
--> 
# Build and Test
This guide gets you started with building and testing Snap. If you have commits you want to contribute, review the [CONTRIBUTING file](../CONTRIBUTING.md) for a shorter list of what we look for and come back here if you need to verify your environment is configured correctly.

## Getting Started
If you prefer a video walkthrough of this process, watch this video: https://vimeo.com/161561815

To build the Snap Framework you'll need:
* [Golang >= 1.5](https://golang.org)
    * Should be [downloaded](https://golang.org/dl/) and [installed](https://golang.org/doc/install)
* [GNU Make](https://www.gnu.org/software/make/)
* [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

The instructions below assume that the `GOPATH` environment variable has been set properly. Many of us use [go version manager (gvm)](https://github.com/moovweb/gvm) to easily switch between Go versions.

Now you can install Snap into your `$GOPATH`: 
```
$ go get github.com/intelsdi-x/snap
$ cd $GOPATH/src/github.com/intelsdi-x/snap
```

[godeps](https://github.com/tools/godep) is a dependency for running the `make` task(s) required for the build process. If it is not already installed, install `godeps` now:
```
$ # first check to see if it is installed
$ which godep
$ # if not installed, do so
$ # then download and set your path
$ go get github.com/tools/godep
$ export PATH=$GOPATH/bin/
```

In the `snap/` directory there's a `Makefile` that builds all dependencies and then the Snap Framework binaries. To get dependencies and build Snap run:  
```
$ cd $GOPATH/src/github.com/intelsdi-x/snap
$ make
```

It runs `make deps` and `make all` commands for you. Alternatively, you can run `make` with any of these other targets:

* `deps`: fetches all dependencies using godeps
* `check`: runs test suite
* `all`: builds snapd, snapctl, and the test plugins
* `snap` builds snapd and snapctl
* `install`: installs snapd and snapctl binaries in /usr/local/bin
* `release`: cuts a Snap release


To see how to use Snap, look at [Running Snap](../README.md#running-snap), [SNAPD.md](SNAPD.md), and [SNAPCTL.md](SNAPCTL.md).

## Test
### Creating Tests
Our tests are written using [smartystreets' GoConvey package](https://github.com/smartystreets/goconvey).  See https://github.com/smartystreets/goconvey/wiki for an introduction to creating a test using this package.

### Running Tests
#### In local machine
First you need to run the following to get all test dependencies (this runs ./scripts/test.sh as well):
```
$ cd $GOPATH/src/github.com/intelsdi-x/snap
$ make test
```

To run all tests in order (it will stop at any failure in a directory before continuing on):  
```
$ ./scripts/test.sh
```

To run all tests *and to continue through all directories even with errors*:  
```
$ go test ./...
``` 
 
Tests can be pruned in go test with the `-run` option followed by the test name (e.g. `TestLoad` from `control_test.go`):
```
$ go test ./... -v -run TestLoad
```

Or using GoConvey ([in a browser](https://github.com/smartystreets/goconvey#in-the-browser)):
```
$ go test -coverprofile=/tmp/coverage.out && go tool cover -html=/tmp/coverage.out
```

#### In Docker
The Snap Framework supports running tests in an isolated container as opposed to your local host. Run the test script, which calls a `Dockerfile` located at `./scripts/Dockerfile`: 

```
$ cd $GOPATH/src/github.com/intelsdi-x/snap
$ scripts/run_tests_with_docker.sh`  
```
