default: build-pulse

deps:
	go get -u github.com/smartystreets/goconvey
	go get -u github.com/smartystreets/assertions
	go get -u golang.org/x/tools/cmd/cover
	go get -u github.com/docker/libcontainer
	cd ../../docker/libcontainer; git checkout tags/v1.4.0; cd -

test: 
	export PULSE_PATH=`pwd`/build; bash -c "./scripts/test.sh"

build-pulse:
	bash -c "./scripts/build.sh $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))"
