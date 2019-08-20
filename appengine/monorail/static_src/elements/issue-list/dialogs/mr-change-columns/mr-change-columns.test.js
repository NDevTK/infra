// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {assert} from 'chai';
import {MrChangeColumns} from './mr-change-columns.js';


let element;

describe('mr-change-columns', () => {
  beforeEach(() => {
    element = document.createElement('mr-change-columns');
    document.body.appendChild(element);
  });

  afterEach(() => {
    document.body.removeChild(element);
  });

  it('initializes', () => {
    assert.instanceOf(element, MrChangeColumns);
  });

  it('input initializes with currently set columns', async () => {
    element.columns = ['ID', 'Summary'];

    await element.updateComplete;

    const input = element.shadowRoot.querySelector('#columnsInput');

    assert.equal(input.value, 'ID Summary');
  });

  it('editing input and saving updates columns in URL', async () => {
    element.columns = ['ID', 'Summary'];
    element.queryParams = {};

    await element.updateComplete;

    const input = element.shadowRoot.querySelector('#columnsInput');

    input.value = 'ID Summary Owner';

    element._page = sinon.stub();
    sinon.stub(element, '_currentPage').get(() => '/test');

    element.save();

    sinon.assert.calledWith(element._page,
        '/test?colspec=ID%2BSummary%2BOwner');
  });
});
