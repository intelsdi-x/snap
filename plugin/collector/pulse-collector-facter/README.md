## Pulse Fact Collector Plugin 

Collect facts from Facter and convert them into Pulse metrics.

Features:

* robust
* configurable
* caching layer that protects system against overuse

#### Entry point

./main.go

facter package content:

* facter.go - implements Pulse plugin API (discover&collect) and caching
* cmd.go - abstraction over external binary to collects fact from Facter 
