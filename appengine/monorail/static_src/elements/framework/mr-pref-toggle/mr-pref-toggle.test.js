// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import 'sinon';
import {assert} from 'chai';
import {MrPrefToggle} from './mr-pref-toggle.js';
import {prpcClient} from 'prpc-client-instance.js';

let element;

describe('mr-pref-toggle', () => {
  beforeEach(() => {
    element = document.createElement('mr-pref-toggle');
    element.label = 'Code';
    element.title = 'Code font';
    element.prefName = 'code_font';
    document.body.appendChild(element);
    sinon.stub(prpcClient, 'call').returns(Promise.resolve({}));
    window.ga = sinon.stub();
  });

  afterEach(() => {
    document.body.removeChild(element);
    prpcClient.call.restore();
  });

  it('initializes', () => {
    assert.instanceOf(element, MrPrefToggle);
  });

  it('toggling does not save when user is not logged in', async () => {
    element.userDisplayName = undefined;
    element.prefs = new Map([]);

    await element.updateComplete;

    const chopsToggle = element.shadowRoot.querySelector('chops-toggle');
    chopsToggle.click();
    await element.updateComplete;

    sinon.assert.notCalled(prpcClient.call);

    assert.isTrue(element.prefs.get('code_font'));
  });

  it('toggling to true saves result', async () => {
    element.userDisplayName = 'test@example.com';
    element.prefs = new Map([['code_font', false]]);

    await element.updateComplete;

    const chopsToggle = element.shadowRoot.querySelector('chops-toggle');

    chopsToggle.click(); // Toggle it on.
    await element.updateComplete;

    sinon.assert.calledWith(
        prpcClient.call,
        'monorail.Users',
        'SetUserPrefs',
        {prefs: [{name: 'code_font', value: 'true'}]});

    assert.isTrue(element.prefs.get('code_font'));
  });

  it('toggling to false saves result', async () => {
    element.userDisplayName = 'test@example.com';
    element.prefs = new Map([['code_font', true]]);

    await element.updateComplete;

    const chopsToggle = element.shadowRoot.querySelector('chops-toggle');

    chopsToggle.click(); // Toggle it off.
    await element.updateComplete;

    sinon.assert.calledWith(
        prpcClient.call,
        'monorail.Users',
        'SetUserPrefs',
        {prefs: [{name: 'code_font', value: 'false'}]});

    assert.isFalse(element.prefs.get('code_font'));
  });
});
