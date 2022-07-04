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

  // This is necessary because FindIt sometimes returns duplicate results
  _computeSuspectedCls(extension) {
    if (!this._haveSuspectCLs(extension)) {
      return [];
    }
    let revisions = {};
    for (var i in extension.suspected_cls) {
      revisions[extension.suspected_cls[i].revision] =
          extension.suspected_cls[i];
    }
    return Object.values(revisions);
  }

  _computeCulprits(extension) {
    if (!this._haveCulprits(extension)) {
      return [];
    }
    const culprits = {};
    for (const culprit of extension.culprits) {
      culprits[culprit.commit.id] = culprit;
    }
    return Object.values(culprits);
  }

  _finditIsRunning(extension) {
    return extension && !extension.suspected_cls && !extension.culprits &&
           !extension.is_finished && !extension.has_findings &&
           extension.is_supported;
  }

  _finditHasNoResult(extension) {
    return extension && !extension.suspected_cls && !extension.culprits &&
           extension.is_finished && !extension.has_findings;
  }

  _finditFoundNoResult(extension) {
    return this._finditHasNoResult(extension) && extension.is_supported;
  }

  _finditNotSupport(extension) {
    return this._finditHasNoResult(extension) && !extension.is_supported;
  }

  _finditHasUrl(extension) {
    return extension && extension.findit_url;
  }

  _finditApproach(cl) {
    if (cl.analysis_approach == 'HEURISTIC') {
      return ' suspects CL ';
    } else {
      return ' found culprit ';
    }
  }

  _finditConfidence(cl) {
    return cl.confidence.toString();
  }

  _haveSuspectCLs(extension) {
    return extension && extension.suspected_cls;
  }

  _haveRevertCL(cl) {
    return cl && cl.revert_cl_url;
  }

  _revertIsCommitted(cl) {
    return this._haveRevertCL(cl) && cl.revert_committed;
  }

  _haveCulprits(extension) {
    return extension && extension.culprits;
  }

  _linkForCulprit(culprit) {
    return 'https://' + culprit.commit.host + '/' + culprit.commit.project + '/+/' + culprit.commit.id;
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
