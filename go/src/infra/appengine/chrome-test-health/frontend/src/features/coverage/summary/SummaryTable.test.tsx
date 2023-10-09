// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { screen } from '@testing-library/react';
import { renderWithContext } from './testUtils';
import SummaryTable from './SummaryTable';
import { MetricData, MetricType, Node } from './LoadSummary';

const mockMetrics: Map<MetricType, MetricData> = new Map<MetricType, MetricData>(
    [
      [MetricType.LINE, { covered: 67, total: 100, percentageCovered: 67 }],
    ],
);

const mockNodes: Node[] = [
  {
    id: 'dir',
    name: 'dir',
    metrics: mockMetrics,
    rows: [],
  },
];

describe('when rendering the SummaryTable', () => {
  it('should render loading screen is isLoading is true', () => {
    renderWithContext(<SummaryTable/>, { data: [], isLoading: true });
    expect(screen.getByTestId('legend')).toBeInTheDocument();
    expect(screen.getByText('Loading...')).toBeInTheDocument();
    expect(screen.getByTestId('loading-bar')).not.toHaveClass('hidden');
  });

  it('should render the table with data', () => {
    renderWithContext(<SummaryTable/>, { data: mockNodes });
    expect(screen.getByTestId('legend')).toBeInTheDocument();
    expect(screen.getByTestId('tableBody')).toBeInTheDocument();
    expect(screen.getByText('Directories/Files')).toBeInTheDocument();
    expect(screen.getByText('Line Coverage')).toBeInTheDocument();
    expect(screen.getByTestId('tablerow-dir')).toBeInTheDocument();
    expect(screen.getByText('67.00% (67/100)')).toBeInTheDocument();
  });
});
