// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';

import { HeuristicAnalysisTableRow } from './heuristic_analysis_table_row';
import { HeuristicSuspect } from '../../../services/analysis_details';
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

    expect(screen.getByRole('link').getAttribute('href')).toBe(
      mockSuspect.cl.reviewURL
    );
    expect(screen.getByText(mockSuspect.confidence)).toBeInTheDocument();
    expect(screen.getByText(mockSuspect.score)).toBeInTheDocument();
    mockSuspect.justification.forEach((reason) => {
      expect(screen.getByText(reason)).toBeInTheDocument();
    });
  });
});
