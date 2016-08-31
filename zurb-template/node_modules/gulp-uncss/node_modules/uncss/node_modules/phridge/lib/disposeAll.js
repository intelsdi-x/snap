"use strict";

var instances = require("./instances.js");

/**
 * Terminates all running PhantomJS processes. Returns a Promises/A+ compliant promise
 * which resolves when a processes terminated cleanly.
 *
 * @returns {Promise}
 */
function disposeAll() {
    var copy = instances.slice(0); // copy the array because phantom.dispose() will modify it

    return Promise.all(copy.map(exit));
}

/**
 * @private
 * @param {Phantom} phantom
 * @returns {Promise}
 */
function exit(phantom) {
    return phantom.dispose();
}

module.exports = disposeAll;
