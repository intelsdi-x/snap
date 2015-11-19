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

**Pulse** is a framework for enabling the gathering of telemetry from systems. The goals of this project are to:

* Empower all servers to expose a consistent set of telemetry data
* Simplify dissemination of data to telemetry ingesting systems
* Improve the deployment model, packaging and flexibility of telemetry toolkits
* Provide dynamic control of collection for small or large clusters of systems
* Allow flexible processing of telemetry data on agent (e.g. machine learning)

The key features of Pulse are:

* **Plugin Architecture**: The three types of plugins (collectors, processors, and publishers) allow Pulse to mix and match functionality based on user need. All plugins are designed with versioning, signing and deployment at scale in mind. The open plugin model allows for loading built-in, community, or proprietary plugins into Pulse.
  * **Collectors** - Consume telemetry data through plugins. Collectors are built-in plugins for leveraging existing telemetry solutions (Facter, CollectD, Ohai) as well as specific plugins for consuming Intel telemetry (Node, DCM, NIC, Disk) and can reach into new architectures through additional plugins. Telemetry data is organized into a dynamically generated catalog of available data points.
  * **Processors** - Extensible workflow injection. Convert telemetry into another data model for consumption by existing telemetry consumption systems (like OpenStack Ceilometer). Allows encryption of all or part of the telemetry payload before publishing. Inject remote queries into workflow for tokens, filtering, or other external calls. Implement filtering at an agent level reducing injection load on telemetry consumer.
  * **Publishers** - Storage telemetry into a wide array of systems. Decouple the collection of telemetry from the implementation of where to send it. A large library of publisher plugins allow exposure to telemetry analytics systems both custom and common. Makes Pulse valuable in the way it enables and cooperates with existing systems. Make Pulse valuable to open source and commercial ecosystems by allow them to build a publisher into their architectures.


* **Dynamic Updates**: Pulse is designed to evolve. Loading a new plugin automatically upgrades running workflows in tasks, unless the collection is pinned to a version (ex: get /intel/server/cpu/load/v1). Each scheduled workflow automatically uses the most mature plugin for that step. This behavior, coupled with dynamic plugin loading, results in instantaneous updates to existing workflows. Helpful for bug fixes, security patching, improving accuracy. Load plugins dynamically, without a restart to the service or server. Dynamically extends the metric catalog when loaded. Swaps a newer version plugin for an old one in a safe transaction.

* **Tribe**: With Pulse Tribes, you have a scalable gossip-based node-to-node communication that allows administrators to control all Pulse nodes by speaking to just one of them. It is cluster configuration management made simple. Tribes are organized by through agreement- or task-based node groups. There is auto-discovery of new nodes and import of tasks & plugins from nodes within a given group (aka tribe).

Some additionally important notes about how Pulse works:

* Adding a newer version of a plugin automatically upgrades the next scheduled actions _(when not pinned to a version)_
* You have the ability to run Pulse in a distributed role deployment model
* Extensible source allows for the addition of customization or improvement of scheduling, plugins, routing, and more
* Multiple management modules including: CLI, REST API, and Web Console (each of which can be turned on or off)
* Secure validation via plugin signing, SSL encryption for APIs and payload encryption for communication between components
* CLI control from Linux or OS X

**Pulse** is not intended to:

* Operate as an analytics platform: the intention is to allow plugins for feeding those platforms
* Compete with existing metric/monitoring/telemetry agents: Pulse is simply a new option to use or reference

Some additional architectural principles:

* Decoupled internal structure with a focus on event-driven handling
* Make consuming telemetry declarative

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

**Pulse** needs:

* _Feedback_: try it and tell us about it.
* _Contributors_: We need plugins, schedules, testing, and more.
* _Integrations_: **Pulse** can feasibly publish to almost any destination. We need publishing plugins for [Ceilometer](https://wiki.openstack.org/wiki/Ceilometer), [vCOPs](http://www.vmware.com/products/vrealize-operations), and more.

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
