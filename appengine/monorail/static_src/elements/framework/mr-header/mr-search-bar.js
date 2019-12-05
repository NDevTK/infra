// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {LitElement, html, css} from 'lit-element';
import page from 'page';
import qs from 'qs';

import '../mr-dropdown/mr-dropdown.js';
import {prpcClient} from 'prpc-client-instance.js';
import ClientLogger from 'monitoring/client-logger';
import {issueRefToUrl} from 'shared/converters.js';

// Search field input regex testing for all digits
// indicating that the user wants to jump to the specified issue.
const JUMP_RE = /^\d+$/;

/**
 * `<mr-search-bar>`
 *
 * The searchbar for Monorail.
 *
 */
export class MrSearchBar extends LitElement {
  /** @override */
  static get styles() {
    return css`
      :host {
        --mr-search-bar-background: white;
        --mr-search-bar-border-radius: 4px;
        --mr-search-bar-border: var(--chops-normal-border);
        --mr-search-bar-chip-color: var(--chops-gray-200);
        height: 30px;
        font-size: var(--chops-large-font-size);
      }
      input#searchq {
        display: flex;
        align-items: center;
        justify-content: flex-start;
        flex-grow: 2;
        min-width: 100px;
        border: none;
        border-top: var(--mr-search-bar-border);
        border-bottom: var(--mr-search-bar-border);
        background: var(--mr-search-bar-background);
        height: 100%;
        box-sizing: border-box;
        padding: 0 2px;
        font-size: inherit;
      }
      mr-dropdown {
        text-align: right;
        display: flex;
        text-overflow: ellipsis;
        box-sizing: border-box;
        background: var(--mr-search-bar-background);
        border: var(--mr-search-bar-border);
        border-left: 0;
        border-radius: 0 var(--mr-search-bar-border-radius)
          var(--mr-search-bar-border-radius) 0;
        height: 100%;
        align-items: center;
        justify-content: center;
        text-decoration: none;
      }
      button {
        font-size: inherit;
        order: -1;
        background: var(--mr-search-bar-background);
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        height: 100%;
        box-sizing: border-box;
        border: var(--mr-search-bar-border);
        border-left: none;
        border-right: none;
        padding: 0 8px;
      }
      form {
        display: flex;
        height: 100%;
        width: 100%;
        align-items: center;
        justify-content: flex-start;
        flex-direction: row;
      }
      i.material-icons {
        font-size: var(--chops-icon-font-size);
        color: var(--chops-primary-icon-color);
      }
      .select-container {
        order: -2;
        max-width: 150px;
        min-width: 50px;
        flex-shrink: 1;
        height: 100%;
        position: relative;
        box-sizing: border-box;
        border: var(--mr-search-bar-border);
        border-radius: var(--mr-search-bar-border-radius) 0 0
          var(--mr-search-bar-border-radius);
        background: var(--mr-search-bar-chip-color);
      }
      .select-container i.material-icons {
        display: flex;
        align-items: center;
        justify-content: center;
        position: absolute;
        right: 0;
        top: 0;
        height: 100%;
        width: 20px;
        z-index: 2;
        padding: 0;
      }
      select {
        display: flex;
        align-items: center;
        justify-content: flex-start;
        -webkit-appearance: none;
        -moz-appearance: none;
        appearance: none;
        text-overflow: ellipsis;
        cursor: pointer;
        width: 100%;
        height: 100%;
        background: none;
        margin: 0;
        padding: 0 20px 0 8px;
        box-sizing: border-box;
        border: 0;
        z-index: 3;
        font-size: inherit;
        position: relative;
      }
      select::-ms-expand {
        display: none;
      }
      select::after {
        position: relative;
        right: 0;
        content: 'arrow_drop_down';
        font-family: 'Material Icons';
      }
    `;
  }

