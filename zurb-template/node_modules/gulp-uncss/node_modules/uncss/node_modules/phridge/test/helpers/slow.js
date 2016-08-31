"use strict";

/**
 * @param {function} fn
 * @returns {function}
 */
function slow(fn) {
    if (fn.length === 1) {
        /**
         * @this Runner
         * @param {function} done
         * @returns {*}
         */
        return function (done) {
            this.slow(2000);
            this.timeout(15000);
            return fn.apply(this, arguments);
        };
    }
    /**
     * @this Runner
     * @returns {*}
     */
    return function () {
        this.slow(2000);
        this.timeout(15000);
        return fn.apply(this, arguments);
    };
}

module.exports = slow;
