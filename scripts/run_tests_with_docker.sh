#!/bin/bash -e

die() {
    echo >&2 $@
    exit 1
}

if [ $# -eq  2 ]; then
	GIT_TOKEN=$1
fi

if [ -z "${GIT_TOKEN}" ]; then
	die "arg missing: github token is required so we can clone a private repo)"
fi

sed s/\<GIT_TOKEN\>/${GIT_TOKEN}/ scripts/Dockerfile > scripts/Dockerfile.tmp
docker build -t sdilabs-x/pulse-test -f scripts/Dockerfile.tmp .
rm scripts/Dockerfile.tmp
docker run -it sdilabs-x/pulse-test scripts/test.sh
