#!/bin/bash

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

die () {
    echo >&2 "$@"
    exit 1
}

[ "$#" -eq 1 ] || die "Error: Expected to get one or more machine names as arguments."
[ "${SNAP_PATH}x" != "x" ] || die "Error: SNAP_PATH must be set"
command -v docker-machine >/dev/null 2>&1 || die "Error: docker-machine is required."
command -v docker-compose >/dev/null 2>&1 || die "Error: docker-compose is required."
command -v docker >/dev/null 2>&1 || die "Error: docker is required."
command -v netcat >/dev/null 2>&1 || die "Error: netcat is required."
file $SNAP_PATH/plugin/snap-plugin-collector-psutil >/dev/null 2>&1 || die "Error: missing $SNAP_PATH/build/plugin/snap-plugin-collector-psutil"
file $SNAP_PATH/plugin/snap-plugin-publisher-influxdb >/dev/null 2>&1 || die "Error: missing $SNAP_PATH/build/plugin/snap-plugin-publisher-influxdb"



#docker machine ip
dm_ip=$(docker-machine ip $1) || die
echo "docker machine ip: ${dm_ip}"

#start containers
docker-compose up -d

echo -n "waiting for influxdb and grafana to start"

# wait for influxdb to start up
while ! curl --silent -G "http://${dm_ip}:8086/query?u=admin&p=admin" --data-urlencode "q=SHOW DATABASES" 2>&1 > /dev/null ; do
  sleep 1
  echo -n "."
done
echo ""

#influxdb IP
influx_ip=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' influxdbgrafana_influxdb_1)
echo "influxdb ip: ${influx_ip}"

# create snap database in influxdb
curl -G "http://${dm_ip}:8086/ping"
echo -n ">>deleting snap influx db (if it exists) => "
curl -G "http://${dm_ip}:8086/query?u=admin&p=admin" --data-urlencode "q=DROP DATABASE snap"
echo ""
echo -n "creating snap influx db => "
curl -G "http://${dm_ip}:8086/query?u=admin&p=admin" --data-urlencode "q=CREATE DATABASE snap"
echo ""

# create influxdb datasource in grafana
echo -n "adding influxdb datasource to grafana => "
COOKIEJAR=$(mktemp -t 'snap-tmp')
curl -H 'Content-Type: application/json;charset=UTF-8' \
	--data-binary '{"user":"admin","email":"","password":"admin"}' \
    --cookie-jar "$COOKIEJAR" \
    "http://${dm_ip}:3000/login"

curl --cookie "$COOKIEJAR" \
	-X POST \
	--silent \
	-H 'Content-Type: application/json;charset=UTF-8' \
	--data-binary "{\"name\":\"influx\",\"type\":\"influxdb\",\"url\":\"http://${influx_ip}:8086\",\"access\":\"proxy\",\"database\":\"snap\",\"user\":\"admin\",\"password\":\"admin\"}" \
	"http://${dm_ip}:3000/api/datasources"
echo ""

dashboard=$(cat $SNAP_PATH/../examples/influxdb-grafana/grafana/psutil.json)
curl --cookie "$COOKIEJAR" \
	-X POST \
	--silent \
	-H 'Content-Type: application/json;charset=UTF-8' \
	--data "$dashboard" \
	"http://${dm_ip}:3000/api/dashboards/db"
echo ""

echo -n "starting snapd"
$SNAP_PATH/bin/snapd --log-level 1 -t 0 --auto-discover $SNAP_PATH/plugin > /tmp/snap.out 2>&1  &
echo ""

sleep 3

echo -n "adding task "
TASK="${TMPDIR}/snap-task-$$.json"
echo "$TASK"
cat $SNAP_PATH/../examples/tasks/psutil-influx.json | sed s/INFLUXDB_IP/${dm_ip}/ > $TASK
$SNAP_PATH/bin/snapctl task create -t $TASK

echo ""
echo "Grafana Dashboard => http://${dm_ip}:3000/dashboard/db/snap-dashboard"
echo "Influxdb UI       => http://${dm_ip}:8083"
echo ""
echo "Press enter to start viewing the snap.log"
read
tail -f /tmp/snap.out
