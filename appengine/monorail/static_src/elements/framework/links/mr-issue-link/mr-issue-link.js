// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {LitElement, html, css} from 'lit-element';
import {ifDefined} from 'lit-html/directives/if-defined';
import {issueRefToString} from 'elements/shared/converters.js';
import {fromShortlink} from 'elements/shared/federated.js';

/**
 * `<mr-issue-link>`
 *
 * Displays a link to an issue.
 *
 */
export class MrIssueLink extends LitElement {
  static get styles() {
    return css`
      a[is-closed] {
        text-decoration: line-through;
      }
    `;
  }

  // TODO(CL): Move to a more appropriate place.
  async connectedCallback() {
    super.connectedCallback();

    if (this.issue && this.issue.extIdentifier) {
      // TODO(CL): If it's a Buganizer issue...
      // Or make the Buganizer class make the call.. actually that's nicer
      const bugID = this.issue.extIdentifier.substr(2);
      console.log('WILL LOAD', bugID)
      const response = await loadGoogleIssueTrackerIssue(bugID);
      console.log('GOT', response)

      const issueStatus = response && response.issueState && response.issueState.status;
      const googIssueTrackerClosedStatuses = new Set(['FIXED']);

      if (googIssueTrackerClosedStatuses.has(issueStatus)) {
        console.log('setting to false')
        // TODO(CL): Probably not the way to do this.
        this.issue.statusRef = {meansOpen: false};
        this.requestUpdate();
      }
    }
  }

  render() {
    return html`
      <a
        id="bugLink"
        href=${this._issueUrl}
        title=${ifDefined(this.issue && this.issue.summary)}
        ?is-closed=${this.isClosed}
      >${this._linkText}</a>
    `;
  }

  static get properties() {
    return {
      // The issue being viewed. Falls back gracefully if this is only a ref.
      issue: {type: Object},
      text: {type: String},
      // The global current project name. NOT the issue's project name.
      projectName: {type: String},
    };
  }

  get isClosed() {
    if (!this.issue || !this.issue.statusRef) return false;

    return this.issue.statusRef.meansOpen === false;
  }

  get _issueUrl() {
    const issue = this.issue;
    if (!issue) return '';

    if (issue.extIdentifier) {
      const extRef = fromShortlink(issue.extIdentifier);
      if (!extRef) {
        console.error(`No tracker found for reference: ${issue.extIdentifier}`);
        return '';
      }
      return extRef.toURL();
    }

    return `/p/${issue.projectName}/issues/detail?id=${issue.localId}`;
  }

  get _linkText() {
    const projectName = this.projectName;
    const issue = this.issue;
    const text = this.text;
    if (text) return text;

    if (issue && issue.extIdentifier) {
      return issue.extIdentifier;
    }

    return `Issue ${issueRefToString(issue, projectName)}`;
  }
}

customElements.define('mr-issue-link', MrIssueLink);
