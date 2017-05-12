#http://www.apache.org/licenses/LICENSE-2.0.txt
#
#
#Copyright 2015 Intel Corporation
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.

OS = $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH = $(shell uname -m)

default:
	$(MAKE) deps
	$(MAKE) snap
	$(MAKE) plugins
	$(MAKE) swagger
deps:
	bash -c "./scripts/deps.sh"
test:
	bash -c "./scripts/test.sh $(SNAP_TEST_TYPE)"
test-legacy:
	bash -c "./scripts/test.sh legacy"
test-small:
	bash -c "./scripts/test.sh small"
test-medium:
	bash -c "./scripts/test.sh medium"
test-large:
	bash -c "./scripts/test.sh large"
test-all:
	$(MAKE) test-legacy
	$(MAKE) test-medium
	$(MAKE) test-small
	$(MAKE) test-large
# NOTE:
# By default compiles will use all cpu cores, use BUILD_JOBS to control number
# of parallel builds: `BUILD_JOBS=2 make plugins`
#
# Build only snapteld/snaptel
snap:
	bash -c "./scripts/build_snap.sh"
# Build only plugins
plugins:
	bash -c "./scripts/build_plugins.sh"
# Build snap and plugins for all platforms
all:
	bash -c "./scripts/build_all.sh"
install:
	mkdir -p /usr/local/sbin
	mkdir -p /usr/local/bin
	cp build/$(OS)/$(ARCH)/snapteld /usr/local/sbin/
	cp build/$(OS)/$(ARCH)/snaptel /usr/local/bin/
proto:
	cd `echo $(GOPATH) | cut -d: -f 1`; bash -c "./src/github.com/intelsdi-x/snap/scripts/gen-proto.sh"
swagger:
	bash -c "./scripts/swagger.sh"
