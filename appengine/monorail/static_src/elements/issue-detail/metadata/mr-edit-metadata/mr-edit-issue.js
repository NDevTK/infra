// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {LitElement, html, css} from 'lit-element';
import debounce from 'debounce';

import {store, connectStore} from 'reducers/base.js';
import * as issueV0 from 'reducers/issueV0.js';
import * as projectV0 from 'reducers/projectV0.js';
import * as ui from 'reducers/ui.js';
import {arrayToEnglish} from 'shared/helpers.js';
import './mr-edit-metadata.js';
import 'shared/typedef.js';
import {migratedTypes} from 'shared/issue-fields.js';
import ClientLogger from 'monitoring/client-logger.js';

const DEBOUNCED_PRESUBMIT_TIME_OUT = 400;

/**
 * `<mr-edit-issue>`
 *
 * Edit form for a single issue. Wraps <mr-edit-metadata>.
 *
 */
export class MrEditIssue extends connectStore(LitElement) {
  /** @override */
  render() {
    const issue = this.issue || {};
    let blockedOnRefs = issue.blockedOnIssueRefs || [];
    if (issue.danglingBlockedOnRefs && issue.danglingBlockedOnRefs.length) {
      blockedOnRefs = blockedOnRefs.concat(issue.danglingBlockedOnRefs);
    }

    let blockingRefs = issue.blockingIssueRefs || [];
    if (issue.danglingBlockingRefs && issue.danglingBlockingRefs.length) {
      blockingRefs = blockingRefs.concat(issue.danglingBlockingRefs);
    }

    let migratedNotice = html``;
    if (this._isMigrated) {
      migratedNotice = html`
        <div class="migrated-banner">
          <i
            class="warning-icon material-icons"
            icon="warning"
          >warning</i>
          ${this._migratedLink}
        </div>
        <chops-button
          class="legacy-edit"
          @click=${this._allowLegacyEdits}
        >
          I want to edit the old version of this issue.
        </chops-button>
      `;
    }

    return html`
      <link href="https://fonts.googleapis.com/icon?family=Material+Icons"
        rel="stylesheet">
      <style>
        mr-edit-issue .migrated-banner {
          width: 100%;
          background-color: var(--chops-orange-50);
          border: var(--chops-normal-border);
          border-top: 0;
          font-size: var(--chops-main-font-size);
          padding: 0.25em 8px;
          box-sizing: border-box;
          display: flex;
          flex-direction: row;
          justify-content: flex-start;
          align-items: center;
          margin-bottom: 1em;
        }
        mr-edit-issue i.material-icons {
          color: var(--chops-primary-icon-color);
          font-size: var(--chops-icon-font-size);
        }
        mr-edit-issue .warning-icon {
          margin-right: 4px;
        }
        mr-edit-issue .legacy-edit {
          margin-bottom: 2em;
        }
      </style>
      <h2 id="makechanges" class="medium-heading">
        <a href="#makechanges">Add a comment and make changes</a>
      </h2>

      ${migratedNotice}

      <mr-edit-metadata
        ?hidden=${this._isMigrated && !this._editLegacyIssue}
        formName="Issue Edit"
        .ownerName=${this._ownerDisplayName(this.issue.ownerRef)}
        .cc=${issue.ccRefs}
        .status=${issue.statusRef && issue.statusRef.status}
        .statuses=${this._availableStatuses(this.projectConfig.statusDefs, this.issue.statusRef)}
        .summary=${issue.summary}
        .components=${issue.componentRefs}
        .fieldDefs=${this._fieldDefs}
        .fieldValues=${issue.fieldValues}
        .blockedOn=${blockedOnRefs}
        .blocking=${blockingRefs}
        .mergedInto=${issue.mergedIntoIssueRef}
        .labelNames=${this._labelNames}
        .derivedLabels=${this._derivedLabels}
        .error=${this.updateError}
        ?saving=${this.updatingIssue}
        @save=${this.save}
        @discard=${this.reset}
        @change=${this._onChange}
      ></mr-edit-metadata>
    `;
  }

  /** @override */
  static get properties() {
    return {
      /**
       * ID of an Issue Tracker issue that the issue migrated to.
       */
      migratedId: {
        type: String,
      },
      /**
       * Type of the issue migrated to.
       */
       migratedType: {
        type: migratedTypes,
      },
      /**
       * All comments, including descriptions.
       */
      comments: {
        type: Array,
      },
      /**
       * The issue being updated.
       */
      issue: {
        type: Object,
      },
      /**
       * The issueRef for the currently viewed issue.
       */
      issueRef: {
        type: Object,
      },
      /**
       * The config of the currently viewed project.
       */
      projectConfig: {
        type: Object,
      },
      /**
       * The Name of the currently viewed project.
       */
      projectName: {
        type: String,
      },
      /**
       * Whether the issue is currently being updated.
       */
      updatingIssue: {
        type: Boolean,
      },
      /**
       * An error response, if one exists.
       */
      updateError: {
        type: String,
      },
      /**
       * Hash from the URL, used to support the 'r' hot key for making changes.
       */
      focusId: {
        type: String,
      },
      _fieldDefs: {
        type: Array,
      },
      _editLegacyIssue: {
        type: Boolean,
      },
    };
  }

