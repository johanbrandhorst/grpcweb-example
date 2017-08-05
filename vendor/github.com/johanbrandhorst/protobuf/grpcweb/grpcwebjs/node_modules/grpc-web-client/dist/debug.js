"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
function debug() {
    var args = [];
    for (var _i = 0; _i < arguments.length; _i++) {
        args[_i] = arguments[_i];
    }
    if (console.debug) {
        console.debug.apply(null, args);
    }
    else {
        console.log.apply(null, args);
    }
}
exports.debug = debug;
function debugBuffer(str, buffer) {
    var asArray = [];
    for (var i = 0; i < buffer.length; i++) {
        asArray.push(buffer[i]);
    }
    debug(str, asArray.join(","));
}
exports.debugBuffer = debugBuffer;
//# sourceMappingURL=debug.js.map