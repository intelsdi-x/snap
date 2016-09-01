var fs    = require('fs');
var path  = require('path');
var utils = require('./utils');

/**
 * Looks for files with .html, .hbs, or .handlebars extensions within the given directory, and adds them as Handlebars partials matching the name of the file.
 * @param {string} dir - Folder to check for partials.
 */
module.exports = function(dir) {
  var partials = utils.loadFiles(dir, '**/*.{html,hbs,handlebars}');

  for (var i in partials) {
    var ext = path.extname(partials[i]);
    var file = fs.readFileSync(partials[i]);
    var name = path.basename(partials[i], ext);
    this.Handlebars.registerPartial(name, file.toString() + '\n');
  }
}
