"use strict";

var chai = require("chai");
var expect = chai.expect;
var phridge = require("../lib/main.js");
var slow = require("./helpers/slow.js");

require("./helpers/setup.js");

describe("disposeAll()", function () {

    it("should exit cleanly all running phantomjs instances", slow(function () {
        var exited = [];

        return Promise.all([
            phridge.spawn(),
            phridge.spawn(),
            phridge.spawn()
        ])
        .then(function (p) {
            p[0].childProcess.on("exit", function () { exited.push(0); });
            p[1].childProcess.on("exit", function () { exited.push(1); });
            p[2].childProcess.on("exit", function () { exited.push(2); });

            return phridge.disposeAll();
        })
        .then(function () {
            exited.sort();
            expect(exited).to.eql([0, 1, 2]);
        });
    }));

});
