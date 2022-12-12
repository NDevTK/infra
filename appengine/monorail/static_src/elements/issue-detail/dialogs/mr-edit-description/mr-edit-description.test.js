// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {assert} from 'chai';
import {MrEditDescription} from './mr-edit-description.js';

let element;

describe('mr-edit-description', () => {
  beforeEach(() => {
    element = document.createElement('mr-edit-description');

    document.body.appendChild(element);
    element.commentsByApproval = new Map([
      ['', [
        {
          descriptionNum: 1,
          content: 'first description',
        },
        {
          content: 'first comment',
        },
        {
          descriptionNum: 2,
          content: '<b>last</b> description',
        },
        {
          content: 'second comment',
        },
        {
          content: 'third comment',
        },
      ]], ['foo', [
        {
          descriptionNum: 1,
          content: 'first foo survey',
          approvalRef: {
            fieldName: 'foo',
          },
        },
        {
          descriptionNum: 2,
          content: 'last foo survey',
          approvalRef: {
            fieldName: 'foo',
          },
        },
      ]], ['bar', [
        {
          descriptionNum: 1,
          content: 'bar survey',
          approvalRef: {
            fieldName: 'bar',
          },
        },
      ]],
    ]);
  });

  afterEach(() => {
    document.body.removeChild(element);
  });

  it('initializes', () => {
    assert.instanceOf(element, MrEditDescription);
  });

  it('selects last issue description', async () => {
    element.fieldName = '';
    element.reset();

    await element.updateComplete;

    assert.equal(element._editedDescription, 'last description');
    assert.equal(element._title, 'Description');
  });

  it('selects last survey', async () => {
    element.fieldName = 'foo';
    element.reset();

    await element.updateComplete;

    assert.equal(element._editedDescription, 'last foo survey');
    assert.equal(element._title, 'foo Survey');
  });

  it('toggle sendEmail', async () => {
    element.reset();
    await element.updateComplete;

    const sendEmail = element.shadowRoot.querySelector('#sendEmail');

    await sendEmail.updateComplete;

    sendEmail.click();
    await element.updateComplete;
    assert.isFalse(element._sendEmail);

    sendEmail.click();
    await element.updateComplete;
    assert.isTrue(element._sendEmail);

    sendEmail.click();
    await element.updateComplete;
    assert.isFalse(element._sendEmail);
  });

  it('renders valid markdown description with preview class', async () => {
    element.projectName = 'monkeyrail';
    element._prefs = new Map([['render_markdown', true]]);
    element.reset();

    element._editedDescription = '# h1';

    await element.updateComplete;

    const previewMarkdown = element.shadowRoot.querySelector('.markdown-preview');
    assert.isNotNull(previewMarkdown);

    const headerText = previewMarkdown.querySelector('h1').textContent;
    assert.equal(headerText, 'h1');
  });

  it('does not show preview when markdown is disabled', async () => {
    element.projectName = 'disabled_project';
    element._prefs = new Map([['render_markdown', true]]);
    element.reset();

    await element.updateComplete;

    const previewMarkdown = element.shadowRoot.querySelector('.markdown-preview');
    assert.isNull(previewMarkdown);
  });
});
