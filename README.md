# **Pulse** <sup><sub>_A powerful telemetry agent framework from Intel Corporation &reg;_</sub></sup>
[![Build Status](https://magnum.travis-ci.com/intelsdilabs/pulse.svg?token=2ujsxEpZo1issFyVWX29&branch=master)](https://magnum.travis-ci.com/intelsdilabs/pulse) [![Coverage Status](https://coveralls.io/repos/intelsdilabs/pulse/badge.svg?branch=HEAD)](https://coveralls.io/r/intelsdilabs/pulse?branch=HEAD)

## Description

**Pulse** is a framework for enabling the gathering of telemetry from systems. It aims to make collection, processing, and distribution of data _(telemetry)_ from running systems easier to schedule and deploy. We want to reduce the friction in gathering important system data into something that provides value with that data.

**Pulse** provides several specific features:

* Three types of plugins: collectors, processors, and publishers.
* Open plugin model allows for loading built-in, community, or proprietary plugins into **Pulse**.
* Dynamic loading, unloading, and swapping of plugins.
* Adding a newer version of a plugin automatically upgrades the next scheduled actions (when not pinned to a version).
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
* Integrations: **Pulse** can feasibly publish to almost any destination. We need publishing plugins for Ceilometer, vCOPs, Riemann, InfluxDB, and more.

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

## Testing

```
make test
```

## Releases

## Contributors

* Joel Cooklin - [@jcooklin](http://github.com/jcooklin) - Intel Corporation
* Justin Guidroz - [@geauxvirtual](http://github.com/jcooklin) - Intel Corporation
* Szymon Konefal - [@skonefal](http://github.com/jcooklin) - Intel Corporation
* Pawel Palucki - [@ppalucki](http://github.com/jcooklin) - Intel Corporation
* Daniel Pittman - [@danielscottt](http://github.com/jcooklin) - Intel Corporation
* Nicholas Weaver - [@lynxbat](http://github.com/jcooklin) - Intel Corporation
