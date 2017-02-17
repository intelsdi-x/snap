#!/usr/bin/env bash

set -e
set -u
set -o pipefail

LOG_LEVEL="${LOG_LEVEL:-6}"
NO_COLOR="${NO_COLOR:-}"

trap_exitcode() {
  exit $?
}

trap trap_exitcode SIGINT

_go_get() {
  local _url=$1
  local _util

  _util=$(basename "${_url}")

  type -p "${_util}" > /dev/null || go get "${_url}" && _debug "go get ${_util} ${_url}"
}

_fmt () {
  local color_debug="\x1b[35m"
  local color_info="\x1b[32m"
  local color_notice="\x1b[34m"
  local color_warning="\x1b[33m"
  local color_error="\x1b[31m"
  local colorvar=color_$1

  local color="${!colorvar:-$color_error}"
  local color_reset="\x1b[0m"
  if [ "${NO_COLOR}" = "true" ] || [[ "${TERM:-}" != "xterm"* ]] || [ -t 1 ]; then
    # Don't use colors on pipes or non-recognized terminals
    color=""; color_reset=""
  fi
  echo -e "$(date -u +"%Y-%m-%d %H:%M:%S UTC") ${color}$(printf "[%9s]" "${1}")${color_reset}";
}

_debug ()   { [ "${LOG_LEVEL}" -ge 7 ] && echo "$(_fmt debug) ${*}" 1>&2 || true; }
_info ()    { [ "${LOG_LEVEL}" -ge 6 ] && echo "$(_fmt info) ${*}" 1>&2 || true; }
_notice ()  { [ "${LOG_LEVEL}" -ge 5 ] && echo "$(_fmt notice) ${*}" 1>&2 || true; }
_warning () { [ "${LOG_LEVEL}" -ge 4 ] && echo "$(_fmt warning) ${*}" 1>&2 || true; }
_error ()   { [ "${LOG_LEVEL}" -ge 3 ] && echo "$(_fmt error) ${*}" 1>&2 || true; exit 1; }

test_dirs=$(find . -type f -name '*.go' -not -path "./.*" -not -path "*/_*" -not -path "./Godeps/*" -not -path "./vendor/*" -not -path "./scripts/*" -print0 | xargs -0 -n1 dirname | sort -u)

_debug "go code directories:
${test_dirs}"

_go_get() {
  local _url=$1
  local _util

  _util=$(basename "${_url}")

  type -p "${_util}" > /dev/null || go get "${_url}" && _debug "go get ${_util} ${_url}"
}

_path_prepend() {
  if [ -d "$1" ] && [[ ":$PATH:" != *":$1:"* ]]; then
    PATH="$1${PATH:+":$PATH"}"
    _debug "Update PATH: ${PATH}"
  fi
}

_go_path() {
  [[ ! -z $GOPATH ]] || _error "Error \$GOPATH unset"

  _debug "GOPATH: ${GOPATH}"
  _debug "PATH: ${PATH}"

  # NOTE: handles colon separated gopath
  go_bin_path=${GOPATH//://bin:}/bin

  _path_prepend "${go_bin_path}"
}

_goimports() {
  _go_get golang.org/x/tools/cmd/goimports
  test -z "$(goimports -l -d $(find . -type f -name '*.go' -not -path "./vendor/*") | tee /dev/stderr)"
}

_gofmt() {
  test -z "$(gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*") | tee /dev/stderr)"
}

_golint() {
  _go_get github.com/golang/lint/golint
  golint ./...
}

_go_vet() {
  go vet ${test_dirs}
}

_go_race() {
  go test -race ./...
}

_go_test() {
  _info "running test type: ${TEST_TYPE}"
  # Standard go tooling behavior is to ignore dirs with leading underscors
  for dir in $test_dirs;
  do
    if [[ -z ${go_cover+x} ]]; then
      _debug "running go test with cover in ${dir}"
      go test --tags="${TEST_TYPE}" -covermode=count -coverprofile="${dir}/profile.tmp" "${dir}"
      if [ -f "${dir}/profile.tmp" ]; then
        tail -n +2 "${dir}/profile.tmp" >> "profile-${TEST_TYPE}.cov"
        rm "${dir}/profile.tmp"
      fi
    else
      _debug "running go test without cover in ${dir}"
      go test --tags="${TEST_TYPE}" "${dir}"
    fi
  done
}

_go_cover() {
  go tool cover -func "profile-${TEST_TYPE}.cov"
}

_git_version() {
  git_branch=$(git symbolic-ref HEAD 2> /dev/null | cut -b 12-)
  git_branch="${git_branch:-test}"
  git_sha=$(git log --pretty=format:"%h" -1)
  git_version=$(git describe --always --exact-match 2> /dev/null || echo "${git_branch}-${git_sha}")
  echo "${git_version}"
}
