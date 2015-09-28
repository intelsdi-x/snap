#http://www.apache.org/licenses/LICENSE-2.0.txt
#
#
#Copyright 2015 Intel Coporation
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

#!/bin/bash -e
# Update path in order for Godeps to function
export PATH=$PATH:$GOPATH/bin

USER="intelsdi-x"
REPO="pulse"

# Set GITVERSION that will be used in release. If tag specified during make release call,
# then use that version and tag the repo so pulse version corresponds to tag.
if [ -n "$1" ]; then
	GITVERSION=$1
else
	GITVERSION=`git describe --always`
fi

# Collect release data before tagging repository if tag specified
LATEST_TAG=$(git describe --tags `git rev-list --tags --max-count=1`)
if [ -n "$2" ]; then
	COMMIT=$2
	COMPARISON="$LATEST_TAG..$COMMIT"
	github-release tag create -u $USER -r $REPO -t $GITVERSION -m $GITVERSION -c $COMMIT
else
	COMPARISON="$LATEST_TAG..HEAD"
	github-release tag create -u $USER -r $REPO -t $GITVERSION -m $GITVERSION
fi
CHANGELOG=`git log $COMPARISON --oneline --no-merges --reverse`

# Fetch the tag we created to version the binaries
git fetch -t --all

git checkout $GITVERSION
# Build Pulse for each OS and arch specified.
# Currently support:
#   - Darwin (Mac OS X) 64bit
#   - Linux 64bit
for GOOS in darwin linux; do
	GOARCH=amd64
	echo "Building Pulse ($GITVERSION) for $GOOS-$GOARCH"
        export GOOS=$GOOS
        export GOARCH=$GOARCH

        make

	echo "Preparing Pulse release $GITVERSION-$GOOS-$GOARCH"
	ARCH_DIR=build/release/$GOOS-$GOARCH
        RELEASE_DIR=$ARCH_DIR/pulse-$GITVERSION
	DIST_DIR=build/release/dist/$GITVERSION
        mkdir -p $RELEASE_DIR/{bin,plugin}
	mkdir -p $DIST_DIR
	cp build/bin/* $RELEASE_DIR/bin
        cp build/plugin/* $RELEASE_DIR/plugin
	tar czf $DIST_DIR/pulse-$GITVERSION-$GOOS-$GOARCH.tar.gz -C $ARCH_DIR pulse-$GITVERSION/bin/
	tar czf $DIST_DIR/pulse-plugins-$GITVERSION-$GOOS-$GOARCH.tar.gz -C $ARCH_DIR pulse-$GITVERSION/plugin/
done

echo "Pushing Pulse release $GITVERSION"
github-release release create \
	-u $USER \
	-r $REPO \
	-t $GITVERSION \
        -n $GITVERSION \
	-o "**CHANGELOG**<br/>$CHANGELOG" \
	-f "$DIST_DIR/*" \
	-p

git checkout master