  /** @override */
  render() {
    return html`
      <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">
      <form
        @submit=${this._searchSubmitted}
        @keypress=${this._submitSearchWithKeypress}
      >
        <div class="select-container">
          <i class="material-icons">arrow_drop_down</i>
          <select id="can" name="can" @change=${this._redirectOnSelect} aria-label="Search scope">
            <optgroup label="Search within">
              <option
                value="1"
                ?selected=${this.currentCan === '1'}
              >All issues</option>
              <option
                value="2"
                ?selected=${this.currentCan === '2'}
              >Open issues</option>
              <option
                value="3"
                ?selected=${this.currentCan === '3'}
              >Open and owned by me</option>
              <option
                value="4"
                ?selected=${this.currentCan === '4'}
              >Open and reported by me</option>
              <option
                value="5"
                ?selected=${this.currentCan === '5'}
              >Open and starred by me</option>
              <option
                value="8"
                ?selected=${this.currentCan === '8'}
              >Open with comment by me</option>
              <option
                value="6"
                ?selected=${this.currentCan === '6'}
              >New issues</option>
              <option
                value="7"
                ?selected=${this.currentCan === '7'}
              >Issues to verify</option>
            </optgroup>
            <optgroup label="Project queries" ?hidden=${!this.userDisplayName}>
              ${this.projectSavedQueries && this.projectSavedQueries.map((query) => html`
                <option
                  class="project-query"
                  value=${query.queryId}
                  ?selected=${this.currentCan === query.queryId}
                >${query.name}</option>
              `)}
              <option data-href="/p/${this.projectName}/adminViews">Manage project queries...</option>
            </optgroup>
            <optgroup label="My saved queries" ?hidden=${!this.userDisplayName}>
              ${this.userSavedQueries && this.userSavedQueries.map((query) => html`
                <option
                  class="user-query"
                  value=${query.queryId}
                  ?selected=${this.currentCan === query.queryId}
                >${query.name}</option>
              `)}
              <option data-href="/u/${this.userDisplayName}/queries">Manage my saved queries...</option>
            </optgroup>
          </select>
        </div>
        <input
          id="searchq"
          type="text"
          name="q"
          placeholder="Search ${this.projectName} issues..."
          value=${this.initialValue || ''}
          autocomplete="off"
          aria-label="Search box"
          @focus=${this._searchEditStarted}
          @blur=${this._searchEditFinished}
          spellcheck="false"
        />
        <button type="submit">
          <i class="material-icons">search</i>
        </button>
        <mr-dropdown
          label="Search options"
          .items=${this._searchMenuItems}
        ></mr-dropdown>
      </form>
    `;
  }

  /** @override */
  static get properties() {
    return {
      projectName: {type: String},
      userDisplayName: {type: String},
      currentCan: {type: String},
      initialValue: {type: String},
      projectSavedQueries: {type: Array},
      userSavedQueries: {type: Array},
      queryParams: {type: Object},
      keptQueryParams: {type: Array},
      _boundFocus: {
        type: Object,
        hasChanged: () => false,
      },
    };
  }

  /** @override */
  constructor() {
    super();
    this.queryParams = {};
    this.keptQueryParams = [
      'sort',
      'groupby',
      'colspec',
      'x',
      'y',
      'mode',
      'cells',
      'num',
    ];
    this.initialValue = '';
    this.currentCan = '2';
    this.projectSavedQueries = [];
    this.userSavedQueries = [];

    this.clientLogger = new ClientLogger('issues');

    this._page = page;
  }

  /** @override */
  connectedCallback() {
    super.connectedCallback();

    // Global event listeners. Make sure to unbind these when the
    // element disconnects.
    this._boundFocus = this.focus.bind(this);
    window.addEventListener('focus-search', this._boundFocus);
  }

  /** @override */
  disconnectedCallback() {
    super.disconnectedCallback();

    window.removeEventListener('focus-search', this._boundFocus);
  }

