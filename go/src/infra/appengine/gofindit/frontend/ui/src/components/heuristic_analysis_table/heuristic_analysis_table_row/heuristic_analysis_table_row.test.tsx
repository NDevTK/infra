// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';

import { HeuristicAnalysisTableRow } from './heuristic_analysis_table_row';
import { HeuristicSuspect } from '../../../services/gofindit';
import { getMockHeuristicSuspect } from '../../../testing_tools/mocks/heuristic_suspect_mock';

describe('Test HeuristicAnalysisTable component', () => {
  test('if the details for a heuristic suspect are displayed', async () => {
    const mockSuspect: HeuristicSuspect = getMockHeuristicSuspect('ac52e3');

    render(
      <Table>
        <TableBody>
          <HeuristicAnalysisTableRow suspect={mockSuspect} />
        </TableBody>
      </Table>
    );

    await screen.findByTestId('heuristic_analysis_table_row');

    // Check there is a link to the suspect's code review
    const suspectReviewLink = screen.getByRole('link');
    expect(suspectReviewLink).toBeInTheDocument();
    expect(suspectReviewLink.getAttribute('href')).toBe(mockSuspect.reviewUrl);
    if (mockSuspect.reviewTitle) {
      expect(suspectReviewLink.textContent).toContain(mockSuspect.reviewTitle);
    }

    // Check confidence level, score and reasons are displayed
    expect(screen.getByText(mockSuspect.confidenceLevel)).toBeInTheDocument();
    expect(screen.getByText(mockSuspect.score)).toBeInTheDocument();
    const reasons = mockSuspect.justification.split('\n');
    reasons.forEach((reason) => {
      expect(screen.getByText(reason)).toBeInTheDocument();
    });
  });
});
