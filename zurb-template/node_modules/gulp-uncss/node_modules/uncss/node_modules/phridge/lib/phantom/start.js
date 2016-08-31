// Not using strict mode here because strict mode has an impact on evaled source code

/* eslint-disable no-unused-vars, camelcase */
// Yep, they are unused intentionally. They are just available for convenience reasons.
var webpage = require("webpage");
var system = require("system");
var fs = require("fs");
var webserver = require("webserver");
var child_process = require("child_process");
var configPath = system.args[1];
var config = JSON.parse(fs.read(configPath));
/*eslint-enable no-unused-vars, camelcase */
var pages = {};
var context = {};
var commandHandlers = {};

/**
 * Reads a line from stdin and calls the requested command handler. Calls itself after setTimeout 0.
 */
function loop() {
    // stdin.readLine() is sync and halts until a whole line is read
    var line = system.stdin.readLine();
    var message = JSON.parse(line);
    var handler = commandHandlers[message.action];

    if (!handler) {
        throw new Error("Unknown action '" + message.action + "'");
    }

    handler(message);

    setTimeout(loop, 0);
}

/**
 * Returns a function that should be called to return the result for this message.
 *
 * @param {Object} message
 * @returns {resolve}
 */
function createResolver(message) {
    /**
     * @param {Object} data
     */
    function resolve(data) {
        write({
            status: "success",
            id: message.done ? null : message.id,
            data: data
        });
        message.done = true;
    }

    return resolve;
}

/**
 * Returns a function that should be called to indicate that this message yielded to an error.
 *
 * @param {Object} message
 * @returns {reject}
 */
function createRejecter(message) {
    /**
     * @param {Object} data
     */
    function reject(data) {
        var stack;

        try {
            throw new Error(data ? data.message || "Error" : "Error");
        } catch (err) {
            stack = err.stack;

            stack += "\n" +
                "    -----------------------------------------------------------------------\n" +
                message.from;

            data = {
                message: err.message,
                stack: stack
            };
        }

        write({
            status: "fail",
            id: message.done ? null : message.id,
            data: data
        });
        message.done = true;
    }

    return reject;
}

/**
 * Runs message.data.src in the given context.
 *
 * @param {Object} message
 * @param {Object} context
 */
function run(message, context) {
    var resolve = createResolver(message);
    var reject = createRejecter(message);

    try {
        evalSrc(message.data.src, context, resolve, reject);
    } catch (err) {
        reject(err);
    }
}

/**
 * Helper function for run() to avoid scope pollution. `context`, `resolve` and `reject` are needed according
 * to the serializeFn-module.
 *
 * @param {string} src
 * @param {object} context
 * @param {Function} resolve
 * @param {Function} reject
 */
function evalSrc(src, context, resolve, reject) {
    eval(src); // eslint-disable-line
}

/**
 * Helper function that stringifies the given object and writes it to system.stdout
 *
 * @param {Object} message
 */
function write(message) {
    system.stdout.writeLine("message to node: " + JSON.stringify(message));
}

/**
 * Collection of request-able commands (as defined in the action-property of the message).
 *
 * @type {Object}
 */
commandHandlers = {

    /**
     * The ping command is a neat trick so PhantomJS isn't stuck in the stdin.readLine()-loop
     * while waiting for an asynchronous event. A ping-command is sent by node as long as it
     * waits for PhantomJS to respond. We're responding with a pong to tell node that we're waiting
     * for the next ping.
     */
    ping: function () {
        write({
            status: "pong"
        });
    },

    /**
     * Runs message.data.src in the default context.
     *
     * @param {Object} message
     */
    run: function (message) {
        run(message, context);
    },

    /**
     * Runs message.data.src in the requested page context. If the page context doesn't exist, a new webpage
     * is created implicitly.
     *
     * @param {Object} message
     */
    "run-on-page": function (message) {
        var pageId = message.data.pageId;
        var page = pages[pageId];

        if (!page) {
            pages[pageId] = page = webpage.create();
        }

        run(message, page);
    }
};

// remove the config as it is not needed anymore
fs.remove(configPath);

// send hi to node so node knows that we're ready
system.stdout.writeLine("message to node: hi");

loop();
