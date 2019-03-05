// Copyright 2018 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

'use strict';

import {LitElement, html} from '@polymer/lit-element/lit-element.js';
import {repeat} from 'lit-html/lib/repeat.js';

import {request} from './prpc.js';

class TriciumRun extends LitElement {
  static get properties() {
    return {
      run: String,
      data: Object,
      error: String,
    };
  }

  _render({run, data, error}) {
    if (error || !data || !data.runId) {
      return html`<p style="color:red">${error}</p>`;
    }
    return html`
      <p>
        <b>Run ID: ${data.runId}</b> (State: ${data.state})
      </p>
      ${repeat(data.functionProgress, (f) => f.name, this._renderFunction)}
    `;
  }

  _renderFunction(f) {
    return html`
      <b>${f.name}</b>
        (State: ${f.state || 'PENDING'},
        ${this._renderLink(f)},
        comments: ${f.numComments || 0})
      </p>
    `;
  }

  _renderLink(f) {
    if (f.swarmingTaskId) {
      return html`
        <a href$=${f.swarmingUrl}/task?id=${f.swarmingTaskId}
          task ${f.swarmingTaskId}
        </a>`;
    } else if (f.buildbucketBuildId) {
      return html`
        <a href$=https://${this._miloHost(f.buildbucketHost)}/b/${f.buildbucketBuildId}
          build ${f.buildbucketBuildId}
        </a>`;
    }
    return html`no link`;
  }

  _miloHost(buildbucketHost) {
    if (buildbucketHost == 'cr-buildbucket-dev.appspot.com') {
      return 'luci-milo-dev.appspot.com';
    }
    return 'ci.chromium.org';
  }

  connectedCallback() {
    super.connectedCallback();
    if (!this.run) {
      console.warn('No run set on tricium-run');
    }
    this._refresh();
  }

  async _refresh() {
    try {
      this.data = await request('Progress', {runId: this.run});
    } catch (error) {
      this.error = error.message;
    }
  }
}

customElements.define('tricium-run', TriciumRun);
