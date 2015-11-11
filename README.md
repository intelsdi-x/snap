<!--
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# **Pulse** <sup><sub>_A powerful telemetry agent framework from Intel Corporation &reg;_</sub></sup>
[![Build Status](https://magnum.travis-ci.com/intelsdi-x/pulse.svg?token=2ujsxEpZo1issFyVWX29&branch=master)](https://magnum.travis-ci.com/intelsdi-x/pulse)

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
* _Integrations_: **Pulse** can feasibly publish to almost any destination. We need publishing plugins for [Ceilometer](https://wiki.openstack.org/wiki/Ceilometer), [vCOPs](http://www.vmware.com/products/vrealize-operations), and more.

**Pulse** architecture

* Decoupled internal structure with a focus on event-driven handling

## License

<DO NOT PUT SOMETHING HERE YET>

## System Requirements

## Building

```
make
```

## Testing

```
make test
```

## Installation

Local installation

```
make install
```

## Running

## Documentation

### REST API

### Pulsed
* [Pulsed](docs/PULSED.md)

### CLI
* [Pulsectl](cmd/pulsectl/README.md)

## Contributing

* [Please read our development guidelines](https://github.com/intelsdilabs/pulse/wiki/Development-guidelines)
* [ ] TODO - CLA

## Releases

## Contributors

* Joel Cooklin - [@jcooklin](http://github.com/jcooklin) - Intel Corporation
* Justin Guidroz - [@geauxvirtual](http://github.com/geauxvirtual) - Intel Corporation
* Tiffany Jernigan - [@tiffanyfj](https://github.com/tiffanyfj) - Intel Corporation
* Szymon Konefal - [@skonefal](http://github.com/skonefal) - Intel Corporation
* Pawel Palucki - [@ppalucki](http://github.com/ppalucki) - Intel Corporation
* Daniel Pittman - [@danielscottt](http://github.com/danielscottt) - Intel Corporation
* Nicholas Weaver - [@lynxbat](http://github.com/lynxbat) - Intel Corporation
