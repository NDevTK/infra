// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@polymer/polymer/polymer-legacy.js';
import {PolymerElement, html} from '@polymer/polymer';

import '../../chops/chops-dialog/chops-dialog.js';
import '../../chops/chops-collapse/chops-collapse.js';
import {store, connectStore} from '../../redux/base.js';
import * as issue from '../../redux/issue.js';
import * as project from '../../redux/project.js';
import * as user from '../../redux/user.js';
import * as ui from '../../redux/ui.js';
import {fieldTypes} from '../../shared/field-types.js';
import '../../mr-comment-content/mr-description.js';
import '../mr-comment-list/mr-comment-list.js';
import '../mr-edit-metadata/mr-edit-metadata.js';
import '../mr-metadata/mr-metadata.js';
import '../../shared/mr-shared-styles.js';

const APPROVER_RESTRICTED_STATUSES = new Set(
  ['NA', 'Approved', 'NotApproved']);

const STATUS_ENUM_TO_TEXT = {
  '': 'NotSet',
  'NEEDS_REVIEW': 'NeedsReview',
  'NA': 'NA',
  'REVIEW_REQUESTED': 'ReviewRequested',
  'REVIEW_STARTED': 'ReviewStarted',
  'NEED_INFO': 'NeedInfo',
  'APPROVED': 'Approved',
  'NOT_APPROVED': 'NotApproved',
};

const TEXT_TO_STATUS_ENUM = {
  'NotSet': 'NOT_SET',
  'NeedsReview': 'NEEDS_REVIEW',
  'NA': 'NA',
  'ReviewRequested': 'REVIEW_REQUESTED',
  'ReviewStarted': 'REVIEW_STARTED',
  'NeedInfo': 'NEED_INFO',
  'Approved': 'APPROVED',
  'NotApproved': 'NOT_APPROVED',
};

const STATUS_CLASS_MAP = {
  'NotSet': 'status-notset',
  'NeedsReview': 'status-pending',
  'NA': 'status-notset',
  'ReviewRequested': 'status-pending',
  'ReviewStarted': 'status-pending',
  'NeedInfo': 'status-pending',
  'Approved': 'status-approved',
  'NotApproved': 'status-rejected',
};

const STATUS_DOCSTRING_MAP = {
  'NotSet': '',
  'NeedsReview': 'Approval gate needs work',
  'NA': 'Approval gate not required',
  'ReviewRequested': 'Approval requested',
  'ReviewStarted': 'Approval in progress',
  'NeedInfo': 'Approval review needs more information',
  'Approved': 'Approved for Launch',
  'NotApproved': 'Not Approved for Launch',
};

const CLASS_ICON_MAP = {
  'status-notset': 'remove',
  'status-pending': 'autorenew',
  'status-approved': 'done',
  'status-rejected': 'close',
};

/**
 * `<mr-approval-card>`
 *
 * This element shows a card for a single approval.
 *
 */
