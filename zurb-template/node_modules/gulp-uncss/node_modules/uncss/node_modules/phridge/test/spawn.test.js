"use strict";

var chai = require("chai");
var getport = require("getport");
var net = require("net");
var expect = chai.expect;
var spawn = require("../lib/spawn.js");
var phridge = require("../lib/main.js");
var Phantom = require("../lib/Phantom.js");
var slow = require("./helpers/slow.js");
var lift = require("../lib/lift.js");

require("./helpers/setup.js");

getport = lift(getport);

describe("spawn(config?)", function () {
    var port;

    function blockGhostDriverPort() {
        return getport(10000)
            .then(function (freePort) {
                var server = net.createServer();
                var listen = lift(server.listen);

                port = freePort;

                // We're blocking the GhostDriver port so phantomjs crashes on startup.
                // Otherwise the phantomjs processes can't be killed because it doesn't
                // listen on our commands in GhostDriver-mode.
                return listen.call(server, freePort);
            });
    }

    after(slow(function () {
        return phridge.disposeAll();
    }));

    it("should resolve to an instance of Phantom", slow(function () {
        return expect(spawn()).to.eventually.be.an.instanceOf(Phantom);
    }));

    it("should pass the provided config to phantomjs", slow(function (done) {
        // We're setting the webdriver option to test if the config is recognized
        // Setting this option does not make any sense because phantomjs is
        // unusable with phantomjs in GhostDriver-mode. But it prints a nice
        // message to the console which causes the promise to be rejected
        blockGhostDriverPort()
            .then(function () {
                // Prevent PhantomJS from printing a disturbing error message to the console
                phridge.config.stdout = null;
                phridge.config.stderr = null;

                return expect(phridge.spawn({
                    webdriver: "localhost:" + port
                })).to.be.rejectedWith(/GhostDriver - main\.fail/);
            })
            .then(function () {
                phridge.config.stdout = process.stdout;
                phridge.config.stderr = process.stderr;

                // Give phantomjs some time to exit
                setTimeout(done, 100);
            });
    }));

    it("should also allow CLI style config options", slow(function (done) {
        // We're setting the webdriver option to test if the config is recognized
        // Setting this option does not make any sense because phantomjs is
        // unusable with phantomjs in GhostDriver-mode. But it prints a nice
        // message to the console which causes the promise to be rejected
        blockGhostDriverPort()
            .then(function () {
                // Prevent PhantomJS from printing a disturbing error message to the console
                phridge.config.stdout = null;
                phridge.config.stderr = null;

                return expect(phridge.spawn({
                    "--load-images": "true",
                    "--webdriver": "localhost:" + port
                })).to.be.rejectedWith("GhostDriver");
            })
            .then(function () {
                phridge.config.stdout = process.stdout;
                phridge.config.stderr = process.stderr;

                // Give phantomjs some time to exit
                setTimeout(done, 100);
            })
            .catch(done);
    }));

});
