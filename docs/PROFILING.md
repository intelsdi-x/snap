# Profiling Snap using pprof and go-torch

Using pprof and go-torch is a good way to bring to light on how your application performs. If you've never heard about profiling with the Go programming language, feel free to take a look at [Debugging performance issues in Go programs](https://software.intel.com/en-us/blogs/2014/05/10/debugging-performance-issues-in-go-programs).

## Set up ...
### FlameGraph
Flame graphs are a visualization of profiled software, allowing the most frequent code-paths to be identified quickly and accurately. FlameGraph needs to be installed with Perl before trying to use go-torch.

You can install Perl with your program manager.

*`brew` on OS X* 
```bash
brew install perl
```
*`apt` on Debian or Ubuntu* 
```bash
apt install perl
```
*`yum` on CentOS* 
```bash
yum install perl
```

Then download flamegraph.pl from the [flamegraph repository](https://github.com/brendangregg/FlameGraph):
```
wget https://raw.githubusercontent.com/brendangregg/FlameGraph/master/flamegraph.pl
```

And move it into your preferred $PATH directory, e.g. `~/bin`:
```bash
mv flamegraph.pl ~/bin/flamegraph
```
*Don't forget to rename it without the extension.*

You'll also have to make flamegraph executable:
```bash
chmod 755 ~/bin/flamegraph
```

### graphviz
Graphviz will allow you to generate graphs with pprof.
*`brew` on OS X* 
```bash
brew install graphviz
```
*`apt` on Debian or Ubuntu* 
```bash
apt install graphviz
```
*`yum` on CentOS* 
```bash
yum install graphviz
```

### go-torch

```bash
go get github.com/uber/go-torch
```

## Generating a profile
Before exploiting any result with go-torch and pprof we need to generate a profile - in our case a CPU profile. In this example, we'll use the package `github.com/pkg/profile`.

### Implement the code
So on your main function start the profile:
```go
import "github.com/pkg/profile"

var p interface {
	Stop()
}

func main() {
	p = profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook)
    ...
}
```

Then you must run `p.Stop()` just before exiting the program. In the snap daemon case it's just after handling a SIGINT (found in [snapd.go](../snapd.go)):
```go
func startInterruptHandling(modules ...coreModule) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

    //Let's block until someone tells us to quit
    go func() {
        sig := <-c
        p.Stop()    // Stop profiling
        ...
```

## Launch your program
Now launch your program - in this example, we started snapd and ran a task. The longer you run your task, the deeper your graph will go into the subroutines.

When your program quits you should have a `cpu.pprof` file generated in your current folder.

(If the size of this file is 0 bytes, it means you didn't properly stop profiling with `p.Stop()` in your code.)


## Use pprof and go-torch
From your terminal, you can now generate the flame graphs. In this example we assume that the `snapd` binary and cpu.pprof are in the same folder:
```
go-torch --binaryinput cpu.pprof --binaryname snapd
```

Same with pprof:
```
go tool pprof snapd cpu.pprof
```

Note that pprof allows you to watch the profile of any Go program executed during the profiling:
```
go tool pprof $SNAP_PLUGIN/snap-plugin-type-myplugin cpu.pprof
```
