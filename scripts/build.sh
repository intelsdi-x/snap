#!/bin/bash

echo "****  Pulse Build  ****"
echo

# Vars
BUILDDIR=build
PLUGINDIR=plugin
BINDIR=$BUILDDIR/bin
COLLECTORDIR=$BUILDDIR/$PLUGINDIR/collector
PUBLISHERDIR=$BUILDDIR/$PLUGINDIR/publisher

# Clean build
rm -rf $BUILDDIR/*

# Make dir
mkdir -p $BINDIR
mkdir -p $COLLECTORDIR
mkdir -p $PUBLISHERDIR

# Binaries
# 
echo " Building Pulse Agent"	
go build -o $BINDIR/pulse-agent . || exit 1

# Built-in Collector Plugin building
cd plugin/collector/
echo " Building Collector Plugin(s)"
for d in *; do
	if [[ -d $d ]]; then
		echo "    $d"		
		go build -o ../../$COLLECTORDIR/$d ./$d/ || exit 2		
		# chmod -x ../../$COLLECTORDIR/$d / for testing non-executable builds
	fi
done
cd ../../

# Built-in Publisher Plugin building

echo
echo "*******************"
