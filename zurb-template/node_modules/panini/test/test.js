import { src, dest } from 'vinyl-fs';
import assert from 'assert';
import equal from 'assert-dir-equal';
import { Panini } from '..';

const FIXTURES = 'test/fixtures/';

describe('Panini', () => {
  it('builds a page with a default layout', done => {
    var p = new Panini({
      root: FIXTURES + 'basic/pages/',
      layouts: FIXTURES + 'basic/layouts'
    });

    p.refresh();

    src(FIXTURES + 'basic/pages/*')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'basic/build'))
      .on('finish', () => {
        equal(FIXTURES + 'basic/expected', FIXTURES + 'basic/build');
        done();
      });
  });

  it('builds a page with a custom layout', done => {
    var p = new Panini({
      root: FIXTURES + 'layouts/pages/',
      layouts: FIXTURES + 'layouts/layouts/'
    });

    p.refresh();

    src(FIXTURES + 'layouts/pages/*')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'layouts/build'))
      .on('finish', () => {
        equal(FIXTURES + 'layouts/expected', FIXTURES + 'layouts/build');
        done();
      });
  });

  it('builds a page with preset layouts by folder', done => {
    var p = new Panini({
      root: FIXTURES + 'page-layouts/pages/',
      layouts: FIXTURES + 'page-layouts/layouts/',
      pageLayouts: {
        'alternate': 'alternate'
      }
    });

    p.refresh();

    src(FIXTURES + 'page-layouts/pages/**/*.html')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'page-layouts/build'))
      .on('finish', () => {
        equal(FIXTURES + 'page-layouts/expected', FIXTURES + 'page-layouts/build');
        done();
      });
  });

  it('builds a page with custom partials', done => {
    var p = new Panini({
      root: FIXTURES + 'partials/pages/',
      layouts: FIXTURES + 'partials/layouts/',
      partials: FIXTURES + 'partials/partials/'
    });

    p.refresh();

    src(FIXTURES + 'partials/pages/*')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'partials/build'))
      .on('finish', () => {
        equal(FIXTURES + 'partials/expected', FIXTURES + 'partials/build');
        done();
      });
  });

  it('builds a page with custom data', done => {
    var p = new Panini({
      root: FIXTURES + 'data-page/pages/',
      layouts: FIXTURES + 'data-page/layouts/',
      partials: FIXTURES + 'data-page/partials/'
    });

    p.refresh();

    src(FIXTURES + 'data-page/pages/*')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'data-page/build'))
      .on('finish', () => {
        equal(FIXTURES + 'data-page/expected', FIXTURES + 'data-page/build');
        done();
      });
  });

  it('builds a page with custom helpers', done => {
    var p = new Panini({
      root: FIXTURES + 'helpers/pages/',
      layouts: FIXTURES + 'helpers/layouts/',
      helpers: FIXTURES + 'helpers/helpers/'
    });

    p.refresh();

    src(FIXTURES + 'helpers/pages/*')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'helpers/build'))
      .on('finish', () => {
        equal(FIXTURES + 'helpers/expected', FIXTURES + 'helpers/build');
        done();
      });
  });

  it('builds a page with external JSON data', done => {
    var p = new Panini({
      root: FIXTURES + 'data-json/pages/',
      layouts: FIXTURES + 'data-json/layouts/',
      data: FIXTURES + 'data-json/data/'
    });

    p.refresh();

    src(FIXTURES + 'data-json/pages/*')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'data-json/build'))
      .on('finish', () => {
        equal(FIXTURES + 'data-json/expected', FIXTURES + 'data-json/build');
        done();
      });
  });

  it('builds a page with an array of external JSON data', done => {
    var p = new Panini({
      root: FIXTURES + 'data-array/pages/',
      layouts: FIXTURES + 'data-array/layouts/',
      data: [FIXTURES + 'data-array/data/', FIXTURES + 'data-array/data-extra/']
    });

    p.refresh();

    src(FIXTURES + 'data-array/pages/*')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'data-array/build'))
      .on('finish', () => {
        equal(FIXTURES + 'data-array/expected', FIXTURES + 'data-array/build');
        done();
      });
  });

  it('builds a page with external YAML data', done => {
    var p = new Panini({
      root: FIXTURES + 'data-yaml/pages/',
      layouts: FIXTURES + 'data-yaml/layouts/',
      data: FIXTURES + 'data-yaml/data/'
    });

    p.refresh();

    src(FIXTURES + 'data-yaml/pages/*')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'data-yaml/build'))
      .on('finish', () => {
        equal(FIXTURES + 'data-yaml/expected', FIXTURES + 'data-yaml/build');
        done();
      });
  });
});

