// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {LitElement, html, css} from 'lit-element';
import page from 'page';
import qs from 'qs';
import Mousetrap from 'mousetrap';

import {store, connectStore} from 'reducers/base.js';
import * as issue from 'reducers/issue.js';
import * as project from 'reducers/project.js';
import 'elements/chops/chops-dialog/chops-dialog.js';
import {issueRefToString} from 'shared/converters.js';


const SHORTCUT_DOC_GROUPS = [
  {
    title: 'Issue list',
    keyDocs: [
      {
        keys: ['k', 'j'],
        tip: 'up/down in the list',
      },
      {
        keys: ['o', 'Enter'],
        tip: 'open the current issue',
      },
      {
        keys: ['Shift-O'],
        tip: 'open issue in new tab',
      },
      {
        keys: ['x'],
        tip: 'select the current issue',
      },
    ],
  },
  {
    title: 'Issue details',
    keyDocs: [
      {
        keys: ['k', 'j'],
        tip: 'prev/next issue in list',
      },
      {
        keys: ['u'],
        tip: 'up to issue list',
      },
      {
        keys: ['r'],
        tip: 'reply to current issue',
      },
    ],
  },
  {
    title: 'Anywhere',
    keyDocs: [
      {
        keys: ['/'],
        tip: 'focus on the issue search field',
      },
      {
        keys: ['c'],
        tip: 'compose a new issue',
      },
      {
        keys: ['s'],
        tip: 'star the current issue',
      },
      {
        keys: ['?'],
        tip: 'show this help dialog',
      },
    ],
  },
];

/**
 * `<mr-keystrokes>`
 *
 * Adds keybindings for Monorail, including a dialog for showing keystrokes.
 * @extends {LitElement}
 */
export class MrKeystrokes extends connectStore(LitElement) {
  /** @override */
  static get styles() {
    return css`
      h2 {
        margin-top: 0;
        display: flex;
        justify-content: space-between;
        font-weight: normal;
        border-bottom: 2px solid white;
        font-size: var(--chops-large-font-size);
        padding-bottom: 0.5em;
      }
      .close-button {
        border: 0;
        background: 0;
        text-decoration: underline;
        cursor: pointer;
      }
      .keyboard-help {
        display: flex;
        align-items: flex-start;
        justify-content: space-around;
        flex-direction: row;
        border-bottom: 2px solid white;
        flex-wrap: wrap;
      }
      .keyboard-help-section {
        width: 32%;
        display: grid;
        grid-template-columns: 40% 60%;
        padding-bottom: 1em;
        grid-gap: 4px;
        min-width: 300px;
      }
      .help-title {
        font-weight: bold;
      }
      .key-shortcut {
        text-align: right;
        padding-right: 8px;
        font-weight: bold;
        margin: 2px;
      }
      kbd {
        background: var(--chops-gray-200);
        padding: 2px 8px;
        border-radius: 2px;
        min-width: 28px;
      }
    `;
  }

  /** @override */
  render() {
    return html`
      <chops-dialog ?opened=${this._opened}>
        <h2>
          Issue tracker keyboard shortcuts
          <button class="close-button" @click=${this._closeDialog}>
            Close
          </button>
        </h2>
        <div class="keyboard-help">
          ${this._shortcutDocGroups.map((group) => html`
            <div class="keyboard-help-section">
              <span></span><span class="help-title">${group.title}</span>
              ${group.keyDocs.map((keyDoc) => html`
                <span class="key-shortcut">
                  ${keyDoc.keys.map((key, i) => html`
                    <kbd>${key}</kbd>
                    <span
                      class="key-separator"
                      ?hidden=${i === keyDoc.keys.length - 1}
                    > / </span>
                  `)}:
                </span>
                <span class="key-tip">${keyDoc.tip}</span>
              `)}
            </div>
          `)}
        </div>
        <p>
          Note: Only signed in users can star issues or add comments, and
          only project members can select issues for bulk edits.
        </p>
      </chops-dialog>
    `;
  }

  /** @override */
  static get properties() {
    return {
      issueEntryUrl: {type: String},
      issueId: {type: Number},
      projectName: {type: String},
      queryParams: {type: Object},
      _fetchingIsStarred: {type: Boolean},
      _isStarred: {type: Boolean},
      _issuePermissions: {type: Array},
      _opened: {type: Boolean},
      _shortcutDocGroups: {type: Array},
      _starringIssues: {type: Object},
    };
  }

  /** @override */
  constructor() {
    super();

    this._shortcutDocGroups = SHORTCUT_DOC_GROUPS;
    this._opened = false;
    this._starringIssues = new Map();
    this.projectName = undefined;
    this.issueId = undefined;
    this.queryParams = undefined;
    this.issueEntryUrl = undefined;
  }

