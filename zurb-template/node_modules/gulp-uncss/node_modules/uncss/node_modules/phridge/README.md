phridge
=======
**A bridge between [node](http://nodejs.org/) and [PhantomJS](http://phantomjs.org/).**

[![](https://img.shields.io/npm/v/phridge.svg)](https://www.npmjs.com/package/phridge)
[![](https://img.shields.io/npm/dm/phridge.svg)](https://www.npmjs.com/package/phridge)
[![Dependency Status](https://david-dm.org/peerigon/phridge.svg)](https://david-dm.org/peerigon/phridge)
[![Build Status](https://travis-ci.org/peerigon/phridge.svg?branch=master)](https://travis-ci.org/peerigon/phridge)
[![Coverage Status](https://img.shields.io/coveralls/peerigon/phridge.svg)](https://Coveralls.io/r/peerigon/phridge?branch=master)

Working with PhantomJS in node is a bit cumbersome since you need to spawn a new PhantomJS process for every single task. However, spawning a new process is quite expensive and thus can slow down your application significantly.

phridge provides an api to easily

- spawn new PhantomJS processes
- run functions with arguments inside PhantomJS
- return results from PhantomJS to node
- manage long-running PhantomJS instances

Unlike other node-PhantomJS bridges phridge provides a way to run code directly inside PhantomJS instead of turning every call and assignment into an async operation.

phridge uses PhantomJS' stdin and stdout for [inter-process communication](http://en.wikipedia.org/wiki/Inter-process_communication). It stringifies the given function, passes it to PhantomJS via stdin, executes it in the PhantomJS environment and passes back the results via stdout. Thus you can write your PhantomJS scripts inside your node modules in a clean and synchronous way.

Instead of ...

```javascript
phantom.addCookie("cookie_name", "cookie_value", "localhost", function () {
    phantom.createPage(function (page) {
        page.set("customHeaders.Referer", "http://google.com", function () {
            page.set(
                "settings.userAgent",
                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5)",
                function () {
                    page.open("http://localhost:9901/cookie", function (status) {
                        page.evaluate(function (selector) {
                            return document.querySelector(selector).innerText;
                        }, function (text) {
                            console.log("The element contains the following text: "+ text)
                        }, "h1");
                    });
                }
            );
        });
    });
});
```

... you can write ...

```javascript
// node
phantom.run("h1", function (selector, resolve) {
    // this code runs inside PhantomJS

    phantom.addCookie("cookie_name", "cookie_value", "localhost");

    var page = webpage.create();
    page.customHeaders = {
        Referer: "http://google.com"
    };
    page.settings = {
        userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5)"
    };
    page.open("http://www.google.com", function () {
        var text = page.evaluate(function (selector) {
            return document.querySelector(selector).innerText;
        }, selector);

        // resolve the promise and pass 'text' back to node 
        resolve(text);
    });
}).then(function (text) {
    // inside node again
    console.log("The element contains the following text: " + text);
});
```

Please note that the `phantom`-object provided by phridge is completely different to the `phantom`-object inside PhantomJS. So is the `page`-object. [Check out the api](#api-phantom) for further information.

<br />

Installation
------------------------------------------------------------------------

`npm install phridge`

<br />

Examples
------------------------------------------------------------------------

### Spawn a new PhantomJS process

```javascript
phridge.spawn({
    proxyAuth: "john:1234",
    loadImages: false,
    // passing CLI-style options does also work
    "--remote-debugger-port": 8888
}).then(function (phantom) {
    // phantom is now a reference to a specific PhantomJS process
});
```

`phridge.spawn()` takes an object which will be passed as config to PhantomJS. Check out [their documentation](http://phantomjs.org/api/command-line.html) for a detailed overview of options. CLI-style options are added as they are, so be sure to escape the space character.

*Please note: There are [known issues](https://github.com/peerigon/phridge/issues/31) of PhantomJS that some config options are only supported in CLI-style.*

### Run any function inside PhantomJS

```javascript
phantom.run(function () {
    console.log("Hi from PhantomJS");
});
```

phridge stringifies the given function, sends it to PhantomJS and evals it again. Hence you can't use scope variables:

```javascript
var someVar = "hi";

phantom.run(function () {
    console.log(someVar); // throws a ReferenceError
});
```

### Passing arguments

You can also pass arguments to the PhantomJS process:

```javascript
phantom.run("hi", 2, {}, function (string, number, object) {
    console.log(string, number, object); // 'hi', 2, [object Object]
});
```

Arguments are stringified by `JSON.stringify()`, so be sure to use JSON-valid objects.

### Returning results

The given function can run sync and async. However, the `run()` method itself will always run async as it needs to wait for the process to respond.

**Sync**

```javascript
phantom.run(function () {
    return Math.PI;
}).then(function (pi) {
    console.log(pi === Math.PI); // true
});
```

**Async**

```javascript
phantom.run(function (resolve) {
    setTimeout(function () {
        resolve("after 500 ms");
    }, 500);
}).then(function (msg) {
    console.log(msg); // 'after 500 ms'
});
```

Results are also stringified by `JSON.stringify()`, so returning application objects with functions won't work.

```javascript
phantom.run(function () {
    ...
    // doesn't work because page is not a JSON-valid object
    return page;
});
```

### Returning errors

Errors can be returned by using the `throw` keyword or by calling the `reject` function. Both ways will reject the promise returned by `run()`.

**Sync**

```javascript
phantom.run(function () {
    throw new Error("An unknown error occured");
}).catch(function (err) {
    console.log(err); // 'An unknown error occured'
});
```

**Async**

```javascript
phantom.run(function (resolve, reject) {
    setTimeout(function () {
        reject(new Error("An unknown error occured"));
    }, 500);
}).catch(function (err) {
    console.log(err); // 'An unknown error occured'
});
```

### Async methods with arguments

`resolve` and `reject` are just appended to the regular arguments:

```javascript
phantom.run(1, 2, 3, function (one, two, three, resolve, reject) {

});
```

### Persisting states inside PhantomJS

Since the function passed to `phantom.run()` can't declare variables in the global scope, it is impossible to maintain state in PhantomJS. That's why `phantom.run()` calls all functions on the same context object. Thus you can easily store state variables.

```javascript
phantom.run(function () {
    this.message = "Hello from the first call";
}).then(function () {
    phantom.run(function () {
        console.log(this.message); // 'Hello from the first call'
    });
});
```

For further convenience all PhantomJS modules are already available in the global scope.

```javascript
phantom.run(function () {
    console.log(webpage);           // [object Object]
    console.log(system);            // [object Object]
    console.log(fs);                // [object Object]
    console.log(webserver);         // [object Object]
    console.log(child_process);     // [object Object]
});
```

### Working in a page context

Most of the time its more useful to work in a specific webpage context. This is done by creating a Page via `phantom.createPage()` which calls internally `require("webpage").create()`. The returned page wrapper will then execute all functions bound to a PhantomJS [webpage instance](http://phantomjs.org/api/webpage/). 

```javascript
var page = phantom.createPage();

page.run(function (resolve, reject) {
    // `this` is now a webpage instance
    this.open("http://example.com", function (status) {
        if (status !== "success") {
            return reject(new Error("Cannot load " + this.url));
        }
        resolve();
    });
});
```

And for the busy ones: You can just call `phantom.openPage(url)` which is basically the same as above:

```javascript
phantom.openPage("http://example.com").then(function (page) {
    console.log("Example loaded");
});
``` 

### Cleaning up

If you don't need a particular page anymore, just call:

```javascript
page.dispose().then(function () {
    console.log("page disposed");
});
```

This will clean up all page references inside PhantomJS.

If you don't need the whole process anymore call

```javascript
phantom.dispose().then(function () {
    console.log("process terminated");
});
```

which will terminate the process cleanly by calling `phantom.exit(0)` internally. You don't need to dispose all pages manuallly when you call `phantom.dispose()`.

However, calling

```javascript
phridge.disposeAll().then(function () {
    console.log("All processes created by phridge.spawn() have been terminated");
});
```

will terminate all processes.

**I strongly recommend to call** `phridge.disposeAll()` **when the node process exits as this is the only way to ensure that all child processes terminate as well.** Since `disposeAll()` is async it is not safe to call it on `process.on("exit")`. It is better to call it on `SIGINT`, `SIGTERM` and within your regular exit flow.

<br />

API
----

### phridge

#### .spawn(config?): Promise → Phantom

Spawns a new PhantomJS process with the given config. [Read the PhantomJS documentation](http://phantomjs.org/api/command-line.html) for all available config options. Use camelCase style for option names. The promise will be fulfilled with an instance of `Phantom`.

#### .disposeAll(): Promise

Terminates all PhantomJS processes that have been spawned. The promise will be fulfilled when all child processes emitted an `exit`-event.

#### .config.stdout: Stream = process.stdout

Destination stream where PhantomJS' [clean stdout](#phantom-childprocess-cleanstdout) will be piped to. Set it `null` if you don't want it. Changing the value does not affect processes that have already been spawned.

#### .config.stderr: Stream = process.stderr

Destination stream where PhantomJS' stderr will be piped to. Set it `null` if you don't want it. Changing the value does not affect processes that have already been spawned.

----

### <a name="api-phantom"></a>Phantom.prototype

#### .childProcess: ChildProcess

A reference to the [ChildProcess](http://nodejs.org/api/child_process.html#child_process_class_childprocess)-instance.

#### <a name="phantom-childprocess-cleanstdout"></a> .childProcess.cleanStdout: ReadableStream

phridge extends the [ChildProcess](http://nodejs.org/api/child_process.html#child_process_class_childprocess)-instance by a new stream called `cleanStdout`. This stream is piped to `process.stdout` by default. It provides all data not dedicated to phridge. Streaming data is considered to be dedicated to phridge when the new line is preceded by the classifier string `"message to node: "`.

#### <a name="phantom-run"></a>.run(args..., fn): Promise → *

Stringifies `fn`, sends it to PhantomJS and executes it there again. `args...` are stringified using `JSON.stringify()` and passed to `fn` again. `fn` may simply `return` a result or `throw` an error or call `resolve()` or `reject()` respectively if it is asynchronous. phridge compares `fn.length` with the given number of arguments to determine whether `fn` is sync or async. The returned promise will be resolved with the result or rejected with the error.

#### .createPage(): Page

Creates a wrapper to execute code in the context of a specific [PhantomJS webpage](http://phantomjs.org/api/webpage/).

#### .openPage(url): Promise → Page

Calls `phantom.createPage()`, then `page.open(url, cb)` inside PhantomJS and resolves when `cb` is called. If the returned `status` is not `"success"` the promise will be rejected.

#### .dispose(): Promise

Calls `phantom.exit(0)` inside PhantomJS and resolves when the child process emits an `exit`-event.

### Events

#### unexpectedExit

Will be emitted when PhantomJS exited without a call to `phantom.dispose()` or one of its std streams emitted an `error` event. This event may be fired on some OS when the process group receives a `SIGINT` or `SIGTERM` (see [#35](https://github.com/peerigon/phridge/pull/35)).

When an `unexpectedExit` event is encountered, the `phantom` instance will be unusable and therefore automatically disposed. Usually you don't need to listen for this event.

---

### Page.prototype

#### .phantom: Phantom

A reference to the parent [`Phantom`](#api-phantom) instance.

#### .run(args..., fn): Promise → *

Calls `fn` on the context of a PhantomJS page object. See [`phantom.run()`](#phantom-run) for further information.

#### .dispose(): Promise

Cleans up this page instance by calling `page.close()`

<br />

Contributing
------------------------------------------------------------------------

From opening a bug report to creating a pull request: **every contribution is appreciated and welcome**. If you're planing to implement a new feature or change the api please create an issue first. This way we can ensure that your precious work is not in vain.

All pull requests should have 100% test coverage (with notable exceptions) and need to pass all tests.

- Call `npm test` to run the unit tests
- Call `npm run coverage` to check the test coverage (using [istanbul](https://github.com/gotwarlost/istanbul))  

<br />

License
------------------------------------------------------------------------

Unlicense
