# Plugin Packaging 

Snap supports the ACI (App Container
Image) format defined in the 
[App Container spec (appc)](https://github.com/appc/spec) for packaging a 
plugin.

When Snap loads a plugin it detects the plugins type.  If the plugin is a binary
the plugin is run by snapteld which handshakes with the plugin via reading its 
standard output.  If the plugin is packaged as an ACI image it is extracted
and Snap executes the program referenced by the `exec` field.

## Why  

In cases where we cannot or do not want to compile our plugin into a statically 
linked binary we can load a plugin packaged as an ACI image.  This provides 
an obvious advantage for plugins written in Python, Ruby, Java, etc where the 
plugins dependencies, potentially including an entire Python virtualenv, could 
be distributed with the plugin.  

## How

Since Snap leverages the appc spec for images we recommend using the 
[acbuild](https://github.com/appc/acbuild) tool for creating images.  

In the example below we package one of the mock plugins creating a plugin 
package which can be loaded achieving the same result as if we had simply 
loaded the binary version of the plugin. 

1. Get the [acbuild](https://github.com/appc/acbuild) tool  
    * Download the latest binary 
    [release](https://github.com/appc/acbuild/releases) and install into your 
    PATH.
2. Make Snap
    * From the root of snap run: `make`
4. Using the acbuild tool create an image containing the mock collector plugin.
    * From the `build/plugin` directory run the following commands.
    ``` 
    acbuild begin
    acbuild set-name intelsdi-x/snap-plugin-collector-mock1
    acbuild copy snap-plugin-collector-mock1 /bin/snap-plugin-collector-mock1
    acbuild set-exec /bin/snap-plugin-collector-mock1
    acbuild write snap-plugin-collector-mock1-linux-x86_64.aci
    acbuild end
    ```
![example](https://cloud.githubusercontent.com/assets/10092554/20983225/8355a382-bc70-11e6-82c6-6ac445e16513.gif)

That's it!