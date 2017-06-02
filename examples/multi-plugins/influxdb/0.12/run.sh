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

# add some color to the output
red=`tput setaf 1`
green=`tput setaf 2`
reset=`tput sgr0`

die () {
    echo >&2 "${red} $@ ${reset}"
    exit 1
}

# verify deps and the env
[ "${SNAP_PATH}x" != "x" ] || die "Error: SNAP_PATH must be set"
type docker-compose >/dev/null 2>&1 || die "Error: docker-compose is required"
type docker >/dev/null 2>&1 || die "Error: docker is required"
type netcat >/dev/null 2>&1 || die "Error: netcat is required"

#start containers
docker run -d -p 8083:8083 -p 8086:8086 --name="influxdb" influxdb
docker run -d -p 3000:3000 --link=influxdb --name="grafana" grafana/grafana

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
echo -n ">>deleting 'telemetry playground' influx db (if it exists) => "
curl -G "http://${dm_ip}:8086/query?u=admin&p=admin" --data-urlencode "q=DROP DATABASE playground"
echo ""
echo -n "creating 'telemetry playground' influx db => "
curl -G "http://${dm_ip}:8086/query?u=admin&p=admin" --data-urlencode "q=CREATE DATABASE playground"
echo ""

# create influxdb datasource in grafana
echo -n "${green}adding influxdb datasource to grafana => ${reset}"
COOKIEJAR=$(mktemp)
curl -H 'Content-Type: application/json;charset=UTF-8' \
	--data-binary '{"user":"admin","email":"","password":"admin"}' \
    --cookie-jar "$COOKIEJAR" \
    "http://${dm_ip}:3000/login"

curl --cookie "$COOKIEJAR" \
	-X POST \
	--silent \
	-H 'Content-Type: application/json;charset=UTF-8' \
	--data-binary "{\"name\":\"snap\",\"type\":\"influxdb\",\"url\":\"http://${influx_ip}:8086\",\"access\":\"direct\",\"database\":\"playground\",\"user\":\"admin\",\"password\":\"admin\"}" \
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

echo "${green}getting and building snap-plugin-publisher-influxdb${reset}"
go get github.com/intelsdi-x/snap-plugin-publisher-influxdb
# try and build; If the build first fails try again also getting deps else stop with an error
(cd $SNAP_PATH/../../snap-plugin-publisher-influxdb && make all) || (cd $SNAP_PATH/../../snap-plugin-publisher-influxdb && make) || die "Error: failed to get and compile influxdb plugin" 

echo "${green}getting and building snap-plugin-collector-cpu${reset}"
go get github.com/intelsdi-x/snap-plugin-collector-cpu
# try and build; If the build first fails try again also getting deps else stop with an error
(cd $SNAP_PATH/../../snap-plugin-collector-cpu && make all) || (cd $SNAP_PATH/../../snap-plugin-collector-cpu && make) || die "Error: failed to get and compile cpu plugin"

echo "${green}getting and building snap-plugin-collector-cpu${reset}"
go get github.com/intelsdi-x/snap-plugin-collector-cgroups
# try and build; If the build first fails try again also getting deps else stop with an error
(cd $SNAP_PATH/../../snap-plugin-collector-cgroups && make all) || (cd $SNAP_PATH/../../snap-plugin-collector-cgroups && make) || die "Error: failed to get and compile cgroups plugin"

echo "${green}getting and building snap-plugin-collector-psutil${reset}"
go get github.com/intelsdi-x/snap-plugin-collector-psutil
# try and build; If the build first fails try again also getting deps else stop with an error
(cd $SNAP_PATH/../../snap-plugin-collector-psutil && make all) || (cd $SNAP_PATH/../../snap-plugin-collector-psutil && make) || die "Error: failed to get and compile psutil plugin"

echo -n "${green}starting snapd${reset}"
$SNAP_PATH/bin/snapd --log-level 1 -t 0  > /tmp/snap.out 2>&1  &
echo ""

sleep 3

echo "${green}loading snap-plugin-publisher-influxdb${reset}"
($SNAP_PATH/bin/snapctl plugin load $SNAP_PATH/../../snap-plugin-publisher-influxdb/build/rootfs/snap-plugin-publisher-influxdb) || die "Error: failed to load influxdb plugin"

echo "${green}loading snap-plugin-collector-psutil${reset}"
($SNAP_PATH/bin/snapctl plugin load $SNAP_PATH/../../snap-plugin-collector-psutil/build/rootfs/snap-plugin-collector-psutil) || die "Error: failed to load psutil plugin"

echo "${green}loading snap-plugin-collector-cpu${reset}"
($SNAP_PATH/bin/snapctl plugin load $SNAP_PATH/../../snap-plugin-collector-psutil/build/rootfs/snap-plugin-collector-psutil) || die "Error: failed to load psutil plugin"

echo "${green}loading snap-plugin-collector-cgroups${reset}"
($SNAP_PATH/bin/snapctl plugin load $SNAP_PATH/../../snap-plugin-collector-psutil/build/rootfs/snap-plugin-collector-psutil) || die "Error: failed to load psutil plugin"

echo -n "${greeN}adding task${reset}"
TMPDIR=${TMPDIR:="/tmp"}
TASK="${TMPDIR}/playground-$$.json"
echo "$TASK"
cat $SNAP_PATH/../examples/tasks/playground.json | sed s/INFLUXDB_IP/${dm_ip}/ > $TASK
$SNAP_PATH/bin/snapctl task create -t $TASK

echo ""${green}
echo "Grafana Dashboard => http://${dm_ip}:3000/dashboard/db/snap-dashboard"
echo "Influxdb UI       => http://${dm_ip}:8083"
echo ""
echo "Press enter to start viewing the snap.log${reset}"
read
tail -f /tmp/snap.out
