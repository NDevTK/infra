'use strict';

/**
 * `<mr-launch-overview>` ....
 *
 *   Element description here.
 *
 * @customElement
 * @polymer
 * @demo
 */
class MrLaunchOverview extends Polymer.Element {
  static get is() {
    return 'mr-launch-overview';
  }

  static get properties() {
    return {
      gates: {
        type: Array,
        value: [],
      },
    };
  }

  _handleGateFocus(e) {
    let idx = e.target.getAttribute('value') * 1;
    this.dispatchEvent(
      new CustomEvent('gate-selected', {detail: {gateIndex: idx}}));
  }
}
customElements.define(MrLaunchOverview.is, MrLaunchOverview);
