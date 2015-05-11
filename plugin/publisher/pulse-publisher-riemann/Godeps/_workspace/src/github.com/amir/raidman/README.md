Raidman
=======

Go Riemann client

```go
package main

import (
        "github.com/amir/raidman"
)

func main() {
        c, err := raidman.Dial("tcp", "localhost:5555")
        if err != nil {
                panic(err)
        }

        var event = &raidman.Event{
                State:   "success",
                Host:    "raidman",
                Service: "raidman-sample",
                Metric:  100,
                Ttl:     10,
        }

        err = c.Send(event)
        if err != nil {
                panic(err)
        }

        events, err := c.Query("host = \"raidman\"")
        if err != nil {
                panic(err)
        }

        if len(events) < 1 {
                panic("Submitted event not found")
        }

        c.Close()
}

```
