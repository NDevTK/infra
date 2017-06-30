(function() {
'use strict';

// Time, in milliseconds, between each refresh of data from the server.
const refreshDelayMs = 60 * 1000;

Polymer({
  is: 'som-alert-view',
  behaviors: [AnnotationManagerBehavior, AlertTypeBehavior],
  properties: {
    _activeRequests: {
      type: Number,
      value: 0,
    },
    _alerts: {
      type: Array,
      value: function() {
        return [];
      },
      computed: `_computeAlerts(_alertsData.*, annotations)`,
    },
    // Map of stream to data, timestamp of latest updated data.
    _alertsData: {
      type: Object,
      value: function() {
        return {};
      },
    },
    alertsTimes: {
      type: Object,
      value: function() {
        return {};
      },
      notify: true,
    },
    _alertStreams: {
      type: Array,
      computed: '_computeAlertStreams(tree)',
      observer: '_updateAlerts',
      value: function() {
        return [];
      },
    },
    annotations: {
      type: Object,
      value: function() {
        return {};
      },
    },
    _bugs: Array,
    _checkedAlerts: {
      type: Array,
      value: function() {
        return [];
      },
    },
    _currentAlertView: {
      type: String,
      computed: '_computeCurrentAlertView(_examinedAlert)',
      value: 'alertListPage',
    },
    _examinedAlert: {
      type: Object,
      computed: '_computeExaminedAlert(_alerts, examinedAlertKey)',
      value: function() {
        return {};
      },
    },
    examinedAlertKey: String,
    _fetchAlertsError: String,
    fetchingAlerts: {
      type: Boolean,
      computed: '_computeFetchingAlerts(_activeRequests)',
      notify: true,
    },
    _fetchedAlerts: {
      type: Boolean,
      value: false,
    },
    _hideJulie: {
      type: Boolean,
      computed:
          '_computeHideJulie(_alerts, _fetchedAlerts, fetchingAlerts, _fetchAlertsError, tree)',
      value: true,
    },
    _isTrooperPage: {
      type: Boolean,
      computed: '_computeIsTrooperPage(tree.name)',
      value: false,
    },
    _pageTitleCount: {
      type: Number,
      computed: '_computePageTitleCount(_alerts, _bugs, _isTrooperPage)',
      observer: '_pageTitleCountChanged',
    },
    _showSwarmingAlerts: {
      type: Boolean,
      computed: '_computeShowSwarmingAlerts(_swarmingAlerts, _isTrooperPage)',
    },
    _swarmingAlerts: {
      type: Object,
      value: function() {
        return {};
      },
    },
    _sections: {
      type: Object,
      value: {
        // The order the sections appear in the array is the order they
        // appear on the page.
        'default': ['notifications', 'bugQueue', 'alertsList'],
        'trooper': ['notifications', 'bugQueue', 'swarmingBots', 'alertsList']
      },
    },
    trees: {
      type: Object,
      value: function() {
        return {};
      },
    },
    tree: {
      type: Object,
      observer: '_treeChanged',
    },
    user: String,
    collapseByDefault: {
      type: Boolean,
      value: false,
    },
    linkStyle: String,
    xsrfToken: String,
  },

  created: function() {
    this.async(this._refreshAsync, refreshDelayMs);
  },

  ////////////////////// Refresh ///////////////////////////

  refresh: function() {
    this.$.annotations.fetch();
    this.$.bugQueue.refresh();
    this.$.masterRestarts.refresh();
    this.$.treeStatus.refresh();
    this._updateAlerts(this._alertStreams);
  },

  _refreshAsync: function() {
    this.refresh();
    this.async(this._refreshAsync, refreshDelayMs);
  },

  ////////////////////// Alerts and path ///////////////////////////

  _computeIsTrooperPage: function(treeName) {
    return treeName === 'trooper';
  },

  _pageTitleCountChanged: function(count) {
    if (count > 0) {
      document.title = '(' + count + ') Sheriff-o-Matic';
    } else {
      document.title = 'Sheriff-o-Matic';
    }
  },

  _computePageTitleCount: function(alerts, bugs, isTrooperPage) {
    if (isTrooperPage && bugs) {
      return bugs.length;
    } else if (!isTrooperPage && alerts) {
      return alerts.length;
    }
    return 0;
  },

  _computeShowSwarmingAlerts: function(swarming, isTrooperPage) {
    return isTrooperPage && swarming && (swarming.dead || swarming.quarantined);
  },

  _treeChanged: function(tree) {
    if (!tree)
      return;

    this._alertsData = {};
    this._fetchedAlerts = false;

    // Reorder sections on page based on per tree priorities.
    let sections = this._sections[tree.name] || this._sections.default;
    for (let i in sections) {
      this.$$('#' + sections[i]).style.order = i;
    }
  },

  _computeAlertStreams: function(tree) {
    if (!tree || !tree.name)
      return [];

    if (tree.alert_streams && tree.alert_streams.length > 0) {
      return tree.alert_streams
    }

    return [tree.name];
  },

  _computeCurrentAlertView: function(examinedAlert) {
    if (examinedAlert && examinedAlert.key) {
      return 'examineAlert';
    }
    return 'alertListPage';
  },

  _computeExaminedAlert: function(alerts, examinedAlertKey) {
    let examinedAlert = alerts.find((alert) => {
      return alert.key == examinedAlertKey;
    });
    // Two possibilities if examinedAlert is undefined:
    // 1. The alert key is bad.
    // 2. Alerts has not been ajaxed in yet.
    if (examinedAlert) {
      return examinedAlert;
    }
    return {};
  },

  _updateAlerts: function(alertStreams) {
    this._fetchAlertsError = '';
    if (alertStreams.length > 0) {
      this._fetchedAlerts = false;
      this._activeRequests += alertStreams.length;

      alertStreams.forEach((stream) => {
        let base = '/api/v1/alerts/';
        if (window.location.href.indexOf('useMilo') != -1) {
          base = base + 'milo.';
        }
        window.fetch(base + stream, {credentials: 'include'})
            .then(
                (response) => {
                  this._activeRequests -= 1;
                  if (this._activeRequests <= 0) {
                    this._fetchedAlerts = true;
                  }
                  if (response.status == 404) {
                    this._fetchAlertsError =
                        'Server responded with 404: ' + stream + ' not found. ';
                    return false;
                  }
                  if (!response.ok) {
                    this._fetchAlertsError = 'Server responded with ' +
                                             response.status + ': ' +
                                             response.statusText;
                    return false;
                  }
                  return response.json();
                },
                (error) => {
                  this._activeRequests -= 1;
                  this._fetchAlertsError =
                      'Could not connect to the server. ' + error;
                })
            .then((json) => {
              // Ignore old requests that finished after tree switch.
              if (!this._alertStreams.includes(stream))
                return;

              if (json) {
                this.set('_swarmingAlerts', json.swarming);
                this.set(['_alertsData', this._alertStreamVarName(stream)],
                         json.alerts);

                this.alertsTimes = {};
                this.set(['alertsTimes', this._alertStreamVarName(stream)],
                         json.timestamp);
              }
            });
      });
    }
  },

  _alertStreamVarName(stream) {
    return stream.replace('.', '_');
  },

  _computeFetchingAlerts: function(activeRequests) {
    return activeRequests !== 0;
  },

  _countBuilders: function(alert) {
    if (alert.grouped && alert.alerts) {
      let count = 0;
      for (let i in alert.alerts) {
        count += this._countBuilders(alert.alerts[i]);
      }
      return count;
    } else if (alert.extension && alert.extension.builders) {
      return alert.extension.builders.length;
    } else {
      return 1;
    }
  },

  // TODO(zhangtiff): Refactor this function.
  _computeAlerts: function(alertsData, annotations) {
    if (!alertsData || !alertsData.base) {
      return [];
    }
    alertsData = alertsData.base;

    let allAlerts = [];
    for (let tree in alertsData) {
      let alerts = alertsData[tree];
      if (!alerts) {
        continue;
      }

      let alertItems = [];
      let groups = {};
      for (let i in alerts) {
        let alert = alerts[i];
        let ann = this.computeAnnotation(annotations, alert);

        if (ann.groupID) {
          if (!(ann.groupID in groups)) {
            let group = {
              key: ann.groupID,
              title: ann.groupID,
              body: ann.groupID,
              severity: alert.severity,
              time: alert.time,
              start_time: alert.start_time,
              links: [],
              tags: [],
              type: alert.type,
              extension: {stages: [], builders: [], grouped: true},
              grouped: true,
              alerts: [],
            }

            // Group name is stored using the groupID annotation.
            let groupAnn = this.computeAnnotation(annotations, group);
            if (groupAnn.groupID) {
              group.title = groupAnn.groupID;
            }

            groups[ann.groupID] = group;
            alertItems.push(group);
          }
          let group = groups[ann.groupID];
          if (alert.severity < group.severity) {
            group.severity = alert.severity;
          }
          if (alert.time > group.time) {
            group.time = alert.time;
          }
          if (alert.start_time < group.start_time) {
            group.start_time = alert.start_time;
          }
          if (alert.links)
            group.links = group.links.concat(alert.links);
          if (alert.tags)
            group.tags = group.tags.concat(alert.tags);

          if (alert.extension) {
            this._mergeStages(group.extension.stages, alert.extension.stages,
                              alert.extension.builders);
            this._mergeBuilders(group.extension.builders,
                                alert.extension.builders,
                                alert.extension.stages);
          }
          group.alerts.push(alert);
        } else {
          // Ungrouped alert.
          alertItems.push(alert);
        }
      }
      allAlerts = allAlerts.concat(alertItems);
    }

    if (!allAlerts) {
      return [];
    }

    allAlerts.sort((a, b) => {
      let aAnn = this.computeAnnotation(annotations, a);
      let bAnn = this.computeAnnotation(annotations, b);

      let aHasBugs = aAnn.bugs && aAnn.bugs.length > 0;
      let bHasBugs = bAnn.bugs && bAnn.bugs.length > 0;

      let aBuilders = this._countBuilders(a);
      let bBuilders = this._countBuilders(b);

      let aHasSuspectedCLs = a.extension && a.extension.suspected_cls;
      let bHasSuspectedCLs = b.extension && b.extension.suspected_cls;
      let aHasFindings = a.extension && a.extension.has_findings;
      let bHasFindings = b.extension && b.extension.has_findings;

      if (a.severity != b.severity) {
        // Note: 3 is the severity number for Infra Failures.
        // We want these at the bottom of the severities for sheriffs.
        if (a.severity == AlertSeverity.InfraFailure) {
          return 1;
        } else if (b.severity == AlertSeverity.InfraFailure) {
          return -1;
        }

        // 7 is the severity for offline builders. Note that we want these to
        // appear above infra failures.
        if (a.severity == 7) {
          return 1;
        } else if (b.severity == 7) {
          return -1;
        }
        return a.severity - b.severity;
      }

      // TODO(davidriley): Handle groups.

      if (aAnn.snoozed == bAnn.snoozed && aHasBugs == bHasBugs) {
        // We want to show alerts with Findit results above.
        // Show alerts with revert CL from Findit first;
        // the alerts with suspected_cls;
        // then alerts with flaky tests;
        // then alerts with no Findit results.
        if (aHasSuspectedCLs && bHasSuspectedCLs) {
          for (let key in b.extension.suspected_cls) {
            if (b.extension.suspected_cls[key].reverting_cl_url) {
              return 1;
            }
          }
          return -1;
        } else if (aHasSuspectedCLs) {
          return -1;
        } else if (bHasSuspectedCLs) {
          return 1;
        } else if (aHasFindings) {
          return -1;
        } else if (bHasFindings) {
          return 1;
        }

        if (aBuilders < bBuilders) {
          return 1;
        }
        if (aBuilders > bBuilders) {
          return -1;
        }
        if (a.title < b.title) {
          return -1;
        }
        if (a.title > b.title) {
          return 1;
        }
        return 0;
      } else if (aAnn.snoozed == bAnn.snoozed) {
        return aHasBugs ? 1 : -1;
      }

      return aAnn.snoozed ? 1 : -1;
    });

    return allAlerts;
  },

  _mergeExtensions: function(extension) {
    if (!this._haveGrouped(extension)) {
      return extension;
    }

    // extension is a list of extensions.
    let mergedExtension = {stages: [], builders: []};
    for (let i in extension) {
      let subExtension = extension[i];
      this._mergeStages(mergedExtension.stages, subExtension.stages,
                        subExtension.builders);
      this._mergeBuilders(mergedExtension.builders, subExtension.builders,
                          subExtension.stages);
    }

    return mergedExtension;
  },

  _mergeStages: function(mergedStages, stages, builders) {
    for (let i in stages) {
      this._mergeStage(mergedStages, stages[i], builders);
    }
  },

  _mergeStage: function(mergedStages, stage, builders) {
    let merged = mergedStages.find((s) => {
      return s.name == stage.name;
    });

    if (!merged) {
      merged = {
        name: stage.name,
        status: stage.status,
        logs: [],
        links: [],
        notes: stage.notes,
        builders: [],
      };

      mergedStages.push(merged);
    }
    if (stage.status != merged.status && stage.status == 'failed') {
      merged.status = 'failed';
    }

    // Only keep notes that are in common between all builders.
    merged.notes = merged.notes.filter(function(n) {
      return stage.notes.indexOf(n) !== -1;
    });

    merged.builders = merged.builders.concat(builders);
  },

  _mergeBuilders: function(mergedBuilders, builders, stages) {
    for (let i in builders) {
      this._mergeBuilder(mergedBuilders, builders[i], stages);
    }
  },

  _mergeBuilder: function(mergedBuilders, builder, stages) {
    let merged = mergedBuilders.find((b) => {
      // TODO: In the future actually merge these into a single entry.
      return b.name == builder.name &&
             b.first_failure == builder.first_failure &&
             b.latest_failure == builder.latest_failure;
    });

    if (!merged) {
      merged = Object.assign({stages: []}, builder);
      mergedBuilders.push(merged);
    }

    merged.start_time = Math.min(merged.start_time, builder.start_time);
    merged.first_failure =
        Math.min(merged.first_failure, builder.first_failure);
    if (builder.latest_failure > merged.latest_failure) {
      merged.url = builder.url;
      merged.latest_failure = builder.latest_failure;
    }

    merged.stages = merged.stages.concat(stages);
  },

  _computeHideJulie: function(alerts, fetchedAlerts, fetchingAlerts,
                              fetchAlertsError, tree) {
    if (fetchingAlerts || !fetchedAlerts || !alerts ||
        fetchAlertsError !== '' || !tree) {
      return true;
    }
    return alerts.length > 0;
  },

  ////////////////////// Alert Categories ///////////////////////////

  _alertItemsWithCategory: function(alerts, category, isTrooperPage) {
    return alerts.filter(function(alert) {
      if (isTrooperPage) {
        return alert.tree == category;
      } else if (category == AlertSeverity.InfraFailure) {
        // Put trooperable alerts into "Infra failures" on sheriff views
        return this.isTrooperAlertType(alert.type) ||
               alert.severity == category;
      }
      return alert.severity == category;
    }, this);
  },

  _computeCategories: function(alerts, isTrooperPage) {
    let categories = [];
    alerts.forEach(function(alert) {
      let cat = alert.severity;
      if (isTrooperPage) {
        cat = alert.tree;
      } else if (this.isTrooperAlertType(alert.type)) {
        // When not on /trooper, collapse all of the trooper alerts into
        // the "Infra failures" category.
        cat = AlertSeverity.InfraFailure;
      }
      if (!categories.includes(cat)) {
        categories.push(cat);
      }
    }, this);

    return categories;
  },

  _getCategoryTitle: function(category, isTrooperPage, trees) {
    if (isTrooperPage) {
      if (category in trees) {
        category = trees[category].display_name;
      }
      return category + ' infra failures';
    }
    return {
      0: 'Tree closers',
      1: 'Stale masters',
      2: 'Probably hung builders',
      3: 'Infra failures',
      4: 'Consistent failures',
      5: 'New failures',
      6: 'Idle builders',
      7: 'Offline builders',
      // Chrome OS alerts
      1000: 'CQ failures',
      1001: 'PFQ failures',
      1002: 'Canary failures',
      1003: 'Release branch failures',
      1004: 'Chrome PFQ informational failures',
      1005: 'Chromium PFQ informational failures',
    }[category];
  },

  _isInfraFailuresSection: function(category, isTrooperPage) {
    return !isTrooperPage && category === AlertSeverity.InfraFailure;
  },

  ////////////////////// Annotations ///////////////////////////

  _computeGroupTargets: function(alert, alerts) {
    // Valid group targets:
    // - must be of same type
    // - must not be with itself
    // - must not consist of two groups
    return alerts.filter((a) => {
      return a.type == alert.type && a.key != alert.key &&
             (!alert.grouped || !a.grouped);
    });
  },

  _handleAnnotation: function(evt) {
    this.$.annotations.handleAnnotation(evt.target.get('alert'), evt.detail);
  },

  _handleComment: function(evt) {
    this.$.annotations.handleComment(evt.target.get('alert'));
  },

  _handleLinkBug: function(evt) {
    this.$.annotations.handleLinkBug([evt.target.get('alert')]);
  },

  _handleLinkBugBulk: function(evt) {
    this.$.annotations.handleLinkBug(this._checkedAlerts,
                                     this._uncheckAll.bind(this));
  },

  _handleRemoveBug: function(evt) {
    this.$.annotations.handleRemoveBug(evt.target.get('alert'), evt.detail);
  },

  _handleSnooze: function(evt) {
    this.$.annotations.handleSnooze([evt.target.get('alert')]);
  },

  _handleSnoozeBulk: function(evt) {
    this.$.annotations.handleSnooze(this._checkedAlerts,
                                    this._uncheckAll.bind(this));
  },

  _handleGroup: function(evt) {
    let alert = evt.target.get('alert');
    this.$.annotations.handleGroup(
        alert, this._computeGroupTargets(alert, this._alerts));
  },

  _handleGroupBulk: function(evt) {
    this.$.annotations.group(this._checkedAlerts);
    // Uncheck all alerts after a group... Otherwise, the checking behavior
    // is weird.
    this._uncheckAll();
  },

  _hasGroupAll: function(checkedAlerts) {
    // If more than two of the checked alerts are a group...
    let groups = checkedAlerts.filter((alert) => {
      return alert && alert.grouped;
    });
    return checkedAlerts.length > 1 && groups.length < 2;
  },

  _handleUngroup: function(evt) {
    this.$.annotations.handleUngroup(evt.target.get('alert'));
  },

  _handleResolve: function(evt) {
    let alert = evt.target.get('alert');
    let tree = evt.target.get('tree');
    if (alert.grouped) {
      this._resolveAlerts(tree, alert.alerts);
    } else {
      this._resolveAlerts(tree, [alert]);
    }
  },

  _resolveAlerts: function(tree, alerts) {
    let url = '/api/v1/resolve/' + encodeURIComponent(tree);
    let keys = alerts.map((a) => {
      return a.key;
    });
    let request = {
      'keys': keys,
      'resolved': true,
    };
    this.$.annotations.postJSON(url, request)
        .then(jsonParsePromise)
        .then(this._resolveResponse.bind(this));
  },

  _resolveResponse: function(response) {
    if (response.resolved) {
      let alerts = this._alertsData[response.tree];
      alerts = alerts.filter(function(alert) {
        return !response.keys.find((key) => {
          return alert.key == key;
        });
      });
      // Ensure that the modification is captured.
      this.set(['_alertsData', response.tree], alerts);
    }
    return response;
  },

  _handleChecked: function(evt) {
    let categoryElements = this.getElementsByClassName('alert-category');
    let checked = [];
    for (let i = 0; i < categoryElements.length; i++) {
      checked = checked.concat(categoryElements[i].checkedAlerts);
    }
    this._checkedAlerts = checked;
  },

  _uncheckAll: function(evt) {
    let categoryElements = this.getElementsByClassName('alert-category');
    for (let i = 0; i < categoryElements.length; i++) {
      categoryElements[i].uncheckAll();
    }
  },
});
})();
