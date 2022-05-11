// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React from 'react';
import {render} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {screen} from '@testing-library/dom';
import {assert} from 'chai';

import SelectMenu from './SelectMenu.tsx';

describe('SelectMenu', () => {
  let container: React.RenderResult;

  beforeEach(() => {
    container = render(<SelectMenu optionsList = {['op1', 'op2']} />).container;
  });

  it('renders', () => {
    const form = container.querySelector('form');
    assert.isNotNull(form)
  });

  it('renders options on click', async () => {
    const input = document.getElementById('outlined-select-category');
    if (!input) {
      throw new Error('Input is undefined');
    }

    userEvent.click(input)

    // 14 is the current number of options in the select menu
    const count = (await screen.findAllByTestId('select-menu-item')).length;

    assert.equal(count, 2);
  });
});
