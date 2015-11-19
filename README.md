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

# **Pulse** <sup><sub>_A powerful telemetry agent framework_</sub></sup>
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

## Getting Started

### Getting Pulse

You can get the pre-built binaries for your OS and architecture at Pulse's [Github Releases](https://github.com/intelsdi-x/pulse/releases) page.  Right now, Pulse only supports Linux and OSX.

If you're looking for the bleeding edge of Pulse, you can build it by cloning down the `master` branch.  To build Pulse from source, you will need [Golang >= 1.4](https://golang.org) and [GNU Make](https://www.gnu.org/software/make/).  More on building Pulse [here](./CONTRIBUTING.md).

### Running Pulse

Start a standalone Pulse agent:

```sh
$ ./bin/pulsed --plugin-trust 0 --log-level 1
```

This will bring up a Pulse agent without requiring plugin signing and set the logging level to debug.  Pulse's REST API will be listening on port 8181.

Next, lets load a few of the demo plugins.  You can do this via Curl, or `pulsectl`, Pulse's CLI:

```sh
curl -X POST -F plugin=@build/plugin/pulse-collector-mock1 http://localhost:8181/v1/plugins
```

And:

```sh
$ ./bin/pulsectl plugin load build/plugin/pulse-processor-passthru
$ ./bin/pulsectl plugin load build/plugin/pulse-publisher-file
```

Let's look at what plugins we have loaded now:

```sh
$ ./bin/pulsectl plugin list
NAME             VERSION         TYPE            SIGNED          STATUS          LOADED TIME
mock1            1               collector       false           loaded          Tue, 17 Nov 2015 14:08:17 PST
passthru         1               processor       false           loaded          Tue, 17 Nov 2015 14:16:12 PST
file             3               publisher       false           loaded          Tue, 17 Nov 2015 14:16:19 PST
```

Next, let's start one of the [example tasks](./examples/tasks/mock-file.json) from the `examples/` directory:

```
$ ./bin/pulsectl task create -t examples/tasks/mock-file.json
Using task manifest to create task
Task created
ID: 8b9babad-b3bc-4a16-9e06-1f35664a7679
Name: Task-8b9babad-b3bc-4a16-9e06-1f35664a7679
State: Running
```

From here, you should be able to do 2 things:

See the data that is being published to the file:
```
$ tail -f /tmp/published
```

Or actually tap into the data that Pulse is collecting:
```
$ ./bin/pulsectl task watch 8b9babad-b3bc-4a16-9e06-1f35664a7679
```

## License

Pulse is open sourced under the Apache 2.0 [License](https://github.com/intelsdi-x/pulse/blob/master/LICENSE).

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

## Maintainers

Amongst the many awesome contributors, there are the maintainers. These maintainers may change over time, but they are:
* Committed to reviewing pull requests, issues, and addressing comments/questions as quickly as possible
* A point of contact for questions

<table border="0" cellspacing="0" cellpadding="0">
  <tr>
  	<td width="125"><a href="https://github.com/candysmurf"><sub>@candysmurf</sub><img src="https://avatars.githubusercontent.com/u/13841563" alt="@candysmurf"></a></td>
	<td width="125"><a href="https://github.com/ConnorDoyle"><sub>@ConnorDoyle</sub><img src="https://avatars.githubusercontent.com/u/379372?" alt="@ConnorDoyle"></a></td>
	<td width="125"><a href="https://github.com/danielscottt"><sub>@danielscottt</sub><img src="https://avatars.githubusercontent.com/u/1194436" alt="@danielscottt"></a></td>
	<td width="125"><a href="https://github.com/geauxvirtual"><sub>@geauxvirtual</sub><img src="https://avatars.githubusercontent.com/u/1395030" alt="@geauxvirtual"></a></td>
  </tr>
  <tr>
	<td width="125"><a href="http://github.com/jcooklin"><sub>@jcooklin</sub><img src="https://avatars.githubusercontent.com/u/862968" alt="@jcooklin"></a></td>
  	<td width="125"><a href="https://github.com/lynxbat"><sub>@lynxbat</sub><img src="https://avatars.githubusercontent.com/u/1278669" width="100" alt="@lynxbat"></a></td>
	<td width="125"><a href="https://github.com/nqn"><sub>@nqn</sub><img src="https://avatars.githubusercontent.com/u/897374" width="100" alt="@nqn"></a></td>
	<td width="125"><a href="https://github.com/tiffanyfj"><sub>@tiffanyfj</sub><img src="https://avatars.githubusercontent.com/u/12282848" width="100" alt="@tiffanyfj"></a></td>
  </tr>
</table>
