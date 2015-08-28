#!/bin/bash -e

BUILDCMD='go build -a -ldflags "-w"'
BUILDDIR=$1
PLUGIN=$2
PLUGINNAME=`echo $PLUGIN | grep -oh "pulse-.*"` 

echo "    $PLUGINNAME => $BUILDDIR"
$BUILDCMD -o $BUILDDIR/$PLUGINNAME $PLUGIN || exit 2
