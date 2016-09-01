"use strict";

/* global pages, config */

var chai = require("chai");
var Page = require("../lib/Page.js");
var expect = chai.expect;
var phridge = require("../lib/main.js");
var slow = require("./helpers/slow.js");

require("./helpers/setup.js");

describe("Page", function () {

    describe(".prototype", function () {
        var phantom;
        var page;

        function createPage() {
            page = phantom.createPage();
        }

        function disposePage() {
            page.dispose();
        }

        before(slow(function () {
            return phridge.spawn({}).then(function (newPhantom) {
                phantom = newPhantom;
            });
        }));

        after(slow(function () {
            return phantom.dispose();
        }));

        describe(".constructor(phantom, id)", function () {

            it("should return an instance of Page with the given arguments applied", function () {
                page = new Page(phantom, 1);
                expect(page).to.be.an.instanceOf(Page);
                expect(page.phantom).to.equal(phantom);
                expect(page._id).to.equal(1);
            });

        });

        describe(".phantom", function () {

            it("should be null by default", function () {
                expect(Page.prototype.phantom).to.equal(null);
            });

        });

        describe(".run(arg1, arg2, arg3, fn)", function () {

            before(createPage);
            after(disposePage);

            describe("with fn being an asynchronous function", function () {

                it("should provide a resolve function", function () {
                    return expect(page.run(function (resolve) {
                        resolve("everything ok");
                    })).to.eventually.equal("everything ok");
                });

                it("should provide the possibility to resolve with any stringify-able data", function () {
                    return Promise.all([
                        expect(page.run(function (resolve) {
                            resolve();
                        })).to.eventually.equal(undefined),
                        expect(page.run(function (resolve) {
                            resolve(true);
                        })).to.eventually.equal(true),
                        expect(page.run(function (resolve) {
                            resolve(2);
                        })).to.eventually.equal(2),
                        expect(page.run(function (resolve) {
                            resolve(null);
                        })).to.eventually.equal(null),
                        expect(page.run(function (resolve) {
                            resolve([1, 2, 3]);
                        })).to.eventually.deep.equal([1, 2, 3]),
                        expect(page.run(function (resolve) {
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
                    return page.run(function (resolve, reject) {
                        reject(new Error("not ok"));
                    }).catch(function (err) {
                        expect(err.message).to.equal("not ok");
                    });
                });

            });

            describe("with fn being a synchronous function", function () {

                it("should resolve to the returned value", function () {
                    return expect(page.run(function () {
                        return "everything ok";
                    })).to.eventually.equal("everything ok");
                });

                it("should provide the possibility to resolve with any stringify-able data", function () {
                    return Promise.all([
                        expect(page.run(function () {
                            // returns undefined
                        })).to.eventually.equal(undefined),
                        expect(page.run(function () {
                            return true;
                        })).to.eventually.equal(true),
                        expect(page.run(function () {
                            return 2;
                        })).to.eventually.equal(2),
                        expect(page.run(function () {
                            return null;
                        })).to.eventually.equal(null),
                        expect(page.run(function () {
                            return [1, 2, 3];
                        })).to.eventually.deep.equal([1, 2, 3]),
                        expect(page.run(function () {
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
                    return page.run(function () {
                        throw new Error("not ok");
                    }).catch(function (err) {
                        expect(err.message).to.equal("not ok");
                    });
                });

            });

            it("should provide all phantomjs default modules as convenience", function () {
                return expect(page.run(function () {
                    return Boolean(webpage && system && fs && webserver && child_process); // eslint-disable-line
                })).to.eventually.equal(true);
            });

            it("should provide the config object to store all kind of configuration", function () {
                return expect(page.run(function () {
                    return config;
                })).to.eventually.deep.equal({});
            });

            it("should provide the possibility to pass params", function () {
                var params = {
                    some: ["param"],
                    withSome: "crazy",
                    values: {
                        number1: 1
                    }
                };

                return expect(page.run(params, params, params, function (params1, params2, params3) {
                    return [params1, params2, params3];
                })).to.eventually.deep.equal([params, params, params]);
            });

            it("should report errors", function () {
                return expect(page.run(function () {
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
                            //console.log(err.stack);
                        })
                ]);
            });

            it("should run the function with the page as context", function () {
                return page.run(/** @this WebPage */function () {
                    if (!this.clipRect) {
                        throw new Error("The function's context is not the web page");
                    }
                });
            });

        });

        describe(".dispose()", function () {

            beforeEach(createPage);

            it("should remove the page from the pages-object", function () {
                var pageId = page._id;

                function checkForPage(pageId) {
                    if (pages[pageId]) {
                        throw new Error("page is still present in the page-object");
                    }
                }

                return page.dispose().then(function () {
                    return phantom.run(pageId, checkForPage);
                });
            });

            it("should remove the phantom reference", function () {
                return page.dispose().then(function () {
                    expect(page.phantom).to.equal(null);
                });
            });

        });

    });

});

