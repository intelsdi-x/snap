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
	./scripts/build_deps.sh

restore:
	# restore all dependencies from net using recurisvlly called godep in pulse agent and all plugins
	./scripts/restore_deps.sh

test: build
	# make use of alread build all plugins a moment ago
	export PULSE_BUILD_HELPER_NO_REBUILD=no
	./scripts/test.sh

build:
	./scripts/build.sh

docker-build:
	docker build --tag pulse_tests .

docker-test:
	docker run --rm --name pulse_tests pulse_tests

docker-bash:
	docker run -ti --rm --name pulse_tests pulse_tests /bin/bash

