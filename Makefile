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
