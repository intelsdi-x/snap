var fs    = require('fs');
var path  = require('path');
var utils = require('./utils');
var yaml  = require('js-yaml');

/**
 * Looks for files with .json or .yml extensions within the given directory, and adds them as Handlebars variables matching the name of the file.
 * @param {string} dir - Folder to check for data files.
 */
module.exports = function(dir) {
  var dataFiles = utils.loadFiles(dir, '*.{json,yml}');

  for (var i in dataFiles) {
    var file = fs.readFileSync(dataFiles[i]);
    var ext = path.extname(dataFiles[i]);
    var name = path.basename(dataFiles[i], ext);
    var data;

    if (ext === '.json') {
      data = require(dataFiles[i])
    }
    else if (ext === '.yml') {
      data = yaml.safeLoad(fs.readFileSync(dataFiles[i]));
    }

    this.data[name] = data;
  }
}
