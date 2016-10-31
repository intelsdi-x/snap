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

<img src="https://cloud.githubusercontent.com/assets/6523391/19813910/24b75b66-9d3c-11e6-98ed-e6897faafd1c.png" width="20%" align="right">
</br>
</br>
</br>
</br>

# Tribe

Tribe is the name of the clustering feature in Snap.  When it is enabled, snapd instances can join to one another through an `agreement`. When an action is taken by one snapd instance that is a member of an agreement, that action will be carried out by all other members of the agreement. When a new snapd joins an existing agreement it will retrieve plugins and tasks from the members of the agreement.

## Usage
This walkthrough assumes you have downloaded a Snap release as described in [Getting Started](../README.md#getting-started).

### Starting the first snapd in tribe mode

Start the first node:
```
$ snapd --tribe -t 0
```

Only `--tribe` and some trust level (`-t 0`) is required to start tribe. This will result in defaults for all other parameters:
* Default `tribe-node-name` will be the same as your hostname
* Default `tribe-seed` is port 6000
* Default `tribe-addr` and `tribe-port` are the same as `snapd` and `tribe-seed` (ex. 127.0.0.1:6000)

See `snapd -h` for all the possible flags.

### Creating an initial agreement

Members of a tribe only share configuration once they join an agreement. To create your first agreement:
```
$ snapctl agreement create all-nodes
Name 	    Number of Members 	 plugins 	 tasks
all-nodes 	0 			         0 		     0
```

Join our running snapd into this agreement:
```
$ snapctl agreement join all-nodes `hostname`
Name 	    Number of Members 	 plugins 	 tasks
all-nodes 	1 			         0   		 0
```

### Joining other snapd into an existing tribe
Since tribe is implemented on top of a gossip based protocol there is no "master." All other nodes who join a tribe by communicating with any existing member.

Start another instance of snapd to join to our existing tribe. The local IP address is 192.168.136.176 in our example. Note that we need a few more parameters to avoid conflicting ports on a single system:
```
$ snapd --tribe -t 0 --tribe-port 6001 --api-port 8182 --tribe-node-name secondnodename --tribe-seed 192.168.136.176:6000 --control-listen-port 8083
```

Both snapd instances will see each other in their member list:
```
$ snapctl member list
Name
secondnodename
firstnode
```

This member needs to join the agreement:
```
$ snapctl agreement join all-nodes secondnodename
Name 		 Number of Members 	 plugins 	 tasks
all-nodes 	 2       			 0 		     0
```

From this point forward, any plugins or tasks you load will load into both members of this agreement.


## Examples

*Starting a 4 node cluster and listing members*
![tribe-start-list-members](http://i.giphy.com/xTk9ZZFdTeIFBFZgPu.gif)

*Note: Once the cluster is started subsequent new nodes can choose to establish
membership through **any** node as there is no "master".*


*Creating an agreement and joining members to it*
![tribe-create-join-agreement](http://i.giphy.com/d2YTZ5P1N0Gh4WJ2.gif)

In the example below an agreement has been created and all members of the
cluster have joined it.  After loading a collector and publishing
plugin and starting a task on one node we demonstrate that the plugins and
tasks are now running on all of the other nodes in the agreement.        


*Loading plugins and starting a task on a node participating in an agreement*
![tribe-load-start](http://i.giphy.com/3o8doZ9e9MX6ZOH4Iw.gif)
