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

- [ ] TODO: Consider branching note ([like this one](https://github.com/Netflix/genie#in-active-development))

1. [Overview](#overview)
2. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Installation](#installation)
  * [Running Pulse](#running-pulse)
  * [Load Plugins](#load-plugins)
  * [Running Tasks](#running-tasks)
  * [Building Tasks](#building-tasks)
  * [Plugin Catalog](#plugin-catalog)
2. [Documentation](#documentation)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
  * [Author a Plugin]
5. [License](#license)
6. [Maintainers](#maintainers)
7. [Thank You](#thank-you)

## Overview

**Pulse** is a framework for enabling the gathering of telemetry from systems. The goals of this project are to:

* Empower systems to expose a consistent set of telemetry data
* Simplify telemetry ingestion across ubiquitous storage system
* Improve the deployment model, packaging and flexibility for collecting telemetry
* Allow flexible processing of telemetry data on agent (e.g. machine learning)
* Provide powerful clustered control of telemetry workflows across small or large clusters

The key features of Pulse are:

* **Plugin Architecture**: Pulse has a simple and smart modular design. The three types of plugins (collectors, processors, and publishers) allow Pulse to mix and match functionality based on user need. All plugins are designed with versioning, signing and deployment at scale in mind. The open plugin model allows for loading built-in, community, or proprietary plugins into Pulse.
  * **Collectors** - Collectors consume telemetry data. Collectors are built-in plugins for leveraging existing telemetry solutions (Facter, CollectD, Ohai) as well as specific plugins for consuming Intel telemetry (Node, DCM, NIC, Disk) and can reach into new architectures through additional plugins (see [Plugin Authoring below](#)). Telemetry data is organized into a dynamically generated catalog of available data points.
  * **Processors** - Extensible workflow injection. Convert telemetry into another data model for consumption by existing consumption systems (like OpenStack Ceilometer). Allows encryption of all or part of the telemetry payload before publishing. Inject remote queries into workflow for tokens, filtering, or other external calls. Implement filtering at an agent level reducing injection load on telemetry consumer.
  * **Publishers** - Store telemetry into a wide array of systems. Pulse decouples the collection of telemetry from the implementation of where to send it. Pulse comes with a large library of publisher plugins that allow exposure to telemetry analytics systems both custom and common. This flexibility allows Pulse to be valuable to open source and commercial ecosystems alike by writing a publisher for their architectures.


* **Dynamic Updates**: Pulse is designed to evolve. Each scheduled workflow automatically uses the most mature plugin for that step, unless the collection is pinned to a specific version (ex: get /intel/server/cpu/load/v1). Loading a new plugin automatically upgrades running workflows in tasks. Load plugins dynamically, without a restart to the service or server. This dynamically extends the metric catalog when loaded, giving access to new measurements immediately. Swapping a newer version plugin for an old one in a safe transaction. All of these behaviors allow for simple and secure bug fixes, security patching, and improving accuracy in production.

* **Pulse Tribe**: Pulse is designed for ease of administration. With Pulse Tribe, nodes work in groups (aka tribes). Requests are made through agreement- or task-based node groups, designed as a scalable gossip-based node-to-node communication process. Administrators can control all Pulse nodes in a tribe agreement by messaging just one of them. There is auto-discovery of new nodes and import of tasks & plugins from nodes within a given tribe. It is cluster configuration management made simple.

Some additionally important notes about how Pulse works:

* Multiple management modules including: CLI, REST API, and Web Console (each of which can be turned on or off)
* Secure validation occurs via plugin signing, SSL encryption for APIs and payload encryption for communication between components
* CLI control from Linux or OS X

**Pulse** is not intended to:

* Operate as an analytics platform: the intention is to allow plugins for feeding those platforms
* Compete with existing metric/monitoring/telemetry agents: Pulse is simply a new option to use or reference

## Getting Started

### System Requirements
Pulse deploys as a binary, which makes requirements quite simple. We've tested on the following versions of linux/os x

{}

### Installation

You can get the pre-built binaries for your OS and architecture at Pulse's [Github Releases](https://github.com/intelsdi-x/pulse/releases) page.  Right now, Pulse only supports Linux and OS X.

If you're looking for the bleeding edge of Pulse, you can build it by cloning down the `master` branch.  To build Pulse from source, you will need [Golang >= 1.4](https://golang.org) and [GNU Make](https://www.gnu.org/software/make/).  More on building Pulse [here](./CONTRIBUTING.md).

### Running Pulse

Start a standalone Pulse agent:

```sh
$ ./bin/pulsed --plugin-trust 0 --log-level 1
```

This will bring up a Pulse agent without requiring plugin signing and set the logging level to debug.  Pulse's REST API will be listening on port 8181.

## Load Plugins
Pulse gets its power from the use of plugins. The [Plugin Catalog](#plugin-catalog) is a collection of all known plugins for Pulse.

Next, lets load a few of the demo plugins.  You can do this via cURL, or `pulsectl`, Pulse's CLI:

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
### Running Tasks

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

### Building Tasks

### Plugin Catalog
All known Plugins are tracked in the [Plugin Catalog](#()) and are tagged as consumers, processors and publishers.

If you would like to write your own, read through [Authoring a Plugin](#). Let us know if you begin to write one by opening an Issue. When you finish, please open a Pull Request to add yours to the list!

## Documentation
Documentation for Pulse will be kept in this repository for now. We would also like to link to external how-to blog posts as people write them. See our [Contributing](#contributing) for more details.

* [Pulsectl](cmd/pulsectl/README.md)
* [Pulsed](docs/PULSED.md)

### Examples
There are interesting examples of using Pulse in every plugin repository:

< TODO -- move these to different repos.
* [configs](examples/configs/) is a basic example of
* [tasks](examples/tasks/) has JSON-encoded execution requests for Pulse plugins

### Roadmap


## Community Support
This repository is one of **many** projects in the **Pulse Framework**. Discuss your questions about Pulse by reaching out to us on:

* Pulse Gitter channel (TODO Link)
* Our Google Group (TODO Link)

The full project lives here, at http://github.com/intelsdi-x/pulse.

## Contributing
We encourage contribution from the community. **Pulse** needs:

* _Feedback_: try it and tell us about it through issues, blog posts or Twitter
* _Contributors_: We need plugins, schedules, testing, and more
* _Integrations_: **Pulse** can feasibly publish to almost any destination. We need publishing plugins for [Ceilometer](https://wiki.openstack.org/wiki/Ceilometer), [vCOPs](http://www.vmware.com/products/vrealize-operations), and more. See [the Plugin Catalog](./docs/PLUGIN_CATALOG.md#wish-list) for the full list

To contribute to the Pulse framework, see [our CONTRIBUTING file](CONTRIBUTING.md). To give back to a specific plugin, open an issue on its repository.

## License
Pulse is an Open Source software released under the Apache 2.0 [License](LICENSE).

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

## Thank You
And **thank you!** Your contribution, through code and participation, is incredibly important to us.
