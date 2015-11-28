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

# Tribe

Tribe is the name of the clustering feature in snap.  When it is enabled nodes 
will agree on plugins and/or tasks when they join what is called an 
`agreement`. When an action is taken on a node that is a member of an agreement
that action will be carried out by all other members of the agreement. When a 
new node joins an existing agreement it will retrieve plugins and tasks from 
the members of the agreement. 

## Usage

## Starting snapd in tribe mode

The first node

```
$SNAP_PATH/bin/snapd --tribe
```

All other nodes who join will need to select any existing member of the cluster.
Since tribe is implemented on top of a gossip based protocol there is no 
"master".

```
$SNAP_PATH/bin/snapd --tribe-seed <ip or name of another tribe member>
```

## Member

After starting in tribe mode all nodes in the cluster can be listed.

```
$SNAP_PATH/bin/snapctl member list
```

*Starting a 4 node cluster and listing members*
![tribe-start-list-members](http://i.giphy.com/xTk9ZZFdTeIFBFZgPu.gif)

*Note: Once the cluster is started subsequent new nodes can choose to establish
membership through **any** node as there is no "master".* 

## Agreement

#### create

```
$SNAP_PATH/bin/snapctl agreement create <agreement_name>
```

#### list

```
$SNAP_PATH/bin/snapctl agreeement list
```

#### join

```
$SNAP_PATH/bin/snapctl agreeement join <agreement_name> <member_name>
```

#### delete

```
$SNAP_PATH/bin/snapctl agreeement delete <agreement_name>
```

#### leave

```
$SNAP_PATH/bin/snapctl agreement leave <agreement_name> <member_name>
```

*Creating an agreement and joining members to it*
![tribe-create-join-agreement](http://i.giphy.com/d2YTZ5P1N0Gh4WJ2.gif)

## Managing nodes in a tribe agreement

After an agreement is created and members join it an action, such 
as loading/unloading plugins and adding/removing and starting/stopping tasks, 
taken on a single node in the agreement will be carried out on all members of 
the agreement.

In the example below an agreement has been created and all members of the 
cluster have joined it.  After loading a collector and publishing 
plugin and starting a task on one node we demonstrate that the plugins and 
tasks are now running on all of the other nodes in the agreement.        


*Loading plugins and starting a task on a node participating in an agreement
![tribe-load-start](http://i.giphy.com/3o8doZ9e9MX6ZOH4Iw.gif)
