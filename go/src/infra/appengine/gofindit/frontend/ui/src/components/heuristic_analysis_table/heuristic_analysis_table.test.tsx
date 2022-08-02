// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';

import { HeuristicAnalysisTable } from './heuristic_analysis_table';
import { getMockHeuristicSuspect } from '../../testing_tools/mocks/heuristic_suspect_mock';

describe('Test HeuristicAnalysisTable component', () => {
  test('if heuristic suspects are displayed', async () => {
    const mockSuspects = [
      getMockHeuristicSuspect('ac52e3'),
      getMockHeuristicSuspect('673e20'),
    ];

    render(<HeuristicAnalysisTable suspects={mockSuspects} />);

    await screen.findByText('Suspect CL');

    expect(screen.queryAllByRole('link')).toHaveLength(mockSuspects.length);
  });

  test('if an appropriate message is displayed for no suspects', async () => {
    render(<HeuristicAnalysisTable suspects={[]} />);

    await screen.findByText('Suspect CL');

    expect(screen.queryAllByRole('link')).toHaveLength(0);
    expect(screen.getByText('No suspects to display')).toBeInTheDocument();
  });
});
