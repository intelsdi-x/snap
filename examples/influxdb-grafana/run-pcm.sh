#!/bin/bash

die () {
    echo >&2 "$@"
    exit 1
}

[ "$#" -eq 1 ] || die "Error: Expected to get one or more machine names as arguments."
[ "${PULSE_PATH}x" != "x" ] || die "Error: PULSE_PATH must be set"
command -v docker-machine >/dev/null 2>&1 || die "Error: docker-machine is required."
command -v docker-compose >/dev/null 2>&1 || die "Error: docker-compose is required."
command -v docker >/dev/null 2>&1 || die "Error: docker is required."
command -v netcat >/dev/null 2>&1 || die "Error: netcat is required."


#docker machine ip
dm_ip=$(docker-machine ip $1) || die 
echo ">>docker machine ip: ${dm_ip}"

#start containers
docker-compose up -d

echo -n ">>waiting for influxdb and grafana to start"

# wait for influxdb to start up
while ! curl --silent -G "http://${dm_ip}:8086/query?u=admin&p=admin" --data-urlencode "q=SHOW DATABASES" 2>&1 > /dev/null ; do   
  sleep 1 
  echo -n "."
done
echo ""

#influxdb IP 
influx_ip=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' influxdbgrafana_influxdb_1)
echo ">>influxdb ip: ${influx_ip}"

# create pulse database in influxdb
curl -G "http://${dm_ip}:8086/ping"
echo -n ">>deleting pulse influx db (if it exists) => "
curl -G "http://${dm_ip}:8086/query?u=admin&p=admin" --data-urlencode "q=DROP DATABASE pulse"
echo ""
echo -n ">>creating pulse influx db => "
curl -G "http://${dm_ip}:8086/query?u=admin&p=admin" --data-urlencode "q=CREATE DATABASE pulse"
echo ""

# create influxdb datasource in grafana
echo -n ">>adding influxdb datasource to grafana => "
COOKIEJAR=$(mktemp -t 'pulse-tmp')
curl -H 'Content-Type: application/json;charset=UTF-8' \
	--data-binary '{"user":"admin","email":"","password":"admin"}' \
    --cookie-jar "$COOKIEJAR" \
    "http://${dm_ip}:3000/login"

curl --cookie "$COOKIEJAR" \
	-X POST \
	--silent \
	-H 'Content-Type: application/json;charset=UTF-8' \
	--data-binary "{\"name\":\"influx\",\"type\":\"influxdb\",\"url\":\"http://${influx_ip}:8086\",\"access\":\"proxy\",\"database\":\"pulse\",\"user\":\"admin\",\"password\":\"admin\"}" \
	"http://${dm_ip}:3000/api/datasources"
echo ""

dashboard=$(cat $PULSE_PATH/../examples/influxdb-grafana/grafana/pcm-psutil.json)
curl --cookie "$COOKIEJAR" \
	-X POST \
	--silent \
	-H 'Content-Type: application/json;charset=UTF-8' \
	--data "$dashboard" \
	"http://${dm_ip}:3000/api/dashboards/db"
echo ""

echo -n ">>starting pulsed"
$PULSE_PATH/bin/pulsed --log-level 1 --auto-discover $PULSE_PATH/plugin -t 0 > /tmp/pulse.out 2>&1  &
echo ""

sleep 3

$PULSE_PATH/bin/pulsectl plugin load /Users/jrcookli/.gvm/pkgsets/go1.5/pulse/src/github.com/intelsdi-x/pulse-plugin-publisher-influxdb/build/rootfs/pulse-plugin-publisher-influxdb

echo -n ">>adding pulse task"
TASK="${TMPDIR}/pulse-task-$$.json"
echo "$TASK"
cat $PULSE_PATH/../examples/tasks/pcm-influx.json | sed s/INFLUXDB_IP/${dm_ip}/ > $TASK 
$PULSE_PATH/bin/pulsectl task create -t $TASK

sleep 1

echo ">>starting pulse task"
$PULSE_PATH/bin/pulsectl task start 1

echo ""
echo "Grafana Dashboard => http://${dm_ip}:3000/dashboard/db/pulse-dashboard"
echo "Influxdb UI       => http://${dm_ip}:8083"
echo ""
echo "Press enter to start viewing the pulse.log" 
read 
tail -f /tmp/pulse.out

