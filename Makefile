default: build-pulse

deps:
	go get -u github.com/smartystreets/goconvey
	go get -u golang.org/x/tools/cmd/cover

test: 
	export PULSE_PATH=`pwd`/build; bash -c "./scripts/test.sh"

build-pulse:
	bash -c "./scripts/build.sh $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))"
