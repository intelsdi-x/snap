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

build_path="${__proj_dir}/build"
go_build=(go build -ldflags "-w -X main.gitversion=${git_version}")

_info "snap build version: ${git_version}"
_info "git commit: $(git log --pretty=format:"%H" -1)"

# Disable CGO for builds.
export CGO_ENABLED=0

# rebuild binaries:
export GOOS=linux
export GOARCH=amd64
bin_path="${build_path}/${GOOS}/x86_64"
mkdir -p "${bin_path}"
_info "building snapd/snapctl for ${GOOS}/${GOARCH}"
"${go_build[@]}" -o "${bin_path}/snapd" . || exit 1
(cd "${__proj_dir}/cmd/snapctl" && "${go_build[@]}" -o "${bin_path}/snapctl" . || exit 1)

_info "building plugins for ${GOOS}/${GOARCH}"
find "${__proj_dir}/plugin/" -type d -iname "snap-*" -print0 | xargs -0 -n 1 -I{} "${__dir}/build_plugin.sh" {}

export GOOS=darwin
export GOARCH=amd64
bin_path="${build_path}/${GOOS}/x86_64"
mkdir -p "${bin_path}"
_info "building snapd/snapctl for ${GOOS}/${GOARCH}"
"${go_build[@]}" -o "${bin_path}/snapd" . || exit 1
(cd "${__proj_dir}/cmd/snapctl" && "${go_build[@]}" -o "${bin_path}/snapctl" . || exit 1)

_info "building plugins for ${GOOS}/${GOARCH}"
find "${__proj_dir}/plugin/" -type d -iname "snap-*" -print0 | xargs -0 -n 1 -I{} "${__dir}/build_plugin.sh" {}
