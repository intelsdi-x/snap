#!/bin/bash -e

#http://www.apache.org/licenses/LICENSE-2.0.txt
#
#
#Copyright 2015 Intel Corporation
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

# This script runs the correct godep sequences for pulse and built-in plugins
# This will rebase back to the committed version. It should be run from pulse/.
ctrl_c()
{
  exit $?
} 
trap ctrl_c SIGINT

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
echo "Checking pulse root for deps"
godep restore
# REST API
echo "Checking pulsectl for deps"
cd cmd/pulsectl
godep restore
# CLI
echo "Checking pulse mgmt/rest for deps"
cd ../../mgmt/rest
godep restore
cd ../../


# Next loop over all plugin types looking for a Godeps dir and loading
for type in ${TYPES[*]}
do	
	checkPluginType $type
done

