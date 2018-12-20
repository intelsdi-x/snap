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

# DISCONTINUATION OF PROJECT 

**This project will no longer be maintained by Intel.  Intel will not provide or guarantee development of or support for this project, including but not limited to, maintenance, bug fixes, new releases or updates.  Patches to this project are no longer accepted by Intel. If you have an ongoing need to use this project, are interested in independently developing it, or would like to maintain patches for the community, please create your own fork of the project.**

# **The Snap Telemetry Framework** [![Build Status](https://travis-ci.org/intelsdi-x/snap.svg?branch=master)](https://travis-ci.org/intelsdi-x/snap) [![Go Report Card](https://goreportcard.com/badge/intelsdi-x/snap)](https://goreportcard.com/report/intelsdi-x/snap)  [![Join the chat on Slack](https://intelsdi-x.herokuapp.com/badge.svg)](https://intelsdi-x.herokuapp.com/)

<p align="center">
<img src="https://cloud.githubusercontent.com/assets/1744971/20331694/e07e9148-ab5b-11e6-856a-e4e956540077.png" width="70%">
</p>

**Snap** is an open telemetry framework designed to simplify the collection, processing and publishing of system data through a single API. The goals of this project are to:

* Empower systems to expose a consistent set of telemetry data
* Simplify telemetry ingestion across ubiquitous storage systems
* Allow flexible processing of telemetry data on agent (e.g. filtering and decoration)
* Provide powerful clustered control of telemetry workflows across small or large clusters

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
3. [Documentation](#documentation)
   * [Examples](#examples)
4. [Community Support](#community-support)
5. [Contributing](#contributing)
   * [Author a Plugin](#author-a-plugin)
   * [Become a Maintainer](#become-a-maintainer)
6. [Code of Conduct](#code-of-conduct)
7. [Security Disclosure](#security-disclosure)
8. [License](#license)
10. [Thank You](#thank-you)

## Overview

**The Snap Telemetry Framework** is a project made up of multiple parts:

* A hardened, extensively tested daemon, `snapteld`, and CLI, `snaptel` (in this repo)
* A growing number of maturing `plugins` (found in the [Plugin Catalog](#plugin-catalog))
* Lots of example `tasks` to gather and publish metrics (found in the [Examples folder](examples/))

These and other terminology are explained in the [glossary](docs/GLOSSARY.md).

![workflow-collect-process-publish](https://cloud.githubusercontent.com/assets/1744971/14644683/be49a6b6-0607-11e6-8621-14f7b54e2192.png)

The key features of Snap are:

* **Plugin Architecture**: Snap has a simple and smart modular design. The four types of plugins (collectors, processors, publishers and streaming collectors) allow Snap to mix and match functionality based on user need. All plugins are designed with versioning, signing and deployment at scale in mind. The **open plugin model** allows for loading built-in, community, or proprietary plugins into Snap.
  * **Collectors** - Collectors gather telemetry data at determined interval. Collectors are plugins for leveraging existing telemetry solutions (Facter, CollectD, Ohai) as well as specific plugins for consuming Intel telemetry (Node, DCM, NIC, Disk) and can reach into new architectures through additional plugins (see [Plugin Authoring below](#author-a-plugin)). Telemetry data is organized into a dynamically generated catalog of available data points.
  * **Processors** - Extensible workflow injection. Convert telemetry into another data model for consumption by existing systems. Allows encryption of all or part of the telemetry payload before publishing. Inject remote queries into workflow for tokens, filtering, or other external calls. Implement filtering at an agent level reducing injection load on telemetry consumer.
  * **Publishers** - Store telemetry into a wide array of systems. Snap decouples the collection of telemetry from the implementation of where to send it. Snap comes with a large library of publisher plugins that allow exposure to telemetry analytics systems both custom and common. This flexibility allows Snap to be valuable to open source and commercial ecosystems alike by writing a publisher for their architectures.
  * **Streaming Collectors** - Streaming collectors act just like collectors, but there is no determined interval of gathering metrics. They send metrics immediately when they are available over a GRPC to Snap daemon. There is also available mechanism of buffering incoming metrics configurable by params MaxMetricsBuffer and MaxCollectDuration. Check out [STREAMING.md](/docs/STREAMING.md) for more details. 

* **Dynamic Updates**: Snap is designed to evolve. Each scheduled workflow automatically uses the most mature plugin for that step, unless the collection is pinned to a specific version (e.g. get `/intel/psutil/load/load1/v1`). Loading a new plugin automatically upgrades running workflows in tasks. Load plugins dynamically, without a restart to the service or server. This dynamically extends the metric catalog when loaded, giving access to new measurements immediately. Swapping a newer version plugin for an old one in a safe transaction. All of these behaviors allow for simple and secure bug fixes, security patching, and improving accuracy in production.

* **Snap tribe**: Snap is designed for ease of administration. With Snap tribe, nodes work in groups (aka tribes). Requests are made through agreement- or task-based node groups, designed as a scalable gossip-based node-to-node communication process. Administrators can control all Snap nodes in a tribe agreement by messaging just one of them. There is auto-discovery of new nodes and import of tasks and plugins from nodes within a given tribe. It is cluster configuration management made simple.

**Snap** is not intended to:

* Operate as an analytics platform: the intention is to allow plugins for feeding those platforms
* Compete with existing metric/monitoring/telemetry agents: Snap is simply a new option to use or reference

## Getting Started

### System Requirements

Snap needs [Swagger for Go](https://github.com/go-swagger/go-swagger)  installed to update OpenAPI specification file after successful build. Swagger will be installed automatically during build process (`make` or `make deps`).

#### To install Swagger manually

Using `go get` (recommended):
```sh
go get -u github.com/go-swagger/go-swagger/cmd/swagger
```

From Debian package:
```sh
echo "deb https://dl.bintray.com/go-swagger/goswagger-debian ubuntu main" | sudo tee -a /etc/apt/sources.list
sudo apt-get update
sudo apt-get install swagger
```

From GitHub release:
```sh
curl -LO https://github.com/go-swagger/go-swagger/releases/download/0.10.0/swagger_linux_amd64
chmod +x swagger_linux_amd64 && sudo mv swagger_linux_amd64 /usr/bin/swagger
```

Snap does not have external dependencies since it is compiled into a statically linked binary. At this time, we build Snap binaries for Linux and MacOS. We also provide Linux RPM/Deb packages and MacOS X .pkg installer.

### Installation

You can obtain Linux RPM/Deb packages from [Snap's packagecloud.io repository](https://packagecloud.io/intelsdi-x/snap). After installation, please check and ensure `/usr/local/bin:/usr/local/sbin` is in your path via `echo $PATH` before executing any Snap commands.

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

We only build and test packages for a limited set of Linux distributions. For distros that are compatible with RedHat/Ubuntu packages, you can use the environment variable `os=` and `dist=` to override the OS detection script. For example Linux Mint 17/17.* (use `dist=xenial` for Linux Mint 18/18.*):
```
$ curl -s https://packagecloud.io/install/repositories/intelsdi-x/snap/script.deb.sh | sudo os=ubuntu dist=trusty bash
$ sudo apt-get install -y snap-telemetry
```

MacOS X:

If you use homebrew, the latest version of Snap package: [![snap-telemetry](https://img.shields.io/homebrew/v/snap-telemetry.svg)](http://brew.sh/)
```
$ brew install snap-telemetry
```

If you do not use homebrew, download and install Mac pkg package:
```
$ curl -sfL mac.pkg.dl.snap-telemetry.io -o snap-telemetry.pkg
$ sudo installer -pkg ./snap-telemetry.pkg -target /
```

Tarball (choose the appropriate version and platform):
```
$ curl -sfL linux.tar.dl.snap-telemetry.io -o snap-telemetry.tar.gz
$ tar xf snap-telemetry.tar.gz
$ cp snapteld /usr/local/sbin
$ cp snaptel /usr/local/bin
```

The intelsdi-x package repo contains additional information regarding:

* [Snap Redhat/Ubuntu packages](https://packagecloud.io/intelsdi-x/snap)
* [Snap `.tar.gz` bundles and MacOS `.pkg` installer](https://github.com/intelsdi-x/snap/releases).
* [installation script](https://packagecloud.io/intelsdi-x/snap/install#bash)
* [manual installation steps](https://packagecloud.io/intelsdi-x/snap/install#manual)
* [repo mirroring](https://packagecloud.io/intelsdi-x/snap/mirror)

NOTE: snap-telemetry packages prior to 0.19.0 installed `/usr/local/bin/{snapctl|snapd}` and these binaries have been renamed to `snaptel` and `snapteld`. snap-telemetry packages prior to 0.18.0 symlinked `/usr/bin/{snapctl|snapd}` to `/opt/snap/bin/{snapctl|snapd}` and may cause conflicts with [Ubuntu's `snapd` package](http://packages.ubuntu.com/xenial-updates/snapd). Ubuntu 16.04.1 [snapd package version 2.13+](https://launchpad.net/ubuntu/+source/snapd) installs snapd/snapctl binary in /usr/bin. These executables are not related to snap-telemetry. Running `snapctl` from snapd package will result in the following error message:

```
$ snapctl
error: snapctl requires SNAP_CONTEXT environment variable
```

NOTE: If you prefer to build from source, follow the steps in the [build documentation](docs/BUILD_AND_TEST.md). The _alpha_ binaries containing the latest master branch are available here for bleeding edge testing purposes:
* snapteld: [linux](http://snap.ci.snap-telemetry.io/snap/latest_build/linux/x86_64/snapteld) | [darwin](http://snap.ci.snap-telemetry.io/snap/latest_build/darwin/x86_64/snapteld)
* snaptel: [linux](http://snap.ci.snap-telemetry.io/snap/latest_build/linux/x86_64/snaptel) | [darwin](http://snap.ci.snap-telemetry.io/snap/latest_build/darwin/x86_64/snaptel)

### Running Snap

If you installed Snap from RPM/Deb package, you can start/stop Snap daemon as a service:

RedHat 6/Ubuntu 14.04:
```
$ service snap-telemetry start
```

RedHat 7/Ubuntu 16.04:
```
$ systemctl start snap-telemetry
```

If you installed Snap from binary, you can start Snap daemon via the command:
```
$ sudo mkdir -p /var/log/snap
$ sudo snapteld --plugin-trust 0 --log-level 1 --log-path /var/log/snap &
```

To view the service logs:
```
$ tail -f /var/log/snap/snapteld.log
```

By default, Snap daemon will be running in standalone mode and listening on port 8181. To enable gossip mode, checkout the [tribe documentation](docs/TRIBE.md). For additional configuration options such as plugin signing and port configuration see [snapteld documentation](docs/SNAPTELD.md).

### Load Plugins

Snap gets its power from the use of plugins. The [plugin catalog](#plugin-catalog) contains a collection of all known Snap plugins with links to their repo and release pages.

First, let's download the file and psutil plugins (also make sure [psutil is installed](https://github.com/giampaolo/psutil/blob/master/INSTALL.rst)):

```
$ export OS=$(uname -s | tr '[:upper:]' '[:lower:]')
$ export ARCH=$(uname -m)
$ curl -sfL "https://github.com/intelsdi-x/snap-plugin-publisher-file/releases/download/2/snap-plugin-publisher-file_${OS}_${ARCH}" -o snap-plugin-publisher-file
$ curl -sfL "https://github.com/intelsdi-x/snap-plugin-collector-psutil/releases/download/8/snap-plugin-collector-psutil_${OS}_${ARCH}" -o snap-plugin-collector-psutil
```

Next load the plugins into Snap daemon using `snaptel`:
```
$ snaptel plugin load snap-plugin-publisher-file
Plugin loaded
Name: file
Version: 2
Type: publisher
Signed: false
Loaded Time: Fri, 14 Oct 2016 10:53:59 PDT

$ snaptel plugin load snap-plugin-collector-psutil
Plugin loaded
Name: psutil
Version: 8
Type: collector
Signed: false
Loaded Time: Fri, 14 Oct 2016 10:54:07 PDT
```

Verify plugins are loaded:
```
$ snaptel plugin list
NAME      VERSION    TYPE         SIGNED     STATUS    LOADED TIME
file      2          publisher    false      loaded    Fri, 14 Oct 2016 10:55:20 PDT
psutil    8          collector    false      loaded    Fri, 14 Oct 2016 10:55:29 PDT
```

See which metrics are available:
```
$ snaptel metric list
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

### Running Tasks

To collect data, you need to create a task by loading a `Task Manifest`. The Task Manifest contains a specification for what interval a set of metrics are gathered, how the data is transformed, and where the information is published. For more information see [task](docs/TASKS.md) documentation.

Now, download and load the [psutil example](examples/tasks/psutil-file.yaml):
```
$ curl https://raw.githubusercontent.com/intelsdi-x/snap/master/examples/tasks/psutil-file.yaml -o /tmp/psutil-file.yaml
$ snaptel task create -t /tmp/psutil-file.yaml
Using task manifest to create task
Task created
ID: 8b9babad-b3bc-4a16-9e06-1f35664a7679
Name: Task-8b9babad-b3bc-4a16-9e06-1f35664a7679
State: Running
```

NOTE: In subsequent commands use the task ID from your CLI output in place of the `<task_id>`.

This starts a task collecting metrics via psutil, then publishes the data to a file. To see the data published to the file (CTRL+C to exit):
```
$ tail -f /tmp/psutil_metrics.log
```

Or directly tap into the data stream that Snap is collecting using `snaptel task watch <task_id>`:
```
$ snaptel task watch 8b9babad-b3bc-4a16-9e06-1f35664a7679
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
* ran `snapteld` manually: `sudo pkill snapteld`

When you're ready to move on, walk through other uses of Snap available in the [Examples folder](examples/).

### Building Tasks
Documentation for building a task can be found [here](docs/TASKS.md).

### Plugin Catalog
All known plugins are tracked in the [plugin catalog](https://github.com/intelsdi-x/snap/blob/master/docs/PLUGIN_CATALOG.md) and are tagged as collectors, processors, publishers and streaming collectors.

If you would like to write your own, read through [Author a Plugin](#author-a-plugin) to get started. Let us know if you begin to write one by [joining our Slack channel](https://intelsdi-x.herokuapp.com/). When you finish, please open a Pull Request to add yours to the catalog!

## Documentation
Documentation for Snap will be kept in this repository for now with an emphasis of filling out the `docs/` directory. We would also like to link to external how-to blog posts as people write them. [Read about contributing to the project](#contributing) for more details.

* [snapteld (Snap agent)](docs/SNAPTELD.md)
* [configuring snapteld](docs/SNAPTELD_CONFIGURATION.md)
* [snaptel (Snap CLI)](docs/SNAPTEL.md)
* [build and test](docs/BUILD_AND_TEST.md)
* [REST API V1](docs/REST_API_V1.md)
* [tasks](docs/TASKS.md)
* [plugin life cycle](docs/PLUGIN_LIFECYCLE.md)
* [plugin signing](docs/PLUGIN_SIGNING.md)
* [tribe](docs/TRIBE.md)
* [secure plugin communication](docs/SECURE_PLUGIN_COMMUNICATION.md)

To learn more about Snap and how others are using it, check out our [blog](https://medium.com/intel-sdi). A good first post to read is [My How-to for the Snap Telemetry Framework](https://medium.com/intel-sdi/my-how-to-for-the-snap-telemetry-framework-e3bb641bc740#.6f5nk543t) by @mjbrender.

### Examples
More complex examples of using Snap Framework configuration, Task Manifest files and use cases are available under the [Examples folder](examples/README.md). There are also interesting examples of using Snap in every plugin repository. For the full list of plugins, review the [Plugin Catalog](docs/PLUGIN_CATALOG.md).

## Community Support
This repository is one of many in the Snap framework and [has maintainers supporting it](docs/MAINTAINERS.md). We love contributions from our community along the way. No improvement is too small.

This note is especially important for plugins. While the Snap framework is hardened through tons of use, **plugins mature at their own pace**. If you have subject matter expertise related to a plugin, please share your feedback on that repository.

## Contributing
We encourage contributions from the community. Snap needs:

* _Contributors_: We always appreciate more eyes on the core framework and plugins
* _Feedback_: try it and tell us about it on [our Slack team](https://intelsdi-x.herokuapp.com/), through [a blog posts](https://medium.com/intel-sdi/) or Twitter with #SnapTelemetry
* _Integrations_: Snap can collect from and publish to almost anything by [authoring a plugin](#author-a-plugin)

To contribute to the Snap framework, see our [CONTRIBUTING.md](CONTRIBUTING.md) file. To give back to a specific plugin, open an issue on its repository. Snap maintainers aim to address comments and questions as quickly as possible. To get some attention on an issue, reach out to us [on Slack](http://slack.snap-telemetry.io), or open an issue to get a conversation started.


### Author a Plugin
The power of Snap comes from its open architecture and its growing community of contributors. You can be one of them:

* The definitive how-to is in [PLUGIN_AUTHORING.md](docs/PLUGIN_AUTHORING.md)

Add to the ecosystem by building your own plugins to collect, process or publish telemetry.

### Become a Maintainer
Snap maintainers are here to help guide Snap, the plugins, and the community forward in a positive direction. Maintainers of Snap and the Intel created plugins are selected based on contributions to the project and recommendations from other maintainers. The full list of active maintainers can be found [here](docs/MAINTAINERS.md).

Interested in becoming a maintainer? Check out [Responsibilities of a Maintainer](docs/MAINTAINERS.md#responsibilities-of-maintainers) and open an issue [here](https://github.com/intelsdi-x/snap/issues/new?title=interested+in+becoming+a+maintainer&body=About+me) to discuss your interest.

## Code of Conduct
All contributors to Snap are expected to be helpful and encouraging to all members of the community, treating everyone with a high level of professionalism and respect. See our [code of conduct](CODE_OF_CONDUCT.md) for more details.

## Security Disclosure

The Snap team takes security very seriously. If you have any issue regarding security, please notify us by sending an email to sys_snaptel-security@intel.com
and not by creating a GitHub issue. We will follow up with you promptly with more information and a plan for remediation.

## License
Snap is Open Source software released under the [Apache 2.0 License](LICENSE).

## Thank You
And **thank you!** Your contribution, through code and participation, is incredibly important to us.
