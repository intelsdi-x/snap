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
5. [Security Disclosure](#security-disclosure)
6. [License](#license)
7. [Contributors](#contributors)
  * [Initial Authors](#initial-authors)
  * [Maintainers](#maintainers)
8. [Thank You](#thank-you)

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

### System Requirements

Snap does not have external dependencies since it is compiled into a statically linked binary. At this time, we build snap binaries for Linux and MacOS. We also provide Linux RPM/Deb packages and MacOS X .pkg installer.

### Installation

You can obtain Linux RPM/Deb packages from [Snap's packagecloud.io respository](https://packagecloud.io/intelsdi-x/snap). Snap binaries in `.tar.gz` bundles and MacOS `.pkg` installer are available at Snap's [GitHub release page](https://github.com/intelsdi-x/snap/releases).

RedHat 6/7:
```
$ curl -s https://packagecloud.io/install/repositories/intelsdi-x/snap/script.rpm.sh | sudo bash
$ sudo yum install -y snap-telemetry
```

Ubuntu 14.04/16.04 (see known issue with Ubuntu 16.04.1 below)
```
$ curl -s https://packagecloud.io/install/repositories/intelsdi-x/snap/script.deb.sh | sudo bash
$ sudo apt-get install -y snap-telemetry
```

MacOS X (homebrew support pending 1.0.0 release, use the appropriate url from github release page):
```
$ curl -sfLO https://github.com/intelsdi-x/snap/releases/download/0.17.0/snap-telemetry-0.17.0.pkg
$ sudo installer -pkg ./snap-telemetry-0.17.0.pkg -target /
```

Tarball (choose the appropriate version and platform):
```
$ curl -sfLO https://github.com/intelsdi-x/snap/releases/download/0.17.0/snap-0.17.0-linux-amd64.tar.gz
$ tar xf snap-0.17.0-linux-amd64.tar.gz
$ cp snapd /usr/local/bin
$ cp snapctl /usr/local/bin
```

Ubuntu 16.04.1 [snapd package version 2.13+](https://launchpad.net/ubuntu/+source/snapd) installs snapd/snapctl binary in /usr/bin. These executables are not related to snap-telemetry. Running `snapctl` from snapd package will result in the following error message:

```
$ snapctl
error: snapctl requires SNAP_CONTEXT environment variable
```

Please make sure you invoke the snap-telemetry snapd/snapctl binary using fully qualified path (i.e. /usr/local/bin/{snapd|snapctl} if you installed the snap-telemetry package).

NOTE: If you prefer to build from source, follow the steps in the [build documentation](docs/BUILD_AND_TEST.md). The _alpha_ binaries containing the latest master branch are available here for bleeding edge testing purposes:
* snapd: [linux](http://snap.ci.snap-telemetry.io/snap/latest_build/linux/x86_64/snapd) | [darwin](http://snap.ci.snap-telemetry.io/snap/latest_build/darwin/x86_64/snapd)
* snapctl: [linux](http://snap.ci.snap-telemetry.io/snap/latest_build/linux/x86_64/snapctl) | [darwin](http://snap.ci.snap-telemetry.io/snap/latest_build/darwin/x86_64/snapctl)

### Running Snap

If you installed snap from RPM/Deb package, you can start/stop Snap daemon as a service:

RedHat 6/Ubuntu 14.04:
```
$ service snap-telemetry start
```

RedHat 7/Ubuntu 16.04:
```
$ systemctl start snap-telemetry
```

If you installed snap from binary, you can start Snap daemon via the command:
```
$ sudo mkdir -p /var/log/snap
$ sudo snapd --plugin-trust 0 --log-level 1 -o /var/log/snap &
```

To view the service logs:
```
$ tail -f /var/log/snap/snapd.log
```

By default Snap daemon will be running in standalone mode and listening on port 8181. To enable gossip mode, checkout the [tribe documentation](docs/TRIBE.md). For additional configuration options such as plugin signing and port configuration see [snapd documentation](docs/SNAPD.md).


### Load Plugins

Snap gets its power from the use of plugins. The [plugin catalog](#plugin-catalog) contains a collection of all known Snap plugins with links to their repo and release pages. (NOTE: Plugin bundles are deprecated in favor of independent plugin releases.)

First, let's download the file and psutil plugins (also make sure [psutil is installed](https://github.com/giampaolo/psutil/blob/master/INSTALL.rst)):

```
$ export OS=$(uname -s | tr '[:upper:]' '[:lower:]')
$ export ARCH=$(uname -m)
$ curl -sfL "https://github.com/intelsdi-x/snap-plugin-publisher-file/releases/download/2/snap-plugin-publisher-file_${OS}_${ARCH}" -o snap-plugin-publisher-file
$ curl -sfL "https://github.com/intelsdi-x/snap-plugin-collector-psutil/releases/download/8/snap-plugin-collector-psutil_${OS}_${ARCH}" -o snap-plugin-collector-psutil
```

Next load the plugins into Snap daemon using `snapctl`:
```
$ snapctl plugin load snap-plugin-publisher-file
Plugin loaded
Name: file
Version: 2
Type: publisher
Signed: false
Loaded Time: Fri, 14 Oct 2016 10:53:59 PDT

$ snapctl plugin load snap-plugin-collector-psutil
Plugin loaded
Name: psutil
Version: 8
Type: collector
Signed: false
Loaded Time: Fri, 14 Oct 2016 10:54:07 PDT
```

Verify plugins are loaded:
```
$ snapctl plugin list
NAME      VERSION    TYPE         SIGNED     STATUS    LOADED TIME
file      2          publisher    false      loaded    Fri, 14 Oct 2016 10:55:20 PDT
psutil    8          collector    false      loaded    Fri, 14 Oct 2016 10:55:29 PDT
```

See which metrics are available:
```
$ snapctl metric list
NAMESPACE                                VERSIONS
/intel/psutil/cpu/cpu-total/guest        8
/intel/psutil/cpu/cpu-total/guest_nice   8
/intel/psutil/cpu/cpu-total/idle         8
/intel/psutil/cpu/cpu-total/iowait       8
/intel/psutil/cpu/cpu-total/irq          8
/intel/psutil/cpu/cpu-total/nice         8
/intel/psutil/cpu/cpu-total/softirq      8
/intel/psutil/cpu/cpu-total/steal        8
/intel/psutil/cpu/cpu-total/stolen       8
/intel/psutil/cpu/cpu-total/system       8
/intel/psutil/cpu/cpu-total/user         8
/intel/psutil/load/load1                 8
/intel/psutil/load/load15                8
/intel/psutil/load/load5                 8
...
```

NOTE: Plugin bundles are available for convenience in the Snap [GitHub release page](https://github.com/intelsdi-x/snap/releases), for the latest up to date version use the release/download in the [plugin catalog](#plugin-catalog).

### Running Tasks

To collect data, you need to create a [task](docs/TASKS.md) by loading a `Task Manifest`. The manifest contains a specification for what interval a set of metrics are gathered, how the data is transformed, and where the information is published. For more information see [task](docs/TASKS.md) documentation.

Now, download and load the [psutil example](examples/psutil-file.yaml):
```
$ curl https://raw.githubusercontent.com/intelsdi-x/snap/master/examples/tasks/psutil-file.yaml -o /tmp/psutil-file.yaml
$ snapctl task create -t /tmp/psutil-file.yaml
Using task manifest to create task
Task created
ID: 8b9babad-b3bc-4a16-9e06-1f35664a7679
Name: Task-8b9babad-b3bc-4a16-9e06-1f35664a7679
State: Running
```

NOTE: in subsequent commands use the task ID from your CLI output, not the example task ID shown below.

This starts a task collecting metrics via psutil, then publishes the data to a file. To see the data published to the file (CTRL+C to exit):
```
$ tail -f /tmp/psutil_metrics.log
```

Or directly tap into the data stream that Snap is collecting using `snapctl task watch <task_id>`:
```
$ snapctl task watch 8b9babad-b3bc-4a16-9e06-1f35664a7679
NAMESPACE                             DATA             TIMESTAMP
/intel/psutil/cpu/cpu-total/idle      451176.5         2016-10-14 11:01:44.666137773 -0700 PDT
/intel/psutil/cpu/cpu-total/system    33749.2734375    2016-10-14 11:01:44.666139698 -0700 PDT
/intel/psutil/cpu/cpu-total/user      65653.2578125    2016-10-14 11:01:44.666145594 -0700 PDT
/intel/psutil/load/load1              1.81             2016-10-14 11:01:44.666072208 -0700 PDT
/intel/psutil/load/load15             2.62             2016-10-14 11:01:44.666074302 -0700 PDT
/intel/psutil/load/load5              2.38             2016-10-14 11:01:44.666074098 -0700 PDT
```

Nice work - you're all done with this example. Depending on how you started `snap-telemetry` service earlier, use the appropriate command to stop the daemon:

* init.d service: `service snap-telemetry stop`
* systemd service: `systemctl stop snap-telemetry`
* ran `snapd` manually: `sudo pkill snapd`

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

## Security Disclosure

The Snap team take security very seriously. If you have any issue regarding security, please notify us by sending an email to snap-security@intel.com
and not by creating a GitHub issue. We will follow up with you promptly with more information and a plan for remediation.

## License
Snap is Open Source software released under the [Apache 2.0 License](LICENSE).

## Contributors
### Initial Authors
All contributors are equally important to us, but we would like to thank the [initial authors](AUTHORS.md#initial-authors) for helping make open sourcing snap possible.

### Maintainers
Amongst the many awesome contributors, there are the maintainers. These maintainers may change over time, but they are all members of the **@intelsdi-x/snap-maintainers** team. This group will help you by:
* Committing to reviewing pull requests, issues, and addressing comments/questions as quickly as possible
* Acting as a point of contact for questions

Just tag **@intelsdi-x/snap-maintainers** if you need to get some attention on an issue. If at any time, you don't get a quick enough response, reach out to us [on Slack](http://slack.snap-telemetry.io)

## Thank You
And **thank you!** Your contribution, through code and participation, is incredibly important to us.
