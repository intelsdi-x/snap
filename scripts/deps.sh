#!/bin/bash
# This script runs the correct godep sequences for pulse and built-in plugins
# This will rebase back to the committed version. It should be run from pulse/.
TYPES=(collector publisher)

function loadDeps() {
	cd $z
	echo "Restoring deps for $z" 
	godep restore
}

function checkPluginType() {	
	for z in plugin/$1/*
	do		
		if [ -d "$z/Godeps" ]; then			 	
			loadDeps $z
		fi
	done
}

# First load pulse deps
godep restore

# Next loop over all plugin types looking for a Godeps dir and loading
for type in ${TYPES[*]}
do	
	checkPluginType $type
done

