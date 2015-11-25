#Contributing

##Build and Test
###Build
In the /snap directory there's a `Makefile` that builds all dependencies and snap.
To get dependencies and build snap just run:  
```
make
```

Alternatively, you can run `make` with any of the following options:

Makefile options:
```
default:
	#runs make deps and make all
	$(MAKE) deps
	$(MAKE) all
deps:
	#gets all dependencies using godeps
	bash -c "./scripts/deps.sh"
test:
	#exports snap build path to env var SNAP_PATH and runs test files
	export SNAP_PATH=`pwd`/build; bash -c "./scripts/test.sh"
check:
	#runs make test
	$(MAKE) test
all:
	#builds snap daemon, CLI, and plugin binaries
	bash -c "./scripts/build.sh $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))) true"
snap:
	#builds snap daemon and CLI binaries, but not plugin binaries
	bash -c "./scripts/build.sh $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))"
install:
	#copies snapd and snapctl binaries into /usr/local/bin
	cp build/bin/snapd /usr/local/bin/
	cp build/bin/snapctl /usr/local/bin/
release:
	#creates a snap release
	bash -c "./scripts/release.sh $(TAG) $(COMMIT)"
```

###Test
####Creating Tests
Our tests are written using [smartystreets' GoConvey package](https://github.com/smartystreets/goconvey)   
File names have the following convention:  
File to be tested: `filename.go`  
Testing file:	   `filename_test.go`  
Each convey statement starts off a new go routine.   
See https://github.com/smartystreets/goconvey/wiki for an introduction to creating a test.

####Running Tests
#####In local machine
To run all tests in order (it will stop at any failure in a directory before continuing on):  
```
./scripts/test.sh
```

TO run all tests *and to continue through all directories even with errors*:  
```
go test ./...
```  

To use advanced functionality possible through GoConvey:
```
go test ./... -v -run <ConveyFunc>
```

e.g. `TestLoad` from `control_test.go`:
```
go test ./... -v -run TestLoad
```

e.g. using GoConvey UX:
```
go test -coverprofile=/tmp/coverage.out && go tool cover -html=/tmp/coverage.out
```

#####In Docker
There's a `Dockerfile` located at `./scripts/Dockerfile`:
```
FROM golang:latest
ENV GOPATH=$GOPATH:/app
ENV SNAP_PATH=/go/src/github.com/intelsdi-x/snap/build
RUN apt-get update && \
    apt-get -y install facter
WORKDIR /go/src/github.com/intelsdi-x/
RUN git clone https://<GIT_TOKEN>@github.com/intelsdi-x/gomit.git
WORKDIR /go/src/github.com/intelsdi-x/snap
ADD . /go/src/github.com/intelsdi-x/snap
RUN go get github.com/tools/godep && \
    go get golang.org/x/tools/cmd/goimports && \
    go get golang.org/x/tools/cmd/vet && \
    go get golang.org/x/tools/cmd/cover && \
    go get github.com/smartystreets/goconvey
RUN scripts/deps.sh
RUN make
```
This is run in the snap directory using `./scripts/run_tests_with_docker.sh`  
First you need a [github personal access token](https://help.github.com/articles/creating-an-access-token-for-command-line-use/)  
Then export the token using `export GIT_TOKEN=<tokenID>`.

```
#!/bin/bash -e

die() {
    echo >&2 $@
    exit 1
}

if [ $# -eq  2 ]; then
	GIT_TOKEN=$1
fi

if [ -z "${GIT_TOKEN}" ]; then
	die "arg missing: github token is required so we can clone a private repo)"
fi

sed s/\<GIT_TOKEN\>/${GIT_TOKEN}/ scripts/Dockerfile > scripts/Dockerfile.tmp
docker build -t intelsdi-x/snap-test -f scripts/Dockerfile.tmp .
rm scripts/Dockerfile.tmp
docker run -it intelsdi-x/snap-test scripts/test.sh
```
