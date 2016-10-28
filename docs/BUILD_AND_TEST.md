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

If you prefer a video walkthrough of this process, watch this [tutorial](https://vimeo.com/161561815).

To build the Snap Framework you'll need:
* [Golang >= 1.6](https://golang.org)
    * Should be [downloaded](https://golang.org/dl/) and [installed](https://golang.org/doc/install)
* [GNU Make](https://www.gnu.org/software/make/)
* [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

The instructions below assume that the `GOPATH` environment variable has been set properly. Many of us use [go version manager (gvm)](https://github.com/moovweb/gvm) to easily switch between Go versions.

Now you can download Snap into your `$GOPATH`:

```
$ # -d is used to download snap without building it
$ go get -d github.com/intelsdi-x/snap
$ cd $GOPATH/src/github.com/intelsdi-x/snap
```

In the `snap/` directory there's a `Makefile` that builds all dependencies and then the Snap Framework binaries. To get dependencies and build Snap run:  
```
$ cd $GOPATH/src/github.com/intelsdi-x/snap
$ make
```

By default `make` runs `make deps`, `make snap`, and `make plugins` commands for you. Alternatively, you can run `make` with any of these other targets:

* `deps`: fetches all dependencies using glide
* `test-(legacy|small|medium|large)`: runs test suite
* `all`: builds snapd, snapctl, and test plugins for all platforms (MacOS and Linux)
* `snap` builds snapd and snapctl for local operating system
* `plugins` builds test plugins for local operating system
* `install`: installs snapd and snapctl binaries in /usr/local/bin

To see how to use Snap, look at [gettings started](../README.md#getting-started), [SNAPD.md](SNAPD.md), and [SNAPCTL.md](SNAPCTL.md).

## Test
### Creating Tests
Our tests are written using [smartystreets' GoConvey package](https://github.com/smartystreets/goconvey).  See https://github.com/smartystreets/goconvey/wiki for an introduction to creating a test using this package.

### Tests in Golang
We follow the Golang methodology of placing tests into files with names that look like `*_test.go`. See [this](https://golang.org/cmd/go/#hdr-Test_packages) section from [go command](https://golang.org/cmd/go/) documentation for more details

### Test Types in Snap
As was mentioned previously, tests in Snap are broken down into `small`, `medium`, and `large` tests. Definitions for these test types can be found in the 'Testing Guidelines' section [CONTRIBUTING.md](../CONTRIBUTING.md) document.

But, how do you actually tag a test as *small*, *medium*, or *large* in Golang? As it turns out, this is a relatively simple process. To tag a given test file by the category of tests that it contains, all we have to do is add a single line to the beginning of the file that provides a *build tag* for that file. For example, a line like this:

    // +build small

would identify that file as a file that contains *small* tests, while a line like this as the first line of the file:

	 // +build medium

would identify that file as a file that contains *medium* tests. Once those build tags are have been added to the test files in the Snap codebase, it is relative simple to run a specific set of tests (by type) by simply adding a `-tags [TAG]` command-line flag to the `go test` command (where the `[TAG]` value is replaced by one of our test types). For example, this command will run all of the tests in the current working directory or any of its subdirectories that have been tagged as *small* tests:

	 $ go test -v  -tags=small ./...

It should be noted here that if there are any untagged tests in the directory referenced by the `go test` command, those untagged tests will also be run, regardless of the `-tags [TAG]` option that is passed into the `go test` command. To deal with this issue, all preexisting tests in the Snap framework have been tagged with a *legacy* tag. We have also modified the current test matrix (in the appropriate TravisCI and test shell-script files) to run both the *small* and *legacy* tests until such time that we feel that there is sufficient test coverage by the *small* tests that are currently being developed).

Once the maintainers feel that the *small* tests provide sufficient code coverage, the existing *legacy* tests will be phased out (or used in the construction of a set of *medium* and *large* tests for the Snap CI/CD toolchain). All new tests being added to the Snap framework by contributors should be marked as either `small`, `medium`, or `large` tests, depending on their scope.

### Running Tests
#### On a local machine
Contributors are assumed to have run the `legacy` and `small` tests that cover their contribution on their local machine prior to submitting the corresponding pull request. There are several ways of accomplishing this depending on the coverage required.

Before any tests can be run, the following command should be executed to pull down all of the test dependencies to the local machine (note; this will also run the `./scripts/test.sh` script, see below for more details):
```
$ cd $GOPATH/src/github.com/intelsdi-x/snap
$ SNAP_TEST_TYPE=[SNAP_TEST_TYPE] make test
```
where the string `[SNAP_TEST_TYPE]` that is shown above is replaced with one of our test types (`legacy`, `small`, `medium`, or `large`). Once the dependencies have been pulled down to the local machine, there are a number of ways to run the various tests included in the Snap framework. To run the complete set of `legacy`, `small`, `medium`, or `large` tests in the Snap framework (and stop at the first failure in a directory before continuing on), you can simply use the same shell-script that is run by the `make test` command that was shown above:  
```
$ ./scripts/test.sh [SNAP_TEST_TYPE]
```
As was the case in the `make test` command that was shown above, the string `[SNAP_TEST_TYPE]` that is shown in this command should be replaced by the type of tests you wish to run (`legacy`, `small`, `medium`, or `large`). Note, that for convenience we have also added four new targets to the Snap `Makefile` that can be used to run the `legacy`, `small`, `medium`, or `large` tests in a simple `make test-*` command. To run the `small` tests, for example, one could simply run a command that looks something like this:
```
$ make test-small
```
To run the other types of tests in the Snap framework, simply replace the `small` type with one of the other types (`legacy`, `medium`, or `large` in the example `make test-*` command shown above).

If you are  interested in running all of the `legacy` tests from the Snap framework *and continuing through to all subdirectories, regardless of any errors that might be encountered*, then you can run a `go test ...` command directly (instead of running a `make test-*` or `scripts/test.sh [SNAP_TEST_TYPE]` command like those shown above). That `go test ...` command would look something like this:  
```
go test -tags=legacy ./...
```
It should be noted here that the `go test ...` command shown above will iterate over all of the subdirectories of the current working directory and run all of the `legacy` tests found. It will not stop if errors are detected during a test run; those errors will simply be reported as part of the output of the `go test ...` command.

To run all of the small tests from the Snap framework, one would simply change the value of the tag passed into the command from `legacy` to `small`:
```
go test -tags=small ./...
```
All tests in the Snap framework are tagged as `legacy`, `small`, `medium`, or `large`, so at least one of those tags must be provided to any `go test ...` command that is executed by a contributor.

Individual tests from the framework can also be selected and run by passing a test name into the `go test...` command using the `-run` command-line flag:
```
go test -v -tags=legacy -run=[TEST_NAME] ./...
```

e.g. to run the `TestLoad` test from `control_test.go`, one would simply run a command that looks something like this:
```
go test -v -tags=legacy -run=TestLoad ./...
```
Finally, coverage reports can be generated that show (on a function-by-function level) how well a given test or set of tests cover the current codebase. For example, the following pair of commands will generate a report showing the coverage of all of the `legacy` tests in the current working directory (as text output to the console):
```
go test -tags=small -coverprofile=/tmp/coverage.out . && go tool cover -func=/tmp/coverage.out
```
A similar pair of commands can be used to generate an HTML view showing the same information, but in a browsable view that shows how the matching tests cover the code on a line-by-line basis for each file in the current working directory (the resulting view will open in the default web browser defined on the local machine)
```
go test -tags=small -coverprofile=/tmp/coverage.out . && go tool cover -html=/tmp/coverage.out
```
As was the case with the previous commands, the `-run` command-line flag can be used to further reduce the scope of the generated report to only show the coverage of named test. It should be noted here that these coverage reports can only be generated for code in a single directory. The `-coverprofile` command-line flag cannot be used to generate a coverage report across all subdirectories as was shown in the `go test ...` command shown previously.

### Building Effective Small Tests

Any `small` tests added to the Snap framework must conform to the following constraints:
* They should test the behavior of a single function or method in the framework
* There should no reliance rely on external systems or system-level resources (eg. networks, filesystem, external systems or services, system properties, multiple threads of execution, or the use of sleep statements) as part of these tests; the expected responses of any such external systems or access to system-level resources should be mocked appropriately.
* They should be independent of any other tests in the test framework
* They should return the same result every time they are run, regardless of the environment they are run in (a failure of any of these tests should be indicative of an issue with the code being tested, not the environment that the test is being run in)

When complete, the full set of `small` tests for any given function or method should provide sufficient code coverage to ensure that any changes made to that function or method will not 'break the build'. This will assure the Snap maintainers that any pull requests that are made to modify or add to the framework can be safely merged (provided that there is sufficient code coverage and the associated tests pass).

It should be noted here that the maintainers will refuse to merge any pull requests that trigger a failure of any of the `small` or `legacy` tests that cover the code being modified or added to the framework. As such, we highly recommend that contributors run the tests that cover their contributions locally before submitting their contribution as a pull request. Maintainers may also  ask that contributors add tests to their pull requests to ensure adequate code coverage before they are willing to accept a given pull request, even if all existing tests pass. Our hope is that you, as a contributor, will understand the need for this requirement.

#### In Docker
The Snap Framework supports running tests in an isolated container as opposed to your local host. Run the test script, which calls a `Dockerfile` located at `./scripts/Dockerfile`:

```
$ cd $GOPATH/src/github.com/intelsdi-x/snap
$ scripts/run_tests_with_docker.sh [SNAP_TEST_TYPE]  
```
