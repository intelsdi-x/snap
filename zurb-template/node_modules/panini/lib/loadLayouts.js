var fs      = require('fs');
var path    = require('path');
var utils   = require('./utils');

/**
 * Looks for files with .html, .hbs, or .handlebars extensions within the given directory, and adds them as layout files to be used by pages.
 * @param {string} dir - Folder to check for layouts.
 */
module.exports = function(dir) {
  var layouts = utils.loadFiles(dir, '**/*.{html,hbs,handlebars}');

  for (var i in layouts) {
    var ext = path.extname(layouts[i]);
    var name = path.basename(layouts[i], ext);
    var file = fs.readFileSync(layouts[i]);
    this.layouts[name] = this.Handlebars.compile(file.toString());
  }
}
