'use strict';

/**
 * `<mr-approval-card>` ....
 *
 *   Element description here.
 *
 * @customElement
 * @polymer
 * @demo
 */
class MrApprovalCard extends Polymer.Element {
  static get is() {
    return 'mr-approval-card';
  }

  ready() {
    super.ready();
  }

  static get properties() {
    return {
      teamTitle: String,
      approvalComments: Array,
      survey: String,
      urls: Array,
      labels: Array,
      users: Array,
      _surveyLines: {
        type: Array,
        value: [],
        computed: '_computeSurveyLines(survey)',
      },
    };
  }

  editData() {
    this.$.editDialog.open();
  }

  toggleCard(evt) {
    let path = evt.path;
    for (let i = 0; i < path.length; i++) {
      let itm = path[i];
      if (itm.classList && itm.classList.contains('no-toggle')) {
        return;
      }
    }
    this.$.cardCollapse.toggle();
  }

  _computeSurveyLines(survey) {
    return survey.trim().split(/\r?\n/);
  }
}
customElements.define(MrApprovalCard.is, MrApprovalCard);
