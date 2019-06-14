// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {assert} from 'chai';
import {MrSiteBanner} from './mr-site-banner.js';


let element;

describe('mr-site-banner', () => {
  beforeEach(() => {
    element = document.createElement('mr-site-banner');
    document.body.appendChild(element);
  });

  afterEach(() => {
    document.body.removeChild(element);
  });

  it('initializes', () => {
    assert.instanceOf(element, MrSiteBanner);
  });

  it('displays a banner message', async () => {
    element.bannerMessage = 'Message';
    await element.updateComplete;
    assert.equal(element.shadowRoot.textContent.trim(), 'Message');
    assert.isNull(element.shadowRoot.querySelector('chops-timestamp'));
  });

  it('displays the banner timestamp', async () => {
    element.bannerMessage = 'Message';
    element.bannerTime = 123456789;
    await element.updateComplete;
    assert.isNotNull(element.shadowRoot.querySelector('chops-timestamp'));
  });

  it('hides when there is no banner message', async () => {
    await element.updateComplete;
    assert.isTrue(element.hidden);
  });
});
