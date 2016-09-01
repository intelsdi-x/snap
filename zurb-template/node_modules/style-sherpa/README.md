# Style Sherpa

Style Sherpa is a simple style guide generator. It takes a single Markdown file and converts it to a pre-made HTML template with tabbed sections. The template is powered by [Foundation for Sites](http://foundation.zurb.com).

## Installation

```bash
npm install style-sherpa --save-dev
```

## Usage

Your style guide lives in a single Markdown file, which can have any name and sit anywhere in your project. Use first-level headings (a single `#` before a title) to mark new sections in the style guide.

```markdown
# Section 1

Lorem ipsum dolor sit amet, consectetur adipisicing elit. Fuga saepe, vero ratione optio illum aliquam. Sint esse velit est voluptatum. Ipsa tempora saepe nostrum quidem voluptatem esse voluptatum quibusdam laboriosam!
```

To create new sections, add four line breaks.

```markdown
# Section 1

Lorem ipsum dolor sit amet, consectetur adipisicing elit. Fuga saepe, vero ratione optio illum aliquam. Sint esse velit est voluptatum. Ipsa tempora saepe nostrum quidem voluptatem esse voluptatum quibusdam laboriosam!



# Section 2

<!-- ... -->
```

To actually run the parser, include the `style-sherpa` library and run the command. At minimum you need file paths for the input and output, but you can also optionally specify a custom template, or supply a callback.

```javascript
var sherpa = require('style-sherpa');

sherpa('./test/fixtures/test.md', {
  output: './test/fixtures/test.html',
  template: './template.hbs'
}, cb());
```

### sherpa(input [, options, callback])

Generates an HTML style guide from a Markdown file.

#### input

**Type:** `String`

Path to the input Markdown file to parse.

#### options

**Type:** `Object`

- `output` (`String`): Path to output the finished HTML file. Defaults to the current working directory.
- `template` (`String`): Path to a custom Handlebars template to use, instead of the default one.

### callback

**Type:** `Function`

Callback to run when the file has been written to disk.
