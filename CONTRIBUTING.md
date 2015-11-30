#Contributing to snap

snap is Apache 2.0 licensed and accepts contributions via GitHub pull requests. This document
will cover how to contribute code, report issues, build the project and run the tests.

## Contributing code

Contributing code to snap is a snap (pun intended).
- Fork the project to your own repository
- Create a topic branch from where you want to base your work (usually master)
- Make commit(s) (following commit guidelines below)
- Add any needed test coverage
- Push your commit(s) to your repository
- Open a pull request against the original repo and follow the pull request guidelines below

The maintainers of the repo utilize a "Looks Good To Me" (LGTM) message in the pull request.

### Commit Guidelines

Commits should have logical groupings. A bug fix should be a single commit. A new feature
should be a single commit. 

Commit messages should be clear on what is being fixed or added to the code base. If a
commit is addressing an open issue, please start the commit message with "Fix #XXX" or 
"Feature #XXX". This will help make the generated changelog for each release easy to read
with what commits were fixes and what commits were features.

### Pull Request Guidelines

Pull requests can contain a single commit or multiple commits. If a pull request adds
a feature but also fixes two bugs, then the pull request should have three commits, one
commit each for the feature and two bug fixes.

Your pull request should be rebased against the current master branch. Please do not merge
the current master branch in with your topic branch, nor use the Update Branch button provided
by GitHub on the pull request page.

## Reporting Issues

Reporting issues are very beneficial to the project. Before reporting an issue, please review current
open issues to see if there are any matches. If there is a match, comment with a +1, or "Also seeing this issue".
If any environment details differ, please add those with your comment to the matching issue.

When reporting an issue, details are key. Include the following:
- OS version
- snap version
- Environment details (virtual, physical, etc.)
- Steps to reproduce
- Actual results
- Expected results
 
##Build and Test
###Build
In the /snap directory there's a `Makefile` that builds all dependencies and snap.
To get dependencies and build snap just run:  
```
make
```

Alternatively, you can run `make` with any of the following targets:

* `default`: runs make deps and make all
* `deps`: fetches all dependencies using godeps
* `check`: runs test suite
* `all`: builds snapd, snapctl, and the test plugins
* `snap` builds snapd and snapctl
* `install`: installs snapd and snapctl binaries in /usr/local/bin
* `release`: cuts a snap release

###Test
####Creating Tests
Our tests are written using [smartystreets' GoConvey package](https://github.com/smartystreets/goconvey).  See https://github.com/smartystreets/goconvey/wiki for an introduction to creating a test.

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
