"use strict";

function noop() {}

function childProcessMock() {
    return {
        on: noop,
        removeListener: noop,
        phridge: {
            on: noop
        },
        stdin: {
            write: noop,
            on: noop
        },
        stdout: {
            on: noop
        },
        stderr: {
            on: noop
        }
    };
}

module.exports = childProcessMock;