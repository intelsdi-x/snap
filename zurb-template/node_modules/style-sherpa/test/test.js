var assert = require('assert');
var sherpa = require('../index');

describe('Style Sherpa', function() {
  it('generates an HTML style guide from a Markdown file', function(done) {
    sherpa('./test/fixtures/test.md', {
      output: './test/fixtures/test.html',
      template: './template.hbs'
    }, done);
  });
});
