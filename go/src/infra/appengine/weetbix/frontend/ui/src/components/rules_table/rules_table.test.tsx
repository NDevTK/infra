// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';
import 'node-fetch';

import { screen } from '@testing-library/react';

import { renderWithRouterAndClient } from '../../testing_tools/libs/mock_router';
import { mockFetchAuthState } from '../../testing_tools/mocks/authstate_mock';
import { mockFetchRules } from '../../testing_tools/mocks/rules_mock';
import RulesTable from './rules_table';

describe('Test RulesTable component', () => {
  it('given a project, should display the active rules', async () => {
    mockFetchAuthState();
    mockFetchRules();

    renderWithRouterAndClient(
        <RulesTable
          project='chromium'/>,
        '/p/chromium/rules',
        '/p/:project/rules',
    );
    await screen.findByText('Rule Definition');

    expect(screen.getByText('crbug.com/90001')).toBeInTheDocument();
    expect(screen.getByText('crbug.com/90002')).toBeInTheDocument();
    expect(screen.getByText('test LIKE "rule1%"')).toBeInTheDocument();
    expect(screen.getByText('reason LIKE "rule2%"')).toBeInTheDocument();
  });
});
