// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {assert} from 'chai';
import * as sitewide from './sitewide.js';

function wrapServerStatus(serverStatus) {
  return {sitewide: {...serverStatus}};
}

describe('sitewide', () => {
  it('selectors', () => {
    const state = wrapServerStatus({
      bannerMessage: 'Message',
      bannerTime: 1234,
      readOnly: true,
    });
    assert.deepEqual(sitewide.bannerMessage(state), 'Message');
    assert.deepEqual(sitewide.bannerTime(state), 1234);
    assert.isTrue(sitewide.readOnly(state));
  });
});
