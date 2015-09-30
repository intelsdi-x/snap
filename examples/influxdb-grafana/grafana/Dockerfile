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

FROM debian:wheezy

RUN apt-get update && apt-get -y install libfontconfig wget adduser openssl ca-certificates

RUN wget http://grafanarel.s3.amazonaws.com/builds/grafana_latest_amd64.deb

RUN dpkg -i grafana_latest_amd64.deb

RUN apt-get install -y curl netcat 

EXPOSE 3000

VOLUME ["/var/lib/grafana"]
VOLUME ["/var/log/grafana"]
VOLUME ["/etc/grafana"]

WORKDIR /usr/share/grafana

ENTRYPOINT ["/usr/sbin/grafana-server", "--config", "/etc/grafana/grafana.ini"]
	
