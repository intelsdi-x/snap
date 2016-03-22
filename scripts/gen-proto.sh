#!/bin/bash -e

#http://www.apache.org/licenses/LICENSE-2.0.txt
#
#
#Copyright 2016 Intel Corporation
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

proto_files=("../control/rpc/control.proto" "../scheduler/rpc/scheduler.proto")
proto_paths=("../control/rpc" "../scheduler/rpc")
pb_go_files=("../control/rpc/control.pb.go" "../scheduler/rpc/scheduler.pb.go")

license='/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
'
echo "Generating pb.go files"
for i in "${!proto_files[@]}"
do
	path="${proto_paths[$i]}"
	file="${proto_files[$i]}"
	pb="${pb_go_files[$i]}"
	protoc --go_out=plugins=grpc:"$path" --proto_path="$path" "$file"
	echo "$license" | cat - "$pb" > temp && mv temp "$pb"
done
