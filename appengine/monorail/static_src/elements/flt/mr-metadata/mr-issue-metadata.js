// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@polymer/polymer/polymer-legacy.js';
import {PolymerElement, html} from '@polymer/polymer';
import '../../chops/chops-timestamp/chops-timestamp.js';
import {ReduxMixin, actionType} from '../../redux/redux-mixin.js';
import * as issue from '../../redux/issue.js';
import * as project from '../../redux/project.js';
import '../../mr-user-link/mr-user-link.js';
import '../../shared/mr-shared-styles.js';
import './mr-metadata.js';

/**
 * `<mr-issue-metadata>`
 *
 * The metadata view for a single issue. Contains information such as the owner.
 *
 */
class MrIssueMetadata extends ReduxMixin(PolymerElement) {
  static get template() {
    return html`
      <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">
      <style include="mr-shared-styles">
        :host {
          box-sizing: border-box;
          padding: 0.5em 8px;
          max-width: 100%;
          display: block;
        }
        a.label {
          color: hsl(120, 100%, 25%);
          text-decoration: none;
        }
        a.label[data-derived] {
          font-style: italic;
        }
        .restricted {
          background: hsl(30, 100%, 93%);
          border: var(--chops-normal-border);
          width: 100%;
          box-sizing: border-box;
          padding: 0.5em 8px;
          margin: 1em auto;
        }
        .restricted i.material-icons {
          color: hsl(30, 5%, 39%);
          display: block;
          margin-right: 4px;
          margin-bottom: 4px;
        }
        .restricted strong {
          display: flex;
          align-items: center;
          justify-content: center;
          text-align: center;
          width: 100%;
          margin-bottom: 0.5em;
        }
        .star-line {
          width: 100%;
          text-align: center;
          display: flex;
          align-items: center;
          justify-content: center;
        }
        /* Wrap the star icon around a button for accessibility. */
        .star-line button {
          background: none;
          border: none;
          cursor: pointer;
          padding: 0;
          margin: 0;
          margin-right: 4px;
        }
        .star-line button[disabled] {
          opacity: 0.5;
          cursor: default;
        }
        .star-line i.material-icons {
          color: hsl(120, 5%, 66%);
        }
        .star-line i.material-icons.starred {
          color: cornflowerblue;
        }
      </style>
      <div class="star-line">
        <button on-click="toggleStar" disabled\$="[[!_canStar]]">
          <template is="dom-if" if="[[isStarred]]">
            <i class="material-icons starred" title="You've starred this issue">star</i>
          </template>
          <template is="dom-if" if="[[!isStarred]]">
            <i class="material-icons" title="Click to star this issue">star_border</i>
          </template>
        </button>
        Starred by [[_renderCount(issue.starCount)]] user[[_renderPluralS(issue.starCount)]]
      </div>
      <mr-metadata
        aria-label="Issue Metadata"
        owner="[[issue.ownerRef]]"
        cc="[[issue.ccRefs]]"
        issue-status="[[issue.statusRef]]"
        components="[[_components]]"
        field-defs="[[_fieldDefs]]"
        blocked-on="[[issue.blockedOnIssueRefs]]"
        blocking="[[issue.blockingIssueRefs]]"
        merged-into="[[issue.mergedIntoIssueRef]]"
        modified-timestamp="[[issue.modifiedTimestamp]]"
      ></mr-metadata>

      <div class="labels-container">
        <template is="dom-repeat" items="[[issue.labelRefs]]" as="label">
          <a href\$="/p/[[projectName]]/issues/list?q=label:[[label.label]]" class="label" data-derived\$="[[label.isDerived]]">[[label.label]]</a>
          <br>
        </template>
      </div>
    `;
  }

  static get is() {
    return 'mr-issue-metadata';
  }

  static get properties() {
    return {
      issue: Object,
      issueId: Number,
      projectName: String,
      projectConfig: String,
      isStarred: {
        type: Boolean,
        value: false,
      },
      fetchingIsStarred: Boolean,
      starringIssue: Boolean,
      _components: Array,
      _fieldDefs: Array,
      _canStar: {
        type: Boolean,
        computed: '_computeCanStar(fetchingIsStarred, starringIssue)',
      },
      _type: String,
    };
  }

  static mapStateToProps(state, element) {
    return {
      issue: state.issue,
      issueId: state.issueId,
      projectName: state.projectName,
      projectConfig: project.project(state).config,
      isStarred: state.isStarred,
      fetchingIsStarred: state.requests.fetchIsStarred.requesting,
      starringIssue: state.requests.starIssue.requesting,
      _components: issue.components(state),
      _fieldDefs: issue.fieldDefs(state),
      _type: issue.type(state),
    };
  }

  edit() {
    this.dispatchAction({
      type: actionType.OPEN_DIALOG,
      dialog: DialogState.EDIT_ISSUE,
    });
  }

  toggleStar() {
    if (!this._canStar) return;

    const newIsStarred = !this.isStarred;
    const issueRef = {
      projectName: this.projectName,
      localId: this.issueId,
    };

    this.dispatchAction(issue.starIssue(issueRef, newIsStarred));
  }

  _computeCanStar(fetching, starring) {
    return !(fetching || starring);
  }

  _renderPluralS(count) {
    return count == 1 ? '' : 's';
  }

  _renderCount(count) {
    return count ? count : 0;
  }
}

customElements.define(MrIssueMetadata.is, MrIssueMetadata);