  /** @override */
  constructor() {
    super();

    this.clientLogger = new ClientLogger('issues');
    this.updateError = '';

    this.presubmitDebounceTimeOut = DEBOUNCED_PRESUBMIT_TIME_OUT;

    this._editLegacyIssue = false;
  }

  /** @override */
  createRenderRoot() {
    return this;
  }

  /** @override */
  disconnectedCallback() {
    super.disconnectedCallback();

    // Prevent debounced logic from running after the component has been
    // removed from the UI.
    if (this._debouncedPresubmit) {
      this._debouncedPresubmit.clear();
    }
  }

  /** @override */
  stateChanged(state) {
    this.migratedId = issueV0.migratedId(state);
    this.migratedType = issueV0.migratedType(state);
    this.issue = issueV0.viewedIssue(state);
    this.issueRef = issueV0.viewedIssueRef(state);
    this.comments = issueV0.comments(state);
    this.projectConfig = projectV0.viewedConfig(state);
    this.projectName = projectV0.viewedProjectName(state);
    this.updatingIssue = issueV0.requests(state).update.requesting;

    const error = issueV0.requests(state).update.error;
    this.updateError = error && (error.description || error.message);
    this.focusId = ui.focusId(state);
    this._fieldDefs = issueV0.fieldDefs(state);
  }

  /** @override */
  updated(changedProperties) {
    if (this.focusId && changedProperties.has('focusId')) {
      // TODO(zhangtiff): Generalize logic to focus elements based on ID
      // to a reuseable class mixin.
      if (this.focusId.toLowerCase() === 'makechanges') {
        this.focus();
      }
    }

    if (changedProperties.has('updatingIssue')) {
      const isUpdating = this.updatingIssue;
      const wasUpdating = changedProperties.get('updatingIssue');

      // When an issue finishes updating, we want to show a snackbar, record
      // issue update time metrics, and reset the edit form.
      if (!isUpdating && wasUpdating) {
        if (!this.updateError) {
          this._showCommentAddedSnackbar();
          // Reset the edit form when a user's action finishes.
          this.reset();
        }

        // Record metrics on when the issue editing event finished.
        if (this.clientLogger.started('issue-update')) {
          this.clientLogger.logEnd('issue-update', 'computer-time', 120 * 1000);
        }
      }
    }
  }

  // TODO(crbug.com/monorail/6933): Remove the need for this wrapper.
  /**
   * Snows a snackbar telling the user they added a comment to the issue.
   */
  _showCommentAddedSnackbar() {
    store.dispatch(ui.showSnackbar(ui.snackbarNames.ISSUE_COMMENT_ADDED,
        'Your comment was added.'));
  }

  /**
   * Resets all form fields to their initial values.
   */
  reset() {
    const form = this.querySelector('mr-edit-metadata');
    if (!form) return;
    form.reset();
  }

  /**
   * Dispatches an action to save issue changes on the server.
   */
  async save() {
    const form = this.querySelector('mr-edit-metadata');
    if (!form) return;

    const delta = form.delta;
    if (!allowRemovedRestrictions(delta.labelRefsRemove)) {
      return;
    }

    const message = {
      issueRef: this.issueRef,
      delta: delta,
      commentContent: form.getCommentContent(),
      sendEmail: form.sendEmail,
    };

    // Add files to message.
    const uploads = await form.getAttachments();

    if (uploads && uploads.length) {
      message.uploads = uploads;
    }

    if (message.commentContent || message.delta || message.uploads) {
      this.clientLogger.logStart('issue-update', 'computer-time');

      store.dispatch(issueV0.update(message));
    }
  }

  /**
   * Focuses the edit form in response to the 'r' hotkey.
   */
  focus() {
    const editHeader = this.querySelector('#makechanges');
    editHeader.scrollIntoView();

    const editForm = this.querySelector('mr-edit-metadata');
    editForm.focus();
  }

  /**
   * Turns all LabelRef Objects attached to an issue into an Array of strings
   * containing only the names of those labels that aren't derived.
   * @return {Array<string>} Array of label names.
   */
  get _labelNames() {
    if (!this.issue || !this.issue.labelRefs) return [];
    const labels = this.issue.labelRefs;
    return labels.filter((l) => !l.isDerived).map((l) => l.label);
  }