  /** @override */
  stateChanged(state) {
    this.projectName = project.viewedProjectName(state);
    this._issuePermissions = issue.permissions(state);

    const starredIssues = issue.starredIssues(state);
    this._isStarred = starredIssues.has(issueRefToString(this._issueRef));
    this._fetchingIsStarred = issue.requests(state).fetchIsStarred.requesting;
    this._starringIssues = issue.starringIssues(state);
  }

  /** @override */
  updated(changedProperties) {
    if (changedProperties.has('projectName') ||
        changedProperties.has('issueEntryUrl')) {
      this._bindProjectKeys(this.projectName, this.issueEntryUrl);
    }
    if (changedProperties.has('projectName') || changedProperties.has('issueId') ||
        changedProperties.has('issuePermissions') ||
        changedProperties.has('queryParams')) {
      this._bindIssueDetailKeys(this.projectName, this.issueId,
          this._issuePermissions, this.queryParams);
    }
  }

  /** @override */
  disconnectedCallback() {
    super.disconnectedCallback();
    this._unbindProjectKeys();
    this._unbindIssueDetailKeys();
  }

  /** @private */
  get _isStarring() {
    const requestKey = issueRefToString(this._issueRef);
    if (this._starringIssues.has(requestKey)) {
      return this._starringIssues.get(requestKey).requesting;
    }
    return false;
  }

  /** @private */
  get _issueRef() {
    return {
      projectName: this.projectName,
      localId: this.issueId,
    };
  }

  /** @private */
  _toggleDialog() {
    this._opened = !this._opened;
  }

  /** @private */
  _openDialog() {
    this._opened = true;
  }

  /** @private */
  _closeDialog() {
    this._opened = false;
  }

  /**
   * @param {string} projectName
   * @param {string} issueEntryUrl
   * @private
   */
  _bindProjectKeys(projectName, issueEntryUrl) {
    this._unbindProjectKeys();

    if (!projectName) return;

    issueEntryUrl = issueEntryUrl || `/p/${projectName}/issues/entry`;

    Mousetrap.bind('/', (e) => {
      e.preventDefault();
      // Focus search.
      this.dispatchEvent(new CustomEvent('focus-search',
          {composed: true, bubbles: true}));
    });

    Mousetrap.bind('?', () => {
      // Toggle key help.
      this._toggleDialog();
    });

    Mousetrap.bind('esc', () => {
      // Close key help dialog if open.
      this._closeDialog();
    });

    Mousetrap.bind('c', () => page(issueEntryUrl));
  }

  /** @private */
  _unbindProjectKeys() {
    Mousetrap.unbind('/');
    Mousetrap.unbind('?');
    Mousetrap.unbind('esc');
    Mousetrap.unbind('c');
  }

  /**
   * @param {string} projectName
   * @param {string} issueId
   * @param {Array<string>} issuePermissions
   * @param {Object} queryParams
   * @private
   */
  _bindIssueDetailKeys(projectName, issueId, issuePermissions, queryParams) {
    this._unbindIssueDetailKeys();

    if (!projectName || !issueId) return;

    const queryString = qs.stringify(queryParams);
    const projectHomeUrl = `/p/${projectName}`;

    issuePermissions = issuePermissions || [];

    // TODO(zhangtiff): Update these links when mr-flipper's async request
    // finishes.
    const listUrl = `${projectHomeUrl}/issues/detail/list?${queryString}`;
    const prevUrl = `${projectHomeUrl}/issues/detail/previous?${queryString}`;
    const nextUrl = `${projectHomeUrl}/issues/detail/next?${queryString}`;
    const canComment = issuePermissions.includes('addissuecomment');
    const canStar = issuePermissions.includes('setstar');


    if (prevUrl) {
      // Previous issue in list.
      Mousetrap.bind('k', () => page(prevUrl));
    }

    if (nextUrl) {
      // Next issue in list.
      Mousetrap.bind('j', () => page(nextUrl));
    }

    if (listUrl) {
      // Back to list.
      Mousetrap.bind('u', () => page(listUrl));
    }

    if (canComment) {
      // Navigate to the form to make changes.
      Mousetrap.bind('r', (e) => {
        e.preventDefault();

        // Force a hash change even the hash is already makechanges.
        if (window.location.hash.toLowerCase() === '#makechanges') {
          window.location.hash = ' ';
        }
        window.location.hash = '#makechanges';
      });
    }

    if (canStar) {
      Mousetrap.bind('s', () => {
        // Star an issue.
        if (!this._fetchingIsStarred && !this._isStarring) {
          const newIsStarred = !this._isStarred;

          store.dispatch(issue.star(this._issueRef, newIsStarred));
        }
      });
    }
  }

  /** @private */
  _unbindIssueDetailKeys() {
    Mousetrap.unbind('k');
    Mousetrap.unbind('j');
    Mousetrap.unbind('u');
    Mousetrap.unbind('r');
    Mousetrap.unbind('s');
  }
}

customElements.define('mr-keystrokes', MrKeystrokes);
