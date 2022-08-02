// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';

import { getMockSuspectSummary } from '../../testing_tools/mocks/suspect_mock';
import { SuspectsOverview } from './suspects_overview';
import { SuspectSummary } from '../../services/analysis_details';

describe('Test SuspectsOverview component', () => {
  test('if all suspect details are displayed', async () => {
    const mockSuspects = [
      getMockSuspectSummary('c234de'),
      getMockSuspectSummary('412533'),
    ];

    render(<SuspectsOverview suspects={mockSuspects} />);

    await screen.findByText('Suspect CL');

    // check there is a link for each suspect CL
    expect(screen.queryAllByRole('link')).toHaveLength(mockSuspects.length);

    // check the target URL for a suspect CL
    expect(screen.getByText(mockSuspects[1].title).getAttribute('href')).toBe(
      mockSuspects[1].url
    );
  });

  test('if an appropriate message is displayed for no suspects', async () => {
    const mockSuspects: SuspectSummary[] = [];

    render(<SuspectsOverview suspects={mockSuspects} />);

    await screen.findByText('Suspect CL');

    expect(screen.queryAllByRole('link')).toHaveLength(0);
    expect(screen.getByText('No suspects to display')).toBeInTheDocument();
  });
});
