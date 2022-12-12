// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

process.env.CHROME_BIN = require('puppeteer').executablePath();

module.exports = function(config) {
  const isDebug = process.argv.some((arg) => arg === '--debug');
  const coverage = process.argv.some((arg) => arg === '--coverage');
  process.env.TZ = 'America/Los_Angeles';
  config.set({

    // base path that will be used to resolve all patterns (eg. files, exclude)
    basePath: '',


    client: {
      mocha: {
        reporter: 'html',
        ui: 'bdd',
        checkLeaks: true,
        globals: [
          // Used for submitting Issue Wizard Feedback.
          'userfeedback',
          'CS_env',
          // __tsMonClient probably shouldn't be allowed to
          // leak between tests, but locating the current source of these
          // leaks has proven quite difficult.
          '__tsMonClient',
          'ga',
          'Color',
          'Chart',
          // TODO(ehmaldonado): Remove once the old autocomplete code is
          // deprecated.
          'TKR_populateAutocomplete',
          // All of the below are necessary for loading gapi.js.
          'gapi',
          '__gapiLoadPromise',
          '___jsl',
          'osapi',
          'gadgets',
          'shindig',
          'googleapis',
          'iframer',
          'ToolbarApi',
          'iframes',
          'IframeBase',
          'Iframe',
          'IframeProxy',
          'IframeWindow',
          '__gapi_jstiming__',
        ],
        timeout: 5000,
      },
    },

    mochaReporter: {
      showDiff: true,
    },


    // frameworks to use
    // available frameworks: https://npmjs.org/browse/keyword/karma-adapter
    frameworks: ['mocha', 'sinon', 'webpack'],


    // list of files / patterns to load in the browser
    files: [
      'static_src/test/setup.js',
      'static_src/test/index.js',
    ],


    // list of files / patterns to exclude
    exclude: [
    ],


    // preprocess matching files before serving them to the browser
    // available preprocessors: https://npmjs.org/browse/keyword/karma-preprocessor
    preprocessors: {
      'static_src/test/setup.js': ['webpack', 'sourcemap'],
      'static_src/test/index.js': ['webpack', 'sourcemap'],
    },

    plugins: [
      'karma-chrome-launcher',
      'karma-coverage',
      'karma-mocha',
      'karma-mocha-reporter',
      'karma-sinon',
      'karma-sourcemap-loader',
      'karma-webpack',
      '@chopsui/karma-reporter',
    ],

    webpack: {
      // webpack configuration
      devtool: 'inline-source-map',
      mode: 'development',
      resolve: {
        modules: ['node_modules', 'static_src'],
      },
      module: {
        rules: [
          {
            test: /\.(ts|tsx)$/,
            exclude: /node_modules/,
            use: ['babel-loader'],
          },
          {
            test: /\.css$/i,
            use: [
              {loader: 'style-loader', options: {injectType: 'styleTag'}},
              {
                loader: 'css-loader',
                options: {
                  modules: true,
                  importLoaders: 1,
                },
              },
              'postcss-loader',
            ],
          },
        ],
      },
    },


    // test results reporter to use
    // possible values: 'dots', 'progress'
    // available reporters: https://npmjs.org/browse/keyword/karma-reporter
    reporters: ['mocha', 'chopsui-json'].concat(
      coverage ? ['coverage'] : []),


    // configure coverage reporter
    coverageReporter: {
      dir: 'coverage',
      reporters: [
        {type: 'lcovonly', subdir: '.'},
        {type: 'json', subdir: '.', file: 'coverage.json'},
        {type: 'html'},
        {type: 'text'},
      ],
    },

    chopsUiReporter: {
      stdout: false,
      buildNumber: String(new Date().getTime()),
      outputFile: 'full_results.json',
    },

    // web server port
    port: 9876,


    // enable / disable colors in the output (reporters and logs)
    colors: true,


    // level of logging
    // possible values: config.LOG_DISABLE || config.LOG_ERROR || config.LOG_WARN || config.LOG_INFO || config.LOG_DEBUG
    logLevel: config.LOG_INFO,


    // enable / disable watching file and executing tests whenever any file changes
    autoWatch: true,


    // start these browsers
    // available browser launchers: https://npmjs.org/browse/keyword/karma-launcher
    browsers: isDebug ? ['Chrome_latest'] : ['ChromeHeadless'],


    customLaunchers: {
      Chrome_latest: {
        base: 'Chrome',
        version: 'latest',
      },
    },


    // Continuous Integration mode
    // if true, Karma captures browsers, runs the tests and exits
    singleRun: isDebug ? false : true,

    // Concurrency level
    // how many browser should be started simultaneous
    concurrency: Infinity,
  });
};
