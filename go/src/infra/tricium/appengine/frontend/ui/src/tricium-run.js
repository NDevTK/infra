// Copyright 2018 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

'use strict';

import { PolymerElement, html } from '@polymer/polymer/polymer-element.js';
import '@polymer/polymer/lib/elements/dom-if.js';
import '@polymer/iron-ajax/iron-ajax.js'

class TriciumRun extends PolymerElement {

  static get properties() {
    return {
      runId: String,
      content: String,
      error: String,
    };
  }

  ready() {
    super.ready();
    if (this.runId) {
      this._refresh();
    }
  }

  _refresh() {
    this.$.ajax.body = JSON.stringify({runId: this.runId});
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
        url="/prpc/tricium.Tricium/Progress"
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

customElements.define('tricium-run', TriciumRun);
