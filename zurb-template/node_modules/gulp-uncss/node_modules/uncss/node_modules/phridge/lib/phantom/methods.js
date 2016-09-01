"use strict";

/**
 * Opens the given page and resolves when PhantomJS called back.
 * Will be executed inside of PhantomJS.
 *
 * @private
 * @this Page
 * @param {string} url
 * @param {Function} resolve
 * @param {Function} reject
 */
function openPage(url, resolve, reject) { /* jshint validthis: true */
    this.open(url, function onPageLoaded(status) {
        if (status !== "success") {
            return reject(new Error("Cannot load " + url + ": PhantomJS returned status " + status));
        }
        resolve();
    });
}

/**
 * Calls phantom.exit() with errorcode 0
 *
 * @private
 */
function exitPhantom() { /* global phantom */
    Object.keys(pages).forEach(function (pageId) {
        // Closing all pages just to cleanup properly
        pages[pageId].close();
    });

    // Using setTimeout(0) to ensure that all JS code still waiting in the queue is executed before exiting
    // Otherwise PhantomJS prints a confusing security warning
    // @see https://github.com/ariya/phantomjs/commit/1eec21ed5c887bf21a1a6833da3c98c68401d90e
    setTimeout(function () {
        phantom.exit(0);
    }, 0);
}

/**
 * Cleans all references to a specific page.
 *
 * @private
 * @param {number} pageId
 */
function disposePage(pageId) { /* global pages */
    pages[pageId].close();
    delete pages[pageId];
}

exports.openPage = openPage;
exports.exitPhantom = exitPhantom;
exports.disposePage = disposePage;
