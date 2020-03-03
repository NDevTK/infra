// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {LitElement, html, css} from 'lit-element';
import {prpcClient} from 'prpc-client-instance.js';

/**
 * `<mr-issue-entry-page>`
 *
 * This is the main details section for a given issue.
 *
 */
export class MrIssueEntryPage extends LitElement {
  /** @override */
  static get styles() {
    return css`
      :host {
        margin: 0;
      }
    `;
  }

  /** @override */
  render() {
    return html`
      <div>SPA issue entry page place holder</div>
    `;
  }

  /** @override */
  async connectedCallback() {
    super.connectedCallback();

    const message = {parent: 'projects/test-project-name'};
    await prpcClient.call(
        'monorail.v1.Projects',
        'ListIssueTemplates',
        message,
    );
  }
}

customElements.define('mr-issue-entry-page', MrIssueEntryPage);
