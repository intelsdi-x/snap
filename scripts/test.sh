#!/bin/bash

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

# Support travis.ci environment matrix:
SNAP_TEST_TYPE="${SNAP_TEST_TYPE:-$1}"

UNIT_TEST="${UNIT_TEST:-"gofmt goimports go_test go_cover"}"

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"

SNAP_PATH="${SNAP_PATH:-"${__proj_dir}/build"}"
export SNAP_PATH

# shellcheck source=scripts/common.sh
. "${__dir}/common.sh"

_debug "script directory ${__dir}"
_debug "project directory ${__proj_dir}"

[[ "$SNAP_TEST_TYPE" =~ ^(small|medium|large|legacy)$ ]] || _error "invalid/missing SNAP_TEST_TYPE (value must be 'small', 'medium', 'large', or 'legacy', received:${SNAP_TEST_TYPE}"

# If the following plugins don't exist, exit
[ -f $SNAP_PATH/plugin/snap-plugin-collector-mock1 ] || { _error 'Error: $SNAP_PATH/plugin/snap-plugin-collector-mock1 does not exist. Run make to build it.'; }
[ -f $SNAP_PATH/plugin/snap-plugin-collector-mock2 ] || { _error 'Error: $SNAP_PATH/plugin/snap-plugin-collector-mock2 does not exist. Run make to build it.';  }
[ -f $SNAP_PATH/plugin/snap-plugin-processor-passthru ] || { _error 'Error: $SNAP_PATH/plugin/snap-plugin-processor-passthru does not exist. Run make to build it.'; }
[ -f $SNAP_PATH/plugin/snap-plugin-publisher-mock-file ] || { _error 'Error: $SNAP_PATH/plugin/snap-plugin-publisher-mock-file does not exist. Run make to build it.'; }

_go_path
# If the following tools don't exist, get them
_go_get github.com/smartystreets/goconvey

# Run test coverage on each subdirectories and merge the coverage profile.
echo "mode: count (${SNAP_TEST_TYPE})" > "profile-${SNAP_TEST_TYPE}.cov"

TEST_TYPE=$SNAP_TEST_TYPE
export TEST_TYPE

go_tests=(gofmt goimports golint go_vet go_race go_test go_cover)

_debug "available unit tests: ${go_tests[*]}"
_debug "user specified tests: ${UNIT_TEST}"

((n_elements=${#go_tests[@]}, max=n_elements - 1))

for ((i = 0; i <= max; i++)); do
	if [[ "${UNIT_TEST}" =~ (^| )"${go_tests[i]}"( |$) ]]; then
		_info "running ${go_tests[i]}"
		_"${go_tests[i]}"
	else
		_debug "skipping ${go_tests[i]}"
	fi
done

_info "test complete: ${SNAP_TEST_TYPE}"
