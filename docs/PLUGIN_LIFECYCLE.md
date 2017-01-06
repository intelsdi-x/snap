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
# Plugin life cycle

A Snap plugin can be in a `Loaded` or `Running` state.  A plugin can be loaded
in the following two ways.

1. `snapteld` was started with an auto discover path `snapteld -a /opt/snap/plugins`
2. A user loads a plugin through the REST API or using the snaptel 
    * `curl -F file=@snap-plugin-publisher-file http://localhost:9191/v1/plugins`
    * `snaptel plugin load snap-plugin-publisher-file` 

A plugin transitions to a `running` state when a task is started that uses the 
plugin.  This is also called a plugin subscription.  

## What happens when a plugin is loaded

When a plugin is loaded snapteld takes the following steps.

1. Handshakes with the plugin by reading it's stdout
2. Updates the metric catalog by calling the plugin over RPC
    * `GetMetricTypes` returns the metrics the plugin provides
    * `GetConfigPolicy` returns the conf policy that the plugin needs
3. The plugin is stopped

It should be emphasized that when a plugin is loaded it is started but stopped 
as soon as the metric catalog has been updated.  

## What happens when a plugin is unloaded

When a plugin is unloaded snapteld removes it from the metric catalog and running
instances of the plugin are stopped.   

## What happens when a task is started

When a task is started the plugins that the task references are started and 
subscribed to.   The following steps are taken when a task is created and 
started.

1. On **task creation** the task is validated (`snaptel task create -t mytask.yml
--no-start`) 
    * The schedule is validated
    * The config provided for the metrics (collectors), processors and 
    publishers are validated
2. On **task starting** the plugins are started (`snaptel task start <TASK_ID>`)
3. Subscriptions for each plugin referenced by the task are incremented 

## Diving deeper

**Task started** - When a task is started the plugins which are referenced by 
the task manifest are subscribed to and a 'subscription group' is created.  A 
'subscription group' is a view that contains an ID, the requested metrics, 
plugins and configuration provided in the workflow.    

![start_task](https://www.dropbox.com/s/p3gj83zti6q7rgc/scheduler_scheduler_startTask_new.png?raw=1)
    
**Task stopped** - When a task is stopped the plugins referenced by the
subscription group are unsubscribed and the subscription group is removed.  

![stop_task](https://www.dropbox.com/s/yzl1b0c15z7tnen/scheduler_scheduler_stopTask.png?raw=1)

**Plugin loaded/unloaded** - When a plugin is loaded or unloaded the event triggers
processing of all subscription groups.  When a subscription group is processed the
requested metrics are evaluated and mapped to collector plugins.  The required
plugins are compared with the previous state of the subscription group 
triggering the appropriate subscribe or unsubscribe calls. Finally, the 
subscription group view is updated with the current plugin dependencies and
the metrics that will be collected based on the requested metrics (query).

![load_unload_plugins](https://www.dropbox.com/s/9xwqg94qo8z8mmq/control_pluginControl_handlePluginEvent.png?raw=1)  

**CollectMetrics** - When a task fires and CollectMetrics is called the 
subscriptionGroup is used to lookup up the plugins that will be called to 
ultimately perform the collection.

![collect_metrics](https://www.dropbox.com/s/s7eo4570vymsfcd/control_pluginControl_CollectMetrics.png?raw=1)
