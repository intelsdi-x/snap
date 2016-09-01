/**
 * Handlebars block helper that repeats the content inside of it n number of times.
 * @param {integer} count - Number of times to repeat.
 * @param {object} options - Handlebars object.
 * @example
 * {{#repeat 5}}<li>List item!</li>{{/repeat}}
 * @returns The content inside of the helper, repeated n times.
 */
module.exports = function(count, options) {
  var str = '';

  for (var i = 0; i < count; i++) {
    str += options.fn(this);
  }

  return str;
}