export class MrApprovalCard extends connectStore(PolymerElement) {
  static get template() {
    return html`
      <link href="https://fonts.googleapis.com/icon?family=Material+Icons"
            rel="stylesheet">
      <style include="mr-shared-styles">
        :host {
          width: 100%;
          background-color: white;
          font-size: var(--chops-main-font-size);
          border-bottom: var(--chops-normal-border);
          box-sizing: border-box;
          display: block;
          border-left: 4px solid var(--approval-bg-color);
          --approval-bg-color: hsl(227, 20%, 92%);
          --approval-accent-color: hsl(227, 80%, 40%);
        }
        :host(.status-approved) {
          --approval-bg-color: hsl(78, 55%, 90%);
          --approval-accent-color: hsl(78, 100%, 30%);
        }
        :host(.status-pending) {
          --approval-bg-color: hsl(40, 75%, 90%);
          --approval-accent-color: hsl(33, 100%, 39%);
        }
        :host(.status-rejected) {
          --approval-bg-color: hsl(5, 60%, 92%);
          --approval-accent-color: hsl(357, 100%, 39%);
        }
        chops-button {
          border: var(--chops-normal-border);
          margin: 0;
        }
        h3 {
          margin: 0;
          padding: 0;
          display: inline;
          font-weight: inherit;
          font-size: inherit;
          line-height: inherit;
        }
        mr-description {
          display: block;
          margin-bottom: 0.5em;
        }
        .approver-notice {
          padding: 0.25em 0;
          width: 100%;
          display: flex;
          flex-direction: row;
          align-items: baseline;
          justify-content: space-between;
          border-bottom: 1px dotted hsl(0, 0%, 83%);
        }
        .card-content {
          box-sizing: border-box;
          padding: 0.5em 16px;
          padding-bottom: 1em;
        }
        .expand-icon {
          display: block;
          margin-right: 8px;
          color: hsl(0, 0%, 45%);
        }
        .header {
          margin: 0;
          width: 100%;
          border: 0;
          font-size: var(--chops-large-font-size);
          font-weight: normal;
          box-sizing: border-box;
          display: flex;
          align-items: center;
          flex-direction: row;
          padding: 0.5em 8px;
          background-color: var(--approval-bg-color);
          cursor: pointer;
        }
        .status {
          font-size: var(--chops-main-font-size);
          color: var(--approval-accent-color);
          display: inline-flex;
          align-items: center;
          margin-left: 32px;
        }
        .survey {
          padding: 0.5em 0;
          max-height: 500px;
          overflow-y: auto;
          max-width: 100%;
          box-sizing: border-box;
        }
        [role="heading"] {
          display: flex;
          flex-direction: row;
          justify-content: space-between;
          align-items: flex-end;
        }
      </style>
      <button class="header" on-click="toggleCard" aria-expanded$="[[_toString(opened)]]">
        <i class="material-icons expand-icon">[[_expandIcon]]</i>
        <h3>[[fieldName]]</h3>
        <span class="status">
          <i class="material-icons status-icon">[[_statusIcon]]</i>
          [[_status]]
        </span>
      </button>
      <chops-collapse class="card-content" opened$="[[opened]]">
        <div class="approver-notice">
          <template is="dom-if" if="[[_isApprover]]">
            You are an approver for this bit.
          </template>
          <template is="dom-if" if="[[user.isSiteAdmin]]">
            Your site admin privileges give you full access to edit this approval.
          </template>
        </div>
        <mr-metadata
          aria-label$="[[fieldName]] Approval Metadata"
          approval-status="[[_status]]"
          approvers="[[approvers]]"
          setter="[[setter]]"
          field-defs="[[fieldDefs]]"
          is-approval="true"
        ></mr-metadata>
        <h4
          class="medium-heading"
          role="heading"
        >
          [[fieldName]] Survey
          <chops-button on-click="_openEditSurvey">
            Edit responses
          </chops-button>
        </h4>
        <mr-description
          class="survey"
          description-list="[[_surveyList]]"
        ></mr-description>
        <mr-comment-list
          heading-level=4
          comments="[[_comments]]"
        >
          <h4 id$="[[_editId]]" class="medium-heading">
            Editing approval: [[phaseName]] &gt; [[fieldName]]
          </h4>
          <mr-edit-metadata
            form-name="[[phaseName]] > [[fieldName]]"
            approvers="[[approvers]]"
            field-defs="[[fieldDefs]]"
            statuses="[[_availableStatuses]]"
            status="[[_status]]"
            has-approver-privileges="[[_hasApproverPrivileges]]"
            is-approval
            disabled="[[updatingApproval]]"
            error="[[updateApprovalError.description]]"
            on-save="save"
            on-discard="reset"
          ></mr-edit-metadata>
        </mr-comment-list>
      </chops-collapse>
    `;
  }

  static get is() {
    return 'mr-approval-card';
  }

  static get properties() {
    return {
      fieldName: String,
      approvers: Array,
      approvalComments: Array,
      phaseName: String,
      setter: Object,
      fieldDefs: {
        type: Array,
        computed:
          '_computeApprovalFieldDefs(fieldDefsByApprovalName, fieldName)',
      },
      fieldDefsByApprovalName: Object,
      focusId: String,
      user: Object,
      issue: {
        type: Object,
        observer: 'reset',
      },
      issueRef: Object,
      projectConfig: Object,
      class: {
        type: String,
        reflectToAttribute: true,
        computed: '_computeClass(_status)',
      },
      comments: Array,
      opened: {
        type: Boolean,
        reflectToAttribute: true,
        value: false,
      },
      statusEnum: {
        type: String,
        value: '',
      },
      statuses: {
        type: Array,
        value: () => {
          return Object.keys(STATUS_CLASS_MAP).map(
            (status) => (
              {status, docstring: STATUS_DOCSTRING_MAP[status], rank: 1}));
        },
      },
      updatingApproval: Boolean,
      updateApprovalError: Object,
      _availableStatuses: {
        type: Array,
        computed: '_filterStatuses(_status, statuses, _hasApproverPrivileges)',
      },
      _comments: {
        type: Array,
        computed: '_filterComments(comments, fieldName)',
      },
      _editId: {
        type: String,
        computed: '_computeEditId(fieldName)',
      },
      _survey: {
        type: Object,
        computed: '_computeSurvey(_surveyList)',
      },
      _surveyList: {
        type: Array,
        computed: '_computeSurveyList(comments, fieldName)',
      },
      _isApprover: {
        type: Boolean,
        computed: '_computeIsApprover(approvers, user.email, user.groups)',
        observer: '_openUserCards',
      },
      _hasApproverPrivileges: {
        type: Boolean,
        computed: `_computeHasApproverPrivileges(user.isSiteAdmin,
          _isApprover)`,
      },
      _expandIcon: {
        type: String,
        computed: '_computeExpandIcon(opened)',
      },
      _status: {
        type: String,
        computed: '_computeStatus(statusEnum)',
      },
      _statusIcon: {
        type: String,
        computed: '_computeStatusIcon(class)',
      },
    };
  }

