(function() {
  'use strict';

  const codeSearchURL = 'https://cs.chromium.org/';
  const testResultsURL = 'https://test-results.appspot.com/';

  Polymer({
    is: 'som-extension-build-failure',
    behaviors: [LinkifyBehavior, TreeBehavior],

    properties: {
      extension: {
        type: Object,
        value: function() {
          return {};
        },
        observer: '_extensionChanged',
      },
      type: {type: String, value: ''},
      _suspectedCls: {
        type: Array,
        computed: '_computeSuspectedCls(extension)',
      },
      tree: String,
    },

    _extensionChanged: function() {
      // De-dupe testnames. TODO: do this in the analyzer.
      if (!(this.extension && this.extension.reason &&
          this.extension.reason.test_names)) {
        return;
      }
      let testNames = this.extension.reason.test_names;
      this.extension.reason.test_names = Array.from(new Set(testNames));
      let tests = this.extension.reason.tests;
      if (!tests) {
        return;
      }
      let seen = new Map();
      this.extension.reason.tests = tests.filter((test) => {
        if (seen.has(test.test_name)) {
          return false;
        }
        seen.set(test.test_name, true);
        return true;
      });
    },

    _isChromium: function(tree) {
      return tree == 'chromium';
    },

    _haveBuilders: function(extension) {
      return extension && extension.builders && extension.builders.length > 0;
    },

    _failureCount: function(builder) {
      // The build number range is inclusive.
      return builder.latest_failure - builder.first_failure + 1;
    },

    _failureCountText: function(builder) {
      let numBuilds = this._failureCount(builder);
      if (numBuilds == 1) {
        return '';
      }

      if (builder.count) {
        return `[${builder.count} out of the last ${
                                                    numBuilds
                                                  } builds have failed]`;
      }

      if (numBuilds > 1) {
        return `[${numBuilds} since first detection]`;
      }
    },

    _classForBuilder: function(builder) {
      let classes = ['builder'];
      if (this._failureCount(builder) > 1) {
        classes.push('multiple-failures');
      }
      if (this.type == 'infra-failure') {
        classes.push('infra-failure');
      }
      return classes.join(' ');
    },

    // This is necessary because FindIt sometimes returns duplicate results
    _computeSuspectedCls: function(extension) {
      if (!this._haveSuspectCLs(extension)) {
        return [];
      }
      let revisions = {};
      for (var i in extension.suspected_cls) {
        revisions[extension.suspected_cls[i].revision] =
            extension.suspected_cls[i];
      }
      return Object.values(revisions);
    },

    _finditIsRunning: function(extension) {
      return extension && !extension.suspected_cls && !extension.is_finished &&
             !extension.has_findings && extension.is_supported;
    },

    _finditHasNoResult: function(extension) {
      return extension && !extension.suspected_cls && extension.is_finished &&
             !extension.has_findings;
    },

    _finditFoundNoResult: function(extension) {
      return this._finditHasNoResult(extension) && extension.is_supported;
    },

    _finditNotSupport: function(extension) {
      return this._finditHasNoResult(extension) && !extension.is_supported;
    },

    _finditHasUrl: function(extension) {
      return extension && extension.findit_url;
    },

    _finditApproach: function(cl) {
      if (cl.analysis_approach == 'HEURISTIC') {
        return ' suspects CL ';
      } else {
        return ' found culprit ';
      }
    },

    _finditConfidence: function(cl) {
      return cl.confidence.toString();
    },

    _haveSuspectCLs: function(extension) {
      return extension && extension.suspected_cls;
    },

    _haveRevertCL: function(cl) {
      return cl && cl.revert_cl_url;
    },

    _haveRegressionRanges: function(regression_ranges) {
      return regression_ranges && regression_ranges.length > 0;
    },

    _haveTests: function(tests) {
      return tests && tests.length > 0;
    },

    _isFlaky: function(test) {
      return test && test.is_flaky;
    },

    _linkForTest: function(reason, testName) {
      return testResultsURL + 'dashboards/' +
             'flakiness_dashboard.html#' +
             'tests=' + encodeURIComponent(testName) +
             '&testType=' + encodeURIComponent(reason.step);
    },

    _linkToCSForTest: function(testName) {
      let url = codeSearchURL + 'search/?q=';
      let query = testName;
      if (testName.includes('#')) {
        // Guessing that it's a java test; the format expected is
        // test.package.TestClass#testMethod. For now, just split around the #
        let split = testName.split('#');

        if (split.length > 2) {
          console.error('invalid java test name', testName);
        } else {
          query = split[0] + ' function:' + split[1];
        }
      }
      return url + encodeURIComponent(query);
    },

    _linkToEditForTest: function(testName) {
      return 'test-expectations/' + encodeURIComponent(testName);
    },

    _linkForCL: function(cl) {
      return 'https://crrev.com/' + cl;
    },

    _showRegressionRange: function(range) {
      return range.positions && range.positions.length > 0 &&
             range.repo != 'v8';
    },

    _sortTests: function(a, b) {
      return a.test_name.localeCompare(b.test_name);
    },

    _testText: function(tests) {
      // NOTE: This really shouldn't happen; we should only be calling this
      // function
      // when tests is actually defined. We are though, for some reason, and it
      // looks
      // like it might be some weird dom-repeat/Polymer bug. So check that tests
      // is ok here anyways.
      if (tests == null) {
        return '';
      }

      let len = tests.length;

      if (len == 1) {
        return '1 test failed';
      }
      return len.toString() + ' tests failed';
    },

    _textForCL: function(commit_position, revision) {
      if (commit_position == null) {
        return revision.substring(0, 7);
      }
      return commit_position;
    },

    _hasSuspect: function(test) {
      return test && test.suspected_cls;
    },

    _makeLogDiffUrl: function(master, name, buildNum1, buildNum2) {
      return '/chromium/logdiff/' + master + '/' + name + '/' + buildNum1 + '/' + buildNum2;
    },
  });
})();
