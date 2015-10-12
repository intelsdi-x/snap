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

## Pulse PCM Collector Plugin

# Description
Collect metrics from PCM (Intel Performance Counter Monitor)

# Assumptions
User is required to have PCM installed in system. It can be downloaded from [Intel PCM](http://www.intel.com/software/pcm) site.

# Suggestions
* Currently, Ubuntu 14.04 users have to manually compile it and add it to $PATH or $PULSE_PCM_PATH to be able to use it.
* pulsed has to be running with root privileges, for ability to collect data from PCM.
* To be able, to use PCM, [NMI watchdog](https://en.wikipedia.org/wiki/Non-maskable_interrupt) needs to be disabled. There are two ways to do this:
 * at running time: `echo 0 > /proc/sys/kernel/nmi_watchdog`
 * permanently: `echo 'kernel.nmi_watchdog=0' >> /etc/sysctl.conf`
