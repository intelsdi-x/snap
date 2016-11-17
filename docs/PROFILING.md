# Profiling Snap using pprof and go-torch

Using pprof and go-torch is a good way to bring to light on how your application performs. If you've never heard about profiling with the Go programming language, feel free to take a look at [Debugging performance issues in Go programs](https://software.intel.com/en-us/blogs/2014/05/10/debugging-performance-issues-in-go-programs).

## Getting started
If you have **Docker** installed, jump to [generate a profile](#generating-a-profile).

### System Requirements
Natively supported OS:
- Linux
- OS X 10.11+ ([patch for 10.6 - 10.10](https://github.com/rsc/pprof_mac_fix))


### go-torch
In order to use go-torch, we need:
- FlameGraph.
- Perl (required for FlameGraph).


Flame graphs are a visualization of profiled software, allowing the most frequent code-paths to be identified quickly and accurately.

#### Install Perl
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

#### Install FlameGraph
Then download `flamegraph.pl` from the [flamegraph repository](https://github.com/brendangregg/FlameGraph):
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

#### Install go-torch
(requires [Go](https://golang.org/doc/install)) 
```bash
go get github.com/uber/go-torch
```

## Generating a profile
### Run snapteld
Start snapteld with the `--pprof` flag:
```bash
snapteld -t 0 --pprof
```

Just doing that will create endpoints on the port used by Snap API. If you use the default port like in the example above, you should be able to access this url from your web browser: http://127.0.0.1:8181/debug/pprof

Next steps of this documentation will be focused on profiling tools, which is one part of what is exposed by `net/http/pprof` package. Find more information about other tools on the [official documentation](https://golang.org/pkg/net/http/pprof/#pkg-overview).

### Give some work to Snap
If you don't give work to Snap before generating profile, the **profile will be empty**. Some example of work:
- Run a task
- Send a lot of request to the API

### Generate a profile
You can now start profiling for a defined period of time (e.g. 120 seconds):
```bash
curl "http://127.0.0.1:8181/debug/pprof/profile?seconds=120" > cpu.pprof
```

During this step you actually have to wait (here 120 seconds), while Snap is doing work. **The longer is the period, the more accurate is the profile.**

## Use pprof and go-torch
From your terminal, you can now generate the flame graphs:

### Without Docker
```bash
go-torch --binaryinput cpu.pprof --binaryname `which snapteld`
```

### For Docker user
```bash
cp `which snapteld` .
docker run -v `pwd`:/tmp uber/go-torch --binaryinput /tmp/cpu.pprof --binaryname /tmp/snapteld -p > torch.svg
```

### Pprof tool
Once again, I highly encourage you to read about pprof in [this blog post](https://software.intel.com/en-us/blogs/2014/05/10/debugging-performance-issues-in-go-programs).
(requires [Go](https://golang.org/doc/install)) 
```bash
go tool pprof `which snapteld` cpu.pprof
```

## BONUS: using pprof tools for plugin
Plugins that use the [new Go library](https://github.com/intelsdi-x/snap-plugin-lib-go/tree/master/v1/plugin) will have their own endpoint.
```bash
$ snaptel plugin list --running
NAME 		 HIT COUNT 	 LAST HIT 			 TYPE 		 PPROF PORT
mock-grpc 	 54 		 Tue, 08 Nov 2016 13:08:38 PST 	 collector 	 62143
statistics 	 266 		 Tue, 08 Nov 2016 13:08:38 PST 	 processor 	 62148
influxdb 	 266 		 Tue, 08 Nov 2016 13:08:38 PST 	 publisher
```

On the column "pprof port" you can find the port used by the plugin to expose pprof tools. e.g.:
```bash
$curl "http://127.0.0.1:62148/debug/pprof/profile?seconds=10" > cpu.pprof
```


If there is no port, it means that plugin is compiled with an old library, or it's not a Go plugin. 