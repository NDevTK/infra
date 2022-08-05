// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';

import {
  fireEvent,
  render,
  screen,
} from '@testing-library/react';

import UserActions from './user_actions';

describe('test UserActions component', () => {
  beforeAll(() => {
    window.email = 'test@google.com';
    window.avatar = '/example.png';
    window.fullName = 'Test Name';
    window.logoutUrl = '/logout';
  });

  it('should display user email and logout url', async () => {
    render(
        <UserActions />,
    );

    await screen.getByText(window.email);

    expect(screen.getByRole('img')).toHaveAttribute('src', window.avatar);
    expect(screen.getByRole('img')).toHaveAttribute('alt', window.fullName);
    expect(screen.getByTestId('useractions_logout')).toHaveAttribute('href', window.logoutUrl);
  });

  it('when clicking on email button then should display logout url', async () => {
    render(
        <UserActions />,
    );

    await screen.getByText(window.email);

    expect(screen.getByTestId('user-settings-menu')).not.toBeVisible();

    await fireEvent.click(screen.getByText(window.email));

    expect(screen.getByTestId('user-settings-menu')).toBeVisible();
  });
});
