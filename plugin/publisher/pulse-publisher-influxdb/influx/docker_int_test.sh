#!/bin/bash

# This script assumes you have docker machine installed, pass the machine name, and have docker client pointed to it.
# After we add docker support to integration testing this should be removed. This is purely for laptop tesing this plugin before PR.

#docker-machine start $1
stop_dm=true
if docker-machine ls | grep "${1}" | grep Running ; then
	echo "docker machine is already running"
	stop_dm=false
else
	echo "starting docker machine"
	docker-machine start $1
	stop_dm=true
fi
eval "$(docker-machine env $1)"
#redirecting stdout and err to cleanup output
docker run --name pulse_int_influxdb -d -p 8083:8083 -p 8086:8086 --expose 8090 --expose 8099 tutum/influxdb:staging-0.9.0-rc 2>&1 > /dev/null || docker start pulse_int_influxdb
export PULSE_INFLUXDB_HOST=`docker-machine ip $1`
#redirect stdout and err to cleanup output
curl -G --fail --silent --show-error http://${PULSE_INFLUXDB_HOST}:8086/query --data-urlencode "q=CREATE DATABASE test" > /dev/null
sleep 3
go test
if $stop_dm ; then
	echo "stopping docker machine"
	docker-machine stop $1
fi
#docker-machine stopte $1
