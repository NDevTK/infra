// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {LitElement, html, css} from 'lit-element';
import {store} from 'elements/reducers/base.js';
import * as issue from 'elements/reducers/issue.js';

import 'elements/chops/chops-button/chops-button.js';
import 'elements/chops/chops-timestamp/chops-timestamp.js';
import 'elements/framework/mr-comment-content/mr-comment-content.js';
import 'elements/framework/mr-comment-content/mr-attachment.js';
import 'elements/framework/mr-dropdown/mr-dropdown.js';
import 'elements/framework/links/mr-issue-link/mr-issue-link.js';
import 'elements/framework/links/mr-user-link/mr-user-link.js';

// Match: projectName:localIdFormat
const ISSUE_ID_REGEX = /(?:-?([a-z0-9-]+):)?(\d+)/i;
const ISSUE_REF_FIELD_NAMES = [
  'Blocking',
  'Blockedon',
  'Mergedinto',
];

/**
 * `<mr-comment>`
 *
 * A component for an individual comment.
 *
 */
export class MrComment extends LitElement {
  constructor() {
    super();

    this._isExpandedIfDeleted = false;
  }

  static get properties() {
    return {
      comment: {type: Object},
      headingLevel: {type: String},
      quickMode: {type: Boolean},
      highlighted: {
        type: Boolean,
        reflect: true,
      },
      _isExpandedIfDeleted: {type: Boolean},
    };
  }

  updated(changedProperties) {
    super.updated(changedProperties);

    if (changedProperties.has('highlighted') && this.highlighted) {
      this.scrollIntoView();
      // TODO(ehmaldonado): Figure out a way to get the height from the issue
      // header, and scroll by that amount.
      window.scrollBy(0, -150);
    }
  }

  static get styles() {
    return css`
      :host {
        display: block;
        margin: 1.5em 0 0 0;
      }
      :host([highlighted]) {
        border: 1px solid var(--chops-primary-accent-color);
        box-shadow: 0 0 4px 4px var(--chops-primary-accent-bg);
      }
      :host([hidden]) {
        display: none;
      }
      .comment-header {
        background: var(--chops-card-heading-bg);
        padding: 3px 1px 1px 8px;
        width: 100%;
        display: flex;
        flex-direction: row;
        justify-content: space-between;
        align-items: center;
        box-sizing: border-box;
      }
      .comment-header a {
        display: inline-flex;
      }
      .comment-options {
        float: right;
        text-align: right;
        text-decoration: none;
      }
      .comment-body {
        margin: 4px;
        box-sizing: border-box;
      }
      .deleted-comment-notice {
        margin-left: 4px;
      }
      .issue-diff {
        background: var(--chops-card-details-bg);
        display: inline-block;
        padding: 4px 8px;
        width: 100%;
        box-sizing: border-box;
      }
    `;
  }

  render() {
    return html`
      ${this._renderHeading()}
      ${_shouldShowComment(this._isExpandedIfDeleted, this.comment) ? html`
        ${this._renderDiff()}
        ${this._renderBody()}
      ` : ''}
    `;
  }

  _renderHeading() {
    return html`
      <div
        role="heading"
        aria-level=${this.headingLevel}
        class="comment-header">
        <div>
          <a href="?id=${this.comment.localId}#c${this.comment.sequenceNum}"
          >Comment ${this.comment.sequenceNum}</a>

          ${this._renderByline()}
        </div>
        ${_shouldOfferCommentOptions(this.quickMode, this.comment) ? html`
          <div class="comment-options">
            <mr-dropdown
              .items=${_commentOptions(
                  this._isExpandedIfDeleted,
                  this.comment,
                  this._toggleHideDeletedComment.bind(this))}
              icon="more_vert"
            ></mr-dropdown>
          </div>
        ` : ''}
      </div>
    `;
  }

  _renderByline() {
    if (_shouldShowComment(this._isExpandedIfDeleted, this.comment)) {
      return html`
        by
        <mr-user-link .userRef=${this.comment.commenter}></mr-user-link>
        ${!this.quickMode ? html`
          on
          <chops-timestamp
              timestamp=${this.comment.timestamp}></chops-timestamp>
        ` : ''}
      `;
    } else {
      return html`<span class="deleted-comment-notice">Deleted</span>`;
    }
  }

