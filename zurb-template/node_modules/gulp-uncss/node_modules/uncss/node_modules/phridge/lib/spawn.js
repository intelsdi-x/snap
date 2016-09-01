"use strict";

var childProcess = require("child_process");
var phantomjs = require("phantomjs-prebuilt");
var config = require("./config.js");
var fs = require("fs");
var temp = require("temp");
var path = require("path");
var Phantom = require("./Phantom.js");
var forkStdout = require("./forkStdout.js");
var lift = require("./lift.js");

var startScript = path.resolve(__dirname, "./phantom/start.js");
var writeFile = lift(fs.writeFile);
var close = lift(fs.close);
var open = lift(temp.open);
var initialMessage = "message to node: hi";

/**
 * Spawns a new PhantomJS process with the given phantom config. Returns a Promises/A+ compliant promise
 * which resolves when the process is ready to execute commands.
 *
 * @see http://phantomjs.org/api/command-line.html
 * @param {Object} phantomJsConfig
 * @returns {Promise}
 */
function spawn(phantomJsConfig) {
    var args;
    var configPath;
    var stdout;
    var stderr;
    var child;

    phantomJsConfig = phantomJsConfig || {};

    // Saving a reference of the current stdout and stderr because this is (probably) the expected behaviour.
    // If we wouldn't save a reference, the config of a later state would be applied because we have to
    // do asynchronous tasks before piping the streams.
    stdout = config.stdout;
    stderr = config.stderr;

    /**
     * Step 1: Write the config
     */
    return open(null)
        .then(function writeConfig(info) {
            configPath = info.path;

            // Pass config items in CLI style (--some-config) separately to avoid Phantom's JSON config bugs
            // @see https://github.com/peerigon/phridge/issues/31
            args = Object.keys(phantomJsConfig)
                .filter(function filterCliStyle(configKey) {
                    return configKey.charAt(0) === "-";
                })
                .map(function returnConfigValue(configKey) {
                    var configValue = phantomJsConfig[configKey];

                    delete phantomJsConfig[configKey];

                    return configKey + "=" + configValue;
                });

            return writeFile(info.path, JSON.stringify(phantomJsConfig))
                .then(function () {
                    return close(info.fd);
                });
        })
    /**
     * Step 2: Start PhantomJS with the config path and pipe stderr and stdout.
     */
        .then(function startPhantom() {
            return new Promise(function (resolve, reject) {
                function onStdout(chunk) {
                    var message = chunk.toString("utf8");

                    child.stdout.removeListener("data", onStdout);
                    child.stderr.removeListener("data", onStderr);

                    if (message.slice(0, initialMessage.length) === initialMessage) {
                        resolve();
                    } else {
                        reject(new Error(message));
                    }
                }

                // istanbul ignore next because there is no way to trigger stderr artificially in a test
                function onStderr(chunk) {
                    var message = chunk.toString("utf8");

                    child.stdout.removeListener("data", onStdout);
                    child.stderr.removeListener("data", onStderr);

                    reject(new Error(message));
                }

                args.push(
                  "--config=" + configPath,
                  startScript,
                  configPath
                );

                child = childProcess.spawn(phantomjs.path, args);

                prepareChildProcess(child);

                child.stdout.on("data", onStdout);
                child.stderr.on("data", onStderr);

                // Our destination streams should not be ended if the childProcesses exists
                // thus { end: false }
                // @see https://github.com/peerigon/phridge/issues/27
                if (stdout) {
                    child.cleanStdout.pipe(stdout, { end: false });
                }
                if (stderr) {
                    child.stderr.pipe(stderr, { end: false });
                }
            });
        })
    /**
     * Step 3: Create the actual Phantom-instance and return it.
     */
        .then(function () {
            return new Phantom(child);
        });
}

/**
 * Prepares childProcess' stdout for communication with phridge. The childProcess gets two new properties:
 * - phridge: A stream which provides all messages to the phridge module
 * - cleanStdout: A stream which provides all the other data piped to stdout.
 *
 * @param {child_process.ChildProcess} childProcess
 * @private
 */
function prepareChildProcess(childProcess) {
    var stdoutFork = forkStdout(childProcess.stdout);

    childProcess.cleanStdout = stdoutFork.cleanStdout;
    childProcess.phridge = stdoutFork.phridge;
    childProcess.on("exit", disposeChildProcess);
}

/**
 * Clean up our childProcess extensions
 *
 * @private
 * @this ChildProcess
 */
function disposeChildProcess() {
    var childProcess = this;

    childProcess.phridge.removeAllListeners();
    childProcess.phridge = null;
    childProcess.cleanStdout.removeAllListeners();
    childProcess.cleanStdout = null;
}

module.exports = spawn;
