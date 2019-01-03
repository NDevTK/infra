/* Copyright 2018 The Chromium Authors. All Rights Reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file or at
 * https://developers.google.com/open-source/licenses/bsd
 */

// TODO(jojwang): Move all useful FLT const to a shared file.
const TEXT_TO_STATUS_ENUM = {
  'NeedsReview': 'NEEDS_REVIEW',
  'NA': 'NA',
  'ReviewRequested': 'REVIEW_REQUESTED',
  'ReviewStarted': 'REVIEW_STARTED',
  'NeedInfo': 'NEED_INFO',
  'Approved': 'APPROVED',
  'NotApproved': 'NOT_APPROVED',
};

const NO_UPDATES_MESSAGE = 'User lacks approver perms for approval in all issues.'

class MrBulkApprovalUpdate extends Polymer.Element {
  static get is() {
    return 'mr-bulk-approval-update';
  }

  static get properties() {
    return {
      approvals: {
        type: Array,
        value: () => [],
      },
      issueRefs: {
        type: Array,
        computed: '_computeIssueRefs(projectName, localIdsStr)',
      },
      localIdsStr: String,
      projectName: String,
      responseMessage: String,
      statusOptions: {
        type: Array,
        value: () => {
          return Object.keys(TEXT_TO_STATUS_ENUM);
        },
      },
    };
  }

  _computeIssueRefs(projectName, localIdsStr) {
    if (!projectName || !localIdsStr) return [];
    let issueRefs = [];
    let localIds = localIdsStr.split(',');
    localIds.forEach(localId => {
      issueRefs.push({projectName: projectName, localId: localId});
    })
    return issueRefs;
  }

  fetchApprovals(evt) {
    let message = {issueRefs: this.issueRefs};
    window.prpcClient.call('monorail.Issues', 'ListApplicableFieldDefs', message).then(
        resp => {
          const root = Polymer.dom(this.root);
          resp.fieldDefs.forEach(fieldDef => {
            if (fieldDef.fieldRef.type == 'APPROVAL_TYPE') {
              this.push('approvals', fieldDef);
            }
          })
          if (!this.approvals.length) {
            root.querySelector('#js-noApprovals').classList.remove('hidden');
          }
          root.querySelector('#js-showApprovals').classList.add('hidden');
        })
  }

  save(evt) {
    this.responseMessage = '';
    this.errorMessage = '';
    this.toggleDisableForm();
    const root = Polymer.dom(this.root);
    let selectedFieldDef = this.approvals.find(
        approval => {
          return approval.fieldRef.fieldName == root.querySelector('#approvalSelect').value;
        }
    );
    let message = {
      issueRefs: this.issueRefs,
      fieldRef: selectedFieldDef.fieldRef,
      send_email: true,
    }
    const commentContent = root.querySelector('#commentText').value;
    if (commentContent) {
      message.commentContent = commentContent;
    }
    let delta = {};
    const statusInput = root.querySelector('#statusInput');
    if (statusInput.value != '---') {
      delta.status = TEXT_TO_STATUS_ENUM[statusInput.value];
    }
    const approversInput = root.querySelector('#approversInput');
    let approversAdded = approversInput.getValuesAdded()
    if (approversAdded.length) {
      delta.approverRefsAdd = approversAdded.map(name => ({'displayName': name}));
    }
    if (Object.keys(delta).length) {
      message.approvalDelta = delta;
    }
    window.prpcClient.call('monorail.Issues', 'BulkUpdateApprovals', message).then(
        resp => {
          if (resp.issueRefs && resp.issueRefs.length) {
            let idsStr = Array.from(resp.issueRefs, ref => ref.localId).join(', ')
            this.responseMessage = `${this.getTimeStamp()}: Updated ${selectedFieldDef.fieldRef.fieldName} in issues: ${idsStr} (${resp.issueRefs.length} of ${this.issueRefs.length}).`;
            this.clearForm();
          } else {
            this.errorMessage = NO_UPDATES_MESSAGE;
          };
          this.toggleDisableForm();
        }, (error) => {
          this.errorMessage = error;
          this.toggleDisableForm();
        });
  }

  getTimeStamp() {
    let date = new Date();
    return `${date.getHours()}:${date.getMinutes()}:${date.getSeconds()}`
  }

  clearForm() {
    const root = Polymer.dom(this.root);
    root.querySelectorAll('input, textarea').forEach(input => {
      input.value = '';
    });
    root.querySelectorAll('select').forEach(select => {
      select.selectedIndex = 0;
    });
  }

  toggleDisableForm() {
    const root = Polymer.dom(this.root);
    root.querySelectorAll('input, textarea, select').forEach(input => {
      input.disabled = !input.disabled;
    })
  }
}

customElements.define(MrBulkApprovalUpdate.is, MrBulkApprovalUpdate);
