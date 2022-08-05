// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
import '@testing-library/jest-dom';

import { screen } from '@testing-library/react';

import {
  renderWithRouter,
  renderWithRouterAndClient,
} from '../../testing_tools/libs/mock_router';
import TopBar from './top_bar';

describe('test TopBar component', () => {
  beforeAll(() => {
    window.email = 'test@google.com';
    window.avatar = '/example.png';
    window.fullName = 'Test Name';
    window.logoutUrl = '/logout';
  });

  it('should render logo and user email', async () => {
    renderWithRouter(
        <TopBar />,
    );

    await screen.findAllByText('Weetbix');

    expect(screen.getByText(window.email)).toBeInTheDocument();
  });

  it('given a route with a project then should display pages', async () => {
    renderWithRouterAndClient(
        <TopBar />,
        '/p/chrome',
        '/p/:project',
    );

    await screen.findAllByText('Weetbix');

    expect(screen.getAllByText('Clusters')).toHaveLength(2);
    expect(screen.getAllByText('Rules')).toHaveLength(2);
  });
});
