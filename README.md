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

# **snap** <sup><sub>_A powerful telemetry framework_</sub></sup>

[![Join the chat at https://gitter.im/intelsdi-x/snap](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/intelsdi-x/snap?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Build Status](https://travis-ci.org/intelsdi-x/snap.svg?branch=master)](https://travis-ci.org/intelsdi-x/snap)

<img src="https://cloud.githubusercontent.com/assets/12282848/11570134/809832de-99aa-11e5-852a-cf599c1d893b.jpg" width="50%">

Here are blog posts related to ***snap*** by the team:

* Nickapedia.com - [What if collecting data center telemetry was a snap?](http://nickapedia.com/2015/12/02/what-if-collecting-data-center-telemetry-was-a-snap/)

----

1. [Overview](#overview)
2. [Getting Started](#getting-started)
  * [In Active Development](#in-active-development)
  * [System Requirements](#system-requirements)
  * [Installation](#installation)
  * [Running snap](#running-snap)
  * [Load Plugins](#load-plugins)
  * [Running Tasks](#running-tasks)
  * [Building Tasks](#building-tasks)
  * [Plugin Catalog](#plugin-catalog)
2. [Documentation](#documentation)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
  * [Author a Plugin](#author-a-plugin)
5. [License](#license)
6. [Contributors](#contributors)
  * [Initial Authors](#initial-authors)
  * [Maintainers](#maintainers)
7. [Thank You](#thank-you)

## Overview

**snap** is a framework for enabling the gathering of telemetry from systems. The goals of this project are to:

* Empower systems to expose a consistent set of telemetry data
* Simplify telemetry ingestion across ubiquitous storage system
* Improve the deployment model, packaging and flexibility for collecting telemetry
* Allow flexible processing of telemetry data on agent (e.g. machine learning)
* Provide powerful clustered control of telemetry workflows across small or large clusters

The key features of snap are:

* **Plugin Architecture**: snap has a simple and smart modular design. The three types of plugins (collectors, processors, and publishers) allow snap to mix and match functionality based on user need. All plugins are designed with versioning, signing and deployment at scale in mind. The open plugin model allows for loading built-in, community, or proprietary plugins into snap.
  * **Collectors** - Collectors consume telemetry data. Collectors are built-in plugins for leveraging existing telemetry solutions (Facter, CollectD, Ohai) as well as specific plugins for consuming Intel telemetry (Node, DCM, NIC, Disk) and can reach into new architectures through additional plugins (see [Plugin Authoring below](#author-a-plugin)). Telemetry data is organized into a dynamically generated catalog of available data points.
  * **Processors** - Extensible workflow injection. Convert telemetry into another data model for consumption by existing consumption systems (like OpenStack Ceilometer). Allows encryption of all or part of the telemetry payload before publishing. Inject remote queries into workflow for tokens, filtering, or other external calls. Implement filtering at an agent level reducing injection load on telemetry consumer.
  * **Publishers** - Store telemetry into a wide array of systems. snap decouples the collection of telemetry from the implementation of where to send it. snap comes with a large library of publisher plugins that allow exposure to telemetry analytics systems both custom and common. This flexibility allows snap to be valuable to open source and commercial ecosystems alike by writing a publisher for their architectures.


* **Dynamic Updates**: snap is designed to evolve. Each scheduled workflow automatically uses the most mature plugin for that step, unless the collection is pinned to a specific version (ex: get /intel/server/cpu/load/v1). Loading a new plugin automatically upgrades running workflows in tasks. Load plugins dynamically, without a restart to the service or server. This dynamically extends the metric catalog when loaded, giving access to new measurements immediately. Swapping a newer version plugin for an old one in a safe transaction. All of these behaviors allow for simple and secure bug fixes, security patching, and improving accuracy in production.

* **snap tribe**: snap is designed for ease of administration. With snap tribe, nodes work in groups (aka tribes). Requests are made through agreement- or task-based node groups, designed as a scalable gossip-based node-to-node communication process. Administrators can control all snap nodes in a tribe agreement by messaging just one of them. There is auto-discovery of new nodes and import of tasks & plugins from nodes within a given tribe. It is cluster configuration management made simple.

Some additionally important notes about how snap works:

* Multiple management modules including: CLI, REST API, and Web Console (each of which can be turned on or off)
* Secure validation occurs via plugin signing, SSL encryption for APIs and payload encryption for communication between components
* CLI control from Linux or OS X

**snap** is not intended to:

* Operate as an analytics platform: the intention is to allow plugins for feeding those platforms
* Compete with existing metric/monitoring/telemetry agents: snap is simply a new option to use or reference

## Getting Started

### In Active Development
The master branch is used for new feature development. Our goal is to keep it in a deployable state. If you're looking for the most recent binary that is versioned, please [see the latest release](https://github.com/intelsdi-x/snap/releases).

### System Requirements
Snap deploys as a binary, which makes requirements quite simple. We've tested on a subset of Linux and OS X versions.

### Installation

You can get the pre-built binaries for your OS and architecture at snap's [GitHub Releases](https://github.com/intelsdi-x/snap/releases) page. This isn't the comprehensive list of plugins, but they will help you get started. Right now, snap only supports Linux and OS X.

If you're looking for the bleeding edge of snap, you can build it by cloning down the `master` branch. To build snap from source, you will need [Golang >= 1.4](https://golang.org) and [GNU Make](https://www.gnu.org/software/make/). More on building snap [here](./CONTRIBUTING.md).

### Running snap

Start a standalone snap agent:

```
export SNAP_PATH=<snapdirectory>/build
$SNAP_PATH/bin/snapd --plugin-trust 0 --log-level 1
```
This will bring up a snap agent without requiring plugin signing and set the logging level to debug. snap's REST API will be listening on port 8181.

### Running snap in tribe (cluster) mode

The first node

```
$SNAP_PATH/bin/snapd --tribe
```

All other nodes who join will need to select any existing member of the cluster.

```
$SNAP_PATH/bin/snapd --tribe-seed <ip or name of another tribe member>
```

Checkout the [tribe doc](docs/TRIBE.md) for more info.

## Load Plugins
snap gets its power from the use of plugins. The [plugin catalog](#plugin-catalog) is a collection of all known plugins for snap.

Next, lets load a few of the demo plugins.  You can do this via cURL, or `snapctl`, snap's CLI.

Using cURL
```sh
curl -X POST -F plugin=@build/plugin/snap-collector-mock1 http://localhost:8181/v1/plugins
curl -X POST -F plugin=@build/plugin/snap-processor-passthru http://localhost:8181/v1/plugins
curl -X POST -F plugin=@build/plugin/snap-publisher-file http://localhost:8181/v1/plugins
```

Or:

Using `snapctl`
```sh
$SNAP_PATH/bin/snapctl plugin load build/plugin/snap-collector-mock1
$SNAP_PATH/bin/snapctl plugin load build/plugin/snap-processor-passthru
$SNAP_PATH/bin/snapctl plugin load build/plugin/snap-publisher-file
```

Let's look at what plugins we have loaded now:

```sh
$SNAP_PATH/bin/snapctl plugin list
NAME             VERSION         TYPE            SIGNED          STATUS          LOADED TIME
mock1            1               collector       false           loaded          Tue, 17 Nov 2015 14:08:17 PST
passthru         1               processor       false           loaded          Tue, 17 Nov 2015 14:16:12 PST
file             3               publisher       false           loaded          Tue, 17 Nov 2015 14:16:19 PST
```
### Running Tasks

Tasks can be in JSON or YAML format. Let's start one of the [example tasks](./examples/tasks/mock-file.yaml) from the `examples/` directory:

```
$SNAP_PATH/bin/snapctl task create -t examples/tasks/mock-file.yaml
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

Or actually tap into the data that snap is collecting:
```
$SNAP_PATH/bin/snapctl task watch 8b9babad-b3bc-4a16-9e06-1f35664a7679
```

### Building Tasks
Documentation for building a task can be found [here](docs/TASKS.md).

### Plugin Catalog
All known plugins are tracked in the [plugin catalog](https://github.com/intelsdi-x/snap/blob/master/docs/PLUGIN_CATALOG.md) and are tagged as collectors, processors and publishers.

If you would like to write your own, read through [Author a Plugin](#author-a-plugin). Let us know if you begin to write one by opening an Issue. When you finish, please open a Pull Request to add yours to the catalog!

## Documentation
Documentation for snap will be kept in this repository for now. We would also like to link to external how-to blog posts as people write them. See our [CONTRIBUTING.md](#contributing) for more details.

* [snapctl](docs/SNAPCTL.md)
* [snapd](docs/SNAPD.md)
* [REST API](docs/REST_API.md)
* [plugin signing](docs/PLUGIN_SIGNING.md)
* [tribe](docs/TRIBE.md)

### Examples
There are interesting examples of using snap in every plugin repository. For the full list of plugins, review the [plugin catalog](https://github.com/intelsdi-x/snap/blob/master/docs/PLUGIN_CATALOG.md):

* [configs](examples/configs/) is a basic example of the configuration file that powers your plugins
* [tasks](examples/tasks/) has JSON-encoded execution requests for snap plugins

### Roadmap
We have a few known features we want to take on from here while we remain open for feedback after public release. They are:

* Windows support (spec pending)
* Distributed Workflows (spec pending)
* Workflow Routing (see issue [#539](https://github.com/intelsdi-x/snap/issues/539))

If you would like to propose a feature, please open a Issue that includes RFC in it (for [request for comments](https://en.wikipedia.org/wiki/Request_for_Comments)).

## Community Support
This repository is one of **many** projects in the **snap framework**. Discuss your questions about snap by reaching out to us on:

* Through GitHub Issues. Issues is our home for **all** needs: Q&A on everything - installation, request for events, integrations, bug issues, futures. [Open up an Issue](https://github.com/intelsdi-x/snap/issues) and know there's no wrong question for us.
* We also have a Gitter channel opened up on this repository that threads directly into our engineering team Slack (thanks to [Sameroom](https://sameroom.io/)).

The full project lives here, at http://github.com/intelsdi-x/snap.

## Contributing
We encourage contribution from the community. **snap** needs:

* _Feedback_: try it and tell us about it through issues, blog posts or Twitter
* _Contributors_: We need plugins, schedules, testing, and more
* _Integrations_: **snap** can feasibly publish to almost any destination. We need publishing plugins for [Ceilometer](https://wiki.openstack.org/wiki/Ceilometer), [vROPs](http://www.vmware.com/products/vrealize-operations), and more. See [the plugin catalog](./docs/PLUGIN_CATALOG.md#wish-list) for the full list

To contribute to the snap framework, see [our CONTRIBUTING file](CONTRIBUTING.md). To give back to a specific plugin, open an issue on its repository.

### Author a Plugin
The power of snap comes from its open architecture. Add to the ecosystem by building your own plugins to collect, process or publish telemetry.

* The definitive how-to is [PLUGIN_AUTHORING.md](docs/PLUGIN_AUTHORING.md)
* Recommendations to make effective, well-designed plugins are in [PLUGIN_BEST_PRACTICES.md](docs/PLUGIN_BEST_PRACTICES.md)

## License
snap is Open Source software released under the Apache 2.0 [License](LICENSE).

## Contributors
### Initial Authors
All contributors are equally important to us, but we would like to thank the [initial authors](AUTHORS.md#initial-authors) for helping make open sourcing snap possible.

### Maintainers
Amongst the many awesome contributors, there are the maintainers. These maintainers may change over time, but they are all members of the **@intelsdi-x/snap-maintainers** team. This group will help you by:
* Committing to reviewing pull requests, issues, and addressing comments/questions as quickly as possible
* Acting as a point of contact for questions

Just tag **@intelsdi-x/snap-maintainers** if you need to get some attention on an issue. If at any time, you don't get a quick enough response, reach out to any of the following team members directly:

<table border="0" cellspacing="0" cellpadding="0">
  <tr>
    <td width="125"><a href="https://github.com/andrzej-k"><sub>@andrzej-k</sub><img src="https://avatars.githubusercontent.com/u/13486250" alt="@andrzej-k"></a></td>
    <td width="125"><a href="https://github.com/candysmurf"><sub>@candysmurf</sub><img src="https://avatars.githubusercontent.com/u/13841563" alt="@candysmurf"></a></td>
  	<td width="125"><a href="https://github.com/ConnorDoyle"><sub>@ConnorDoyle</sub><img src="https://avatars.githubusercontent.com/u/379372" alt="@ConnorDoyle"></a></td>
  	<td width="125"><a href="https://github.com/danielscottt"><sub>@danielscottt</sub><img src="https://avatars.githubusercontent.com/u/1194436" alt="@danielscottt"></a></td>
  	<td width="125"><a href="https://github.com/geauxvirtual"><sub>@geauxvirtual</sub><img src="https://avatars.githubusercontent.com/u/1395030" alt="@geauxvirtual"></a></td>
  </tr>
  <tr>
    <td width="125"><a href="http://github.com/jcooklin"><sub>@jcooklin</sub><img src="https://avatars.githubusercontent.com/u/862968" alt="@jcooklin"></a></td>
    <td width="125"><a href="https://github.com/lynxbat"><sub>@lynxbat</sub><img src="https://avatars.githubusercontent.com/u/1278669" width="100" alt="@lynxbat"></a></td>
    <td width="125"><a href="https://github.com/mjbrender"><sub>@mjbrender</sub><img src="https://avatars.githubusercontent.com/u/1744971" width="100" alt="@mjbrender"></a></td>
    <td width="125"><a href="https://github.com/nqn"><sub>@nqn</sub><img src="https://avatars.githubusercontent.com/u/897374" width="100" alt="@nqn"></a></td>
    <td width="125"><a href="https://github.com/tiffanyfj"><sub>@tiffanyfj</sub><img src="https://avatars.githubusercontent.com/u/12282848" width="100" alt="@tiffanyfj"></a></td>
  </tr>
</table>

We're also looking for new maintainers from the community. Please let us know if you would like to become one by opening an Issue titled "interested in becoming a maintainer."

## Thank You
And **thank you!** Your contribution, through code and participation, is incredibly important to us. 
