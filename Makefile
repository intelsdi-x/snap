.PHONY: build test

# `make test` first builds agent and all plugins using vendored packages using help ./scripts/build.sh
# then calls helper scripts ./scripts/tests.sh that actually goes recursively through all src folders (including plugin) call `go test` in each.
# Additionally the flag PULSE_BUILD_HELPER_NO_REBUILD is set to not prevent rebuild each plugin binarry before test.

default: build

# It should reassemble process that happens with CI env
# 1. we should have empty GOPATH with only src/github.com/intelsdi-pulse cloned inside
# 2. then we are install all build & testing dependencies like (goling/gocov/cover/imports/vet/convey)
# 3. we build all plugins one by one using their Godeps/_workspace/src as GOPATH
# 4. we build pulse itself with Godeps/_workspace appended to GOPATH
# 5. then tests
# assumptions because we are vendoring, we are not "go getting" any depedencies other that required strictly for building

clean:
	rm -rf build/

deps:
	echo "installing build & testing dependecies"
	echo "in GOPATH=${GOPATH}"
	#cd ../../docker/libcontainer; git checkout tags/v1.4.0; cd - going to Godeps
	#go get github.com/docker/libcontainer
	go get github.com/golang/lint/golint
	go get github.com/axw/gocov/gocov
	go get github.com/mattn/goveralls
	go get github.com/smartystreets/assertions
	go get github.com/smartystreets/goconvey
	go get github.com/smartystreets/goconvey/convey
	go get github.com/streadway/amqp
	go get github.com/tools/godep
	# cover will be in standard lib from 1.5 for 1.4 we have to live with such kind of hack
	# this hack probably  was required for 1.3
	#if ! go get code.google.com/p/go.tools/cmd/cover; then go get golang.org/x/tools/cmd/cover; fi
	go get golang.org/x/tools/cmd/cover
	go get golang.org/x/tools/cmd/goimports
	go get golang.org/x/tools/cmd/vet

test: build
	# make use of alread build all plugins a moment ago
	export PULSE_BUILD_HELPER_NO_REBUILD=no
	./scripts/test.sh

build:
	./scripts/build.sh
