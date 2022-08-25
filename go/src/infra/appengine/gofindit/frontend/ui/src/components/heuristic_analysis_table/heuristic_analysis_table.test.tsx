// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';

import { HeuristicAnalysisTable } from './heuristic_analysis_table';
import { getMockHeuristicSuspect } from '../../testing_tools/mocks/heuristic_suspect_mock';
import { HeuristicAnalysisResult } from '../../services/gofindit';

describe('Test HeuristicAnalysisTable component', () => {
  test('if heuristic suspects are displayed', async () => {
    const mockSuspects = [
      getMockHeuristicSuspect('ac52e3'),
      getMockHeuristicSuspect('673e20'),
    ];

    const mockHeuristicAnalysisResult: HeuristicAnalysisResult = {
      status: 'FOUND',
      suspects: mockSuspects,
    };

    render(<HeuristicAnalysisTable result={mockHeuristicAnalysisResult} />);

    await screen.findByText('Suspect CL');

    expect(screen.queryAllByRole('link')).toHaveLength(mockSuspects.length);
  });

  test('if an appropriate message is displayed for no suspects', async () => {
    const mockHeuristicAnalysisResult: HeuristicAnalysisResult = {
      status: 'NOTFOUND',
      suspects: [],
    };
    render(<HeuristicAnalysisTable result={mockHeuristicAnalysisResult} />);

    await screen.findByText('Suspect CL');

    expect(screen.queryAllByRole('link')).toHaveLength(0);
    expect(screen.getByText('No suspects to display')).toBeInTheDocument();
  });

  test('if no misleading message is shown for an incomplete analysis', async () => {
    const mockHeuristicAnalysisResult: HeuristicAnalysisResult = {
      status: 'RUNNING',
      suspects: [],
    };
    render(<HeuristicAnalysisTable result={mockHeuristicAnalysisResult} />);

    await screen.findByText('Suspect CL');

    expect(screen.queryAllByRole('link')).toHaveLength(0);
    expect(
      screen.queryByText('No suspects to display')
    ).not.toBeInTheDocument();
  });
});
