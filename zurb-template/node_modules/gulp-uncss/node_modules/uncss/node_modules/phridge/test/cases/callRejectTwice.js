"use strict";

var phridge = require("../../lib/main.js");

phridge.spawn()
    .then(function (phantom) {
        // It's necessary to exit cleanly because otherwise phantomjs doesn't exit on Windows
        process.on("uncaughtException", function (err) {
            console.error(err);
            phantom.dispose().then(function () {
                process.exit(0);
            });
        });

        phantom.run(function (resolve, reject) {
            reject();
            reject();
        });
    });