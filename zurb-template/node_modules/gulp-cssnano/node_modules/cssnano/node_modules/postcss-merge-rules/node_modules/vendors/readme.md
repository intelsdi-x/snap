# vendors [![Build Status][build-badge]][build-page] [![Coverage Status][coverage-badge]][coverage-page]

<!--lint ignore no-html-->

List of (real<sup>†</sup>) vendor prefixes known to the web platform.
From [Wikipedia][wiki] and the [CSS 2.1 spec][spec].

† — real, as in, `mdo-` and `prince-` are not included because they are
not [valid][spec].

## Installation

[npm][]:

```bash
npm install vendors
```

**vendors** is also available as an AMD, CommonJS, and globals module,
[uncompressed and compressed][releases].

## Usage

Dependencies:

```javascript
var vendors = require('vendors');
```

Yields:

```js
[ 'ah',
  'apple',
  'atsc',
  'epub',
  'hp',
  'khtml',
  'moz',
  'ms',
  'o',
  'rim',
  'ro',
  'tc',
  'wap',
  'webkit',
  'xv' ]
```

## API

### `vendors`

`Array.<string>` — List of lower-case prefixes without dashes.

## License

[MIT][license] © [Titus Wormer][author]

<!-- Definition -->

[build-badge]: https://img.shields.io/travis/wooorm/vendors.svg

[build-page]: https://travis-ci.org/wooorm/vendors

[coverage-badge]: https://img.shields.io/codecov/c/github/wooorm/vendors.svg

[coverage-page]: https://codecov.io/github/wooorm/vendors?branch=master

[npm]: https://docs.npmjs.com/cli/install

[releases]: https://github.com/wooorm/vendors/releases

[license]: LICENSE

[author]: http://wooorm.com

[wiki]: https://en.wikipedia.org/wiki/CSS_filter#Prefix_filters

[spec]: https://www.w3.org/TR/CSS21/syndata.html#vendor-keyword-history
