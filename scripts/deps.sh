#!/bin/bash
# This script runs the correct godep sequences for pulse and built-in plugins
# This will rebase back to the committed version. It should be run from pulse/.
declare -a TYPES=(collector publisher)

function loadDeps() {
	cd $z
	echo "Restoring deps for $z"
	godep restore
    cd ..
}

function checkPluginType() {	
	cd plugin/$1
	for z in *;
	do		
		echo "Checking $z for deps"
		if [ -d "$z/Godeps" ]; then			 	
			loadDeps $z
		fi
	done
	cd ../..
}

# First load pulse deps
godep restore

# Next loop over all plugin types looking for a Godeps dir and loading
for type in ${TYPES[*]}
do	
	checkPluginType $type
done

