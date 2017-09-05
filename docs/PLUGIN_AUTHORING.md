# Plugin Authoring

### Table of Content

* [Overview](#overview)
   * [Plugin Library](#plugin-library)
* [Developing Plugins](#developing-plugins)
   * [Plugin Type](#plugin-type)
   * [Plugin Name](#plugin-name)
   * [Plugin Metric Namespace](#plugin-metric-namespace)
   * [Plugin Interface](#plugin-interface)
   * [Plugin Version](#plugin-version)
   * [Plugin Release](#plugin-release)
   * [Plugin Metadata](#plugin-metadata)
   * [Plugin Catalog](#plugin-catalog)
   * [Plugin Status](#plugin-status)
   * [Plugin Tests](#plugin-tests)
   * [Documentation](#documentation)

## Overview

Snap daemon runs as a service which provides the core functionality such as plugin management and task scheduling. Snap plugins extend Snap by providing additional capability to gather, transform, or publish metrics. Since plugins are standalone programs, they can be written in any language supported by [gRPC](http://grpc.io).

### Plugin Library

The following libraries are available to simplify the process of writing a plugin:

* [snap-plugin-lib-go](https://github.com/intelsdi-x/snap-plugin-lib-go)
* [snap-plugin-lib-cpp](https://github.com/intelsdi-x/snap-plugin-lib-cpp)
* [snap-plugin-lib-py](https://github.com/intelsdi-x/snap-plugin-lib-py)

A few notes before we get started:

Communication between Snap daemon and plugins use gRPC. So even if a plugin library isn't available in the language of your choice, you can still write a plugin using the [gRPC library](http://grpc.io/docs) as a starting point. However this requires additional knowledge about Snap API, [gRPC/protobuf](../control/plugin/rpc/plugin.proto), so it is beyond the scope of this document.

Before writing a new Snap plugin, please check out the [Plugin Catalog](./PLUGIN_CATALOG.md) to see if any existing plugins meet your needs. If you need any assistance, please reach out on [Slack #snap-developers channel](https://intelsdi-x.herokuapp.com/).

## Developing Plugins

### Plugin Type

Snap supports four types of plugins:

* collector: gathering metrics based on the specified interval
* processor: transforming metrics
* publisher: publishing metrics
* streaming collector: gathering metrics when they are available 

### Plugin Name

The plugin repo name should follow this convention: `snap-plugin-[type]-[name]`

For example:
* `snap-plugin-collector-hana`
* `snap-plugin-processor-movingaverage`
* `snap-plugin-publisher-influxdb`

### Plugin Metric Namespace

When gathering data in collector plugins, each metric requires a namespace and a description of what data is being gathered. The metric namespace should contain the org, name of plugin, and the metric's name:

`/[organization]/[plugin name]/[plugin internal namespace(s)]/[metric name]`

For example `ACME` org writing a `water` plugin, gather `usage` in gallons from multiple locations.

```
/acme/water/usage
/acme/water/NewYork/usage
/acme/water/NewYork/rainfall
/acme/water/NewYork/Bronx/usage
/acme/water/NewYork/Chelsea/usage
/acme/water/Portland/usage
/acme/water/Portland/rainfall
/acme/water/Portland/Hollywood/usage
/acme/water/Portland/Pearl/usage
```

A plugin can have any number of internal plugin namespace. It's up to the plugin author on how to organize the metrics in a meaningful way. This information should be documented in the README for ease of usage.

NOTE: The intelsdi-x project reserves the org namespace `/intel`.

### Plugin Interface

Depending on the type of plugin, they must implement several methods to satisfy the appropriate interfaces. Please see the [plugin library](#plugin-library) for language specific examples and documentation.

### Plugin Version

Currently plugin versions are integer numbers and registered when a plugin is loaded. Whenever the source code is modified, please update the plugin version.

The following plugin version is taken from [Go lib example](https://github.com/intelsdi-x/snap-plugin-lib-go/blob/master/examples/collector/main.go).

```go
const (
	pluginName    = "rand-collector"
	pluginVersion = 1
)

func main() {
	plugin.StartCollector(rand.RandCollector{}, pluginName, pluginVersion)
}
```

Whenever the version changes, we also recommend:

* git tag the repository with the new plugin version
* publish binaries in github release page

For intelsdi-x repos, the binaries publishing is automated. If we update the rand plugin to version 2, simply tag the commit with the new version and push the new tags to github:

```
$ git tag -a 2 -m 'snap plugin collector rand v2'
$ git push origin --tags
```

NOTE: We are planning to adapt [Semantic Versioning](http://semver.org/). This requires changes to the internal framework, and we will provide a transition path when this is ready.

### Plugin Release

We recommend releasing new binaries to Github Release page whenever the plugin version is updated. This process can be automated via [Travis CI](https://docs.travis-ci.com/user/deployment/releases/). Please check out the file plugin's [.travis.yml](https://github.com/intelsdi-x/snap-plugin-publisher-file/blob/master/.travis.yml) file for a working example.

### Plugin Metadata

In the plugin repo root directory, the `metadata.yml` file provides Snap project additional information about your plugin when we generate the [plugin catalog](./PLUGIN_CATALOG.md) page.

* **name**: plugin full name
* **type**: plugin type
* **maintainer**: your github org or username
* **license**: the plugin software license
* **description**: paragraph describing the plugin's purpose
* **badge**: a list of [badges](https://shields.io/) to display
* **ci**: a list of ci services running for this repo
* **status**: one of the four statuses [described below](#plugin-status)

All metadata fields are optional, but recommended to help users discover your plugin. Please check out the file plugin's [metadata.yml](https://github.com/intelsdi-x/snap-plugin-publisher-file/blob/master/metadata.yml) file for a working example.

We recommend sharing your plugins early and often by adding them to the list of known plugins. To list your plugin in the plugin catalog, please submit a PR and update [plugins.yml](./plugins.yml) file to include the plugin's github `organization/repo_name`.

### Plugin Catalog

We provide a list of Snap plugins at [snap-telemetry.io](http://snap-telemetry.io/plugins.html) and in [this repo](PLUGIN_CATALOG.md). To keep these catalogs in sync, we do the following:

* Update [plugins.yml](plugins.yml) file to include a line containing `- org/repo_name`
* Periodically run [pluginsync tool](https://github.com/intelsdi-x/snap-pluginsync#update-plugin-metadata) to update the catalog
* [Plugin Metadata](#plugin-metadata) will be scanned and used accordingly when the pluginsync tool generates the catalog
* We do _not_ make changes directly to [PLUGIN_CATALOG.md](plugin_catalog.md) or the [parsed_plugin_list.js](https://github.com/intelsdi-x/snap/blob/gh-pages/assets/catalog/parsed_plugin_list.js) file since they are generated from [templates](https://github.com/intelsdi-x/snap-pluginsync/blob/master/PLUGIN_CATALOG.md.erb)

### Plugin Status

While the Snap framework is hardened through tons of testing, **plugins mature at their own pace**. We also want our community to share plugins early and update them often. To help both of these goals, we have tiers of maturity defined for plugins being added to the Plugin Catalog:
* [**Supported**](PLUGIN_STATUS.md#supported-plugins) - Created by a company with the intent of supporting customers
* [**Approved**](PLUGIN_STATUS.md#approved-plugins) - Vetted by Snap maintainers to meet our best practices for design
* [**Experimental**](PLUGIN_STATUS.md#experimental) - Early plugins ready for testing but not known to work as intended
* [**Unlabeled**](PLUGIN_STATUS.md#all-other-plugins-unlabeled) - Shared for reference or extension

 Further details to these definitions are available in [Plugin Status](PLUGIN_STATUS.md).

### Plugin Tests

For a plugin to be labeled `Approved` or `Supported`, it must have reasonable test coverage. At a minimum we require small tests, but large tests are also encouraged. To learn more about our testing best practices visit [BUILD_AND_TEST.md](BUILD_AND_TEST.md) and [LARGE_TESTS.md](LARGE_TESTS.md).

### Documentation

We request that all plugins include a README with the following information:

1. What It Does (and who it's for)
1. Supported Platforms
1. Known Issues
1. Snap Version dependencies
1. Installation
1. Usage
1. Contributors
1. License

You are welcome to copy an existing README.md (and CONTRIBUTING.md) to get started. I recommend looking at [psutil](https://github.com/intelsdi-x/snap-plugin-collector-psutil).
