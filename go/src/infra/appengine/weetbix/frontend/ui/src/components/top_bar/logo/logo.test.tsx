// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
import '@testing-library/jest-dom';

import {
  render,
  screen,
} from '@testing-library/react';

import Logo from './logo';

describe('test Logo component', () => {
  it('should display logo image', async () => {
    render(
        <Logo />,
    );
    await screen.findByRole('img');

    expect(screen.getByRole('img')).toHaveAttribute('alt', 'logo');
  });
});
