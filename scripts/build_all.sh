#!/usr/bin/env bash

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# shellcheck source=scripts/common.sh
. "${__dir}/common.sh"

export GOOS=linux
export GOARCH=amd64
"${__dir}/build_snap.sh" &
"${__dir}/build_plugins.sh" &

export GOOS=darwin
export GOARCH=amd64
"${__dir}/build_snap.sh" &
"${__dir}/build_plugins.sh" &

export GOOS=windows
export GOARCH=amd64
"${__dir}/build_snap.sh" &
"${__dir}/build_plugins.sh" &

export GOOS=windows
export GOARCH=386
"${__dir}/build_snap.sh" &
"${__dir}/build_plugins.sh" &

wait
