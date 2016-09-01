"use strict";

var slice = Array.prototype.slice;

function lift(fn) {
    /**
     * @this ctx
     * @returns {Promise}
     */
    return function () {
        var args = slice.call(arguments);
        var ctx = this;

        return new Promise(function (resolve, reject) {
            args.push(function (err, result) {
                err ? reject(err) : resolve(result);
            });
            fn.apply(ctx, args);
        });
    };
}

module.exports = lift;
