<!--
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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

## Pulse Perf Events Collector Plugin

# Description
Collect following HW metrics for Cgroups from "perf" - Performance Counters for Linux:
*  cycles
*  instructions
*  cache-references
*  cache-misses
*  branch-instructions
*  branch-misses
*  stalled-cycles-frontend
*  stalled-cycles-backend
*  ref-cycles

 By default metrics are gathered once per second.

# Assumptions
* "perf" - performance monitoring tool installed.
* /proc/sys/kernel/perf_event_paranoid set to 0 (echo 0 > /proc/sys/kernel/perf_event_paranoid) 
* Linux kernel version 2.6.31+

# Tips
Creating sample cgroup for testing:
* create sample process
- dd if=/dev/zero of=/dev/null &
- pid=$!

* create cgroup and move process into cgroup
- sudo cgcreate -g perf_event:A -g cpu:A -g cpuset:A -g cpuacct:A
- sudo cgclassify -g perf_event:A -g cpu:A -g cpuacct:A $pid
- sudo cgset -r cpuset.cpus=0-7 A
- sudo cgset -r cpuset.mems=0 A
- sudo cgclassify -g cpuset:A $pid
- sudo cgset -r cpu.shares=20 A

* list cgroup
- lscgroup | grep perf_event
