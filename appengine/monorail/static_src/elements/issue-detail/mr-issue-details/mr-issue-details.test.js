// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {assert} from 'chai';
import {MrIssueDetails} from './mr-issue-details.js';

let element;

describe('mr-issue-details', () => {
  beforeEach(() => {
    element = document.createElement('mr-issue-details');
    document.body.appendChild(element);
  });

  afterEach(() => {
    document.body.removeChild(element);
  });

  it('initializes', () => {
    assert.instanceOf(element, MrIssueDetails);
  });

  it('mr-edit-issue is displayed if user has addissuecomment', async () => {
    element.issuePermissions = ['addissuecomment'];

    await element.updateComplete;

    assert.isNotNull(element.querySelector('mr-edit-issue'));
  });

  it('mr-edit-issue is hidden if user has no addissuecomment', async () => {
    element.issuePermissions = [];

    await element.updateComplete;

    assert.isNull(element.querySelector('mr-edit-issue'));
  });
});
