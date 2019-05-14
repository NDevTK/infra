// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {LitElement, html, css} from 'lit-element';
import '../mr-dropdown/mr-dropdown.js';
import {prpcClient} from 'prpc-client-instance.js';
import ClientLogger from 'monitoring/client-logger';


let clientLogger;
const getClientLogger = () => {
  if (!clientLogger) {
    clientLogger = new ClientLogger('issues');
  }
  return clientLogger;
};

const searchEditStarted = () => {
  getClientLogger().logStart('query-edit', 'user-time');
  getClientLogger().logStart('issue-search', 'user-time');
};

const searchEditFinished = () => {
  getClientLogger().logEnd('query-edit');
};

const searchSubmitted = () => {
  getClientLogger().logEnd('query-edit');
  getClientLogger().logPause('issue-search', 'user-time');
  getClientLogger().logStart('issue-search', 'computer-time');
};

/**
 * `<mr-search-bar>`
 *
 * The searchbar for Monorail.
 *
 */
export class MrSearchBar extends LitElement {
  static get styles() {
    return css`
      :host {
        --mr-search-bar-background: white;
        --mr-search-bar-border-radius: 4px;
        --mr-search-bar-border: var(--chops-accessible-border);
        --mr-search-bar-chip-color: var(--chops-gray-200);
        height: 26px;
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

  render() {
    return html`
      <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">
      <form action="/p/${this.projectName}/issues/list" method="GET" @submit=${searchSubmitted}>
        <div class="select-container">
          <i class="material-icons">arrow_drop_down</i>
          <select id="can" name="can" @change=${this._redirectOnSelect} aria-label="Search scope">
            <optgroup label="Search within">
              <option value="1" ?selected=${this.defaultCan === '1'}>All issues</option>
              <option value="2" ?selected=${this.defaultCan === '2'}>Open issues</option>
              <option value="3" ?selected=${this.defaultCan === '3'}>Open and owned by me</option>
              <option value="4" ?selected=${this.defaultCan === '4'}>Open and reported by me</option>
              <option value="5" ?selected=${this.defaultCan === '5'}>Open and starred by me</option>
              <option value="8" ?selected=${this.defaultCan === '8'}>Open with comment by me</option>
              <option value="6" ?selected=${this.defaultCan === '6'}>New issues</option>
              <option value="7" ?selected=${this.defaultCan === '7'}>Issues to verify</option>
            </optgroup>
            <optgroup label="Project queries" ?hidden=${!this.userDisplayName}>
              ${this.projectSavedQueries && this.projectSavedQueries.map((query) => html`
                <option
                  class="project-query"
                  value=${query.queryId}
                  ?selected=${this.defaultCan === query.queryId}
                >${query.name}</option>
              `)}
              <option data-href="/p/${this.projectName}/adminViews">Manage project queries...</option>
            </optgroup>
            <optgroup label="My saved queries" ?hidden=${!this.userDisplayName}>
              ${this.userSavedQueries && this.userSavedQueries.map((query) => html`
                <option
                  class="user-query"
                  value=${query.queryId}
                  ?selected=${this.defaultCan === query.queryId}
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
          @focus=${searchEditStarted}
          @blur=${searchEditFinished}
        />
        ${this.keptQueryParams.map((param) => (this.queryParams && (param in this.queryParams) ? html`
            <input type="hidden" id=${param} name=${param} value=${this.queryParams[param]} />
          ` : ''))}
        <button type="submit">
          <i class="material-icons">search</i>
        </button>
        <mr-dropdown
          .items=${this._searchMenuItems}
        ></mr-dropdown>
      </form>
    `;
  }

  static get properties() {
    return {
      projectName: {type: String},
      userDisplayName: {type: String},
      defaultCan: {type: String},
      initialValue: {type: String},
      projectSavedQueries: {type: Array},
      userSavedQueries: {type: Array},
      queryParams: {type: Object},
      keptQueryParams: {type: Array},
    };
  }

  constructor() {
    super();
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
    this.defaultCan = '2';
    this.projectSavedQueries = [];
    this.userSavedQueries = [];
  }

  connectedCallback() {
    super.connectedCallback();

    // Global event listeners. Make sure to unbind these when the
    // element disconnects.

    window.addEventListener('focus-search', this.focus);
  }

  disconnectedCallback() {
    super.disconnectedCallback();

    window.removeEventListener('focus-search', this.focus);
  }

  updated(changedProperties) {
    if (this.userDisplayName && changedProperties.has('userDisplayName')) {
      const userSavedQueriesPromise = prpcClient.call('monorail.Users',
        'GetSavedQueries', {});
      userSavedQueriesPromise.then((resp) => {
        this.userSavedQueries = resp.savedQueries;
      });
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
