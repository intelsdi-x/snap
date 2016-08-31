var format = require('string-template');
var hljs = require('highlight.js');
var marked = require('marked');

var renderer = new marked.Renderer();

// Prevents Marked from adding an ID to headings
renderer.heading = function(text, level) {
  return format('<h{1}>{0}</h{1}>', [text, level]);
}

// Allows for the creation of HTML examples and live code in one snippet
renderer.code = function(code, language) {
  var extraOutput = '';

  if (typeof language === 'undefined') language = 'html';

  // If the language is *_example, live code will print out along with the sample
  if (language.match(/_example$/)) {
    extraOutput = format('\n\n<div class="ss-code-live">{0}</div>', [code]);
    language = language.replace(/_example$/, '');
  }

  var renderedCode = hljs.highlight(language, code).value;
  var output = format('<div class="ss-code"><pre><code class="{0}">{1}</code></pre></div>', [language, renderedCode]);

  return output + extraOutput;
}

module.exports = renderer;
