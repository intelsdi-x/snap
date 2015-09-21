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


#start virtual machine
docker-machine start $1 || die "Error: cannot start VM $1"

#source docker env
eval "$(docker-machine env $1)" || die "Error: cannot source VM env"

#docker machine ip
dm_ip=$(docker-machine ip $1) || die 
echo "docker machine ip: ${dm_ip}"

#start containers
echo "prepare docker container"
docker stop opentsdb_opentsdb_1 opentsdb_grafana_1 >/dev/null 2>&1 
docker rm opentsdb_opentsdb_1 opentsdb_grafana_1 >/dev/null 2>&1 
docker run -d --name opentsdb_opentsdb_1 -p 4242:4242 opower/opentsdb || die "Error: cannot run the opentsdb container"

echo -n "waiting for opentsdb to start"

# wait for opentsdb to start up
while ! curl --silent -G "http://${dm_ip}:4242" 2>&1 > /dev/null ; do   
  sleep 1 
  echo -n "."
done
echo ""
 
docker run -d --name opentsdb_grafana_1 -p 3000:3000 grafana/grafana || die "Error: cannot run the grafana container"

echo -n "waiting for grafana to start"

# wait for grafana to start up
while ! curl --silent -G "http://${dm_ip}:3000" 2>&1 > /dev/null ; do   
  sleep 1 
  echo -n "."
done
echo ""

# create opentsdb datasource in grafana
opentsdb_ip=${dm_ip}
echo -n "adding opentsdb datasource to grafana => "
COOKIEJAR=$(mktemp -t 'intel-pulse-opentsdb-tmp')
curl -H 'Content-Type: application/json;charset=UTF-8' \
	--data-binary '{"user":"admin","email":"","password":"admin"}' \
    --cookie-jar "$COOKIEJAR" \
    "http://${dm_ip}:3000/login"

curl --cookie "$COOKIEJAR" \
	-X POST \
	--silent \
	-H 'Content-Type: application/json;charset=UTF-8' \
	--data-binary "{\"name\":\"opentsdb\",\"type\":\"opentsdb\",\"url\":\"http://${opentsdb_ip}:4242\",\"access\":\"proxy\"}" \
	"http://${dm_ip}:3000/api/datasources"
echo ""

dashboard=$(cat $PULSE_PATH/../examples/opentsdb/opentsdb-psutil.json)
curl --cookie "$COOKIEJAR" \
  -X POST \
  --silent \
  -H 'Content-Type: application/json;charset=UTF-8' \
  --data "$dashboard" \
  "http://${dm_ip}:3000/api/dashboards/db"
echo ""


echo -n "starting pulsed"
$PULSE_PATH/bin/pulsed --log-level 1 -t 0 --auto-discover $PULSE_PATH/plugin > /tmp/pulse.out 2>&1  &
echo ""

sleep 3

echo -n "adding task "
TASK="${TMPDIR}/pulse-task-$$.json"
echo "$TASK"
cat $PULSE_PATH/../examples/tasks/psutil-opentsdb.json | sed s/OPENTSDB_IP/${dm_ip}/ > $TASK 
$PULSE_PATH/bin/pulsectl task create -t $TASK

echo "start task"
$PULSE_PATH/bin/pulsectl task start 1

echo ""
echo "Grafana Dashboard 	=> http://${dm_ip}:3000/dashboard"
echo "Opentsdb Dashboard 	=> http://${dm_ip}:4242"

echo ""
echo "Press enter to start viewing the pulse.log" 
read 
tail -f /tmp/pulse.out

