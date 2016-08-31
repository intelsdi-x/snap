'use strict';

var uncss       = require('uncss'),
    gutil       = require('gulp-util'),
    assign      = require('object-assign'),
    transform   = require('stream').Transform,

    PLUGIN_NAME = 'gulp-uncss';

module.exports = function (options) {
    var stream = new transform({objectMode: true});

    // Ignore stylesheets in the HTML files; only use those from the stream
    options.ignoreSheets = [/\s*/];

    stream._transform = function (file, encoding, cb) {
        if (file.isStream()) {
            var error = 'Streaming not supported';
            return cb(new gutil.PluginError(PLUGIN_NAME, error));
        } else if (file.isBuffer()) {
            var contents = String(file.contents);
            if (!contents.length) {
                // Don't crash on empty files
                return cb(null, file);
            }
            options = assign(options, {raw: contents});
            uncss(options.html, options, function (err, output) {
                if (err) {
                    return cb(new gutil.PluginError(PLUGIN_NAME, err));
                }
                file.contents = new Buffer(output);
                cb(null, file);
            });
        } else {
            // Pass through when null
            cb(null, file);
        }
    };

    return stream;
};
