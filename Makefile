default: build-pulse

test:
	go get -u github.com/smartystreets/goconvey
	go get -u golang.org/x/tools/cmd/cover

build-pulse:
	bash -c ./scripts/build.sh