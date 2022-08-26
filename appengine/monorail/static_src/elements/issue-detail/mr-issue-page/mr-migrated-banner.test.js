// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {assert} from 'chai';
import {MrMigratedBanner} from './mr-migrated-banner.js';
import {migratedTypes} from 'shared/issue-fields.js';

let element;

describe('mr-migrated-banner', () => {
  beforeEach(() => {
    element = document.createElement('mr-migrated-banner');
    document.body.appendChild(element);
  });

  afterEach(() => {
    document.body.removeChild(element);
  });

  it('initializes', () => {
    assert.instanceOf(element, MrMigratedBanner);
  });

  it('hides element by default', async () => {
    await element.updateComplete;

    assert.isTrue(element.hasAttribute('hidden'));
  });

  it('hides element when migratedId is empty', async () => {
    element.migratedId = '';
    await element.updateComplete;

    assert.isTrue(element.hasAttribute('hidden'));
  });

  it('shows element when migratedId and migratedType is set', async () => {
    element.migratedId = '1234';
    element.migratedType = migratedTypes.BUGANIZER_TYPE
    await element.updateComplete;

    assert.isFalse(element.hasAttribute('hidden'));
  });

  it('shows bugnizer link when migrate to bugnizer', async () => {
    element.migratedId = '1234';
    element.migratedType = migratedTypes.BUGANIZER_TYPE
    await element.updateComplete;

    const link = element.shadowRoot.querySelector('a');
    assert.include(link.textContent, 'b/1234');
  });

  it('shows launch link when migrate to launch', async () => {
    element.migratedId = '1234';
    element.migratedType = migratedTypes.LAUNCH_TYPE
    await element.updateComplete;

    const link = element.shadowRoot.querySelector('p');
    assert.include(link.textContent, 'This issue has been migrated to Launch, see link in final comment below');
  });
});