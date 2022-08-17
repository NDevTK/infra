// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';

import { AnalysisOverview, AnalysisSummary } from './analysis_overview';
import { getMockAnalysisSummary } from '../../testing_tools/mocks/analysis_summary_mock';

describe('Test AnalysisOverview component', () => {
  test('if all analysis summary details are displayed', async () => {
    const mockSummary = getMockAnalysisSummary('1');

    render(<AnalysisOverview analysis={mockSummary} />);

    await screen.findByTestId('analysis_overview_table_body');

    const expectedStaticFields = [
      ['analysis ID', 'analysisID'],
      ['status', 'status'],
      ['buildbucket ID', 'buildID'],
      ['builder', 'builder'],
      ['failure type', 'failureType'],
    ];

    // check static field labels and values are displayed
    expectedStaticFields.forEach(([label, property]) => {
      const fieldLabel = screen.getByText(new RegExp(`^(${label})$`, 'i'));
      expect(fieldLabel).toBeInTheDocument();
      expect(fieldLabel.nextSibling?.textContent).toBe(
        `${mockSummary[property as keyof AnalysisSummary]}`
      );
    });

    // check the suspect range is displayed correctly
    const suspectRangeLabel = screen.getByText(
      new RegExp('^(suspect range)$', 'i')
    );
    expect(suspectRangeLabel).toBeInTheDocument();
    const suspectRangeValue = mockSummary.suspectRange.linkText;
    expect(suspectRangeLabel.nextElementSibling?.textContent).toBe(
      suspectRangeValue
    );
    expect(screen.getByText(suspectRangeValue).getAttribute('href')).toBe(
      mockSummary.suspectRange.url
    );

    // check related bug links are displayed
    expect(
      screen.getByText(new RegExp('^(related bugs)$', 'i'))
    ).toBeInTheDocument();
    mockSummary.bugs.forEach((bug) => {
      expect(screen.getByText(bug.linkText).getAttribute('href')).toBe(bug.url);
    });
  });

  test('if there are no bugs, then related bugs section is not shown', async () => {
    let mockSummary = getMockAnalysisSummary('2');
    mockSummary.bugs = [];

    const { container } = render(<AnalysisOverview analysis={mockSummary} />);

    await screen.findByTestId('analysis_overview_table_body');

    // check there are no bug links
    expect(screen.queryByText('Related bugs')).not.toBeInTheDocument();
    expect(container.getElementsByClassName('bugLink')).toHaveLength(0);
  });
});
