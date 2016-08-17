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

# **snap** <sup><sub>_the open telemetry framework_</sub></sup>

[![Join the chat on Slack](https://intelsdi-x.herokuapp.com/badge.svg)](https://intelsdi-x.herokuapp.com/)
[![Build Status](https://travis-ci.org/intelsdi-x/snap.svg?branch=master)](https://travis-ci.org/intelsdi-x/snap)
[![Go Report Card](https://goreportcard.com/badge/intelsdi-x/snap)](https://goreportcard.com/report/intelsdi-x/snap)

<img src="https://cloud.githubusercontent.com/assets/1744971/16677455/f1d4e9de-448a-11e6-9afb-c31dcc7e3274.png" width="50%">

----

1. [Overview](#overview)
2. [Getting Started](#getting-started)
  * [In Active Development](#in-active-development)
  * [System Requirements](#system-requirements)
  * [Installation](#installation)
  * [Running Snap](#running-snap)
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

![workflow-collect-process-publish](https://cloud.githubusercontent.com/assets/1744971/14644683/be49a6b6-0607-11e6-8621-14f7b54e2192.png)

**Snap** is an open telemetry framework designed to simplify the collection, processing and publishing of system data through a single API. The goals of this project are to:

* Empower systems to expose a consistent set of telemetry data
* Simplify telemetry ingestion across ubiquitous storage systems
* Improve the deployment model, packaging and flexibility for collecting telemetry
* Allow flexible processing of telemetry data on agent (e.g. filtering and decoration)
* Provide powerful clustered control of telemetry workflows across small or large clusters

The key features of snap are:

* **Plugin Architecture**: snap has a simple and smart modular design. The three types of plugins (collectors, processors, and publishers) allow snap to mix and match functionality based on user need. All plugins are designed with versioning, signing and deployment at scale in mind. The **open plugin model** allows for loading built-in, community, or proprietary plugins into snap.
  * **Collectors** - Collectors consume telemetry data. Collectors are built-in plugins for leveraging existing telemetry solutions (Facter, CollectD, Ohai) as well as specific plugins for consuming Intel telemetry (Node, DCM, NIC, Disk) and can reach into new architectures through additional plugins (see [Plugin Authoring below](#author-a-plugin)). Telemetry data is organized into a dynamically generated catalog of available data points.
  * **Processors** - Extensible workflow injection. Convert telemetry into another data model for consumption by existing consumption systems (like OpenStack Ceilometer). Allows encryption of all or part of the telemetry payload before publishing. Inject remote queries into workflow for tokens, filtering, or other external calls. Implement filtering at an agent level reducing injection load on telemetry consumer.
  * **Publishers** - Store telemetry into a wide array of systems. snap decouples the collection of telemetry from the implementation of where to send it. snap comes with a large library of publisher plugins that allow exposure to telemetry analytics systems both custom and common. This flexibility allows snap to be valuable to open source and commercial ecosystems alike by writing a publisher for their architectures.


* **Dynamic Updates**: snap is designed to evolve. Each scheduled workflow automatically uses the most mature plugin for that step, unless the collection is pinned to a specific version (e.g. get /intel/psutil/load/load1/v1). Loading a new plugin automatically upgrades running workflows in tasks. Load plugins dynamically, without a restart to the service or server. This dynamically extends the metric catalog when loaded, giving access to new measurements immediately. Swapping a newer version plugin for an old one in a safe transaction. All of these behaviors allow for simple and secure bug fixes, security patching, and improving accuracy in production.

* **Snap tribe**: snap is designed for ease of administration. With snap tribe, nodes work in groups (aka tribes). Requests are made through agreement- or task-based node groups, designed as a scalable gossip-based node-to-node communication process. Administrators can control all snap nodes in a tribe agreement by messaging just one of them. There is auto-discovery of new nodes and import of tasks and plugins from nodes within a given tribe. It is cluster configuration management made simple.

Some additionally important notes about how snap works:

* Multiple management modules including: [CLI](docs/SNAPCTL.md) (snapctl) and [REST API](docs/REST_API.md) (each of which can be turned on or off)
* Secure validation occurs via plugin signing, SSL encryption for APIs and payload encryption for communication between components
* CLI control from Linux or OS X

**Snap** is not intended to:

* Operate as an analytics platform: the intention is to allow plugins for feeding those platforms
* Compete with existing metric/monitoring/telemetry agents: Snap is simply a new option to use or reference

## Getting Started

### In Active Development
The master branch is used for new feature development. Our goal is to keep it in a deployable state. If you're looking for the most recent binary that is versioned, please [see the latest release](https://github.com/intelsdi-x/snap/releases).

### System Requirements
Snap deploys as a binary, which makes requirements quite simple. We've tested on a subset of Linux and OS X versions.

### Installation

You can get the pre-built binaries for your OS and architecture at Snap's [GitHub Releases](https://github.com/intelsdi-x/snap/releases) page. This isn't the comprehensive list of plugins. Right now, snap only supports Linux and OS X (Darwin).

### Running Snap
We're going to assume you downloaded the latest packaged release of Snap and its plugins from here on. If you prefer to build from source, follow the steps in [BUILD_AND_TEST.md](docs/BUILD_AND_TEST.md).

Untar the version of Snap you downloaded and move its binaries, `snapd` and `snapctl`, to a place in your path. Here is an example of doing so (on Linux):
```
$ tar -xvf snap-v0.13.0-beta-linux-amd64.tar
$ mv snap-v0.13.0-beta/bin/* /usr/local/bin/
$ rm -rf snap-v0.13.0-beta
```

Start a standalone Snap agent (`snapd`):
```
$ snapd --plugin-trust 0 --log-level 1
```
This will bring up a Snap agent without requiring plugin signing (trust-level 0) and set the logging level to debug (log level 1). Snap's REST API will be listening on port 8181. To learn more about the snap agent and how to use it look at [SNAPD.md](docs/SNAPD.md) and/or run `snapd -h`.

Snap can also be run in a clustered mode called `tribe`. Checkout the [tribe documentation](docs/TRIBE.md) for more info.

### Load Plugins
Snap gets its power from the use of plugins. The [Plugin Catalog](#plugin-catalog) is a collection of all known plugins for Snap with links to the binaries. You can download individual plugins or pull down a package of starter plugins for your operating system under [GitHub Releases](https://github.com/intelsdi-x/snap/releases). This isn't the comprehensive list of plugins, but they will help you get started.

Open a separate window from the one running `snapd`, then unpack the downloaded plugins (example is on Linux):
```
$ tar -xvf snap-plugins-v0.13.0-beta-linux-amd64.tar.gz
$ mkdir -p ~/snap/plugins/
$ mv snap-v0.13.0-beta/plugin/* ~/snap/plugins/
$ rm -rf snap-v0.13.0-beta
```

Next, load the plugins. This can be achieved through the REST API directly or by using the helper command `snapctl`.

Using the API directly with cURL:
```
$ cd ~/snap/plugins/
$ curl -X POST -F plugin=@snap-plugin-collector-mock1 http://localhost:8181/v1/plugins
$ curl -X POST -F plugin=@snap-plugin-processor-passthru http://localhost:8181/v1/plugins
$ curl -X POST -F plugin=@snap-plugin-publisher-file http://localhost:8181/v1/plugins
```

Every interaction with `snapd` can be done through the REST API. To see what else you can do with the API, view our [API Documentation](docs/REST_API.md). We will continue on using `snapctl`:
```
$ cd ~/snap/plugins/
$ snapctl plugin load snap-plugin-collector-mock1
$ snapctl plugin load snap-plugin-processor-passthru
$ snapctl plugin load snap-plugin-publisher-file
```

Let's look at what plugins you have loaded now:
```
$ snapctl plugin list
NAME             VERSION         TYPE            SIGNED          STATUS          LOADED TIME
mock             1               collector       false           loaded          Tue, 17 Nov 2015 14:08:17 PST
passthru         1               processor       false           loaded          Tue, 17 Nov 2015 14:16:12 PST
file             3               publisher       false           loaded          Tue, 17 Nov 2015 14:16:19 PST
```

You now have one of each plugin type loaded into the framework. To begin collecting data, you need to create a task.

### Running Tasks
Tasks are most often shared as a Task Manifest and is written in JSON or YAML format. Make a copy of [this example task](./examples/tasks/mock-file.yaml) from the `examples/tasks/` directory on your local system and then start the task:

```
$ cd ~/snap
$ curl https://raw.githubusercontent.com/intelsdi-x/snap/master/examples/tasks/mock-file.yaml > mock-file.yaml
$ snapctl task create -t mock-file.yaml
Using task manifest to create task
Task created
ID: 8b9babad-b3bc-4a16-9e06-1f35664a7679
Name: Task-8b9babad-b3bc-4a16-9e06-1f35664a7679
State: Running
```

This task generates mock data, "processes" it through a passthrough mechanism and then publishes it to a file. Now with a running task, you should be able to do two things:

See the data that is being published to the file:
```
$ tail -f /tmp/snap_published_mock_file.log
^C
```

Or tap into the data that Snap is collecting using the Task ID to watch the task (note: your Task ID will be different):
```
$ snapctl task watch 8b9babad-b3bc-4a16-9e06-1f35664a7679
```

If the Task ID already scrolled by, you can list it with:
```
$ snapctl task list
```

Nice work - you're all done with this example. You ran `snapd` manually, so stopping the daemon process will stop any running tasks and unload any plugins we loaded.
```
$ pkill snapd
```

Or you can continue to run more tasks using the loaded plugins (why not create a new Task Manifest that publishes mock data to another file?). Alternatively, you can stop any running tasks and unload any plugins you no longer wish to use manually:
```
$ snapctl task stop 8b9babad-b3bc-4a16-9e06-1f35664a7679
$ snapctl plugin unload processor:passthru:1
Plugin unloaded
Name: passthru
Version: 1
Type: processor
$ snapctl plugin unload publisher:file:3
Plugin unloaded
Name: file
Version: 3
Type: publisher
$ snapctl plugin unload collector:mock:1
Plugin unloaded
Name: mock
Version: 1
Type: collector
```

When you're ready to move on, walk through other uses of Snap available in the [Examples folder](examples/).

### Building Tasks
Documentation for building a task can be found [here](docs/TASKS.md).

### Plugin Catalog
All known plugins are tracked in the [plugin catalog](https://github.com/intelsdi-x/snap/blob/master/docs/PLUGIN_CATALOG.md) and are tagged as collectors, processors and publishers.

If you would like to write your own, read through [Author a Plugin](#author-a-plugin) to get started. Let us know if you begin to write one by [joining our Slack channel](https://intelsdi-x.herokuapp.com/). When you finish, please open a Pull Request to add yours to the catalog!

## Documentation
Documentation for Snap will be kept in this repository for now with an emphasis of filling out the `docs/` directory. We would also like to link to external how-to blog posts as people write them. [Read about contributing to the project](#contributing) for more details.

* [snapd (snap agent)](docs/SNAPD.md)
* [configuring snapd](docs/SNAPD_CONFIGURATION.md)
* [snapctl (Snap CLI)](docs/SNAPCTL.md)
* [build and test](docs/BUILD_AND_TEST.md)
* [REST API](docs/REST_API.md)
* [tasks](docs/TASKS.md)
* [plugin life cycle](docs/PLUGIN_LIFECYCLE.md)
* [plugin signing](docs/PLUGIN_SIGNING.md)
* [tribe](docs/TRIBE.md)


Here are blog posts related to Snap by the team:

* [What if collecting data center telemetry was a snap?](http://nickapedia.com/2015/12/02/what-if-collecting-data-center-telemetry-was-a-snap/) by @lynxbat
* [My How-to for the Snap Telemetry Framework](https://medium.com/intel-sdi/my-how-to-for-the-snap-telemetry-framework-e3bb641bc740#.6f5nk543t) by @mjbrender
* [Snap's first GUI is Grafana!](https://medium.com/intel-sdi/snap-first-gui-is-grafana-40bb92df2660#.rgsnfx3w1) by @mjbrender
* [Adding a global configuration to Snap](https://medium.com/intel-sdi/adding-a-global-configuration-to-snap-b95d8fd8b5e0#.hq3cqgt6t) by @tjmcs1


### Examples
More complex examples of using Snap Framework configuration, Task Manifest files and use cases are available under the [Examples folder](examples/README.md). There are also interesting examples of using Snap in every plugin repository. For the full list of plugins, review the [Plugin Catalog](docs/PLUGIN_CATALOG.md).

### Roadmap
We have a few known features we want to take on from here while we remain open for feedback after public release. They are:

* Authentication, authorization, and auditing (see issue [#286](https://github.com/intelsdi-x/snap/issues/286))
* Workflow Routing (see issue [#539](https://github.com/intelsdi-x/snap/issues/539))
* Windows support (see [#671](https://github.com/intelsdi-x/snap/issues/671))
* Distributed Workflows (see [#539](https://github.com/intelsdi-x/snap/issues/539) and [#640](https://github.com/intelsdi-x/snap/issues/640))

If you would like to propose a feature, please [open an Issue](https://github.com/intelsdi-x/snap/issues)) that includes RFC in it (for [request for comments](https://en.wikipedia.org/wiki/Request_for_Comments)).

## Community Support
This repository is one of **many** projects in the **Snap framework**. Discuss your questions about snap by reaching out to us:

* Through GitHub Issues. Issues is our home for **all** needs: Q&A on everything - installation, request for events, integrations, bug issues, futures. [Open up an Issue](https://github.com/intelsdi-x/snap/issues) and know there's no wrong question for us.
* We also have a [Slack team](https://intelsdi-x.herokuapp.com/) where we hang out
* Submit a blog post on your use of Snap to [our Medium.com publication](https://medium.com/intel-sdi)

The full project lives here, at http://github.com/intelsdi-x/snap.

## Contributing
We encourage contributions from the community. No improvement is too small. Snap needs:

* _Feedback_: try it and tell us about it through issues, blog posts or Twitter
* _Contributors_: We can always use more eyes on the core framework and its testing
* _Integrations_: Snap can collect from and publish to almost anything by [authoring a plugin](#author-a-plugin)

To contribute to the Snap framework, see [our CONTRIBUTING file](CONTRIBUTING.md). To give back to a specific plugin, open an issue on its repository.

### Author a Plugin
The power of Snap comes from its open architecture. Add to the ecosystem by building your own plugins to collect, process or publish telemetry.

* The definitive how-to is in [PLUGIN_AUTHORING.md](docs/PLUGIN_AUTHORING.md)
* Recommendations to make effective, well-designed plugins are in [PLUGIN_BEST_PRACTICES.md](docs/PLUGIN_BEST_PRACTICES.md)

## License
Snap is Open Source software released under the Apache 2.0 [License](LICENSE).

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
  	<td width="125"><a href="http://github.com/jcooklin"><sub>@jcooklin</sub><img src="https://avatars.githubusercontent.com/u/862968" alt="@jcooklin"></a></td>
  </tr>
  <tr>
    <td width="125"><a href="https://github.com/lynxbat"><sub>@lynxbat</sub><img src="https://avatars.githubusercontent.com/u/1278669" width="100" alt="@lynxbat"></a></td>
    <td width="125"><a href="https://github.com/marcin-krolik"><sub>@marcin-krolik</sub><img src="https://avatars.githubusercontent.com/u/14905131" width="100" alt="@marcin-krolik"></a></td>
    <td width="125"><a href="https://github.com/mjbrender"><sub>@mjbrender</sub><img src="https://avatars.githubusercontent.com/u/1744971" width="100" alt="@mjbrender"></a></td>
    <td width="125"><a href="https://github.com/nqn"><sub>@nqn</sub><img src="https://avatars.githubusercontent.com/u/897374" width="100" alt="@nqn"></a></td>
    <td width="125"><a href="https://github.com/tiffanyfj"><sub>@tiffanyfj</sub><img src="https://avatars.githubusercontent.com/u/12282848" width="100" alt="@tiffanyfj"></a></td>
    <td width="125"><a href="https://github.com/IzabellaRaulin"><sub>@IzabellaRaulin</sub><img src="https://avatars0.githubusercontent.com/u/11335874" width="100" alt="@IzabellaRaulin"></a></td>
  </tr>
</table>

We're also looking to have maintainers from the community. Please let us know if you would like to become one by opening an Issue titled "interested in becoming a maintainer." We are currently working on a more official process.

## Thank You
And **thank you!** Your contribution, through code and participation, is incredibly important to us.