  /**
   * Finds only the derived labels attached to an issue and returns only
   * their names.
   * @return {Array<string>} Array of label names.
   */
  get _derivedLabels() {
    if (!this.issue || !this.issue.labelRefs) return [];
    const labels = this.issue.labelRefs;
    return labels.filter((l) => l.isDerived).map((l) => l.label);
  }

  /**
   * @return {boolean} Whether this issue is migrated or not.
   */
  get _isMigrated() {
    return this.migratedId && this.migratedId !== '';
  }

  /**
   * @return {string} the link of the issue in Issue Tracker or Launch.
   */
   get _migratedLink() {
    if (this.migratedType === migratedTypes.BUGANIZER_TYPE) {
      const link_url = this.projectName === 'chromium' ? 'issues.chromium.org' : 'issuetracker.google.com';
      const link =
        html`<a href="https://${link_url}/issues/${this.migratedId}">b/${this.migratedId}</a>`;
      return html`<p>This issue has moved to ${link}. Updates should be posted in ${link}.</p>`;
    } else {
      return html`<p>This issue has been migrated to Launch, see link in final comment below.</p>`;
    }
  }

  /**
   * Let the user override th edit form being hidden, in case of mistakes or
   * similar.
   */
  _allowLegacyEdits() {
    this._editLegacyIssue = true;
  }

  /**
   * Gets the displayName of the owner. Only uses the displayName if a
   * userId also exists in the ref.
   * @param {UserRef} ownerRef The owner of the issue.
   * @return {string} The name of the owner for the edited issue.
   */
  _ownerDisplayName(ownerRef) {
    return (ownerRef && ownerRef.userId) ? ownerRef.displayName : '';
  }

  /**
   * Dispatches an action against the server to run "issue presubmit", a feature
   * that warns the user about issue changes that violate configured rules.
   * @param {Object=} issueDelta Changes currently present in the edit form.
   * @param {string} commentContent Text the user is inputting for a comment.
   */
  _presubmitIssue(issueDelta = {}, commentContent) {
    // Don't run this functionality if the element has disconnected. Important
    // for preventing debounced code from running after an element no longer
    // exists.
    if (!this.isConnected) return;

    if (Object.keys(issueDelta).length || commentContent) {
      // TODO(crbug.com/monorail/8638): Make filter rules actually process
      // the text for comments on the backend.
      store.dispatch(issueV0.presubmit(this.issueRef, issueDelta));
    }
  }

  /**
   * Form change handler that runs presubmit on the form.
   * @param {CustomEvent} evt
   */
  _onChange(evt) {
    const {delta, commentContent} = evt.detail || {};

    if (!this._debouncedPresubmit) {
      this._debouncedPresubmit = debounce(
          (delta, commentContent) => this._presubmitIssue(delta, commentContent),
          this.presubmitDebounceTimeOut);
    }
    this._debouncedPresubmit(delta, commentContent);
  }

  /**
   * Creates the list of statuses that the user sees in the status dropdown.
   * @param {Array<StatusDef>} statusDefsArg The project configured StatusDefs.
   * @param {StatusRef} currentStatusRef The status that the issue currently
   *   uses. Note that Monorail supports free text statuses that do not exist in
   *   a project config. Because of this, currentStatusRef may not exist in
   *   statusDefsArg.
   * @return {Array<StatusRef|StatusDef>} Array of statuses a user can edit this
   *   issue to have.
   */
  _availableStatuses(statusDefsArg, currentStatusRef) {
    let statusDefs = statusDefsArg || [];
    statusDefs = statusDefs.filter((status) => !status.deprecated);
    if (!currentStatusRef || statusDefs.find(
        (status) => status.status === currentStatusRef.status)) {
      return statusDefs;
    }
    return [currentStatusRef, ...statusDefs];
  }
}

/**
 * Asks the user for confirmation when they try to remove retriction labels.
 * eg. Restrict-View-Google.
 * @param {Array<LabelRef>} labelRefsRemoved The labels a user is removing
 *   from this issue.
 * @return {boolean} Whether removing these labels is okay. ie: true if there
 *   are either no restrictions being removed or if the user approved the
 *   removal of the restrictions.
 */
export function allowRemovedRestrictions(labelRefsRemoved) {
  if (!labelRefsRemoved) return true;
  const removedRestrictions = labelRefsRemoved
      .map(({label}) => label)
      .filter((label) => label.toLowerCase().startsWith('restrict-'));
  const removeRestrictionsMessage =
    'You are removing these restrictions:\n' +
    arrayToEnglish(removedRestrictions) + '\n' +
    'This might allow more people to access this issue. Are you sure?';
  return !removedRestrictions.length || confirm(removeRestrictionsMessage);
}

customElements.define('mr-edit-issue', MrEditIssue);
