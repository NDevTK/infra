// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';

import { Home } from '@mui/icons-material';
import {
  fireEvent,
  screen,
} from '@testing-library/react';

import { renderWithRouter } from '../../../testing_tools/libs/mock_router';
import { AppBarPage } from '../top_bar';
import { TopBarContextProvider } from '../top_bar_context';
import CollapsedMenu from './collapsed_menu';

describe('test CollapsedMenu component', () => {
  const pages: AppBarPage[] = [
    {
      title: 'Clusters',
      url: '/Clusters',
      icon: Home,
    },
  ];

  it('given a set of pages, then should display them in a menu', async () => {
    renderWithRouter(
        <CollapsedMenu pages={pages}/>,
    );

    await screen.findByText('Weetbix');

    expect(screen.getByText('Clusters')).toBeInTheDocument();
  });

  it('when clicking on menu button then the menu should be visible', async () => {
    renderWithRouter(
        <TopBarContextProvider >
          <CollapsedMenu pages={pages}/>
        </TopBarContextProvider>,
    );

    await screen.findByText('Weetbix');

    expect(screen.getByTestId('collapsed-menu')).not.toBeVisible();

    await fireEvent.click(screen.getByTestId('collapsed-menu-button'));

    expect(screen.getByTestId('collapsed-menu')).toBeVisible();
  });
});
