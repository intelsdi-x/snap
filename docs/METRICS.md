
# Snap Static and Dynamic Metrics

Snap Framework supports two types of metrics. They are static and dynamic metrics. A Snap metric consists of a namespace and value pair.

### Static Metrics
A namespace having no wildcard and only string literals separated by slashes is a static metric.

String representation examples:
```
/intel/cassandra/node/zeus/type/Cache/scope/KeyCache/name/Requests/OneMinuteRate
/intel/cassandra/node/zeus/type/ClientRequest/scope/CASRead/name/Unavailables/FiveMinuteRate
/intel/cassandra/node/apollo/type/Cache/scope/RowCache/name/Hits/OneMinuteRate
/intel/cassandra/node/apollo/type/ClientRequest/scope/RangeSlice/name/Latency/OneMinuteRate
```
### Dynamic Metrics
A namespace including at least one wildcard is a dynamic metric.

String representation examples:
```
/intel/cassandra/node/*/type/*/scope/*/name/*/OneMinuteRate
/intel/cassandra/node/*/type/*/scope/*/name/*/FiveMinuteRate
/intel/cassandra/node/*/type/*/keyspace/*/name/*/OneMinuteRate
/intel/cassandra/node/*/type/*/keyspace/*/name/*/FiveMinuteRate
```
### Namespace Element
Namespace element in Snap is defined as a struct. Both static and dynamic elements share the same definition.
 ```
 type NamespaceElement struct {
	Value       string
	Description string
	Name        string
}
 ```
The _`NamespaceElement`_ forms each cell of a namespace and those cells are separated by slashes. 

Create a dynamic element:
```
ns := core.NewNamespace("intel", "cassandra", "node")
    .AddDynamicElement("node name", "description")
```

Create multiple dynamic elements:
```
ns := core.NewNamespace("intel", "cassandra", "node")
    .AddDynamicElement("node name", "description")
    .AddStaticElement("type")
    .AddDynamicElement("type value", "description")
    .AddStaticElement("scope")
    .AddDynamicElement("scope value", "description")
    .AddStaticElement("name")
    .AddDynamicElement("name value", "description")
    .AddStaticElement("50thPercentile")
```

### Why Dynamic Element
By defining dynamic elements inside a namespace, you'll have the capability to treat them differently at a later time. For instance, you have metric namespaces:
```
/foo/bar/host-alice/status
/foo/bar/host-bob/status
/foo/bar/host-nicole/status
/foo/bar/host-david/status
```
To get the measurement `/foo/bar/status`, specifying the _`hostname`_ element as a dynamic element and create filtering friendly _`hostname`_ tags.

>The key takeaway is that the _`named element`_ is a _`dynamic element`_. A static element has an _`empty Name`_ field.

### Collector Plugins
Building a Snap collector plugin involves two primary tasks. One is to create a collector metric catalog. Another is to collect metric data.

##### Create Metric Catalog
Creating a collector having dynamic metric catalog by utilizing the following methods from Snap _`core`_ package, Snap CLI(snapctl) verbose output could display the definition and the description of a wildcard along with a namespace's `Measurement Unit` if it's defined.

Methods:
```
(n Namespace) AddStaticElement(value string) Namespace
(n Namespace) AddDynamicElement(name, description string) Namespace
(n Namespace) AddStaticElements(values ...string) Namespace
```

Create Metric Catalog:
```
metricType := plugin.MetricType{
    Namespace_: ns,
    Unit_:   <namespace measurement unit>,
}
```

Snap CLI verbose output:
```
$ $SNAP_PATH/bin/snapctl metric list --verbose

/intel/cassandra/node/[node name]/type/[type value]/scope/[scope value]/name/[name value]/OneMinuteRate  float64
/intel/cassandra/node/[node name]/type/[type value]/scope/[scope value]/name/[name value]/FiveMinuteRate  float64
/intel/cassandra/node/[node name]/type/[type value]/keyspace/[keyspace value]/name/[name value]/OneMinuteRate  float64
/intel/cassandra/node/[node name]/type/[type value]/keyspace/[keyspace value]/name/[name value]/FiveMinuteRate  float64
```
##### Collecting metric data
While collecting metric data, collector plugin authors may leverage Snap framework by specifying `Name` field for every _`dynamic`_ element but leaving an empty `Name` field for each static element.

>If you use the _`(n Namespace) AddDynamicElement(value string) Namespace`_ method to build dynamic elements for a metric, remember setting the actual metric element values.

### Publisher Plugins
Snap publisher plugins may leverage following methods from Snap to determine if an element is dynamic or static _`if and only if`_ a collector correctly defined the _`Name`_ field for every dynamic element.

Get all dynamic element positions:
```
isDynamic, indexes := ns.IsDynamic() 
```
Where `isDynamic` is a bool type which indicates if a namespace contains at least one dynamic element and `indexes` is an array of dynamic element positions.

Loop through each element:
```
for _, elt := range ns {
    elt.IsDynamic() {
        // strip off dynamic element to create searchable metric tags
    }
}
```
It's even better that plugin [utilities](https://github.com/intelsdi-x/snap-plugin-utilities/blob/master/mts/mts.go#L32) may save your time.

> Snap checks the Name field of a namespace element to determine if an item is static or dynamic.  The _`Name`_ field should be empty as a static element but not empty as a dynamic element.

Appropriately define your dynamic element to leverage Snap framework!
