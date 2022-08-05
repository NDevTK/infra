// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';

import fetchMock from 'fetch-mock-jest';

import { screen } from '@testing-library/react';

import { identityFunction } from '../../testing_tools/functions';
import { renderWithRouterAndClient } from '../../testing_tools/libs/mock_router';
import { mockFetchAuthState } from '../../testing_tools/mocks/authstate_mock';
import { mockFetchProjectConfig } from '../../testing_tools/mocks/projects_mock';
import BugPicker from './bug_picker';

describe('Test BugPicker component', () => {
  beforeEach(() => {
    mockFetchAuthState();
    mockFetchProjectConfig();
  });

  afterEach(() => {
    fetchMock.mockClear();
    fetchMock.reset();
  });

  it('given a bug and a project, should display select and a text box for writing the bug id', async () => {
    renderWithRouterAndClient(
        <BugPicker
          bugId="chromium/123456"
          bugSystem="monorail"
          handleBugSystemChanged={identityFunction}
          handleBugIdChanged={identityFunction}/>, '/p/chromium', '/p/:project');
    await screen.findByText('Bug tracker');
    expect(screen.getByTestId('bug-system')).toHaveValue('monorail');
    expect(screen.getByTestId('bug-number')).toHaveValue('123456');
  });

  it('given a buganizer bug, should select the bug system correctly', async () => {
    renderWithRouterAndClient(
        <BugPicker
          bugId="123456"
          bugSystem="buganizer"
          handleBugSystemChanged={identityFunction}
          handleBugIdChanged={identityFunction}/>, '/p/chromium', '/p/:project');
    await screen.findByText('Bug tracker');
    expect(screen.getByTestId('bug-system')).toHaveValue('buganizer');
    expect(screen.getByTestId('bug-number')).toHaveValue('123456');
  });
});
