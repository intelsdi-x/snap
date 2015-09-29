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

GITVERSION=`git describe --always`
SOURCEDIR=$1
BUILDPLUGINS=$2
SPLUGINFOLDER=$3
SPLUGIN=$4
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

# pulsed
echo "Source Dir = $SOURCEDIR"
echo " Building Pulse Daemon"
go build -ldflags "-w -X main.gitversion=$GITVERSION" -o $BINDIR/pulsed . || exit 1

# pulsectl
echo " Building Pulse Command Line"
cd $SOURCEDIR/cmd
for d in *; do
	if [[ -d $d ]]; then
		echo "    $d => $BINDIR/$d"
		go build -ldflags "-w -X main.gitversion=$GITVERSION" -o $BINDIR/$d ./$d/ || exit 3
	fi
done


if [ "$BUILDPLUGINS" == "true" ]; then
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
		# Build source plugins into build dir
		rm -rf $BUILDDIR/$PLUGINDIR/*
		cd $SOURCEDIR
		echo " Building Plugin(s)"
		find ./plugin/* -iname "pulse-*" -print0 | xargs -0 -P 4 -n 1 $SOURCEDIR/scripts/build-plugin.sh $BUILDDIR/$PLUGINDIR/ || exit 2		

		cd $SOURCEDIR
	fi
fi

echo
echo "*******************"
