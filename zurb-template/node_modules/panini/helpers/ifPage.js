/**
 * Generates a Handlebars block helper called #ifpage for use in templates. This helper must be re-generated for every page that's rendered, because the return value of the function is dependent on the name of the current page.
 * @param {string} pageName - Name of the page to use in the helper function.
 * @returns {function} A Handlebars helper function.
 */
module.exports = function(pageName) {
  /**
   * Handlebars block helper that renders the content inside of it based on the current page.
   * @param {string...} pages - One or more pages to check.
   * @param (object) options - Handlebars object.
   * @example
   * {{#ifpage 'index', 'about'}}This must be the index or about page.{{/ifpage}}
   * @return The content inside the helper if a page matches, or an empty string if not.
   */
  return function() {
    var params = Array.prototype.slice.call(arguments);
    var pages = params.slice(0, -1);
    var options = params[params.length - 1];

    for (var i in pages) {
      if (pages[i] === pageName) {
        return options.fn(this);
      }
    }

    return '';
  }
}