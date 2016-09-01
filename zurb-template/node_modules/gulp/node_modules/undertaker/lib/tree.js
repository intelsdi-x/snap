'use strict';

var defaults = require('lodash.defaults');
var map = require('lodash.map');

var metadata = require('./helpers/metadata');

function tree(opts) {
  opts = defaults(opts || {}, {
    deep: false,
  });

  var tasks = this._registry.tasks();
  var nodes = map(tasks, function(task) {
    var meta = metadata.get(task);

    if (opts.deep) {
      return meta.tree;
    }

    return meta.tree.label;
  });

  return {
    label: 'Tasks',
    nodes: nodes,
  };
}

module.exports = tree;
