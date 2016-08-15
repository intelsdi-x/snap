This is the master catalog of plugins for snap. The plugins in this list may be written by multiple sources. Please examine the license and documentation of each plugin for more information.

## Maintained plugins
| Name  | Type  | Description | Link | Download |
| :---- | :---- | :---------- | :--- | :------- |
| Apache | Collector | Collects metrics from the Apache Webserver for mod_status| [snap-plugin-collector-apache](https://github.com/intelsdi-x/snap-plugin-collector-apache) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-apache)
| Cassandra | Collector | Collect statistics from Cassandra nodes | [snap-plugin-collector-cassandra](https://github.com/intelsdi-x/snap-plugin-collector-cassandra) |
| CEPH | Collector | Collects from CEPH cluster | [snap-plugin-collector-ceph](https://github.com/intelsdi-x/snap-plugin-collector-ceph) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-ceph)
| Cinder | Collector | Collects from OpenStack Cinder | [snap-plugin-collector-cinder](https://github.com/intelsdi-x/snap-plugin-collector-cinder) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-cinder)
| cgroups | Collector | Collecting cgroups metrics using libcontainer | [snap-plugin-collector-cgroups](https://github.com/intelsdi-x/snap-plugin-collector-cgroups) |
| CPU | Collector | Collects CPU metrics from Linux procfs | [snap-plugin-collector-cpu](https://github.com/intelsdi-x/snap-plugin-collector-cpu) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-cpu)
| DBI | Collector | Collects metrics as a result of executing SQL statements on a DB (MySQL and PostgreSQL supported) | [snap-plugin-collector-dbi](https://github.com/intelsdi-x/snap-plugin-collector-dbi) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-dbi)
| Df | Collector | Collects disk space metrics from `df` Linux tool | [snap-plugin-collector-df](https://github.com/intelsdi-x/snap-plugin-collector-df) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-df)
| Disk | Collector | Collects disk related metrics from Linux procfs | [snap-plugin-collector-disk](https://github.com/intelsdi-x/snap-plugin-collector-disk) |
| Docker | Collector | Collects from Docker engine | [snap-plugin-collector-docker](https://github.com/intelsdi-x/snap-plugin-collector-docker) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-docker)
| Elasticsearch | Collector | Collects metrics from Elasticsearch cluster | [snap-plugin-collector-elasticsearch](https://github.com/intelsdi-x/snap-plugin-collector-elasticsearch) |
| Etcd | Collector | Collects metrics from the Etcd `/metrics` endpoint. | [snap-plugin-collector-etcd](https://github.com/intelsdi-x/snap-plugin-collector-etcd) |
| Ethtool | Collector | Collect from ethtool stats & registry dump |[snap-plugin-collector-ethtool](https://github.com/intelsdi-x/snap-plugin-collector-ethtool) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-ethtool)
| Exec | Collector | Collect output from any executable |[snap-plugin-collector-exec](https://github.com/intelsdi-x/snap-plugin-collector-exec) |
| Facter | Collector | Collects from Facter | [snap-plugin-collector-facter](https://github.com/intelsdi-x/snap-plugin-collector-facter) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-facter) &#124; [Darwin](https://s3-us-west-1.amazonaws.com/snap-plugins-darwin-latest/snap-plugin-collector-facter)
| File | Publisher | Publishes snap metrics to a file as JSON | [snap-plugin-publisher-file](https://github.com/intelsdi-x/snap-plugin-publisher-file) |
| Glance | Collector | Collects metrics from OpenStack Glance | [snap-plugin-collector-glance](https://github.com/intelsdi-x/snap-plugin-collector-glance) |
| HAProxy | Collector | Collects metrics from HAProxy | [snap-plugin-collector-haproxy](https://github.com/intelsdi-x/snap-plugin-collector-haproxy) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-haproxy)
| HEKA | Publisher | Publishes snap metrics into heka via TCP | [snap-plugin-publisher-heka](https://github.com/intelsdi-x/snap-plugin-publisher-heka) |
| InfluxDB | Collector | Collects internal statistics from Influx database. | [snap-plugin-collector-influxdb](https://github.com/intelsdi-x/snap-plugin-collector-influxdb) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-influxdb)
| Interface | Collector | Collects network interfaces metrics from Linux procfs | [snap-plugin-collector-interface](https://github.com/intelsdi-x/snap-plugin-collector-interface) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-interface)
| IOstat | Collector | Collect from IOstat | [snap-plugin-collector-iostat](https://github.com/intelsdi-x/snap-plugin-collector-iostat) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-iostat)
| Keystone | Collector | Collects from OpenStack Keystone | [snap-plugin-collector-keystone](https://github.com/intelsdi-x/snap-plugin-collector-keystone) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-keystone)
| Libvirt | Collector | Collects from libvirt | [snap-plugin-collector-libvirt](https://github.com/intelsdi-x/snap-plugin-collector-libvirt) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-libvirt)
| Load | Collector | Collects plaform load metrics from Linux procfs | [snap-plugin-collector-load](https://github.com/intelsdi-x/snap-plugin-collector-load) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-load)
| Meminfo | Collector | Collects memory related metrics from Linux procfs | [snap-plugin-collector-meminfo](https://github.com/intelsdi-x/snap-plugin-collector-meminfo) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-meminfo)
| Mesos | Collector | Collects metrics from an Apache Mesos cluster | [snap-plugin-collector-mesos](https://github.com/intelsdi-x/snap-plugin-collector-mesos) | 
| MySQL | Collector | Collects metrics from MySQL DB | [snap-plugin-collector-mysql](https://github.com/intelsdi-x/snap-plugin-collector-mysql) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-publisher-mysql) &#124; [Darwin](https://s3-us-west-1.amazonaws.com/snap-plugins-darwin-latest/snap-plugin-publisher-mysql)
| Neutron | Collector | Collect from OpenStack Neutron | [snap-plugin-collector-neutron](https://github.com/intelsdi-x/snap-plugin-collector-neutron) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-neutron)
| NFS Client | Collector | Collect NFS client counters and RPC data | [snap-plugin-collector-nfsclient](https://github.com/intelsdi-x/snap-plugin-collector-nfsclient) |
| NodeManager | Collector | Collects from Intel Node Manager | [snap-plugin-collector-node-manager](https://github.com/intelsdi-x/snap-plugin-collector-node-manager) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-node-manager)
| Nova | Collector | Collect from OpenStack Nova | [snap-plugin-collector-nova](https://github.com/intelsdi-x/snap-plugin-collector-nova) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-nova)
| OSv | Collector | Collect from OSv | [snap-plugin-collector-osv](https://github.com/intelsdi-x/snap-plugin-collector-osv) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-osv)
| PCM | Collector | Collects from PCM.x | [snap-plugin-collector-pcm](https://github.com/intelsdi-x/snap-plugin-collector-pcm)| [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-pcm) &#124; [Darwin](https://s3-us-west-1.amazonaws.com/snap-plugins-darwin-latest/snap-plugin-collector-pcm)
| Perfevents | Collector | Collects perfevents from Linux | [snap-plugin-collector-perfevents](https://github.com/intelsdi-x/snap-plugin-collector-perfevents)| [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-perfevents)
| Processes | Collector | Collects processes metrics from Linux procfs | [snap-plugin-collector-processes](https://github.com/intelsdi-x/snap-plugin-collector-processes) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-processes)
| PSUtil | Collector | Collects from psutil | [snap-plugin-collector-psutil](https://github.com/intelsdi-x/snap-plugin-collector-psutil) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-psutil) &#124; [Darwin](https://s3-us-west-1.amazonaws.com/snap-plugins-darwin-latest/snap-plugin-collector-psutil)
| RabbitMQ | Collector | Collects from RabbitMQ | [snap-plugin-collector-rabbitmq](https://github.com/intelsdi-x/snap-plugin-collector-rabbitmq) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-rabbitmq) &#124; [Darwin](https://s3-us-west-1.amazonaws.com/snap-plugins-darwin-latest/snap-plugin-collector-rabbitmq)
| SMART | Collector | Collects SMART metrics from Intel SSDs | [snap-plugin-collector-smart](https://github.com/intelsdi-x/snap-plugin-collector-smart) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-smart)
| SNMP | Collector | This plugin collects metrics using SNMP (Simple Network Management Protocol) | [snap-plugin-collector-snmp](https://github.com/intelsdi-x/snap-plugin-collector-snmp) 
| Swap | Collector | Collects swap related metrics from Linux procfs | [snap-plugin-collector-swap](https://github.com/intelsdi-x/snap-plugin-collector-swap) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-swap)
| Users | Collector | Collects users related metrics from Linux utmp | [snap-plugin-collector-users](https://github.com/intelsdi-x/snap-plugin-collector-users) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-users)
| Movingaverage | Processor | Processes data and outputs moving average | [snap-plugin-processor-movingaverage](https://github.com/intelsdi-x/snap-plugin-processor-movingaverage) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-processor-movingaverage) &#124; [Darwin](https://s3-us-west-1.amazonaws.com/snap-plugins-darwin-latest/snap-plugin-processor-movingaverage)
| Etcd | Publisher | Publishes metrics to Etcd | [snap-plugin-publisher-etcd](https://github.com/intelsdi-x/snap-plugin-publisher-etcd) |
| Graphite | Publisher | Publishes snap metrics to graphite | [snap-plugin-publisher-graphite](https://github.com/intelsdi-x/snap-plugin-publisher-graphite) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-publisher-graphite) &#124; [Darwin](https://s3-us-west-1.amazonaws.com/snap-plugins-darwin-latest/snap-plugin-publisher-graphite)
| HANA | Publisher | Writes to SAP HANA Database | [snap-plugin-publisher-hana](https://github.com/intelsdi-x/snap-plugin-publisher-hana) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-publisher-hana) &#124; [Darwin](https://s3-us-west-1.amazonaws.com/snap-plugins-darwin-latest/snap-plugin-publisher-hana)
| InfluxDB | Publisher | Writes to Influx Database | [snap-plugin-publisher-influxdb](https://github.com/intelsdi-x/snap-plugin-publisher-influxdb) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-publisher-influxdb)
| Kafka | Publisher | Writes to Kafka messaging system | [snap-plugin-publisher-kafka](https://github.com/intelsdi-x/snap-plugin-publisher-kafka) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-publisher-kafka) &#124; [Darwin](https://s3-us-west-1.amazonaws.com/snap-plugins-darwin-latest/snap-plugin-publisher-kafka)
| KairosDB | Publisher | Writes to KairosDB Database | [snap-plugin-publisher-kairosdb](https://github.com/intelsdi-x/snap-plugin-publisher-kairosdb) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-publisher-opentsdb) &#124; [Darwin](https://s3-us-west-1.amazonaws.com/snap-plugins-darwin-latest/snap-plugin-publisher-kairosdb)
| MySQL | Publisher | Writes to MySQL Database | [snap-plugin-publisher-mysql](https://github.com/intelsdi-x/snap-plugin-publisher-mysql) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-publisher-mysql)
| OpenFOAM | Collector | Collect metrics from OpenFOAM | [snap-plugin-collector-openfoam](https://github.com/intelsdi-x/snap-plugin-collector-openfoam) |
| OpenTSDB | Publisher | Writes to OpenTSDB Database | [snap-plugin-publisher-opentsdb](https://github.com/intelsdi-x/snap-plugin-publisher-opentsdb) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-publisher-opentsdb) &#124; [Darwin](https://s3-us-west-1.amazonaws.com/snap-plugins-darwin-latest/snap-plugin-publisher-opentsdb)
| PostgreSQL | Publisher | Writes to PostgreSQL Database | [snap-plugin-publisher-postgresql](https://github.com/intelsdi-x/snap-plugin-publisher-postgresql) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-publisher-postgresql) &#124; [Darwin](https://s3-us-west-1.amazonaws.com/snap-plugins-darwin-latest/snap-plugin-publisher-postgresql)
| RabbitMQ | Publisher | Writes to RabbitMQ | [snap-plugin-publisher-rabbitmq](https://github.com/intelsdi-x/snap-plugin-publisher-rabbitmq) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-collector-rabbitmq) &#124; [Darwin](https://s3-us-west-1.amazonaws.com/snap-plugins-darwin-latest/snap-plugin-publisher-rabbitmq)
| Riemann | Publisher | Writes to Riemann monitoring system | [snap-plugin-publisher-riemann](https://github.com/intelsdi-x/snap-plugin-publisher-riemann) | [Linux](https://s3-us-west-1.amazonaws.com/snap-plugins-linux-latest/snap-plugin-publisher-riemann) &#124; [Darwin](https://s3-us-west-1.amazonaws.com/snap-plugins-darwin-latest/snap-plugin-publisher-riemann)
| Statistics | Processor | Processes data and return statistics over a sliding window | [snap-plugin-processor-statistics](https://github.com/intelsdi-x/snap-plugin-processor-statistics) |
| Tag | Processor | Processes data and add tags | [snap-plugin-processor-tag](https://github.com/intelsdi-x/snap-plugin-processor-tag) |
| Anomaly Detection | Processor | Process data and hightlight outliers | [snap-plugin-processor-anomalydetection](https://github.com/intelsdi-x/snap-plugin-processor-anomalydetection) |
| USE | Collector | Collect Linux utilization, saturation metrics | [snap-plugin-collector-use](https://github.com/intelsdi-x/snap-plugin-collector-use) |

## Community Plugins
| Name  | Type  | Description | Link |
| :---- | :---- | :---------- | :--- |
| CloudWatch | Publisher | Publishes snap metrics to AWS CloudWatch | [snap-plugin-publisher-cloudwatch](https://github.com/Ticketmaster/snap-plugin-publisher-cloudwatch) |
| Ping | Collector | Collects Ping latency measurements | [snap-plugin-collector-ping](https://github.com/raintank/snap-plugin-collector-ping) |
| Memcached | Collector | Collect Memcached performance stats | [snap-plugin-collector-memcache](https://github.com/raintank/snap-plugin-collector-memcache)|
| Blueflood | Publisher | Publishes metrics to the Blueflood metrics processing system | [snap-plugin-publisher-blueflood](https://github.com/Staples-Inc/snap-plugin-publisher-blueflood)|

## Committed plugins
These plugins are in planned/active development. This list is useful if you want to reach out and contribute to the development.

| Name  | Type  | Description | Link | Authors |
| :---- | :---- | :---------- | :--- | :------ |
| Open vSwitch | Collector | Collects Open vSwitch performance data | -| [@sandlbn](https://github.com/sandlbn) |
| Redfish | Collector | Collects metrics from Redfish API | - | [@candysmurf](https://github.com/candysmurf) |
| Cassandra | Publisher | Publishes snap metrics into Cassandra | - | [@candysmurf](https://github.com/candysmurf) |

## Wish List
This is a wish list of plugins for Snap. If you see one here and want to start on it please let us know.
#### Collector

- CollectD native
- Prometheus
- snap App Endpoint (needs event spec)
- Intel NIC
- Kubernetes Minion
- JVM (via JMX)

#### Processor

- Caffe
- Oslo

#### Publisher

- 0MQ
- ActiveMQ
- SQLite
- Ceilometer (possibly just OSLO proc + RMQ)
