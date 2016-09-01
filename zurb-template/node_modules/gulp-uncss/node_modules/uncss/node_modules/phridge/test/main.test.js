"use strict";

var chai = require("chai");
var expect = chai.expect;
var config = require("../lib/config");
var spawn = require("../lib/spawn");
var disposeAll = require("../lib/disposeAll");
var phridge = require("../lib/main.js");

require("./helpers/setup.js");

describe("phridge", function () {

    describe(".config", function () {

        it("should be the config-module", function () {
            expect(phridge.config).to.equal(config);
        });

    });

    describe(".spawn", function () {

        it("should be the spawn-module", function () {
            expect(phridge.spawn).to.equal(spawn);
        });

    });

    describe(".disposeAll", function () {

        it("should be the exitAll-module", function () {
            expect(phridge.disposeAll).to.equal(disposeAll);
        });

    });

});