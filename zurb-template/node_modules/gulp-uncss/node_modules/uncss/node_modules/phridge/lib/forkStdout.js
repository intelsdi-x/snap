"use strict";

var os = require("os");
var util = require("util");
var ForkStream = require("fork-stream");
var Linerstream = require("linerstream");
var Transform = require("stream").Transform;

var messageToNode = "message to node: ";

/**
 * Creates a fork stream which pipes messages starting with 'message to node: ' to our phridge stream
 * and any other message to the other stream. Thus console.log() inside PhantomJS is still printed to the
 * console while using stdout as communication channel for phridge.
 *
 * @param {stream.Readable} stdout
 * @returns {{phridge: stream.Readable, cleanStdout: stream.Readable}}
 */
function forkStdout(stdout) {
    var fork;
    var phridgeEndpoint;
    var cleanStdoutEndpoint;

    // Expecting a character stream because we're splitting messages by an EOL-character
    stdout.setEncoding("utf8");

    fork = new ForkStream({
        classifier: function (chunk, done) {
            chunk = chunk
                .slice(0, messageToNode.length);
            done(null, chunk === messageToNode);
        }
    });

    stdout
        .pipe(new Linerstream())
        .pipe(fork);

    // Removes the 'message to node: '-prefix from every chunk.
    phridgeEndpoint = fork.a.pipe(new CropPhridgePrefix({
        encoding: "utf8"
    }));

    // We need to restore EOL-character in stdout stream
    cleanStdoutEndpoint = fork.b.pipe(new RestoreLineBreaks({
        encoding: "utf8"
    }));

    return {
        phridge: phridgeEndpoint,
        cleanStdout: cleanStdoutEndpoint
    };
}

/**
 * Appends an EOL-character to every chunk.
 *
 * @param {Object} options stream options
 * @constructor
 * @private
 */
function RestoreLineBreaks(options) {
    Transform.call(this, options);
}
util.inherits(RestoreLineBreaks, Transform);

RestoreLineBreaks.prototype._transform = function (chunk, enc, cb) {
    this.push(chunk + os.EOL);
    cb();
};

/**
 * Removes the 'message to node: '-prefix from every chunk.
 *
 * @param {Object} options stream options
 * @constructor
 * @private
 */
function CropPhridgePrefix(options) {
    Transform.call(this, options);
}
util.inherits(CropPhridgePrefix, Transform);

CropPhridgePrefix.prototype._transform = function (chunk, enc, cb) {
    this.push(chunk.slice(messageToNode.length));
    cb();
};

module.exports = forkStdout;
