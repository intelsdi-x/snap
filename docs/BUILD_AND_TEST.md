 
# Build and Test
## Getting Started
To build snap you'll need:
* [Golang >= 1.4](https://golang.org)
    * An option to look into is using the [go version manager (gvm)](https://github.com/moovweb/gvm) if you want to easily switch between Go versions.
* [GNU Make](https://www.gnu.org/software/make/)
* [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
* snap:
    * `go get github.com/intelsdi-x/snap`. This will put snap in your `$GOPATH`.
    * Fork the repository
    * `cd` into your /intelsdi-x/snap directory on your machine and do the following to set your fork as the origin and make a branch on your fork
        ```
        git remote rename origin upstream
        git remote add origin git@github.com:<yourGithubID>/snap.git
        ```
        To push to your fork later, either do a `git push origin master` or do `git checkout -b <someBranchName>` then you can do `git push`.
## Build
In the /snap directory there's a `Makefile` that builds all dependencies and snap.
To get dependencies and build snap just run:  
```
make
```
It runs the following `make deps` and `make all` commands.

Alternatively, you can run `make` with any of the following targets:

* `default`: runs make deps and make all
* `deps`: fetches all dependencies using godeps
* `check`: runs test suite
* `all`: builds snapd, snapctl, and the test plugins
* `snap` builds snapd and snapctl
* `install`: installs snapd and snapctl binaries in /usr/local/bin
* `release`: cuts a snap release

To update your branch with changes from the intelsdi-x master branch run:
```
git pull --rebase upstream master
```

To see how to run snap, look at [running snap](../README.md#running-snap), [SNAPD.md](SNAPD.md), and [SNAPCTL.md](SNAPCTL.md).

## Test
### Creating Tests
Our tests are written using [smartystreets' GoConvey package](https://github.com/smartystreets/goconvey).  See https://github.com/smartystreets/goconvey/wiki for an introduction to creating a test.

### Running Tests
#### In local machine
First you need to run the following to get all test dependencies (this runs ./scripts/test.sh as well):
```
make test
```
To run all tests in order (it will stop at any failure in a directory before continuing on):  
```
./scripts/test.sh
```
To run all tests *and to continue through all directories even with errors*:  
```
go test ./...
```  
Tests can be pruned in go test with the `-run` option:
```
go test ./... -v -run <TestName>
```

e.g. `TestLoad` from `control_test.go`:
```
go test ./... -v -run TestLoad
```

e.g. using GoConvey UX:
```
go test -coverprofile=/tmp/coverage.out && go tool cover -html=/tmp/coverage.out
```

#### In Docker
You need to have Docker Engine installed. Visit [Install Docker Engine](https://docs.docker.com/engine/installation/) for detailed instructions on how to do it.
If you're using Docker on OS X(Darwin) you need to have an active Docker host and the env var, `$DOCKER_HOST`, needs to be exported. 

There's a `Dockerfile` located at `./scripts/Dockerfile`:
```
FROM golang:latest  
ENV GOPATH=$GOPATH:/app
ENV SNAP_PATH=/go/src/github.com/intelsdi-x/snap/build
RUN apt-get update && \
    apt-get -y install facter
WORKDIR /go/src/github.com/intelsdi-x/
RUN git clone https://github.com/intelsdi-x/gomit.git
WORKDIR /go/src/github.com/intelsdi-x/snap
ADD . /go/src/github.com/intelsdi-x/snap
RUN go get github.com/tools/godep && \
    go get golang.org/x/tools/cmd/goimports && \
    go get golang.org/x/tools/cmd/cover && \
    go get github.com/smartystreets/goconvey
RUN scripts/deps.sh
RUN make
```
This is run in the snap directory using `./scripts/run_tests_with_docker.sh`  
```
#!/bin/bash -e

docker build -t intelsdi-x/snap-test -f scripts/Dockerfile .
docker run -it intelsdi-x/snap-test scripts/test.sh
```
