#!/bin/bash

#http://www.apache.org/licenses/LICENSE-2.0.txt
#
#
#Copyright 2016 Intel Corporation
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

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"

# shellcheck source=scripts/common.sh
. "${__dir}/common.sh"

build_path="${SNAP_PATH:-"${__proj_dir}/build"}"
git_sha=$(git log --pretty=format:"%H" -1)
git_path="${build_path}/${TRAVIS_BRANCH}/${git_sha}"
latest_path="${build_path}/${TRAVIS_BRANCH}/latest"

mkdir -p "${git_path}"
mkdir -p "${latest_path}"

_info "copying snap binaries to ${git_path}"
cp "${build_path}/bin/"* "${git_path}"
cp "${build_path}/plugin/"* "${git_path}"
_info "copying snap binaries to ${latest_path}"
mv "${build_path}/bin/"* "${latest_path}"
mv "${build_path}/plugin/"* "${latest_path}"
