#!/bin/bash

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
