#!/bin/bash

echo "****  Pulse Build  ****"

SOURCEDIR=${1:-`pwd`}
# aka pluginType when called from BuildHelper
SPLUGINFOLDER=$2
# aka pluginName
SPLUGIN=$3
BUILDDIR=$SOURCEDIR/build
PLUGINDIR=plugin
BINDIR=$BUILDDIR/bin
AGENT=$BUILDDIR/bin/pulse-agent
COLLECTORDIR=$BUILDDIR/$PLUGINDIR/collector
PUBLISHERDIR=$BUILDDIR/$PLUGINDIR/publisher

# Exit immediately if a command exits with a non-zero status.
set -e

# Make dir
mkdir -p $BINDIR
mkdir -p $COLLECTORDIR
mkdir -p $PUBLISHERDIR

# Binaries
#
echo
echo "Source dir = $SOURCEDIR"
echo "Plugin = $SPLUGIN"
echo "Plugin dir = $SPLUGINFOLDER"
echo "base GOPATH = $GOPATH"
echo 

PULSE_GOPATH=$GOPATH
function mangleGoPath {
  export GOPATH=`godep path`:$PULSE_GOPATH
}

if [ "$SPLUGIN" ] && [ -n "$SPLUGINFOLDER" ]
then
    echo " Building Plugin: $SPLUGIN"
    # Built-in single plugin building
    cd $SOURCEDIR/plugin/$SPLUGINFOLDER/$SPLUGIN
    destination=$BUILDDIR/$PLUGINDIR/$SPLUGINFOLDER/$SPLUGIN
    echo "    $SPLUGIN => $destination"    
    mangleGoPath
    go build -o $destination . || exit 2
else
    # Clean build bin dir and plugin outputs
    echo " Building Pulse Agent"    
    echo "    . => $AGENT"
    rm -f $AGENT
    mangleGoPath
    go build -ldflags "-X main.gitversion `git describe --always`" -o $BINDIR/pulse-agent . || exit 1

    echo " Building Collector Plugin(s)"
    # Built-in Collector Plugin building
    cd $SOURCEDIR/$PLUGINDIR/collector
    for d in *; do
        if [[ -d $d ]]; then
            cd $d
            echo "    $d => $COLLECTORDIR/$d"        
            destination=$COLLECTORDIR/$d
            rm -f $destination
            mangleGoPath
            go build -o $destination  . || exit 2
            # chmod -x ../../$COLLECTORDIR/$d / for testing non-executable builds
            cd $SOURCEDIR/$PLUGINDIR/collector
        fi
    done
 
    echo " Building Publisher Plugin(s)"
    cd $SOURCEDIR/$PLUGINDIR/publisher
    for d in *; do
        if [[ -d $d ]]; then
            cd $d
            echo "    $d => $PUBLISHERDIR/$d"        
            destination=$PUBLISHERDIR/$d
            rm -f $destination
            mangleGoPath
            go build -o $destination . || exit 2
            # chmod -x ../../$PUBLISHERDIR/$d / for testing non-executable builds
            cd $SOURCEDIR/$PLUGINDIR/publisher
        fi
    done
fi

cd $SOURCEDIR

echo
echo "***********************"
