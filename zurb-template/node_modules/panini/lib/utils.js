'use strict';

var glob = require('glob');
var path = require('path');

/**
 * Load a set of files
 * @param  {string|array} dir
 * @param  {string}       pattern
 * @return {array}
 */
exports.loadFiles = function(dir, pattern) {
  var files = [];

  dir = !Array.isArray(dir) ? [dir] : dir;

  for (var i in dir) {
    files = files.concat(glob.sync(path.join(process.cwd(), dir[i], pattern)));
  }

  return files;
}