var path = require('path');

/**
 * Creates a string that can be prepended to a file path that needs to resolve to the root.
 * @param {string} page - Path to the page.
 * @param {string} root - Path to the page root.
 * @returns {string} An empty string if the page is at the root, or a series of `../` characters if the page is in a subdirectory relative to the root.
 */
module.exports = function(page, root) {
  var pagePath = path.dirname(page);
  var rootPath = path.join(process.cwd(), root);

  var relativePath = path.relative(pagePath, rootPath);

  if (relativePath.length > 0) {
    relativePath += '/';
  }

  // On Windows, Node uses a "\" for the separator
  // However, web browsers use "/" no matter the platform
  return relativePath.replace('\\', '/');
}
