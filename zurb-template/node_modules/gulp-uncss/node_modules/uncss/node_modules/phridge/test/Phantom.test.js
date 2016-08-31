"use strict";

/* eslint-env browser */
/* global config */

var chai = require("chai");
var sinon = require("sinon");
var EventEmitter = require("events").EventEmitter;
var childProcess = require("child_process");
var ps = require("ps-node");
var expect = chai.expect;
var phridge = require("../lib/main.js");
var Phantom = require("../lib/Phantom.js");
var Page = require("../lib/Page.js");
var instances = require("../lib/instances.js");
var slow = require("./helpers/slow.js");
var testServer = require("./helpers/testServer.js");
var Writable = require("stream").Writable;
var createChildProcessMock = require("./helpers/createChildProcessMock.js");

require("./helpers/setup.js");

describe("Phantom", function () {
    var childProcessMock = createChildProcessMock();
    var phantom;
    var spawnPhantom;
    var exitPhantom;
    var stdout;
    var stderr;

    function mockConfigStreams() {
        stdout = phridge.config.stdout;
        stderr = phridge.config.stderr;
        phridge.config.stdout = new Writable();
        phridge.config.stderr = new Writable();
    }

    function unmockConfigStreams() {
        phridge.config.stdout = stdout;
        phridge.config.stderr = stderr;
    }

    spawnPhantom = slow(function () {
        if (phantom && phantom._isDisposed === false) {
            return undefined;
        }
        return phridge.spawn({ someConfig: true })
            .then(function (newPhantom) {
                phantom = newPhantom;
            });
    });
    exitPhantom = slow(function () {
        if (!phantom) {
            return undefined;
        }
        return phantom.dispose();
    });

    before(testServer.start);
    after(exitPhantom);
    after(testServer.stop);

    describe(".prototype", function () {

        beforeEach(spawnPhantom);

        it("should inherit from EventEmitter", function () {
            expect(phantom).to.be.instanceOf(EventEmitter);
        });

        describe("when an unexpected error on the childProcess occurs", function () {

            it("should emit an 'unexpectedExit'-event", function (done) {
                var error = new Error("Something bad happened");

                phantom.on("unexpectedExit", function (err) {
                    expect(err).to.equal(error);
                    done();
                });
                phantom.childProcess.emit("error", error);
            });

        });

        describe("when the childProcess was killed autonomously", function () {

            it("should be safe to call .dispose() after the process was killed", function () {
                phantom.childProcess.kill();
                return phantom.dispose();
            });

            it("should emit an 'unexpectedExit'-event", function (done) {
                phantom.on("unexpectedExit", function () {
                    done();
                });
                phantom.childProcess.kill();
            });

            it("should not emit an 'unexpectedExit'-event when the phantom instance was disposed in the meantime", function (done) {
                phantom.on("unexpectedExit", function () {
                    done(); // Will trigger an error that done() has been called twice
                });
                phantom.childProcess.kill();
                phantom.dispose().then(done, done);
            });

        });

        describe(".constructor(childProcess)", function () {
            var phantom; // It's important to shadow the phantom variable inside this describe block.
                         // This way spawnPhantom and exitPhantom don't use our mocked Phantom instance.

            after(function () {
                // Remove mocked Phantom instances from the instances-array
                instances.length = 0;
            });

            it("should return an instance of Phantom", function () {
                phantom = new Phantom(childProcessMock);
                expect(phantom).to.be.an.instanceof(Phantom);
            });

            it("should set the childProcess", function () {
                phantom = new Phantom(childProcessMock);
                expect(phantom.childProcess).to.equal(childProcessMock);
            });

            it("should add the instance to the instances array", function () {
                expect(instances).to.contain(phantom);
            });

        });

        describe(".childProcess", function () {

            it("should provide a reference on the child process object created by node", function () {
                expect(phantom.childProcess).to.be.an("object");
                expect(phantom.childProcess.stdin).to.be.an("object");
                expect(phantom.childProcess.stdout).to.be.an("object");
                expect(phantom.childProcess.stderr).to.be.an("object");
            });

        });

        describe(".run(arg1, arg2, arg3, fn)", function () {

            describe("with fn being an asynchronous function", function () {

                it("should provide a resolve function", function () {
                    return expect(phantom.run(function (resolve) {
                        resolve("everything ok");
                    })).to.eventually.equal("everything ok");
                });

                it("should provide the possibility to resolve with any stringify-able data", function () {
                    return Promise.all([
                        expect(phantom.run(function (resolve) {
                            resolve();
                        })).to.eventually.equal(undefined),
                        expect(phantom.run(function (resolve) {
                            resolve(true);
                        })).to.eventually.equal(true),
                        expect(phantom.run(function (resolve) {
                            resolve(2);
                        })).to.eventually.equal(2),
                        expect(phantom.run(function (resolve) {
                            resolve(null);
                        })).to.eventually.equal(null),
                        expect(phantom.run(function (resolve) {
                            resolve([1, 2, 3]);
                        })).to.eventually.deep.equal([1, 2, 3]),
                        expect(phantom.run(function (resolve) {
                            resolve({
                                someArr: [1, 2, 3],
                                otherObj: {}
                            });
                        })).to.eventually.deep.equal({
                            someArr: [1, 2, 3],
                            otherObj: {}
                        })
                    ]);
                });

                it("should provide a reject function", function () {
                    return phantom.run(function (resolve, reject) {
                        reject(new Error("not ok"));
                    }).catch(function (err) {
                        expect(err.message).to.equal("not ok");
                    });
                });

                it("should print an error when resolve is called and the request has already been finished", slow(function (done) {
                    var execPath = '"' + process.execPath + '" ';

                    childProcess.exec(execPath + require.resolve("./cases/callResolveTwice"), function (error, stdout, stderr) {
                        expect(error).to.equal(null);
                        expect(stderr).to.contain("Cannot call resolve() after the promise has already been resolved or rejected");
                        done();
                    });
                }));

                it("should print an error when reject is called and the request has already been finished", slow(function (done) {
                    var execPath = '"' + process.execPath + '" ';

                    childProcess.exec(execPath + require.resolve("./cases/callRejectTwice"), function (error, stdout, stderr) {
                        expect(error).to.equal(null);
                        expect(stderr).to.contain("Cannot call reject() after the promise has already been resolved or rejected");
                        done();
                    });
                }));

            });

            describe("with fn being a synchronous function", function () {

                it("should resolve to the returned value", function () {
                    return expect(phantom.run(function () {
                        return "everything ok";
                    })).to.eventually.equal("everything ok");
                });

                it("should provide the possibility to resolve with any stringify-able data", function () {
                    return Promise.all([
                        expect(phantom.run(function () {
                            // returns undefined
                        })).to.eventually.equal(undefined),
                        expect(phantom.run(function () {
                            return true;
                        })).to.eventually.equal(true),
                        expect(phantom.run(function () {
                            return 2;
                        })).to.eventually.equal(2),
                        expect(phantom.run(function () {
                            return null;
                        })).to.eventually.equal(null),
                        expect(phantom.run(function () {
                            return [1, 2, 3];
                        })).to.eventually.deep.equal([1, 2, 3]),
                        expect(phantom.run(function () {
                            return {
                                someArr: [1, 2, 3],
                                otherObj: {}
                            };
                        })).to.eventually.deep.equal({
                            someArr: [1, 2, 3],
                            otherObj: {}
                        })
                    ]);
                });

                it("should reject the promise if fn throws an error", function () {
                    return phantom.run(function () {
                        throw new Error("not ok");
                    }).catch(function (err) {
                        expect(err.message).to.equal("not ok");
                    });
                });

            });

            it("should provide all phantomjs default modules as convenience", function () {
                return expect(phantom.run(function () {
                    return Boolean(webpage && system && fs && webserver && child_process); // eslint-disable-line
                })).to.eventually.equal(true);
            });

            it("should provide the config object to store all kind of configuration", function () {
                return expect(phantom.run(function () {
                    return config;
                })).to.eventually.deep.equal({
                    someConfig: true
                });
            });

            it("should provide the possibility to pass params", function () {
                var params = {
                    some: ["param"],
                    withSome: "crazy",
                    values: {
                        number1: 1
                    }
                };

                return expect(phantom.run(params, params, params, function (params1, params2, params3) {
                    return [params1, params2, params3];
                })).to.eventually.deep.equal([params, params, params]);
            });

            it("should report errors", function () {
                return expect(phantom.run(function () {
                    undefinedVariable; // eslint-disable-line
                })).to.be.rejectedWith("Can't find variable: undefinedVariable");
            });

            it("should preserve all error details like stack traces", function () {
                return Promise.all([
                    phantom
                        .run(function brokenFunction() {
                            undefinedVariable; // eslint-disable-line
                        }).catch(function (err) {
                            expect(err).to.have.property("message", "Can't find variable: undefinedVariable");
                            expect(err).to.have.property("stack");
                            //console.log(err.stack);
                        }),
                    phantom
                        .run(function (resolve, reject) {
                            reject(new Error("Custom Error"));
                        })
                        .catch(function (err) {
                            expect(err).to.have.property("message", "Custom Error");
                            expect(err).to.have.property("stack");
                        })
                ]);
            });

            it("should run all functions on the same empty context", function () {
                return phantom.run(/** @this Object */function () {
                    if (JSON.stringify(this) !== "{}") {
                        throw new Error("The context is not an empty object");
                    }
                    this.message = "Hi from the first run";
                }).then(function () {
                    return phantom.run(/** @this Object */function () {
                        if (this.message !== "Hi from the first run") {
                            throw new Error("The context is not persistent");
                        }
                    });
                });
            });

            it("should reject with an error if PhantomJS process is killed", function () {
                // Phantom will eventually emit an error event when the childProcess was killed
                // In order to prevent node from throwing the error, we need to add a dummy error event listener
                phantom.on("error", Function.prototype);
                phantom.childProcess.kill();
                return phantom.run(function () {})
                    .then(function () {
                        throw new Error("There should be an error");
                    }, function (err) {
                        expect(err).to.be.an.instanceOf(Error);
                        expect(err.message).to.contain("Cannot communicate with PhantomJS process");
                        expect(err.originalError).to.be.an.instanceOf(Error);
                        expect(err.message).to.contain(err.originalError.message);
                    });
            });

        });

        describe(".createPage()", function () {

            it("should return an instance of Page", function () {
                expect(phantom.createPage()).to.be.an.instanceof(Page);
            });

        });

        describe(".openPage(url)", function () {

            it("should resolve to an instance of Page", slow(/** @this Runner */function () {
                return expect(phantom.openPage(this.testServerUrl)).to.eventually.be.an.instanceof(Page);
            }));

            it("should resolve when the given page has loaded", slow(/** @this Runner */function () {
                return phantom.openPage(this.testServerUrl).then(function (page) {
                    return page.run(/** @this WebPage */function () {
                        var headline;
                        var imgIsLoaded;

                        headline = this.evaluate(function () {
                            return document.querySelector("h1").innerText;
                        });
                        imgIsLoaded = this.evaluate(function () {
                            return document.querySelector("img").width > 0;
                        });

                        if (headline !== "This is a test page") {
                            throw new Error("Unexpected headline: " + headline);
                        }
                        if (imgIsLoaded !== true) {
                            throw new Error("The image has not loaded yet");
                        }
                    });
                });
            }));

            it("should reject when the page is not available", slow(function () {
                return expect(
                    phantom.openPage("http://localhost:1")
                ).to.be.rejectedWith("Cannot load http://localhost:1: PhantomJS returned status fail");
            }));

        });

        describe(".dispose()", function () {

            before(mockConfigStreams);
            beforeEach(spawnPhantom);
            after(unmockConfigStreams);

            it("should terminate the child process with exit-code 0 and then resolve", slow(function () {
                var exit = false;

                phantom.childProcess.on("exit", function (code) {
                    expect(code).to.equal(0);
                    exit = true;
                });

                return phantom.dispose().then(function () {
                    expect(exit).to.equal(true);
                    phantom = null;
                });
            }));

            it("should remove the instance from the instances array", slow(function () {
                return phantom.dispose().then(function () {
                    expect(instances).to.not.contain(phantom);
                    phantom = null;
                });
            }));

            // @see https://github.com/peerigon/phridge/issues/27
            it("should neither call end() on config.stdout nor config.stderr", function () {
                phridge.config.stdout.end = sinon.spy();
                phridge.config.stderr.end = sinon.spy();

                return phantom.dispose().then(function () {
                    expect(phridge.config.stdout.end).to.have.callCount(0);
                    expect(phridge.config.stderr.end).to.have.callCount(0);
                    phantom = null;
                });
            });

            it("should be safe to call .dispose() multiple times", slow(function () {
                return Promise.all([
                    phantom.dispose(),
                    phantom.dispose(),
                    phantom.dispose()
                ]);
            }));

            it("should not be possible to call .run() after .dispose()", function () {
                expect(phantom.dispose().then(function () {
                    return phantom.run(function () {});
                })).to.be.rejectedWith("Cannot run function: Phantom instance is already disposed");
            });

            it("should not be possible to call .run() after an unexpected exit", function () {
                phantom.childProcess.emit("error");
                return phantom.run(function () {})
                    .then(function () {
                        throw new Error("There should be an error");
                    }, function (err) {
                        expect(err).to.be.an.instanceOf(Error);
                        expect(err.message).to.contain("Cannot run function");
                        expect(err.originalError).to.be.an.instanceOf(Error);
                        expect(err.message).to.contain(err.originalError.message);
                    });
            });

            // @see https://github.com/peerigon/phridge/issues/41
            it("should not schedule a new ping when a pong message is received right after calling dispose()", function () {
                var message = JSON.stringify({ status: "pong" });
                var promise = phantom.dispose();

                // Simulate a pong message from PhantomJS
                phantom._receive(message);

                return promise;
            });

        });

    });

});

// This last test checks for the presence of PhantomJS zombies that might have been spawned during tests.
// We don't want phridge to leave zombies at all circumstances.
after(slow(function (done) {
    setTimeout(function () {
        ps.lookup({
            command: "phantomjs"
        }, function onLookUp(err, phantomJsProcesses) {
            if (err) {
                throw new Error(err);
            }
            if (phantomJsProcesses.length > 0) {
                throw new Error("PhantomJS zombies detected");
            }
            done();
        });
    }, 2000);
}));
