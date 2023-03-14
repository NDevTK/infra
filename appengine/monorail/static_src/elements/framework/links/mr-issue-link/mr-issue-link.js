// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {LitElement, html, css} from 'lit-element';
import {ifDefined} from 'lit-html/directives/if-defined';
import {issueRefToString, issueRefToUrl} from 'shared/convertersV0.js';
import {SHARED_STYLES} from 'shared/shared-styles.js';

/**
 * `<mr-issue-link>`
 *
 * Displays a link to an issue.
 *
 */
export class MrIssueLink extends LitElement {
  /** @override */
  static get styles() {
    return [
      SHARED_STYLES,
      css`
        a[is-closed] {
          text-decoration: line-through;
        }
      `,
    ];
  }

  /** @override */
  render() {
    return html`
      <a
        id="bugLink"
        href=${this.href}
        title=${ifDefined(this.issue && this.issue.summary)}
        ?is-closed=${this.isClosed}
      >${this._linkText}</a>`;
  }

  /** @override */
  static get properties() {
    return {
      // The issue being viewed. Falls back gracefully if this is only a ref.
      issue: {type: Object},
      text: {type: String},
      // The global current project name. NOT the issue's project name.
      projectName: {type: String},
      queryParams: {type: Object},
      short: {type: Boolean},
    };
  }

  /** @override */
  constructor() {
    super();

    this.issue = {};
    this.queryParams = {};
    this.short = false;
  }

  click() {
    const link = this.shadowRoot.querySelector('a');
    if (!link) return;
    link.click();
  }

  /**
   * @return {string} Where this issue links to.
   */
  get href() {
    return issueRefToUrl(this.issue, this.queryParams);
  }

  get isClosed() {
    if (!this.issue || !this.issue.statusRef) return false;

    return this.issue.statusRef.meansOpen === false;
  }

  get _linkText() {
    const {projectName, issue, text, short} = this;
    if (text) return text;

    if (issue && issue.extIdentifier) {
      return issue.extIdentifier;
    }

    const prefix = short ? '' : 'Issue ';

    return prefix + issueRefToString(issue, projectName);
  }
}

customElements.define('mr-issue-link', MrIssueLink);
