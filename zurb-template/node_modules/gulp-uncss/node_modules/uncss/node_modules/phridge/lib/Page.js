"use strict";

var serializeFn = require("./serializeFn.js");
var phantomMethods = require("./phantom/methods.js");

var slice = Array.prototype.slice;

/**
 * A wrapper to run code within the context of a specific PhantomJS webpage.
 *
 * @see http://phantomjs.org/api/webpage/
 * @param {Phantom} phantom the parent PhantomJS instance
 * @param {number} id internal page id
 * @constructor
 */
function Page(phantom, id) {
    Page.prototype.constructor.apply(this, arguments);
}

/**
 * The parent phantom instance.
 *
 * @type {Phantom}
 */
Page.prototype.phantom = null;

/**
 * The internal page id.
 *
 * @private
 * @type {number}
 */
Page.prototype._id = null;

/**
 * Initializes the page instance.
 *
 * @param {Phantom} phantom
 * @param {number} id
 */
Page.prototype.constructor = function (phantom, id) {
    this.phantom = phantom;
    this._id = id;
};

/**
 * Stringifies the given function fn, sends it to PhantomJS and runs it in the context of a particular PhantomJS webpage.
 * The PhantomJS webpage will be available as `this`. You may prepend any number of arguments which will be passed
 * to fn inside of PhantomJS. Please note that all arguments should be stringifyable with JSON.stringify().
 *
 * @param {...*} args
 * @param {Function} fn
 * @returns {Promise}
 */
Page.prototype.run = function (args, fn) {
    args = slice.call(arguments);

    fn = args.pop();

    return this.phantom._send({
        action: "run-on-page",
        data: {
            src: serializeFn(fn, args),
            pageId: this._id
        }
    }, args.length === fn.length);
};

/**
 * Runs a function inside of PhantomJS to cleanup memory. Call this function if you intent to not use the page-object
 * anymore.
 *
 * @see http://msdn.microsoft.com/en-us/library/system.idisposable.aspx
 * @returns {Promise}
 */
Page.prototype.dispose = function () {
    var self = this;

    return this.run(this._id, phantomMethods.disposePage)
        .then(function () {
            self.phantom = null;
        });
};

module.exports = Page;
