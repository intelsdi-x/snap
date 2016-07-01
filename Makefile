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

default:
	$(MAKE) deps
	$(MAKE) all
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
check:
	$(MAKE) test
all:
	bash -c "./scripts/build_snap.sh"
install:
	cp build/bin/snapd /usr/local/bin/
	cp build/bin/snapctl /usr/local/bin/
proto:
	cd `echo $(GOPATH) | cut -d: -f 1`; bash -c "./src/github.com/intelsdi-x/snap/scripts/gen-proto.sh"
