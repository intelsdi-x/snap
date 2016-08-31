/*!
 * rewrite-ext <https://github.com/jonschlinkert/rewrite-ext>
 *
 * Copyright (c) 2014-2015, Jon Schlinkert.
 * Licensed under the MIT License.
 */

'use strict';

var path = require('path');
var exts = require('ext-map');

module.exports = function rewrite(fp, ext) {
  var extname = path.extname(fp);
  var len = extname.length;

  ext = ext || exts[extname] || extname;
  if (ext.charAt(0) !== '.') {
    ext = '.' + ext;
  }

  return fp.slice(0, fp.length - len) + ext;
};
