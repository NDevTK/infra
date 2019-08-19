// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {LitElement, html, css} from 'lit-element';
import page from 'page';
import qs from 'qs';
import {store, connectStore} from 'elements/reducers/base.js';
import * as issue from 'elements/reducers/issue.js';
import {prpcClient} from 'prpc-client-instance.js';
import 'elements/framework/mr-dropdown/mr-dropdown.js';
import 'elements/framework/mr-issue-list/mr-issue-list.js';
// eslint-disable-next-line max-len
import 'elements/issue-detail/dialogs/mr-update-issue-hotlists/mr-update-issue-hotlists.js';

const COLSPEC_DELIMITER_REGEX = /[\s\+]+/;
const SITEWIDE_DEFAULT_COLUMNS = ['ID', 'Type', 'Status',
  'Priority', 'Milestone', 'Owner', 'Summary'];

export class MrListPage extends connectStore(LitElement) {
  static get styles() {
    return css`
      :host {
        display: block;
        box-sizing: border-box;
        width: 100%;
        padding: 0.5em 8px;
        --monorail-action-bar-height: 24px;
      }
      .container-no-issues {
        width: 100%;
        padding: 0 8px;
        font-size: var(--chops-large-font-size);
      }
      .list-controls {
        width: 100%;
        box-sizing: border-box;
        display: flex;
        align-items: center;
        height: var(--monorail-action-bar-height);
        padding: 0 8px;
        margin-bottom: 0.5em;
      }
      .list-controls button {
        height: 100%;
        background: none;
        display: flex;
        align-items: center;
        justify-content: center;
        border: none;
        border-right: var(--chops-normal-border);
        cursor: pointer;
        transition: 0.2s background ease-in-out;
        color: var(--chops-link-color);
        padding: 0.1em 8px;
      }
      .list-controls button:hover {
        background: var(--chops-blue-50);
      }
    `;
  }

  render() {
    if (this.fetchingIssueList) {
      return html`
        <div class="container-no-issues">
          Loading...
        </div>
      `;
    }

    const selectedRefs = this.selectedIssues.map(
        ({localId, projectName}) => ({localId, projectName}));
    return html`
      <div class="list-controls">
        <button @click=${this.bulkEdit}>
          Bulk edit
        </button>
        <button @click=${this.addToHotlist}>
          Add to hotlist
        </button>
        <mr-dropdown
          icon="more_vert"
          menuAlignment="left"
          title="More actions..."
          .items=${this._moreActions}
        ></mr-dropdown>
      </div>
      <mr-issue-list
        .issues=${this.issues}
        .projectName=${this.projectName}
        .queryParams=${this.queryParams}
        .columns=${this.columns}
        selectionEnabled
        @selectionChange=${this._setSelectedIssues}
      ></mr-issue-list>
      <mr-update-issue-hotlists
        .issueRefs=${selectedRefs}
      ></mr-update-issue-hotlists>
    `;
  }

  static get properties() {
    return {
      issues: {type: Array},
      queryParams: {type: Object},
      projectName: {type: String},
      fetchingIssueList: {type: Boolean},
      selectedIssues: {type: Array},
      columns: {type: Array},
      userDisplayName: {type: String},
    };
  };

  constructor() {
    super();
    this.issues = [];
    this.fetchingIssueList = false;
    this.selectedIssues = [];

    this._boundRefresh = this.refresh.bind(this);

    this._moreActions = [
      {
        text: 'Flag as spam',
        handler: () => this._flagIssues(true),
      },
      {
        text: 'Un-flag as spam',
        handler: () => this._flagIssues(false),
      },
    ];

    // Expose page.js for test stubbing.
    this.page = page;
  };

  connectedCallback() {
    super.connectedCallback();

    window.addEventListener('refreshList', this._boundRefresh);
  }

  disconnectedCallback() {
    super.disconnectedCallback();

    window.removeEventListener('refreshList', this._boundRefresh);
  }

  update(changedProperties) {
    if (changedProperties.has('queryParams')) {
      if (this.queryParams && this.queryParams.colspec) {
        this.columns = this.queryParams.colspec.split(COLSPEC_DELIMITER_REGEX);
      } else {
        // TODO(zhangtiff): Change to project default columns.
        this.columns = SITEWIDE_DEFAULT_COLUMNS;
      }
    }

    super.update(changedProperties);
  }

  updated(changedProperties) {
    if (changedProperties.has('userDisplayName')) {
      store.dispatch(issue.fetchStarredIssues());
    }
    if (changedProperties.has('projectName') ||
        changedProperties.has('queryParams')) {
      this.refresh();
    }
  }

  refresh() {
    store.dispatch(issue.fetchIssueList(this.queryParams, this.projectName,
        {maxItems: 100, start: 0}));
  }

  stateChanged(state) {
    this.issues = (issue.issueList(state) || []);
    this.fetchingIssueList = issue.requests(state).fetchIssueList.requesting;
  }

  noneSelectedAlert(action) {
    // TODO(zhangtiff): Replace this with a modal for a more modern feel.
    alert(`Please select some issues to ${action}.`);
  }

  addToHotlist() {
    const issues = this.selectedIssues;
    if (!issues || !issues.length) {
      return this.noneSelectedAlert('add to hotlists');
    }
    this.shadowRoot.querySelector('mr-update-issue-hotlists').open();
  }

  bulkEdit() {
    const issues = this.selectedIssues;
    if (!issues || !issues.length) return this.noneSelectedAlert('edit');
    const params = {
      ids: issues.map((issue) => issue.localId).join(','),
      q: this.queryParams && this.queryParams.q,
    };
    this.page(`/p/${this.projectName}/issues/bulkedit?${qs.stringify(params)}`);
  }

  async _flagIssues(flagAsSpam = true) {
    const issues = this.selectedIssues;
    if (!issues || !issues.length) {
      return this.noneSelectedAlert(
          `${flagAsSpam ? 'flag' : 'un-flag'} as spam`);
    }
    const refs = issues.map((issue) => ({
      localId: issue.localId,
      projectName: issue.projectName,
    }));

    // TODO(zhangtiff): Refactor this into a shared action creator and
    // display the error on the frontend.
    try {
      await prpcClient.call('monorail.Issues', 'FlagIssues', {
        issueRefs: refs,
        flag: flagAsSpam,
      });
      this.refresh();
    } catch (e) {
      console.error(e);
    }
  }

  _setSelectedIssues() {
    const issueListRef = this.shadowRoot.querySelector('mr-issue-list');
    if (!issueListRef) return [];

    this.selectedIssues = issueListRef.selectedIssues;
  }
};
customElements.define('mr-list-page', MrListPage);
