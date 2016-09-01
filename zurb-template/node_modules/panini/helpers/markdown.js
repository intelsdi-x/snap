var marked = require('marked');

/**
 * Handlebars block helper that converts Markdown to HTML.
 * @param {object} options - Handlebars object.
 * @example
 * {{#markdown}}Welcome to [zombo.com](http://zombo.com){{/markdown}}
 * @returns The Markdown inside the helper, converted to HTML.
 */
module.exports = function(options) {
  return marked(options.fn(this));
}
