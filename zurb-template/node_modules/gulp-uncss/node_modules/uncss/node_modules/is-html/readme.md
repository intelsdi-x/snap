# is-html [![Build Status](https://travis-ci.org/sindresorhus/is-html.svg?branch=master)](https://travis-ci.org/sindresorhus/is-html)

> Check if a string is HTML


## Install

```sh
$ npm install --save is-html
```


## Usage

```js
isHtml('<p>I am HTML</p>');
//=> true

isHtml('<!doctype><html><body><h1>I ❤ unicorns</h1></body></html>');
//=> true

isHtml('<cake>I am XML</cake>');
//=> false

isHtml('>+++++++>++++++++++>+++>+<<<<-');
//=> false
```


## License

MIT © [Sindre Sorhus](http://sindresorhus.com)
