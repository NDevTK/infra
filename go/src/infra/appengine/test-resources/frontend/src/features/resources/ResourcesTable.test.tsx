// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { screen } from '@testing-library/react';
import { Test } from '../context/MetricsContext';
import { MetricType } from '../../api/resources';
import { renderWithContext } from '../../utils/testUtils';
import ResourcesTable from './ResourcesTable';

const mockMetrics: Map<MetricType, number> = new Map<MetricType, number>(
    [
      [MetricType.NUM_RUNS, 1],
      [MetricType.NUM_FAILURES, 2],
      [MetricType.AVG_RUNTIME, 3],
      [MetricType.TOTAL_RUNTIME, 4],
      [MetricType.AVG_CORES, 5],
    ],
);

const tests: Test[] = [{
  id: 'testId',
  name: 'testName',
  fileName: 'fileName',
  metrics: mockMetrics,
  isLeaf: false,
  nodes: [
    {
      id: 'v1',
      name: 'suite',
      subname: 'builder',
      metrics: mockMetrics,
      isLeaf: true,
      nodes: [],
    },
    {
      id: 'v1',
      name: 'suite',
      subname: 'builder',
      metrics: mockMetrics,
      isLeaf: true,
      nodes: [],
    },
  ],
}];

describe('when rendering the ResourcesTable', () => {
  it('should render the TableContainer', () => {
    renderWithContext(<ResourcesTable/>, { data: tests });
    expect(screen.getByTestId('tableBody')).toBeInTheDocument();
    expect(screen.getByText('Test Suite')).toBeInTheDocument();
    expect(screen.getByText('# Runs')).toBeInTheDocument();
    expect(screen.getByText('# Failures')).toBeInTheDocument();
    expect(screen.getByText('Avg Runtime')).toBeInTheDocument();
    expect(screen.getByText('Total Runtime')).toBeInTheDocument();
    expect(screen.getByText('Avg Cores')).toBeInTheDocument();
  });
});

describe('when rendering the ResourcesTable', () => {
  it('should render loading screen', () => {
    renderWithContext(<ResourcesTable/>, { data: [], isLoading: true });
    expect(screen.getByText('Loading...')).toBeInTheDocument();
    expect(screen.getByTestId('loading-bar')).not.toHaveClass(
        'hidden',
    );
  });
});
