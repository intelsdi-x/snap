#!/bin/bash
# Update path in order for Godeps to function
export PATH=$PATH:$GOPATH/bin

REPO="intelsdi-x/pulse"

# Collect release data before tagging repository if tag specified
LATEST_TAG=$(git describe --tags `git rev-list --tags --max-count=1`)
COMPARISON="$LATEST_TAG..HEAD"
CHANGELOG=`git log $COMPARISON --oneline --no-merges --reverse`
BRANCH=`git rev-parse --abbrev-ref HEAD`

# Set GITVERSION that will be used in release. If tag specified during make release call,
# then use that version and tag the repo so pulse version corresponds to tag.
if [ -n "$1" ]; then
	GITVERSION=$1
	git tag -a $GITVERSION -m $GITVERSION
else
	GITVERSION=`git describe --always`
fi

if [ -n "$2" ];  then 
	RELEASE=$2
fi

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
github-release $REPO $GITVERSION $BRANCH "**CHANGELOG**<br/>$CHANGELOG" "$DIST_DIR/*" $RELEASE 
