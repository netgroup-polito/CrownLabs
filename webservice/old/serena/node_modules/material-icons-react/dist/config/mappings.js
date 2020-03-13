(function (global, factory) {
    if (typeof define === "function" && define.amd) {
        define(['exports'], factory);
    } else if (typeof exports !== "undefined") {
        factory(exports);
    } else {
        var mod = {
            exports: {}
        };
        factory(mod.exports);
        global.mappings = mod.exports;
    }
})(this, function (exports) {
    'use strict';

    Object.defineProperty(exports, "__esModule", {
        value: true
    });
    var sizes = exports.sizes = {
        tiny: 'md-18',
        small: 'md-24',
        medium: 'md-36',
        large: 'md-48'
    };
    var light = exports.light = 'md-light';
    var dark = exports.dark = 'md-dark';
    var mdInactive = exports.mdInactive = 'md-inactive';
});