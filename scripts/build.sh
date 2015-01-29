#!/bin/bash

echo "****  Pulse Build  ****"
echo

SOURCEDIR=$1
SPLUGINFOLDER=$2
SPLUGIN=$3
BUILDDIR=$SOURCEDIR/build
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
echo "Source Dir = $SOURCEDIR"
echo "$SPLUGIN"
echo "$SPLUGINFOLDER"
echo " Building Pulse Agent"	
go build -o $BINDIR/pulse-agent . || exit 1

if [ "$SPLUGIN" ] && [ -n "$SPLUGINFOLDER" ]
then
	echo " Building Plugin: $SPLUGIN"
	# Built-in single plugin building
	cd $SOURCEDIR/plugin/$SPLUGINFOLDER/
	target=./$SPLUGIN/
	destination=$BUILDDIR/$PLUGINDIR/$SPLUGINFOLDER/$SPLUGIN
	echo "    $SPLUGIN => $destination"	
	go build -o $destination $target || exit 2
	cd $SOURCEDIR
else
	echo " Building Collector Plugin(s)"
	# Built-in Collector Plugin building
	cd $SOURCEDIR/plugin/collector/
	for d in *; do
		if [[ -d $d ]]; then
			echo "    $d => $COLLECTORDIR/$d"		
			go build -o $COLLECTORDIR/$d ./$d/ || exit 2
			# chmod -x ../../$COLLECTORDIR/$d / for testing non-executable builds
		fi
	done
	cd $SOURCEDIR

	# Built-in Publisher Plugin building
fi



echo
echo "*******************"
