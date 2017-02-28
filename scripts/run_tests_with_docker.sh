#!/usr/bin/env bash

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

SNAP_TEST_TYPE="${SNAP_TEST_TYPE:-$1}"

UNIT_TEST="${UNIT_TEST:-"gofmt goimports go_test go_cover"}"

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"

# shellcheck source=scripts/common.sh
. "${__dir}/common.sh"

_debug "script directory ${__dir}"
_debug "project directory ${__proj_dir}"

[[ "$SNAP_TEST_TYPE" =~ ^(small|medium|large|legacy)$ ]] || _error "invalid TEST_TYPE (value must be 'small', 'medium', 'large', or 'legacy', received:${SNAP_TEST_TYPE}"

(cd ${__proj_dir} && docker build -t intelsdi-x/snap-test -f "${__dir}/Dockerfile" .)
docker run -it intelsdi-x/snap-test scripts/test.sh "${SNAP_TEST_TYPE}"