  static get observers() {
    return [
      '_onFocusId(_comments, focusId)',
    ];
  }

  stateChanged(state) {
    this.setProperties({
      fieldDefsByApprovalName: project.fieldDefsByApprovalName(state),
      focusId: ui.focusId(state),
      user: user.user(state),
      issue: issue.issue(state),
      issueRef: issue.issueRef(state),
      projectConfig: project.project(state).config,
      comments: issue.comments(state),
      updatingApproval: issue.requests(state).updateApproval.requesting,
      updateApprovalError: issue.requests(state).updateApproval.error,
    });
  }

  reset() {
    const form = this.shadowRoot.querySelector('mr-edit-metadata');
    if (!form) return;
    form.reset();
  }

  async save() {
    const form = this.shadowRoot.querySelector('mr-edit-metadata');

    const commentContent = form.getCommentContent();
    const approvalDelta = form.getDelta();
    if (approvalDelta.status) {
      approvalDelta.status = TEXT_TO_STATUS_ENUM[approvalDelta.status];
    }

    const uploads = await form.getAttachments();
    if (commentContent || Object.keys(approvalDelta).length > 0 ||
        uploads.length > 0) {
      store.dispatch(issue.updateApproval({
        issueRef: this.issueRef,
        fieldRef: {
          type: fieldTypes.APPROVAL_TYPE,
          fieldName: this.fieldName,
        },
        sendEmail: form.sendEmail,
        commentContent,
        approvalDelta,
        uploads,
      }));
    }
  }

  toggleCard(evt) {
    this.opened = !this.opened;
  }

  openCard(evt) {
    this.opened = true;

    if (evt && evt.detail && evt.detail.callback) {
      evt.detail.callback();
    }
  }

  _displayNamesToUserRefs(list) {
    return list.map((name) => ({'displayName': name}));
  }

  _computeClass(status) {
    return STATUS_CLASS_MAP[status];
  }

  _computeExpandIcon(opened) {
    if (opened) {
      return 'expand_less';
    }
    return 'expand_more';
  }

  _computeStatus(statusEnum) {
    return STATUS_ENUM_TO_TEXT[statusEnum || ''];
  }

  _computeStatusIcon(cl) {
    return CLASS_ICON_MAP[cl];
  }

  _computeIsApprover(approvers, userEmail, userGroups) {
    if (!userEmail || !approvers) return false;
    userGroups = userGroups || [];
    return !!approvers.find((a) => {
      return a.displayName === userEmail || userGroups.find(
        (group) => group.displayName === a.displayName
      );
    });
  }

  _computeHasApproverPrivileges(isSiteAdmin, isApprover) {
    return isSiteAdmin || isApprover;
  }

  // TODO(zhangtiff): Change data flow here so that this is only computed
  // once for all approvals.
  _filterComments(comments, fieldName) {
    if (!comments || !fieldName) return;
    return comments.filter((c) => (
      c.approvalRef && c.approvalRef.fieldName === fieldName
    )).splice(1);
  }

  _computeApprovalFieldDefs(fdMap, approvalName) {
    if (!fdMap) return [];
    return fdMap.get(approvalName) || [];
  }

  _computeEditId(fieldName) {
    return `edit${fieldName}`;
  }

  _computeSurvey(surveyList) {
    if (!surveyList || !surveyList.length) return;
    return surveyList[surveyList.length - 1];
  }

  // TODO(zhangtiff): Change data flow here so that this is only computed
  // once for all approvals.
  _computeSurveyList(comments, fieldName) {
    if (!comments || !fieldName) return;
    return comments.filter((comment) => comment.approvalRef
        && comment.approvalRef.fieldName === fieldName
        && comment.descriptionNum);
  }

  _filterStatuses(status, statuses, hasApproverPrivileges) {
    return statuses.filter((s) => {
      if (s.status === status) {
        // The current status should always appear as an option.
        return true;
      }

      if (!hasApproverPrivileges
          && APPROVER_RESTRICTED_STATUSES.has(s.status)) {
        // If you are not an approver and and this status is restricted,
        // you can't change to this status.
        return false;
      }

      // No one can set statuses to NotSet, not even approvers.
      return s.status !== 'NotSet';
    });
  }

  _openUserCards(isApprover) {
    if (!this.opened && isApprover) {
      this.opened = true;
    }
  }

  _toString(bool) {
    return bool.toString();
  }

  _openEditSurvey() {
    this.dispatchEvent(new CustomEvent('open-dialog', {
      bubbles: true,
      composed: true,
      detail: {
        dialogId: 'edit-description',
        fieldName: this.fieldName,
      },
    }));
  }

  _onFocusId(comments, focusId) {
    for (const comment of (comments || [])) {
      const commentId = 'c' + comment.sequenceNum;
      if (commentId === focusId) {
        this.opened = true;
        break;
      }
    }
  }
}

customElements.define(MrApprovalCard.is, MrApprovalCard);
