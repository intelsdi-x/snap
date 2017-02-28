#!/usr/bin/env bash

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

set -e

echo "Checking for proto"
if ! which protoc > /dev/null
then
	echo "Error: protoc not installed" >&2
	exit 1
fi

if ! protoc --version | grep 'libprotoc 3\.' > /dev/null
then
	echo "Error: this project requires protobuf 3" >&2
	exit 1
fi

if ! which protoc-gen-go > /dev/null
then
	echo "Error: protoc-gen-go not installed. try : go get github.com/golang/protobuf/protoc-gen-go" >&2
	exit 1
fi

proto_files=("grpc/controlproxy/rpc/control.proto" "control/plugin/rpc/plugin.proto")
pb_go_files=("grpc/controlproxy/rpc/control.pb.go" "control/plugin/rpc/plugin.pb.go")

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
common_path="src/github.com/intelsdi-x/snap"
#generate common
echo "Generating common proto files"
protoc --go_out=plugins:src/ --proto_path=src/ "$common_path"/grpc/common/*.proto
echo "$license" | cat - "$common_path"/grpc/common/common.pb.go > temp && mv temp "$common_path"/grpc/common/common.pb.go
#generate all others

for i in "${!proto_files[@]}"
do
	file="${proto_files[$i]}"
	pb="${pb_go_files[$i]}"    
	echo "Generating $pb"
	protoc --go_out=plugins=grpc:src/ --proto_path=src/ "$common_path"/"$file"
	echo "$license" | cat - "$common_path"/"$pb" > temp && mv temp "$common_path"/"$pb"
done
