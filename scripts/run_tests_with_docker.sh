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

if [[ $# -ne 1 ]]; then
	echo "ERROR; missing SNAP_TEST_TYPE (Usage: $0 SNAP_TEST_TYPE)"
	exit -2
elif [[ “$1” != “legacy” && "$1" != "small" && “$1” != “medium” && “$1” != “large” ]]; then
	echo "Error; invalid SNAP_TEST_TYPE (value must be one of 'legacy', 'small', 'medium', or 'large'; received $1)"
	exit -1
fi
SNAP_TEST_TYPE=$1

docker build -t intelsdi-x/snap-test -f scripts/Dockerfile .
docker run -it intelsdi-x/snap-test scripts/test.sh $SNAP_TEST_TYPE
