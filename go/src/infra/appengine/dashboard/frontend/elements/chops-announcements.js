// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {LitElement, html, css} from 'lit-element';
import {prpcClient} from 'prpc.js';

import 'announcements-table.js';
import 'announcement-input.js';

import {SHARED_STYLES} from 'shared-styles.js';

export class ChopsAnnouncements extends LitElement {
  static get properties() {
    return {
      isTrooper: {type: Boolean},
      liveAnnouncements: {type: Array},
      retiredAnnouncements: {type: Array},
      liveErrorMessage: {type: String},
      retiredErrorMessage: {type: String},
    };
  }

  constructor() {
    super();
    this.liveAnnouncements = [];
    this.retiredAnnouncement = [];
  }

  firstUpdated() {
    console.log('firstUpdated');
    this._fetchAnnouncements(0);
  }

  static get styles() {
    return [SHARED_STYLES, css`
      .round-icon {
        border-radius: 25px;
        display: table;
        width: 48px;
        height: 24px;
        margin: 2px;
      }
      .round-icon p {
        display: table-cell;
        text-align: center;
        vertical-align: middle;
        color: white;
        font-weight: bolder;
      }
      .live {
        background-color: red;
      }
    `];
  }
  render() {
    return html`
      <div class="round-icon live small"><p>LIVE</p></div>
      <announcements-table
        .isTrooper="${this.isTrooper}"
        .announcements="${this.liveAnnouncements}"
        @announcements-changed=${this._fetchAnnouncements(0)}
      ></announcements-table>
      ${this.liveErrorMessage ? html`
          <span class=error>${this.liveErrorMessage}</span>
        ` : ''}
      ${this.retiredErrorMessage ? html`
          <span class=error>${this.retiredErrorMessage}</span>
        ` : ''}
      ${this.isTrooper ? html `
        <announcement-input
          @announcement-created=${this._fetchAnnouncements(0)}
        ></announcement-input>
      ` : ''}
    `;
  }

  _fetchAnnouncements(retiredOffset) {
    console.log('this is being called');
    console.log(this.retiredAnnouncements);
    console.log(this.liveAnnouncements);
    this._fetchLiveAnnouncements();
    this._fetchRetiredAnnouncements(retiredOffset);
  }

  _fetchLiveAnnouncements() {
    const fetchLiveMessage = {
      retired: false,
    };
    const promise = this._searchAnnouncements(fetchLiveMessage);
    promise.then((resp) => {
      this.liveAnnouncements = resp.announcements;
    }).catch((reason) => {
      this.liveErrorMessage = `Failed to fetch live announcements: ${reason}`;
    });

  }
  _fetchRetiredAnnouncements(offset) {
    const fetchRetiredMessage = {
      retired: true,
      offset: offset,
      limit: 5,
    };
    const promise = this._searchAnnouncements(fetchRetiredMessage);
    promise.then((resp) => {
      this.retiredAnnouncements = resp.announcements;
    }).catch((reason) => {
      this.retiredErrorMessage = `Failed to fetch retired announcements: ${reason}`;
    });
  }

  async _searchAnnouncements(message) {
    return prpcClient.call(
	'dashboard.ChopsAnnouncements', 'SearchAnnouncements', message);
  }
}

customElements.define('chops-announcements', ChopsAnnouncements);
