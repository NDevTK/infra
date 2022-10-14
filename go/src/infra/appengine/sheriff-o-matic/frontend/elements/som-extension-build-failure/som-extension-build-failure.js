'use strict';

const codeSearchURL = 'https://cs.chromium.org/';
const testResultsURL = 'https://test-results.appspot.com/';

class SomExtensionBuildFailure extends Polymer.Element {
  static get is() {
    return 'som-extension-build-failure';
  }

  static get properties() {
    return {
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
      bugs: Array,
      _culprits: {
        type: Array,
        computed: '_computeCulprits(extension)',
      },
    };
  }

  _extensionChanged() {
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
  }

  _isChromium(tree) {
    return tree == 'chromium';
  }

  _haveBuilders(extension) {
    return extension && extension.builders && extension.builders.length > 0;
  }

  _failure_bbid(extension) {
    if (!this._haveBuilders(extension)) {
      return ''
    }
    // We cannot use the latest_failure directly due to crbug.com/1366166
    // So we need to work around
    const url = extension.builders[0].latest_failure_url
    if (!url) {
      return ""
    }
    // url is of the form https://ci.chromium.org/.../b<bbid>
    return url.substring(url.lastIndexOf("/b") + 2)
  }

  _failureCount(builder) {
    // The build number range is inclusive.
    return builder.latest_failure_build_number
        - builder.first_failure_build_number + 1;
  }

  _failureCountText(builder) {
    const numBuilds = this._failureCount(builder);

    // first_failure_build_number == 0 means we do not have information about
    // the first failure. In this case, we do not want to display anything.
    if (numBuilds == 1 || builder.first_failure_build_number == 0) {
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
  }

  _classForBuilder(builder) {
    let classes = ['builder'];
    if (this._failureCount(builder) > 1) {
      classes.push('multiple-failures');
    }
    if (this.type == 'infra-failure'
        || builder.build_status === "INFRA_FAILURE") {
      classes.push('infra-failure');
    }
    return classes.join(' ');
  }

  _displayName(builder) {
    if (this.tree === 'chrome_browser_release') {
      if (builder.bucket != 'ci') { // For M85 and before
        return builder.bucket + '.' + builder.name;
      } else { // For M86 onwards
        return builder.project + '.' + builder.name;
      }
    }
    return builder.name;
  }

  _haveRegressionRanges(regression_ranges) {
    return regression_ranges && regression_ranges.length > 0;
  }

  _linkForCL(cl) {
    return 'https://crrev.com/' + cl;
  }

  _showRegressionRange(range) {
    return !!range;
  }

  _textForCL(commit_position, revision) {
    if (commit_position == null) {
      return revision.substring(0, 7);
    }
    return commit_position;
  }
}

customElements.define(SomExtensionBuildFailure.is, SomExtensionBuildFailure);
