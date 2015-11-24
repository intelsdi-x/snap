#!/bin/bash -e

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
docker build -t intelsdi-x/snap-test -f scripts/Dockerfile.tmp .
rm scripts/Dockerfile.tmp
docker run -it intelsdi-x/snap-test scripts/test.sh
