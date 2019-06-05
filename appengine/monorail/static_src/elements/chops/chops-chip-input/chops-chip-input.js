// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {LitElement, html, css} from 'lit-element';
import deepEqual from 'deep-equal';
import 'elements/chops/chops-chip/chops-chip.js';
import {immutableSplice} from 'elements/shared/helpers.js';

const DELIMITER_REGEX = /[,;\s]+/;

/**
 * `<chops-chip-input>`
 *
 * A chip input.
 *
 */
export class ChopsChipInput extends LitElement {
  static get styles() {
    return css`
      :host {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        justify-content: flex-start;
        border-bottom: var(--chops-accessible-border);
      }
      :host([hidden]) {
        display: none;
      }
      :host([focused]) {
        border-bottom: 1px solid var(--chops-primary-accent-color);
      }
      .immutable {
        font-style: italic;
      }
      chops-chip {
        flex-grow: 0;
        flex-shrink: 0;
        margin: 4px 0;
        margin-right: 4px;
      }
      chops-chip[focusable] {
        cursor: pointer;
      }
      input {
        flex-grow: 1;
        border: 0;
        outline: none;
        /* Give inputs the same vertical sizing styles as chips. */
        padding: 0.1em 4px;
        line-height: 140%;
        margin: 0 2px;
        font-size: var(--chops-main-font-size);
      }
    `;
  }

  render() {
    return html`
      ${this.immutableValues.map((value) => html`
        <chops-chip class="immutable" title="Derived: ${value}">
          ${value}
        </chops-chip>
      `)}
      ${this.values.map((value, i) => html`
        <input
          ?hidden=${i !== this.collapsedChipIndex}
          class="edit-value edit-value-${i}"
          data-ac-type=${this.acType}
          autocomplete=${this.autocomplete}
          .value=${value}
          @blur=${this._stopEditingChip}
          @focus=${this._changeFocus}
        />
        <chops-chip
          ?hidden=${i === this.collapsedChipIndex}
          icon="close"
          class="chip-${i}"
          data-index=${i}
          @click-icon=${this._removeValue}
          @dblclick=${this._editChip}
          @keydown=${this._interactWithChips}
          @blur=${this._changeFocus}
          @focus=${this._changeFocus}
          focusable>${value}</chops-chip>
      `)}
      <input
        class="add-value"
        placeholder=${this.placeholder}
        data-ac-type=${this.acType}
        autocomplete=${this.autocomplete}
        @keydown=${this._navigateByKeyboard}
        @keyup=${this._createChipsWhileTyping}
        @blur=${this._onBlur}
        @focus=${this._changeFocus}
      />
    `;
  }

  static get properties() {
    return {
      immutableValues: {type: Array},
      initialValues: {
        type: Array,
        hasChanged(newVal, oldVal) {
          // Prevent extra recomputations of the same initial value cause
          // values to be reset.
          return !deepEqual(newVal, oldVal);
        },
      },
      values: {type: Array},
      // TODO(zhangtiff): Change autocomplete binding once Monorail's
      // autocomplete is rewritten.
      acType: {type: String},
      autocomplete: {type: String},
      placeholder: {type: String},
      focused: {
        type: Boolean,
        reflect: true,
      },
      collapsedChipIndex: {type: Number},
      delimiterRegex: {type: Object},
      undoStack: {type: Array},
      undoLimit: {type: Number},
      _addValueInput: {type: Object},
    };
  }

  constructor() {
    super();

    this.values = [];
    this.initialValues = [];
    this.immutableValues = [];

    this.delimiterRegex = DELIMITER_REGEX;
    this.collapsedChipIndex = -1;
    this.placeholder = 'Add value...';

    this.undoStack = [];
    this.undoLimit = 50;
  }

  connectedCallback() {
    super.connectedCallback();

    this.addEventListener('keydown', this._onKeyDown.bind(this));
  }

  firstUpdated() {
    this._addValueInput = this.shadowRoot.querySelector('.add-value');
  }

  update(changedProperties) {
    if (changedProperties.has('initialValues')) {
      this.reset();
    }

    super.update(changedProperties);
  }

  updated(changedProperties) {
    if (changedProperties.has('values')) {
      this.dispatchEvent(new CustomEvent('change'));
    }
  }

  reset() {
    this.setValues(this.initialValues);
    this.undoStack = [];
  }

  focus() {
    this._addValueInput.focus();
  }

  undo() {
    if (!this.undoStack.length) return;

    const prevValues = this.undoStack.pop();

    // TODO(zhangtiff): Make undo work for values that aren't
    // chips yet as well.
    this.values = prevValues;
  }

  _onKeyDown(e) {
    if (!this.focused) return;
    if (e.key === 'z' && e.ctrlKey) {
      this.undo();
    }
  }

