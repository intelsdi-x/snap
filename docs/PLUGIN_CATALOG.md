This is the master catalog of plugins for Pulse. The plugins in this list may be written by multiple sources. Please examine the license and documentation of each plugin for more information.

## Maintained plugins

### Intel®

| Name  | Type  | Description | Link |
| :---- | :---- | :---------- | :--- |
| CEPH | Collector | Collects from CEPH cluster | [pulse-plugin-collector-ceph](https://github.com/intelsdi-x/pulse-plugin-collector-ceph)
| Docker | Collector | Collects from Docker engine | [pulse-plugin-collector-docker](https://github.com/intelsdi-x/pulse-plugin-collector-docker)
| Facter | Collector | Collects from Facter | [pulse-plugin-collector-facter](https://github.com/intelsdi-x/pulse-plugin-collector-facter) |
| Libvirt | Collector | Collects from libvirt | [pulse-plugin-collector-libvirt](https://github.com/intelsdi-x/pulse-plugin-collector-libvirt)
| NodeManager | Collector | Collects from Intel Node Manager | [pulse-plugin-collector-node-manager](https://github.com/intelsdi-x/pulse-plugin-collector-node-manager)
| PCM | Collector | Collects from PCM.x | [pulse-plugin-collector-pcm](https://github.com/intelsdi-x/pulse-plugin-collector-pcm)|
| Perfevents | Collector | Collects perfevents from Linux | [pulse-plugin-collector-perfevents](https://github.com/intelsdi-x/pulse-plugin-collector-perfevents)|
| PSUtil | Collector | Collects from psutil | [pulse-plugin-collector-psutil](https://github.com/intelsdi-x/pulse-plugin-collector-psutil) |
| SMART | Collector | Collects SMART metrics from Intel SSDs | [pulse-plugin-collector-smart](https://github.com/intelsdi-x/pulse-plugin-collector-smart) |
 | 
| Movingaverage | Processor | Processes data and outputs movingaverage | [pulse-plugin-processor-movingaverage](https://github.com/intelsdi-x/pulse-plugin-processor-movingaverage) |
 | 
| HANA | Publisher | Writes to SAP HANA Database | [pulse-plugin-publisher-hana](https://github.com/intelsdi-x/pulse-plugin-publisher-hana) | 
| InfluxDB | Publisher | Writes to Influx Database | [pulse-plugin-publisher-influxdb](https://github.com/intelsdi-x/pulse-plugin-publisher-influxdb) |
| Kafka | Publisher | Writs to Kafka messaging system | [pulse-plugin-publisher-kafka](https://github.com/intelsdi-x/pulse-plugin-publisher-kafka) |
| MySQL | Publisher | Writes to MySQL Database | [pulse-plugin-publisher-mysql](https://github.com/intelsdi-x/pulse-plugin-publisher-mysql) |
| OpenTSDB | Publisher | Writes to Opentsdb Database | [pulse-plugin-publisher-opentsdb](https://github.com/intelsdi-x/pulse-plugin-publisher-opentsdb) |
| PostgreSQL | Publisher | Writes to PostgreSQL Database | [pulse-plugin-publisher-postgresql](https://github.com/intelsdi-x/pulse-plugin-publisher-postgresql) |
| RabbitMQ | Publisher | Writes to RabbitMQ | [pulse-plugin-publisher-rabbitmq](https://github.com/intelsdi-x/pulse-plugin-publisher-rabbitmq) |
| Riemann | Publisher | Writes to Riemann monitoring system | [pulse-plugin-publisher-riemann](https://github.com/intelsdi-x/pulse-plugin-publisher-riemann) |

### Third-party

TBD

## Committed plugins
These plugins are in planned/active development. This list is useful if you want to reach out and contribute to the development.

| Name  | Type  | Description | Link | Authors |
| :---- | :---- | :---------- | :--- | :------ |
| Ethtool | Collector | Collect from ethtool stats & registry dump | [@lmroz](https://github.com/lmroz)|
| IOstat | Collector | Collect from IOstat | [pulse-plugin-collector-iostat](https://github.com/intelsdi-x/pulse-plugin-collector-iostat) | [@IzabellaRaulin](https://github.com/IzabellaRaulin) |
| Nova | Collector | Collect from Nova/Libvirt | -| [@sandlbn](https://github.com/sandlbn) |
| Open vSwitch | Collector | Collect Open vSwitch performance data | -| [@sandlbn](https://github.com/sandlbn) |
| OSv | Collector | Collect from OSv | [pulse-plugin-collector-osv](https://github.com/intelsdi-x/pulse-plugin-collector-osv) | [@sandlbn](https://github.com/sandlbn) |

## Wish List
This is a wish list of plugins for Pulse. If you see one here and want to start on it please let us know.
#### Collector

- CollectD native
- Prometheus
- Pulse App Endpoint (needs event spec)
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
