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
# 3. golint        (https://github.com/golang/lint) (disabled)
# 4. go vet        (http://golang.org/cmd/vet) (disabled)
# 5. race detector (http://blog.golang.org/race-detector) (disabled)
# 6. test coverage (http://blog.golang.org/cover)

# If the following plugins don't exist, exit
[ -f $SNAP_PATH/plugin/snap-collector-mock1 ] || { echo 'Error: $SNAP_PATH/plugin/snap-collector-mock1 does not exist. Run make to build it.' ; exit 1; }
[ -f $SNAP_PATH/plugin/snap-collector-mock2 ] || { echo 'Error: $SNAP_PATH/plugin/snap-collector-mock2 does not exist. Run make to build it.' ; exit 1; }
[ -f $SNAP_PATH/plugin/snap-processor-passthru ] || { echo 'Error: $SNAP_PATH/plugin/snap-processor-passthru does not exist. Run make to build it.' ; exit 1; }
[ -f $SNAP_PATH/plugin/snap-publisher-file ] || { echo 'Error: $SNAP_PATH/plugin/snap-publisher-file does not exist. Run make to build it.' ; exit 1; }

TEST_DIRS="cmd/ control/ core/ mgmt/ pkg/ snapd.go scheduler/"
# VET_DIRS="./cmd/... ./control/... ./core/... ./mgmt/... ./pkg/... ./scheduler/... ."

set -e

# If the following tools don't exist, get them
echo "Getting GoConvey if not found"
go get github.com/smartystreets/goconvey
echo "Getting goimports if not found"
go get golang.org/x/tools/cmd/goimports
echo "Getting cover if not found"
go get golang.org/x/tools/cmd/cover

# Automatic checks
echo "gofmt"
test -z "$(gofmt -s -l -d $TEST_DIRS | tee /dev/stderr)"

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

# Disabling running go vet currently due to the inconsistency in it's reported
# outputs depending on whether the project was pulled down via 'go get' or 'git clone'.
# We will look to re-enable go vet checking when we can provide consistency in outputs
# no matter how a developer gets the project.
#echo "go vet"
#go vet $VET_DIRS

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
