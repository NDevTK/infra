// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is govered by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

'use strict';

/**
 * `<chops-checkbox>`
 *
 * A checkbox component.
 *
 */
class ChopsCheckbox extends Polymer.Element {
  static get is() {
    return 'chops-checkbox';
  }

  static get properties() {
    return {
      label: String,
      checked: Boolean,
    };
  }

  _checkedChange(e) {
    const checked = e.target.checked;
    const customEvent = new CustomEvent('checked-change', {
      detail: {
        checked: checked,
      },
    });
    this.dispatchEvent(customEvent);
  }
}
customElements.define(ChopsCheckbox.is, ChopsCheckbox);
