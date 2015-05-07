# **Pulse** <sup><sub>_A powerful telemetry agent framework from Intel Corporation &reg;_</sub></sup>
[![Build Status](https://magnum.travis-ci.com/intelsdilabs/pulse.svg?token=2ujsxEpZo1issFyVWX29&branch=master)](https://magnum.travis-ci.com/intelsdilabs/pulse) [![Coverage Status](https://coveralls.io/repos/intelsdilabs/pulse/badge.svg?branch=master&t=rVMsC3)](https://coveralls.io/r/intelsdilabs/pulse?branch=master)

## Description

**Pulse** is a framework for enabling the gathering of telemetry from systems. The goal of this project is:

* Improve the deployment model and flexibility of a telemetry tools.
* Allow dynamic control of collection for one system or many at once.
* Make it easy to inject processing on telemetry at the agent for use cases like machine learning.
* Reduce the friction of pointing Pulse to any telemetry ingesting system.
* Provide operational improvements that make collecting on a very large cluster of systems easier.

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

* _Feedback_: try it and tell us about it.
* _Contributors_: We need plugins, schedules, testing, and more.
* _Integrations_: **Pulse** can feasibly publish to almost any destination. We need publishing plugins for [Ceilometer](https://wiki.openstack.org/wiki/Ceilometer), [vCOPs](http://www.vmware.com/products/vrealize-operations), [Riemann](https://github.com/aphyr/riemann), [InfluxDB](https://github.com/influxdb/influxdb), and more.

**Pulse** architecture

* Written in Go
* Decoupled internal structure with a focus on event-driven handling

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
