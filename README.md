# **Pulse** <sup><sub>_A powerful telemetry agent framework from Intel Corporation &reg;_</sub></sup>
[![Build Status](https://magnum.travis-ci.com/intelsdilabs/pulse.svg?token=2ujsxEpZo1issFyVWX29&branch=master)](https://magnum.travis-ci.com/intelsdilabs/pulse) [![Coverage Status](https://coveralls.io/repos/intelsdilabs/pulse/badge.svg?branch=HEAD)](https://coveralls.io/r/intelsdilabs/pulse?branch=HEAD)

## Description

**Pulse** is a framework for enabling the gathering of telemetry from systems. It aims to make collection, processing, and distribution of data _(telemetry)_ from running systems easier to schedule and deploy. Pulse aims to reduce the friction in gathering important system data and getting it to other services that need it.

**Pulse** provides several specific features:

* Three types of plugins: collectors, processors, and publishers.
* Open plugin model allows for loading built-in, community, or proprietary plugins into **Pulse**.
* Dynamic loading, unloading, and swapping of plugins.
* Adding a newer version of a plugin automatically upgrades the next scheduled actions _(when not pinned to a version)_.
* Powerful control using viral clustering of **Pulse** instances into a Tribe.
* Ability to run **Pulse** in distributed role deployment models.
* Extensible source allowing addition of extended scheduling, plugins, routing, and more.
* Multiple management modules including: CLI, REST API, and Web Console. Each of which can be turned on or off.
* CLI control from OSX, Windows, or Linux.

**Pulse** is not intended to:

* Operate as an analytics platform. The intention is to allow plugins for feeding those platforms.
* Compete with existing metric/monitoring/telemetry agents. It is simply a new option to use or reference.

**Pulse** needs:

* Feedback: try it and tell us about it.
* Contributors: We need plugins, schedules, testing, and more.
* Integrations: **Pulse** can feasibly publish to almost any destination. We need publishing plugins for [Ceilometer](https://wiki.openstack.org/wiki/Ceilometer), [vCOPs](http://www.vmware.com/products/vrealize-operations), [Riemann](https://github.com/aphyr/riemann), [InfluxDB](https://github.com/influxdb/influxdb), and more.

## License

<DO NOT PUT SOMETHING HERE YET>

## System Requirements

## Installation

## Building

```
make
```

## Running

## Documentation

### REST API

### CLI

## Contributing

* [Please read our development guidelines](https://github.com/intelsdilabs/pulse/wiki/Development-guidelines)
* [ ] TODO - CLA

## Testing

```
make test
```

## Releases

## Contributors

* Joel Cooklin - [@jcooklin](http://github.com/jcooklin) - Intel Corporation
* Justin Guidroz - [@geauxvirtual](http://github.com/geauxvirtual) - Intel Corporation
* Szymon Konefal - [@skonefal](http://github.com/skonefal) - Intel Corporation
* Pawel Palucki - [@ppalucki](http://github.com/ppalucki) - Intel Corporation
* Daniel Pittman - [@danielscottt](http://github.com/danielscottt) - Intel Corporation
* Nicholas Weaver - [@lynxbat](http://github.com/lynxbat) - Intel Corporation
