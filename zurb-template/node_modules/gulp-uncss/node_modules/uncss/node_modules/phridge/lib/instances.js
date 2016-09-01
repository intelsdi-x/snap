"use strict";

/**
 * Stores all active Phantom instances. After phantom.dispose() has been called, the phantom instance will
 * be removed from this array.
 *
 * @private
 * @type {Array}
 */
module.exports = [];