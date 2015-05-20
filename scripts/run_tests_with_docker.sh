#!/bin/bash

die() {
    echo >&2 $@
    exit 1
}

[ $# -eq 1 ] || die "arg missing: github token is required so we can clone a private repo)"

sed s/\<GIT_TOKEN\>/$1/ scripts/Dockerfile > scripts/Dockerfile.tmp
docker build -t sdilabs-x/pulse-test -f scripts/Dockerfile.tmp .
rm scripts/Dockerfile.tmp
docker run -it sdilabs-x/pulse-test scripts/test.sh