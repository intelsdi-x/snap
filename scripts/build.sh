#!/bin/bash

GITVERSION=`git describe --always`
SOURCEDIR=$1
SPLUGINFOLDER=$2
SPLUGIN=$3
BUILDDIR=$SOURCEDIR/build
PLUGINDIR=plugin
BINDIR=$BUILDDIR/bin
COLLECTORDIR=$BUILDDIR/$PLUGINDIR/collector
PUBLISHERDIR=$BUILDDIR/$PLUGINDIR/publisher
PROCESSORDIR=$BUILDDIR/$PLUGINDIR/processor
BUILDCMD='go build -a -ldflags "-w"'

echo
echo "****  Pulse Build ($GITVERSION)  ****"
echo

# Disable CGO for builds
export CGO_ENABLED=0

# Clean build bin dir
rm -rf $BINDIR/*

# Make dir
mkdir -p $BINDIR
mkdir -p $COLLECTORDIR
mkdir -p $PUBLISHERDIR
mkdir -p $PROCESSORDIR

# Binaries
#
echo "Source Dir = $SOURCEDIR"
echo " Building Pulse Agent"	
go build -ldflags "-w -X main.gitversion $GITVERSION" -o $BINDIR/pulse-agent . || exit 1

if [ "$SPLUGIN" ] && [ -n "$SPLUGINFOLDER" ]
then
	echo " Building Plugin: $SPLUGIN"
	# Built-in single plugin building
	cd $SOURCEDIR/plugin/$SPLUGINFOLDER/
	target=./$SPLUGIN/
	destination=$BUILDDIR/$PLUGINDIR/$SPLUGINFOLDER/$SPLUGIN
	echo "    $SPLUGIN => $destination"	
	$BUILDCMD -o $destination $target || exit 2
	cd $SOURCEDIR
else
	# Clean build
	rm -rf $COLLECTORDIR/*
	echo " Building Collector Plugin(s)"
	# Built-in Collector Plugin building
	cd $SOURCEDIR/$PLUGINDIR/collector
	for d in *; do
		if [[ -d $d ]]; then
			echo "    $d => $COLLECTORDIR/$d"		
			$BUILDCMD -o $COLLECTORDIR/$d ./$d/ || exit 2
		fi
	done
	
	# Publisher build
	rm -rf $PUBLISHERDIR/*
	echo " Building Publisher Plugin(s)"
	cd $SOURCEDIR/$PLUGINDIR/publisher
	for d in *; do
		if [[ -d $d ]]; then
			echo "    $d => $PUBLISHERDIR/$d"		
			$BUILDCMD -o $PUBLISHERDIR/$d ./$d/ || exit 2
		fi
	done
	
	# Processor build
	rm -rf $PROCESSORDIR/*
	echo " Building Processor Plugin(s)"
	cd $SOURCEDIR/$PLUGINDIR/processor
	for d in *; do
		if [[ -d $d ]]; then
			echo "    $d => $PROCESSORDIR/$d"		
			$BUILDCMD -o $PROCESSORDIR/$d ./$d/ || exit 2			
		fi
	done

	# pulse-ctl
	echo " Building cmd(s)"
	cd $SOURCEDIR/cmd
	for d in *; do
		if [[ -d $d ]]; then
			echo "    $d => $BINDIR/$d"					
			$BUILDCMD -o $BINDIR/$d ./$d/ || exit 3
		fi
	done

	cd $SOURCEDIR
fi

echo
echo "*******************"