  _renderDiff() {
    if (!(this.comment.descriptionNum || this.comment.amendments)) return '';

    return html`
      <div class="issue-diff">
        ${(this.comment.amendments || []).map((delta) => html`
          <strong>${delta.fieldName}:</strong>
          ${_issuesForAmendment(delta, this.comment.projectName).map((issue) => html`
            <mr-issue-link
              projectName=${this.comment.projectName}
              .issue=${issue.issue}
              text=${issue.text}
            ></mr-issue-link>
          `)}
          ${!_amendmentHasIssueRefs(delta.fieldName) ? delta.newOrDeltaValue : ''}
          ${delta.oldValue ? `(was: ${delta.oldValue})` : ''}
          <br>
        `)}
        ${this.comment.descriptionNum ? 'Description was changed.' : ''}
      </div><br>
    `;
  }

  _renderBody() {
    return html`
      <div class="comment-body">
        <mr-comment-content
          ?hidden=${this.comment.descriptionNum}
          .content=${this.comment.content}
          ?isDeleted=${this.comment.isDeleted}
        ></mr-comment-content>
        ${this.quickMode ? '' : html`
          <div ?hidden=${this.comment.descriptionNum}>
            ${(this.comment.attachments || []).map((attachment) => html`
              <mr-attachment
                .attachment=${attachment}
                projectName=${this.comment.projectName}
                localId=${this.comment.localId}
                sequenceNum=${this.comment.sequenceNum}
                ?canDelete=${this.comment.canDelete}
              ></mr-attachment>
            `)}
          </div>
        `}
      </div>
    `;
  }

  _toggleHideDeletedComment() {
    this._isExpandedIfDeleted = !this._isExpandedIfDeleted;
  }
}

function _shouldShowComment(isExpandedIfDeleted, comment) {
  return !comment.isDeleted || isExpandedIfDeleted;
}

function _shouldOfferCommentOptions(quickMode, comment) {
  return !quickMode && (comment.canDelete || comment.canFlag);
}

function _commentOptions(isExpandedIfDeleted, comment, toggleHandler) {
  const options = [];
  if (_canExpandDeletedComment(comment)) {
    const text =
      (isExpandedIfDeleted ? 'Hide' : 'Show') + ' comment content';
    options.push({
      text: text,
      handler: toggleHandler,
    });
    options.push({separator: true});
  }
  if (comment.canDelete) {
    const text =
      (comment.isDeleted ? 'Undelete' : 'Delete') + ' comment';
    options.push({
      text: text,
      handler: _deleteComment.bind(null, comment),
    });
  }
  if (comment.canFlag) {
    const text = (comment.isSpam ? 'Unflag' : 'Flag') + ' comment';
    options.push({
      text: text,
      handler: _flagComment.bind(null, comment),
    });
  }
  return options;
}

function _canExpandDeletedComment(comment) {
  return ((comment.isSpam && comment.canFlag)
          || (comment.isDeleted && comment.canDelete));
}

async function _deleteComment(comment) {
  const issueRef = {
    projectName: comment.projectName,
    localId: comment.localId,
  };
  await window.prpcClient.call('monorail.Issues', 'DeleteIssueComment', {
    issueRef,
    sequenceNum: comment.sequenceNum,
    delete: comment.isDeleted === undefined,
  });
  store.dispatch(issue.fetchComments({issueRef}));
}

async function _flagComment(comment) {
  const issueRef = {
    projectName: comment.projectName,
    localId: comment.localId,
  };
  await window.prpcClient.call('monorail.Issues', 'FlagComment', {
    issueRef,
    sequenceNum: comment.sequenceNum,
    flag: comment.isSpam === undefined,
  });
  store.dispatch(issue.fetchComments({issueRef}));
}

function _issuesForAmendment(delta, projectName) {
  if (!_amendmentHasIssueRefs(delta.fieldName)
      || !delta.newOrDeltaValue) {
    return [];
  }
  // TODO(ehmaldonado): Request the issue to check for permissions and display
  // the issue summary.
  return delta.newOrDeltaValue.split(' ').map((issueRef) => {
    const matches = issueRef.match(ISSUE_ID_REGEX);
    return {
      issue: {
        projectName: matches[1] ? matches[1] : projectName,
        localId: matches[2],
      },
      text: issueRef,
    };
  });
}

function _amendmentHasIssueRefs(fieldName) {
  return ISSUE_REF_FIELD_NAMES.includes(fieldName);
}

customElements.define('mr-comment', MrComment);
