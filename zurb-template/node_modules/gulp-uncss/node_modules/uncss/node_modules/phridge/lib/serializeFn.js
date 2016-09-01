"use strict";

/**
 * Serializes the function and its arguments to send it to stdin.
 *
 * @param {Function} fn
 * @param {Array} args
 * @returns {string}
 */
function serializeFn(fn, args) {
    var fnIsSync = args.length === fn.length;
    var src;

    args = args.map(JSON.stringify);
    args.unshift("context");

    if (fnIsSync) {
        src = "resolve((" + fn.toString() + ").call(" + args.join() + "));";
    } else {
        args.push("resolve", "reject");
        src = "(" + fn.toString() + ").call(" + args.join() + ");";
    }

    // Currently sourceURLs aren't supported by PhantomJS but maybe in the future
    return src + "//# sourceURL=phridge.js";
}

module.exports = serializeFn;