"use strict";

var http = require("http");
var getport = require("getport");
var fs = require("fs");
var path = require("path");

var testPage = fs.readFileSync(path.join(__dirname, "/testPage.html"), "utf8");
var alamidLogo = fs.readFileSync(path.join(__dirname, "/alamid.png"), "utf8");
var server;

/**
 * @this Runner
 * @returns {Promise}
 */
function start() {
    var self = this;

    return new Promise(function (resolve, reject) {
        getport(30000, function (err, port) {
            if (err) {
                return reject(err);
            }
            if (server) {
                stop();
            }
            server = http
                .createServer(serveTestFiles)
                .listen(port, function onListen(err) {
                    if (err) {
                        return reject(err);
                    }
                    self.testServerUrl = "http://localhost:" + port;
                    resolve(port);
                });
        });
    });
}

function stop() {
    server.close();
    server.removeAllListeners();
}

function serveTestFiles(req, res) {
    if (req.url.indexOf("alamid") > -1) {
        res.setHeader("Content-Type", "image/png");
        res.end(alamidLogo);
        return;
    }
    res.setHeader("Content-Type", "text/html; charset=utf8");
    res.end(testPage, "utf8");
}

exports.start = start;
exports.stop = stop;
