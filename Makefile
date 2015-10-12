#http://www.apache.org/licenses/LICENSE-2.0.txt
#
#
#Copyright 2015 Intel Coporation
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
	export PULSE_PATH=`pwd`/build; bash -c "./scripts/test.sh"
check:
	$(MAKE) test
all:
	bash -c "./scripts/build.sh $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))) true"
pulse:
	bash -c "./scripts/build.sh $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))"
install:
	cp build/bin/pulsed /usr/local/bin/
	cp build/bin/pulsectl /usr/local/bin/
release:
	bash -c "./scripts/release.sh $(TAG) $(COMMIT)"