  /** @override */
  updated(changedProperties) {
    if (this.userDisplayName && changedProperties.has('userDisplayName')) {
      const userSavedQueriesPromise = prpcClient.call('monorail.Users',
          'GetSavedQueries', {});
      userSavedQueriesPromise.then((resp) => {
        this.userSavedQueries = resp.savedQueries;
      });
    }
  }

  _searchEditStarted() {
    this.clientLogger.logStart('query-edit', 'user-time');
    this.clientLogger.logStart('issue-search', 'user-time');
  }

  _searchEditFinished() {
    this.clientLogger.logEnd('query-edit');
  }

  _submitSearchWithKeypress(e) {
    if (e.key === 'Enter' && (e.shiftKey)) {
      const form = e.currentTarget;
      this._submitSearch(form, true);
    }
    // In all other cases, we want to let the submit handler do the work.
  }

  _searchSubmitted(e) {
    e.preventDefault();

    const form = e.target;
    this._submitSearch(form);
  }

  _submitSearch(form, newTab) {
    this.clientLogger.logEnd('query-edit');
    this.clientLogger.logPause('issue-search', 'user-time');
    this.clientLogger.logStart('issue-search', 'computer-time');

    const params = {};

    this.keptQueryParams.forEach((param) => {
      if (param in this.queryParams) {
        params[param] = this.queryParams[param];
      }
    });

    params.q = form.q.value.trim();
    params.can = form.can.value;

    this._navigateToNext(params, newTab);
  }

  /**
   * Attempt to jump-to-issue, otherwise continue to list view
   * @param {Object} params URL navigation parameters
   * @param {boolean} newTab
   */
  async _navigateToNext(params, newTab = false) {
    let resp;
    if (JUMP_RE.test(params.q)) {
      const message = {
        issueRef: {
          projectName: this.projectName,
          localId: params.q,
        },
      };

      try {
        resp = await prpcClient.call(
            'monorail.Issues', 'GetIssue', message,
        );
      } catch (error) {
        // Fall through to navigateToList
      }
    }
    if (resp && resp.issue) {
      const link = issueRefToUrl(resp.issue, params);
      this._page(link);
    } else {
      this._navigateToList(params, newTab);
    }
  }

  /**
   * Navigate to list view, currently splits on old and new view
   * @param {Object} params URL navigation parameters
   * @param {boolean} newTab
   */
  _navigateToList(params, newTab = false) {
    // TODO(zhangtiff): Remove this check once list_new is removed
    // when the new list page switches to default.
    const isNewPage = window.location.pathname.endsWith('list_new');

    const pathname = `/p/${this.projectName}/issues/${isNewPage ?
      'list_new' : 'list'}`;

    const hasChanges = !window.location.pathname.startsWith(pathname) ||
      this.queryParams.q !== params.q ||
      this.queryParams.can !== params.can;

    const url =`${pathname}?${qs.stringify(params)}`;

    if (newTab) {
      window.open(url, '_blank', 'noopener');
    } else if (hasChanges) {
      this._page(url);
    } else {
      if (isNewPage) {
        // TODO(zhangtiff): Replace this event with Redux once all of Monorail
        // uses Redux.
        this.dispatchEvent(new Event('refreshList',
            {'composed': true, 'bubbles': true}));
      } else {
        location.reload();
      }
    }
  }

  focus() {
    const search = this.shadowRoot.querySelector('#searchq');
    search.focus();
  }

  get _searchMenuItems() {
    const projectName = this.projectName;
    return [
      {
        text: 'Advanced search',
        url: `/p/${projectName}/issues/advsearch`,
      },
      {
        text: 'Search tips',
        url: `/p/${projectName}/issues/searchtips`,
      },
    ];
  }

  _redirectOnSelect(evt) {
    const target = evt.target;
    const option = target.options[target.selectedIndex];

    if (option.dataset.href) {
      window.location.href = option.dataset.href;
    }
  }
}

customElements.define('mr-search-bar', MrSearchBar);
