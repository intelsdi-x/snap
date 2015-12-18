This is the master catalog of plugins for snap. The plugins in this list may be written by multiple sources. Please examine the license and documentation of each plugin for more information.

## Maintained plugins

### Intel®

| Name  | Type  | Description | Link |
| :---- | :---- | :---------- | :--- |
| CEPH | Collector | Collects from CEPH cluster | [snap-plugin-collector-ceph](https://github.com/intelsdi-x/snap-plugin-collector-ceph)
| Docker | Collector | Collects from Docker engine | [snap-plugin-collector-docker](https://github.com/intelsdi-x/snap-plugin-collector-docker)
| Facter | Collector | Collects from Facter | [snap-plugin-collector-facter](https://github.com/intelsdi-x/snap-plugin-collector-facter) |
| Libvirt | Collector | Collects from libvirt | [snap-plugin-collector-libvirt](https://github.com/intelsdi-x/snap-plugin-collector-libvirt)
| NodeManager | Collector | Collects from Intel Node Manager | [snap-plugin-collector-node-manager](https://github.com/intelsdi-x/snap-plugin-collector-node-manager)
| PCM | Collector | Collects from PCM.x | [snap-plugin-collector-pcm](https://github.com/intelsdi-x/snap-plugin-collector-pcm)|
| Perfevents | Collector | Collects perfevents from Linux | [snap-plugin-collector-perfevents](https://github.com/intelsdi-x/snap-plugin-collector-perfevents)|
| PSUtil | Collector | Collects from psutil | [snap-plugin-collector-psutil](https://github.com/intelsdi-x/snap-plugin-collector-psutil) |
| SMART | Collector | Collects SMART metrics from Intel SSDs | [snap-plugin-collector-smart](https://github.com/intelsdi-x/snap-plugin-collector-smart) |
| OSv | Collector | Collect from OSv | [snap-plugin-collector-osv](https://github.com/intelsdi-x/snap-plugin-collector-osv) |
 | 
| Movingaverage | Processor | Processes data and outputs movingaverage | [snap-plugin-processor-movingaverage](https://github.com/intelsdi-x/snap-plugin-processor-movingaverage) |
 | 
| HANA | Publisher | Writes to SAP HANA Database | [snap-plugin-publisher-hana](https://github.com/intelsdi-x/snap-plugin-publisher-hana) | 
| InfluxDB | Publisher | Writes to Influx Database | [snap-plugin-publisher-influxdb](https://github.com/intelsdi-x/snap-plugin-publisher-influxdb) |
| Kafka | Publisher | Writes to Kafka messaging system | [snap-plugin-publisher-kafka](https://github.com/intelsdi-x/snap-plugin-publisher-kafka) |
| MySQL | Publisher | Writes to MySQL Database | [snap-plugin-publisher-mysql](https://github.com/intelsdi-x/snap-plugin-publisher-mysql) |
| OpenTSDB | Publisher | Writes to Opentsdb Database | [snap-plugin-publisher-opentsdb](https://github.com/intelsdi-x/snap-plugin-publisher-opentsdb) |
| PostgreSQL | Publisher | Writes to PostgreSQL Database | [snap-plugin-publisher-postgresql](https://github.com/intelsdi-x/snap-plugin-publisher-postgresql) |
| RabbitMQ | Publisher | Writes to RabbitMQ | [snap-plugin-publisher-rabbitmq](https://github.com/intelsdi-x/snap-plugin-publisher-rabbitmq) |
| Riemann | Publisher | Writes to Riemann monitoring system | [snap-plugin-publisher-riemann](https://github.com/intelsdi-x/snap-plugin-publisher-riemann) |

### Third-party

TBD

## Committed plugins
These plugins are in planned/active development. This list is useful if you want to reach out and contribute to the development.

| Name  | Type  | Description | Link | Authors |
| :---- | :---- | :---------- | :--- | :------ |
| Ethtool | Collector | Collect from ethtool stats & registry dump |[snap-plugin-collector-ethtool](https://github.com/intelsdi-x/snap-plugin-collector-ethtool)| [@lmroz](https://github.com/lmroz)|
| IOstat | Collector | Collect from IOstat | [snap-plugin-collector-iostat](https://github.com/intelsdi-x/snap-plugin-collector-iostat) | [@IzabellaRaulin](https://github.com/IzabellaRaulin) |
| Nova | Collector | Collect from Nova/Libvirt | -| [@sandlbn](https://github.com/sandlbn) |
| Open vSwitch | Collector | Collect Open vSwitch performance data | -| [@sandlbn](https://github.com/sandlbn) |
| NFS Client | Collector | Collect NFS client counters and RPC data | [snap-plugin-collector-nfsclient](https://github.com/thomastaylor312/snap-plugin-collector-nfsclient) | [@thomastaylor312](https://github.com/thomastaylor312)

## Wish List
This is a wish list of plugins for snap. If you see one here and want to start on it please let us know.
#### Collector

- CollectD native
- Prometheus
- snap App Endpoint (needs event spec)
- Intel NIC
- Kubernetes Minion
- Mesos Slave
- Mesos Master
- OpenStack Nova

#### Processor

- Caffe
- Oslo

#### Publisher

- 0MQ
- ActiveMQ
- SQLite
- Ceilometer (possibly just OSLO proc + RMQ)
