'use strict';

import { PolymerElement, html } from '@polymer/polymer/polymer-element.js';
import '@polymer/iron-ajax/iron-ajax.js'

class TriciumFeedback extends PolymerElement {

  static get properties() {
    return {
      category: String,
      startTime: Date,
      endTime: Date,
      content: String,
      error: String
    };
  }

  ready() {
    super.ready();
    if (!this.category) {
      this.category = "Spacey";
    }
    this._refresh();
  }

  _refresh() {
    this.$.ajax.body = JSON.stringify({category: this.category});
    this.$.ajax.generateRequest();
  }

  _showResponse(event) {
    this.content = JSON.stringify(event.target.lastResponse, null, 2);
  }

  _showError(event) {
    this.error = event.detail.error;
  }

  static get template() {
    return html`
      <iron-ajax
        id="ajax"
        url="/prpc/tricium.Tricium/Feedback"
        method="POST"
        content-type="application/json"
        accept="json"
        json-prefix=")]}'"
        on-response="_showResponse"
        on-error="_showError"
        handle-as="json"
        last-response="{{ajaxResponse}}">
      </iron-ajax>
      <pre>[[content]]</pre>
      <p style="color:red">[[error]]</p>
    `;
  }
}

customElements.define('tricium-feedback', TriciumFeedback);
