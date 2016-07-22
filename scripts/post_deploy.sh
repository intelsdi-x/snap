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

git_sha=$(git log --pretty=format:"%H" -1)
SNAP_BUILD_URL="https://s3-us-west-2.amazonaws.com/intelsdi-x/snap/${git_sha}"
export SNAP_BUILD_URL

_info "updating package URL: ${SNAP_BUILD_URL}"
go get github.com/dnsimple/dnsimple-go/dnsimple
go run "${__proj_dir}/scripts/update_dns.go"
