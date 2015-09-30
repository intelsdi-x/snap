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

## Pulse Fact Collector Plugin 


## Description

Collect facts from Facter and convert them into Pulse metrics.

Features:

- robust
- configurable (timeout)

## Assumptions

- returns nothing when asked for nothing
- always return what was asked for - may return nil as value
- does not check cohesion between return metrics type GetMetricType and what is asked for

### Entry point

./main.go

facter package content:

* facter.go - implements Pulse plugin API (discover&collect)
* cmd.go - abstraction over external binary to collects fact from Facter 
