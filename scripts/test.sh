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

# The script does automatic checking on a Go package and its sub-packages, including:
# 1. gofmt         (http://golang.org/cmd/gofmt/)
# 2. goimports     (https://github.com/bradfitz/goimports)
# 3. golint        (https://github.com/golang/lint)
# 4. go vet        (http://golang.org/cmd/vet)
# 5. race detector (http://blog.golang.org/race-detector)
# 6. test coverage (http://blog.golang.org/cover)

COVERALLS_TOKEN=t47LG6BQsfLwb9WxB56hXUezvwpED6D11
TEST_DIRS="cmd/ control/ core/ mgmt/ pkg/ snapd.go scheduler/"
VET_DIRS="./cmd/... ./control/... ./core/... ./mgmt/... ./pkg/... ./scheduler/... ."

set -e

# Automatic checks
echo "gofmt"
test -z "$(gofmt -l -d $TEST_DIRS | tee /dev/stderr)"

echo "goimports"
test -z "$(goimports -l -d $TEST_DIRS | tee /dev/stderr)"

# Useful but should not fail on link per: https://github.com/golang/lint
# "The suggestions made by golint are exactly that: suggestions. Golint is not perfect,
# and has both false positives and false negatives. Do not treat its output as a gold standard.
# We will not be adding pragmas or other knobs to suppress specific warnings, so do not expect
# or require code to be completely "lint-free". In short, this tool is not, and will never be,
# trustworthy enough for its suggestions to be enforced automatically, for example as part of
# a build process"
# echo "golint"
# golint ./...

echo "go vet"
go vet $VET_DIRS
# go test -race ./... - Lets disable for now
 
# Run test coverage on each subdirectories and merge the coverage profile.
echo "mode: count" > profile.cov
 
# Standard go tooling behavior is to ignore dirs with leading underscors
for dir in $(find . -maxdepth 10 -not -path './.git*' -not -path '*/_*' -not -path './examples/*' -not -path './scripts/*' -type d);
do
	if ls $dir/*.go &> /dev/null; then
	    go test -covermode=count -coverprofile=$dir/profile.tmp $dir
	    if [ -f $dir/profile.tmp ]
	    then
	        cat $dir/profile.tmp | tail -n +2 >> profile.cov
	        rm $dir/profile.tmp
	    fi
	fi
done
 
go tool cover -func profile.cov
 
# Disabled Coveralls.io for now
# To submit the test coverage result to coveralls.io,
# use goveralls (https://github.com/mattn/goveralls)
# goveralls -coverprofile=profile.cov -service=travis-ci -repotoken t47LG6BQsfLwb9WxB56hXUezvwpED6D11
#
# If running inside Travis we update coveralls. We don't want his happening on Macs
# if [ "$TRAVIS" == "true" ]
# then
#     n=1
#     until [ $n -ge 6 ]
#     do
#         echo "posting to coveralls attempt $n of 5"
#         goveralls -v -coverprofile=profile.cov -service travis.ci -repotoken $COVERALLS_TOKEN && break
#         n=$[$n+1]
#         sleep 30
#     done
# fi
