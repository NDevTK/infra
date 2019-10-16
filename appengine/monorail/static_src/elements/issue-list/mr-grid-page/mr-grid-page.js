// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// TODO(juliacordero): Handle pRPC errors with a FE page

import {LitElement, html, css} from 'lit-element';
import {store, connectStore} from 'reducers/base.js';
import * as issue from 'reducers/issue.js';
import 'elements/framework/links/mr-issue-link/mr-issue-link.js';
import './mr-grid-controls.js';
import './mr-grid.js';

export class MrGridPage extends connectStore(LitElement) {
  /** @override */
  render() {
    const displayedProgress = this.progress || 0.02;
    const doneLoading = this.progress === 1;
    const noMatches = this.totalIssues === 0 && doneLoading;
    return html`
      <div id="grid-area">
        <mr-grid-controls
          .projectName=${this.projectName}
          .queryParams=${this.queryParams}
          .issueCount=${this.issues.length}>
        </mr-grid-controls>
        ${noMatches ? html`
          <div class="empty-search">
            Your search did not generate any results.
          </div>` : html`
          <progress
            title="${Math.round(displayedProgress * 100)}%"
            value=${displayedProgress}
            ?hidden=${doneLoading}
          ></progress>`}
        <br>
        <mr-grid
          .issues=${this.issues}
          .xAttr=${this.queryParams.x}
          .yAttr=${this.queryParams.y}
          .cellMode=${this.queryParams.cells ? this.queryParams.cells : 'tiles'}
          .queryParams=${this.queryParams}
          .projectName=${this.projectName}
        ></mr-grid>
      </div>
    `;
  }

  /** @override */
  static get properties() {
    return {
      projectName: {type: String},
      issueEntryUrl: {type: String},
      queryParams: {type: Object},
      userDisplayName: {type: String},
      issues: {type: Array},
      fields: {type: Array},
      progress: {type: Number},
      totalIssues: {type: Number},
    };
  };

  /** @override */
  constructor() {
    super();
    this.issues = [];
    this.progress = 0;
    this.queryParams = {};
  };

  /** @override */
  updated(changedProperties) {
    if (changedProperties.has('userDisplayName')) {
      store.dispatch(issue.fetchStarredIssues());
    }
    // TODO(zosha): Abort sets of calls to ListIssues when
    // queryParams.q is changed.
    if (changedProperties.has('projectName')) {
      this._fetchMatchingIssues();
    } else if (changedProperties.has('queryParams')) {
      const oldParams = changedProperties.get('queryParams');
      const oldQ = oldParams ? oldParams.q : '';
      const newQ = this.queryParams.q;
      if (oldQ !== newQ) {
        this._fetchMatchingIssues();
      }
    }
  }

  _fetchMatchingIssues() {
    store.dispatch(issue.fetchIssueList(this.queryParams,
        this.projectName, {maxItems: 500}, 12));
  }

  /** @override */
  stateChanged(state) {
    this.issues = (issue.issueList(state) || []);
    this.progress = (issue.issueListProgress(state) || 0);
    this.totalIssues = (issue.totalIssues(state) || 0);
  }

  /** @override */
  static get styles() {
    return css `
      progress {
        background-color: white;
        border: 1px solid var(--chops-gray-500);
        width: 40%;
        margin-left: 1%;
        margin-top: 0.5em;
        visibility: visible;
      }
      ::-webkit-progress-bar {
        background-color: white;
      }
      progress::-webkit-progress-value {
        transition: width 1s;
        background-color: var(--chops-blue-700);
      }
      .empty-search {
        text-align: center;
        padding-top: 2em;
      }
    `;
  }
};
customElements.define('mr-grid-page', MrGridPage);
