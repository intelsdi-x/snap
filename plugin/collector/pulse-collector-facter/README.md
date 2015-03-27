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
