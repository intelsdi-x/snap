'use strict';
var htmlTags = require('html-tags');

module.exports = function (str) {
	if (/\s?<!doctype html>|(<html\b[^>]*>|<body\b[^>]*>|<x-[^>]+>)+/i.test(str)) {
		return true;
	}

	var re = new RegExp(htmlTags.map(function (el) {
		return '<' + el + '\\b[^>]*>';
	}).join('|'), 'i');

	if (re.test(str)) {
		return true;
	}

	return false;
};
