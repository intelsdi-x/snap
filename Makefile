default: build-pulse

test:
	go get -u github.com/smartystreets/goconvey
	go get -u github.com/smartystreets/assertions
	go get -u golang.org/x/tools/cmd/cover

build-pulse:
	bash -c "./scripts/build.sh $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))"