var format = require('util').format;
var hljs = require('highlight.js');

/**
 * Handlebars block helper that highlights code samples.
 * @param {string} language - Language of the code sample.
 * @param {object} options - Handlebars object.
 * @example
 * {{#code 'html'}}<a class="button">Button!</a>{{/code}}
 * @returns The HTML inside the helper, with highlight.js classes added.
 */
module.exports = function(language, options) {
  if (typeof language === 'undefined') language = 'html';

  var code = hljs.highlight(language, options.fn(this)).value;

  return format('<div class="code-example"><pre><code class="%s">%s</code></pre></div>', language, code);
}