  getValues() {
    // Make sure to include any values that haven't yet been chipified as well.
    const newValues = this._readCollapsedValues(this._addValueInput);
    return this.values.concat(newValues);
  }

  setValues(values) {
    this._saveValues(values);

    if (this._addValueInput) {
      this._addValueInput.value = '';
    }
  }

  _saveValues(values) {
    this.undoStack.push(this.values);

    if (this.undoStack.length > this.undoLimit) {
      this.undoStack.shift();
    }

    this.values = [...values];
  }

  async _editChip(e) {
    const target = e.target;
    const index = Number.parseInt(target.dataset.index);
    if (index < 0 || index >= this.values.length) return;

    this.collapsedChipIndex = index;

    await this.updateComplete;

    const input = this.shadowRoot.querySelector(`.edit-value-${index}`);

    input.focus();

    // Move cursor to the end of the input.
    const value = input.value;
    input.value = '';
    input.value = value;

    // TODO(zhangtiff): Remove this once autocomplete is rewritten and can be
    // triggered programmatically in a cleaner way.
    // See: http://crbug.com/monorail/5301
    if (window.ac_keyevent_) {
      ac_keyevent_({target: input});
    }
  }

  _removeValue(e) {
    const target = e.target;
    const index = Number.parseInt(target.dataset.index);
    if (index < 0 || index >= this.values.length) return;

    this._saveValues(immutableSplice(this.values, index, 1));
  }

  _stopEditingChip(e) {
    if (this.collapsedChipIndex < 0) return;
    const target = e.target;

    const pieces = this._readCollapsedValues(target);

    this._saveValues(immutableSplice(
      this.values, this.collapsedChipIndex, 1, ...pieces));

    this.collapsedChipIndex = -1;
  }

  _onBlur(e) {
    this._convertNewValuesToChips();
    this._changeFocus(e);
  }

  _createChipsWhileTyping(e) {
    const input = this._addValueInput;
    if (input.value.match(this.delimiterRegex)) {
      this._convertNewValuesToChips();
    }
  }

  _convertNewValuesToChips() {
    const input = this._addValueInput;
    const values = this._readCollapsedValues(input);
    if (values.length) {
      this._saveValues([...this.values, ...values]);
      input.value = '';
    }
  }

  _changeFocus(e) {
    // Check if any element in this shadowRoot is focused.
    const active = this.shadowRoot.activeElement;
    if (active) {
      this.focused = true;
      this.dispatchEvent(new CustomEvent('focus'));
    } else {
      this.focused = false;
      this.dispatchEvent(new CustomEvent('blur'));
    }
  }

  _interactWithChips(e) {
    const chip = e.target;
    const index = Number.parseInt(chip.dataset.index);
    if (index < 0 || index >= this.values.length) return;
    const input = this._addValueInput;

    if (e.key === 'Backspace') {
      // Delete the current chip then focus the one before it.
      this._saveValues(immutableSplice(this.values, index, 1));

      if (this.values.length > 0) {
        const chipBefore = this._getChipElement(Math.max(0, index - 1));
        chipBefore.focus();
      } else {
        // Move to the input if there are no chips left.
        input.focus();
      }
    } else if (e.key === 'ArrowLeft' && index > 0) {
      const prevChip = this._getChipElement(index - 1);
      prevChip.focus();
    } else if (e.key === 'ArrowRight') {
      if (index >= this.values.length - 1) {
        // Move to the input if there are no chips to the right.
        input.focus();
      } else {
        const nextChip = this._getChipElement(index + 1);
        nextChip.focus();
      }
    }
  }

  _navigateByKeyboard(e) {
    const input = e.target;
    const atStartOfInput = input.selectionEnd === input.selectionStart
        && input.selectionStart === 0;
    if (atStartOfInput) {
      if (e.key === 'Backspace') {
        // Delete last chip.
        this._saveValues(immutableSplice(this.values, this.values.length - 1, 1));

        // Prevent autocomplete menu from opening.
        // TODO(zhangtiff): Remove this when reworking autocomplete as a
        // web component.
        e.stopPropagation();
      } else if (e.key === 'ArrowLeft' && this.values.length) {
        const lastChip = this._getChipElement(this.values.length - 1);
        lastChip.focus();

        // Prevent autocomplete menu from opening.
        // TODO(zhangtiff): Remove this when reworking autocomplete as a
        // web component.
        e.stopPropagation();
      }
    }
  }

  _readCollapsedValues(input) {
    const values = input.value.split(this.delimiterRegex);

    // Filter out empty strings.
    const pieces = values.filter(Boolean);
    return pieces;
  }

  _getChipElement(index) {
    return this.shadowRoot.querySelector(`.chip-${index}`);
  }
}

customElements.define('chops-chip-input', ChopsChipInput);
