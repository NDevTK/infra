/* Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
*/

import { act } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { renderWithComponents } from '../../components/testUtils';
import { renderWithContext } from './testUtils';
import { Params } from './LoadSummary';
import SummarySearchParams, { PLATFORM, REVISION, UNIT_TESTS_ONLY } from './SummarySearchParams';

describe('when rendering the SummarySearchParams', () => {
  const params: Params = {
    host: 'chromium.googlesource.com',
    project: 'chromium',
    gitilesRef: 'main',
    revision: 'abc123',
    unitTestsOnly: true,
    bucket: 'bucket1',
    builder: 'builder1',
    platform: 'linux',
    platformList: [
      {
        platform: 'linux',
        bucket: 'test-bucket',
        builder: 'test-builder',
        coverageTool: 'test-cov-tool',
        uiName: 'Linux',
        availableRevision: '12345',
        avaialbleModifierId: '0',
      },
    ],
  };

  it('should render url corrently', async () => {
    await act(async () => {
      renderWithContext(<>
        <SummarySearchParams/>
      </>
      , { params },
      );
    });
    const searchParams = new URLSearchParams(window.location.search);
    expect(searchParams.get(REVISION)).toBe('abc123');
    expect(searchParams.get(UNIT_TESTS_ONLY)).toBe('true');
    expect(searchParams.get(PLATFORM)).toBe('linux');
  });

  it('should render components in url', async () => {
    await act(async () => {
      renderWithComponents((
        <>
          <BrowserRouter>
            <SummarySearchParams/>
          </BrowserRouter>
        </>
      ), { components: ['a', 'b'] },
      );
    });
    const searchParams = new URLSearchParams(window.location.search);
    expect(searchParams.getAll('c')).toEqual(['a', 'b']);
  });
});
