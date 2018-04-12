'use strict';

/**
 * `<mr-gate>` ....
 *
 *   Element description here.
 *
 * @customElement
 * @polymer
 * @demo
 */
class MrGate extends Polymer.Element {
  static get is() {
    return 'mr-gate';
  }

  static get properties() {
    return {
      approvals: {
        type: Array,
        value: [],
      },
    };
  }
}
customElements.define(MrGate.is, MrGate);
