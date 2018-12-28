const STATUS_ENUM_TO_TEXT = {
  'NEEDS_REVIEW': 'NeedsReview',
  'NA': 'NA',
  'REVIEW_REQUESTED': 'ReviewRequested',
  'REVIEW_STARTED': 'ReviewStarted',
  'NEED_INFO': 'NeedInfo',
  'APPROVED': 'Approved',
  'NOT_APPROVED': 'NotApproved',
};

const TEXT_TO_STATUS_ENUM = {
  'NeedsReview': 'NEEDS_REVIEW',
  'NA': 'NA',
  'ReviewRequested': 'REVIEW_REQUESTED',
  'ReviewStarted': 'REVIEW_STARTED',
  'NeedInfo': 'NEED_INFO',
  'Approved': 'APPROVED',
  'NotApproved': 'NOT_APPROVED',
};

class MrBulkApprovalUpdate extends Polymer.Element {
  static get is() {
    return 'mr-bulk-approval-update';
  }

  static get properties() {
    return {
      projectName: String,
      localIdsStr: String,
      issueRefs: {
	type: Array,
	computed: '_computeIssueRefs(projectName, localIdsStr)',
      },
      approvals: {
	type: Array,
	value: () => [],
      },
      statusOptions: {
	type: Array,
	value: () => {
	  return Object.values(STATUS_ENUM_TO_TEXT);
	},
      },
      updatedIssueRefs: {
	type: Array,
	value: () => [],
      }
    }
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
      delta.approversAdded = approversAdded;
    }
    if (Object.keys(delta).length) {
      message.approvalDelta = delta;
    }
    window.prpcClient.call('monorail.Issues', 'BulkUpdateApprovals', message).then(
	resp => {
	  this.set('updatedIssueRefs', resp.issueRefs);
	}
    );
  }
}

customElements.define(MrBulkApprovalUpdate.is, MrBulkApprovalUpdate);
