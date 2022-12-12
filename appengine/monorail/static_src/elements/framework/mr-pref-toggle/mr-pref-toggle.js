// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {LitElement, html} from 'lit-element';

import {store, connectStore} from 'reducers/base.js';
import * as userV0 from 'reducers/userV0.js';
import * as projectV0 from 'reducers/projectV0.js';
import 'elements/chops/chops-toggle/chops-toggle.js';
import {logEvent} from 'monitoring/client-logger.js';

/**
 * `<mr-pref-toggle>`
 *
 * Toggle button for any user pref, including code font and
 * rendering markdown.  For our purposes, pressing it causes
 * issue description and comment text to switch either to
 * monospace font or to render in markdown and the setting
 * is saved in the user's preferences.
 */
export class MrPrefToggle extends connectStore(LitElement) {
  /** @override */
  render() {
    return html`
        <chops-toggle
          ?checked=${this._checked}
          ?disabled=${this._prefsInFlight}
          @checked-change=${this._togglePref}
          title=${this.title}
        >${this.label}</chops-toggle>
      `;
  }

  /** @override */
  static get properties() {
    return {
      prefs: {type: Object},
      userDisplayName: {type: String},
      initialValue: {type: Boolean},
      _prefsInFlight: {type: Boolean},
      label: {type: String},
      title: {type: String},
      prefName: {type: String},
    };
  }

  /** @override */
  stateChanged(state) {
    this.prefs = userV0.prefs(state);
    this._prefsInFlight = userV0.requests(state).fetchPrefs.requesting ||
      userV0.requests(state).setPrefs.requesting;
    this._projectName = projectV0.viewedProjectName(state);
  }

  /** @override */
  constructor() {
    super();
    this.initialValue = false;
    this.userDisplayName = '';
    this.label = '';
    this.title = '';
    this.prefName = '';
    this._projectName = '';
  }

  // Used by the legacy EZT page to interact with Redux.
  fetchPrefs() {
    store.dispatch(userV0.fetchPrefs());
  }

  get _checked() {
    const {prefs, initialValue} = this;
    if (prefs && prefs.has(this.prefName)) return prefs.get(this.prefName);
    return initialValue;
  }

  /**
   * Toggles the code font in response to the user activating the button.
   * @param {Event} e
   * @fires CustomEvent#font-toggle
   * @private
   */
  _togglePref(e) {
    const checked = e.detail.checked;
    this.dispatchEvent(new CustomEvent('font-toggle', {detail: {checked}}));

    const newPrefs = [{name: this.prefName, value: '' + checked}];
    store.dispatch(userV0.setPrefs(newPrefs, !!this.userDisplayName));

    logEvent('mr-pref-toggle', `${this.prefName}: ${checked}`, this._projectName);
  }
}
customElements.define('mr-pref-toggle', MrPrefToggle);
