
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

makeBinary(){
  go get github.com/intelsdi-x/$1
  echo "${green}loading $1 ${reset}"
  cd $GOPATH/src/github.com/intelsdi-x/$1
  if [ ! -f build/rootfs/$1 ] ; then
   	  make
  fi
  snapctl plugin load $GOPATH/src/github.com/intelsdi-x/$1/build/rootfs/$1 || die "Error: failed to load $1"
}

red=`tput setaf 1`
green=`tput setaf 2`
reset=`tput sgr0`

die () {
    echo >&2 "${red} $@ ${reset}"
    exit 1
}

# verify deps and the env

if [ "${SNAP_PATH}x" == "x" ] 
then 
	   echo "Error: SNAP_PATH must be set, please enter your SNAP_PATH below (Please use an absolute path, no variables)"
	   read SNAP_PATH
	   sleep 1s
	   echo "SNAP_PATH:" $SNAP_PATH
fi


type docker-compose >/dev/null 2>&1 || die "Error: docker-compose is required"
type docker >/dev/null 2>&1 || die "Error: docker is required"
type netcat >/dev/null 2>&1 || die "Error: netcat is required"
type snapd >/dev/null 2>&1 || die "Error: snapd is required"
dm_ip="127.0.0.1"	


#start containers
docker pull grafana/grafana
docker pull influxdb
docker-compose up -d

# wait for influxdb to start up
while ! curl --silent -G "http://${dm_ip}:8086/query?u=admin&p=admin" --data-urlencode "q=SHOW DATABASES" 2>&1 > /dev/null ; do
  sleep 1
  echo -n "."
done
echo ""

#influxdb IP
influx_ip=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' multiplugins_influxdb_1)
echo "influxdb ip: ${influx_ip}"

# create influxdb datasource in grafana
echo "${green}adding influxdb datasource to grafana => ${reset}"
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

dashboard=$(cat $SNAP_PATH/grafana/multi-dashboard.json)
curl --cookie "$COOKIEJAR" \
	-X POST \
	--silent \
	-H 'Content-Type: application/json;charset=UTF-8' \
	--data "$dashboard" \
	"http://${dm_ip}:3000/api/dashboards/db"
echo ""

#start Snap
snapd --log-level 1 --plugin-trust 0 &

#load plugins
makeBinary "snap-plugin-publisher-influxdb"
makeBinary "snap-plugin-collector-cgroups"
makeBinary "snap-plugin-collector-psutil"
makeBinary "snap-plugin-collector-cpu"

echo -n "${greeN}adding task${reset}"
TMPDIR=${TMPDIR:="/tmp"}
TASK="${TMPDIR}/multi-plugins-$$.json"
echo "$TASK"
echo $SNAP_PATH
cat "$SNAP_PATH/../tasks/multi-plugins.json" | sed s/INFLUXDB_IP/${dm_ip}/ > $TASK
snapctl task create -t $TASK

echo ""${green}
echo "Grafana Dashboard => http://${dm_ip}:3000/dashboard/db/snap-dashboard"
echo "Influxdb UI       => http://${dm_ip}:8083"
echo ""
echo "Press enter to start viewing the snap.log${reset}"
read
tail -f /tmp/snap.out

read 

pkill snapd