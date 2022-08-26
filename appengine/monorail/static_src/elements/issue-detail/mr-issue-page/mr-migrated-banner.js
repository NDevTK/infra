// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {LitElement, html, css} from 'lit-element';

import {connectStore} from 'reducers/base.js';
import * as issueV0 from 'reducers/issueV0.js';
import {SHARED_STYLES} from 'shared/shared-styles.js';
import {migratedTypes} from 'shared/issue-fields.js';

/**
 * `<mr-migrated-banner>`
 *
 * Display for showing whether an issue is restricted.
 *
 */
export class MrMigratedBanner extends connectStore(LitElement) {
  /** @override */
  static get styles() {
    return [
      SHARED_STYLES,
      css`
        :host {
          width: 100%;
          margin-top: 0;
          background-color: var(--chops-orange-50);
          border-bottom: var(--chops-normal-border);
          font-size: var(--chops-main-font-size);
          padding: 0.25em 8px;
          box-sizing: border-box;
          display: flex;
          flex-direction: row;
          justify-content: flex-start;
          align-items: center;
        }
        :host([hidden]) {
          display: none;
        }
        i.material-icons {
          color: var(--chops-primary-icon-color);
          font-size: var(--chops-icon-font-size);
        }
        .warning-icon {
          margin-right: 4px;
        }
      `,
    ];
  }

  /** @override */
  render() {
    return html`
      <link href="https://fonts.googleapis.com/icon?family=Material+Icons"
            rel="stylesheet">
      <i
        class="warning-icon material-icons"
        icon="warning"
      >warning</i>
      ${this._link}
    `;
  }

  /** @override */
  static get properties() {
    return {
      migratedId: {type: String},
      hidden: {
        type: Boolean,
        reflect: true,
      },
      migratedType: {type: migratedTypes}
    };
  }

  /** @override */
  constructor() {
    super();

    this.hidden = true;
  }

  /** @override */
  stateChanged(state) {
    this.migratedId = issueV0.migratedId(state);
    this.migratedType = issueV0.migratedType(state);
  }

   /** @override */
   update(changedProperties) {
    if (changedProperties.has('migratedId')) {
      this.hidden = !this.migratedId || this.migratedId === '';
    }

    super.update(changedProperties);
  }

  /**
   * @return {string} the link of the issue in Issue Tracker.
   */
  get _link() {
    if (this.migratedType === migratedTypes.BUGANIZER_TYPE) {
      const link = 
        html`<a href="https://issuetracker.google.com/issues/${this.migratedId}">b/${this.migratedId}</a>`;
      return html`<p>This issue has moved to ${link}. Updates should be posted in ${link}.</p>`;
    } else {
      return html`<p>This issue has been migrated to Launch, see link in final comment below.</p>`;
    }
  }
}

customElements.define('mr-migrated-banner', MrMigratedBanner);
