#!/bin/bash

# This script assumes you have docker machine installed, pass the machine name, and have docker client pointed to it.
# After we add docker support to integration testing this should be removed. This is purely for laptop tesing this plugin before PR.

docker-machine start $1
eval "$(docker-machine env $1)"
docker run -d -p 2181:2181 -p 9092:9092 --env ADVERTISED_HOST=`docker-machine ip $1` --env ADVERTISED_PORT=9092 spotify/kafka
export PULSE_TEST_ZOOKEEPER=`docker-machine ip $1`:2181
export PULSE_TEST_KAFKA=`docker-machine ip $1`:9092
sleep 3
go test
docker-machine stop $1
