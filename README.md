
# snap landing page


<img src="https://cloud.githubusercontent.com/assets/1744971/16677455/f1d4e9de-448a-11e6-9afb-c31dcc7e3274.png" width="50%">

visit us at: http://snap-telemetry.io

##Contributing:
Contributions are always welcome and accepted via Github pull requests.

##to download repo:
* have npm installed ([install directions](https://docs.npmjs.com/getting-started/installing-node))
* install bower (`npm install bower`)
* `git clone` repo

##to test & edit:
* view changes as you edit with: `npm start` (inside zurb-template/ folder via [BrowserSync](https://www.browsersync.io/))
* compile a production build with: `npm run build` (inside zurb-template/ folder)
    * this will populate the zurb-template/dist/ folder with finished website files
    * you should copy over these files to the top layer of your repo (same level as zurb-template/ folder)


##troubleshooting:
* if you run into errors, see [troubleshooting](troubleshooting.md)


##Design Notes:
* npm version 2.15.8
* bower version 1.7.9
* uses front end framework: [ZURB foundation 6](http://foundation.zurb.com/)
* uses the [ZURB template](http://foundation.zurb.com/sites/docs/starter-projects.html#zurb-template)
* uses [Panini Library](http://foundation.zurb.com/sites/docs/panini.html) (A flat file compiler that powers the ZURB prototyping template)
    * check out the [Panini helper functions](http://foundation.zurb.com/sites/docs/panini.html#helpers) to help manipulate page content
        * includes [markdown function](http://foundation.zurb.com/sites/docs/panini.html#markdown), so you can write your page content in markdown
* uses iconic font & css toolkit: [Font Awesome](http://fontawesome.io/) (used for github and slack links)

* zurb_template/ folder contains all ZURB Template files 
    * includes dist/ folder which contains finished website files 
        * populates when you run `npm run build`
        * you should copy these files over to the top layer of your repo (same level as zurb-template/ folder)
* edits and additions to default foundation css and js can be found in zurb-template/src/assets/changes
    * can also edit default foundation css and js found in zurb-template/bower_components/foundation-sites/scss and  zurb-template/bower_components/foundation-sites/js
* images and graphics can be found in zurb-template/src/assets/img
* font awesome files in zurb-template/src/assets/design
* overall page template used for every page on site
    * file in zurb-template/src/layouts/default.html
    * default.html pg template includes links to css and js file updates and google analytics code
* each page of site has a separate file that includes the main content/body of the html page
    * files in zurb-template/src/pages/
    * e.g. download.html, index.html, plugins.html
    * you can add a new page here and update the page navigation header bar in zurb-template/src/partials/
* partials/ folder includes separate files for page navigation and footer.

* more details for panini file formatting: zurb-template/node_modules/panini/readme.md
* buttons use intel blue #0071C5
* Snap font: London Between
* file structure:
```
|__zurb-template/
|   |__dist/
|   |__src/
|       |__assets/
|       |   |__img/
|       |   |__design/
|       |   |__changes/
|       |__layouts/
|       |   |__default.html
|       |
|       |__pages/
|       |   |__download.html
|       |   |__index.html
|       |   |__plugins.html
|       |
|       |__partials/
|           |__footer.html
|           |__navigation.html
|
|__index.html
|__download.html
|__plugins.html
|__assets/
|   |__changes/
|   |__img/
|   |__
|
|__README.md
```
