test:
	go get -u github.com/smartystreets/goconvey
	go get -u golang.org/x/tools/cmd/cover

all: build-pulse

build-pulse:
	bash -c ./scripts/build.sh