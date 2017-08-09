
# Snap Metrics

A metric in Snap has the following fields.

* Namespace `[]core.NamespaceElement`
 * Uniquely identifies the metric
* LastAdvertisedTime `time.Time`
 * Describes when the metric was added to the metric catalog
* Version `int`
 * Is bound to the version of the plugin
 * Multiple versions of the same metric can be added to the catalog
  * Unless specified in the Task Manifest, the latest available metric will be collected
* Config `*cdata.ConfigDataNode`
 * Contains data needed to collect a metric
  * Examples include 'uri', 'username', 'password', 'paths'
* Data `interface{}`
 * The collected data
* Tags `map[string]string`
 * Are key value pairs that provide additional metadata about the metric
 * May be added by the framework or other plugins (processors)
  * The framework currently adds the following standard tag to all metrics
   * `plugin_running_on` describing on which host the plugin is running. This value is updated every hour due to a TTL set internally.
 * May be added by a task manifests as described [here](https://github.com/intelsdi-x/snap/pull/941)
 * May be added by the snapteld config as described [here](https://github.com/intelsdi-x/snap/issues/827)
* Unit `string`
 * Describes the magnitude being measured
 * Can be an empty string for unitless data
 * See [Metrics20.org](http://metrics20.org/spec/) for more guidance on units
* Description `string`
 * Is stored in the metric catalog and meant to give the user more details about the metric such as how it is derived
* Timestamp `time.Time`
 * Describes when the metric was collected  

## Static Metrics

A metric is described as static when the string representation of the namespace has no wildcards.

String representation examples:
```
/intel/cassandra/node/zeus/type/Cache/scope/KeyCache/name/Requests/OneMinuteRate
/intel/cassandra/node/zeus/type/ClientRequest/scope/CASRead/name/Unavailables/FiveMinuteRate
/intel/cassandra/node/apollo/type/Cache/scope/RowCache/name/Hits/OneMinuteRate
/intel/cassandra/node/apollo/type/ClientRequest/scope/RangeSlice/name/Latency/OneMinuteRate
```
## Dynamic Metrics

A metric is described as dynamic when it includes one or more wildcards.

String representation examples:
```
/intel/cassandra/node/*/type/*/scope/*/name/*/OneMinuteRate
/intel/cassandra/node/*/type/*/scope/*/name/*/FiveMinuteRate
/intel/cassandra/node/*/type/*/keyspace/*/name/*/OneMinuteRate
/intel/cassandra/node/*/type/*/keyspace/*/name/*/FiveMinuteRate
```

## Metric Namespace

As described above a metrics `Namespace` is an array of NamespaceElements (`[]core.NamespaceElement`).

A `NamespaceElement` has the following fields.

* Value `string`
* Description `string`
 * *Only* used to describe dynamic components of a namespace
* Name `string`
 * *Only* used to describe dynamic components of a namespace

### Static Metric Namespace Example

Given a static metric identified by the namespace `/intel/psutil/load/load1` the `NamespaceElement`s would
have values of 'intel', 'psutil', 'load' and "load1" respectively.  The `Name` and `Description` fields would have
empty values.

The metric's namespace could be created using the following constructor function.

```
namespace := core.NewNamespace("intel", "psutil", "load", "load1")
```  

### Dynamic Metric Namespace Example

Dynamic namespaces enable collector plugins to embed runtime data in the namespace with just enough metadata to enable
downstream plugins (processors and publishers) the ability to extract the data and transform the namespace into its
 canonical form often required by some back ends.     

Given a dynamic metric identified by the namespace `/intel/libvirt/*/disk/*/wrreq` the `NamespaceElement`s would
have values of 'intel', 'libvirt', '\*', 'disk', '\*' and 'wrreq' respectively.  The `Name` and `Description` fields
of the 2nd and 4th elements would also contain non empty values.  

The metric's namespace could be created using the following constructor function.

```
ns := core.NewNamespace("intel", "libvirt")
    .AddDynamicElement("domain_name", "Domain Name")
    .AddStaticElement("disk")
    .AddDynamicElement("disk_name", "Disk Name")
    .AddStaticElement("wrreq")
```

It is *important* to note that the `NamespaceElement` fields `Name` and `Description` should *only* have non-empty string
values when the element they are describing is dynamic in which case the `Value` field would contain the string value "*".

You can find an example of the influxdb publisher creating tags out of the dynamic elements of the namespace and publishing
to a time series [here](https://github.com/intelsdi-x/snap-plugin-publisher-influxdb/blob/b253302ddfc94e3b444780328d0f503a6d73e3e0/influx/influx.go#L164-L176).
Using the example above we can expect a datapoint published to a time series with the name `/intel/libvirt/disk/wrreq`
with tags describing `domain_name` and `disk_name`.  
