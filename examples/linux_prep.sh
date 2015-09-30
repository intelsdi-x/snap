#!/bin/bash

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

HOST_PUBLIC_KEY_FILE="/root/.ssh/id_rsa.pub"
DOCKER_IMAGE_NAME="pulse_image_ssh_container"
DOCKER_CONTAINER_NAME="pulse_ssh_container"
DOCKER_MACHINE_NAME="pulsecontainer"

# Copy host's public key file here. It will be used as authorized_key in docker container (see Dockerfile).
if [[ -f $HOST_PUBLIC_KEY_FILE ]]
	then
		cp $HOST_PUBLIC_KEY_FILE .
	else
		echo "Missing host's public key file. Paswordless communication with container wouldn't be possible. Exiting."
		exit 1
fi

echo "Building Docker image ..."
DOCKER_IMAGE_PRESENT=`docker images | grep "$DOCKER_IMAGE_NAME" | wc -l`
if [ $DOCKER_IMAGE_PRESENT -eq 0 ]
	then
		docker build -t $DOCKER_IMAGE_NAME . || exit 1
	else
		echo "Docker image: $DOCKER_IMAGE_NAME already present. Skipping."
fi

echo "Starting Docker container ..."
DOCKER_CONTAINER_RUNNING=`docker ps | grep "$DOCKER_CONTAINER_NAME" | wc -l`
if [ $DOCKER_CONTAINER_RUNNING -eq 0 ]
	then
		docker run -d -P --name $DOCKER_CONTAINER_NAME $DOCKER_IMAGE_NAME || exit 1
	else
		echo "Docker container: $DOCKER_CONTAINER_NAME already running. Skipping."
fi

# Host's public key no longer needed here.
rm ./id_rsa.pub

echo "Building new docker-machine ..."
DOCKER_MACHINE_RUNNING=`docker-machine ls | grep "$DOCKER_MACHINE_NAME" | wc -l`
if [ $DOCKER_MACHINE_RUNNING -eq 0 ]
	then
		SSH_MAPPING_PORT=`docker port $DOCKER_CONTAINER_NAME 22 | cut -f2 -d":"`
		docker-machine -D create -d generic --generic-ip-address=localhost --generic-ssh-port=$SSH_MAPPING_PORT $DOCKER_MACHINE_NAME
		echo "Container name: $DOCKER_MACHINE_NAME"
	else
		echo "Docker machine: $DOCKER_MACHINE_NAME already running. Skipping."
fi