describe('Panini variables', () => {
  it('{{page}} variable that stores the current page', done => {
    var p = new Panini({
      root: FIXTURES + 'variable-page/pages/',
      layouts: FIXTURES + 'variable-page/layouts/',
    });

    p.refresh();

    src(FIXTURES + 'variable-page/pages/*')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'variable-page/build'))
      .on('finish', () => {
        equal(FIXTURES + 'variable-page/expected', FIXTURES + 'variable-page/build');
        done();
      });
  });

  it('{{layout}} variable that stores the current layout', done => {
    var p = new Panini({
      root: FIXTURES + 'variable-layout/pages/',
      layouts: FIXTURES + 'variable-layout/layouts/',
    });

    p.refresh();

    src(FIXTURES + 'variable-layout/pages/*')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'variable-layout/build'))
      .on('finish', () => {
        equal(FIXTURES + 'variable-layout/expected', FIXTURES + 'variable-layout/build');
        done();
      });
  });

  it('{{root}} variable that stores a relative path to the root folder', done => {
    var p = new Panini({
      root: FIXTURES + 'variable-root/pages/',
      layouts: FIXTURES + 'variable-root/layouts/',
      partials: FIXTURES + 'variable-root/partials/'
    });

    p.refresh();

    src(FIXTURES + 'variable-root/pages/**/*.html')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'variable-root/build'))
      .on('finish', () => {
        equal(FIXTURES + 'variable-root/expected', FIXTURES + 'variable-root/build');
        done();
      });
  });
});

describe('Panini helpers', () => {
  it('#code helper that renders code blocks', done => {
    var p = new Panini({
      root: FIXTURES + 'helper-code/pages/',
      layouts: FIXTURES + 'helper-code/layouts/',
    });

    p.loadBuiltinHelpers();
    p.refresh();

    src(FIXTURES + 'helper-code/pages/**/*.html')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'helper-code/build'))
      .on('finish', () => {
        equal(FIXTURES + 'helper-code/expected', FIXTURES + 'helper-code/build');
        done();
      });
  });

  it('#ifEqual helper that compares two values', done => {
    var p = new Panini({
      root: FIXTURES + 'helper-ifequal/pages/',
      layouts: FIXTURES + 'helper-ifequal/layouts/',
    });

    p.loadBuiltinHelpers();
    p.refresh();

    src(FIXTURES + 'helper-ifequal/pages/**/*.html')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'helper-ifequal/build'))
      .on('finish', () => {
        equal(FIXTURES + 'helper-ifequal/expected', FIXTURES + 'helper-ifequal/build');
        done();
      });
  });

  it('#ifpage helper that checks the current page', done => {
    var p = new Panini({
      root: FIXTURES + 'helper-ifpage/pages/',
      layouts: FIXTURES + 'helper-ifpage/layouts/',
    });

    p.loadBuiltinHelpers();
    p.refresh();

    src(FIXTURES + 'helper-ifpage/pages/**/*.html')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'helper-ifpage/build'))
      .on('finish', () => {
        equal(FIXTURES + 'helper-ifpage/expected', FIXTURES + 'helper-ifpage/build');
        done();
      });
  });

  it('#markdown helper that converts Markdown to HTML', done => {
    var p = new Panini({
      root: FIXTURES + 'helper-markdown/pages/',
      layouts: FIXTURES + 'helper-markdown/layouts/',
    });

    p.loadBuiltinHelpers();
    p.refresh();

    src(FIXTURES + 'helper-markdown/pages/**/*.html')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'helper-markdown/build'))
      .on('finish', () => {
        equal(FIXTURES + 'helper-markdown/expected', FIXTURES + 'helper-markdown/build');
        done();
      });
  });

  it('#repeat helper that prints content multiple times', done => {
    var p = new Panini({
      root: FIXTURES + 'helper-repeat/pages/',
      layouts: FIXTURES + 'helper-repeat/layouts/',
    });

    p.loadBuiltinHelpers();
    p.refresh();

    src(FIXTURES + 'helper-repeat/pages/**/*.html')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'helper-repeat/build'))
      .on('finish', () => {
        equal(FIXTURES + 'helper-repeat/expected', FIXTURES + 'helper-repeat/build');
        done();
      });
  });

  it('#unlesspage helper that checks the current page', done => {
    var p = new Panini({
      root: FIXTURES + 'helper-unlesspage/pages/',
      layouts: FIXTURES + 'helper-unlesspage/layouts/',
    });

    p.loadBuiltinHelpers();
    p.refresh();

    src(FIXTURES + 'helper-unlesspage/pages/**/*.html')
      .pipe(p.render())
      .pipe(dest(FIXTURES + 'helper-unlesspage/build'))
      .on('finish', () => {
        equal(FIXTURES + 'helper-unlesspage/expected', FIXTURES + 'helper-unlesspage/build');
        done();
      });
  });
});
