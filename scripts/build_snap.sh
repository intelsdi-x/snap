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

git_branch=$(git symbolic-ref HEAD 2> /dev/null | cut -b 12-)
git_branch="${git_branch:-test}"
git_sha=$(git log --pretty=format:"%h" -1)
git_version=$(git describe --always --exact-match 2> /dev/null || echo "${git_branch}-${git_sha}")

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"

# shellcheck source=scripts/common.sh
. "${__dir}/common.sh"

_info "project path: ${__proj_dir}"

build_dir="${__proj_dir}/build"
bin_dir="${build_dir}/bin"
plugin_dir="${build_dir}/plugin"
go_build=(go build -ldflags "-w -X main.gitversion=${git_version}")

_info "snap build version: ${git_version}"
_info "git commit: $(git log --pretty=format:"%H" -1)"

# Disable CGO for builds.
export CGO_ENABLED=0

# rebuild binaries:
_debug "removing: ${bin_dir:?}/*"
rm -rf "${bin_dir:?}/"*
mkdir -p "${bin_dir}"

_info "building snapd"
"${go_build[@]}" -o "${bin_dir}/snapd" . || exit 1

_info "building snapctl"
(cd "${__proj_dir}/cmd/snapctl" && "${go_build[@]}" -o "${bin_dir}/snapctl" . || exit 1)

# rebuild plugins:
_debug "removing: ${plugin_dir:?}/*"
rm -rf "${plugin_dir:?}/"*
mkdir -p "${plugin_dir}"

_info "building plugins"
find "${__proj_dir}/plugin/" -type d -iname "snap-*" -print0 | xargs -0 -n 1 -I{} "${__dir}/build_plugin.sh" {}
