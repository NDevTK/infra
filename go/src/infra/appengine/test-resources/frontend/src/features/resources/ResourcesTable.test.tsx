// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { screen } from '@testing-library/react';
import { Params, Test } from '../context/MetricsContext';
import { MetricType, Period, SortType } from '../../api/resources';
import { renderWithContext } from '../../utils/testUtils';
import ResourcesTable from './ResourcesTable';

const mockMetricTypeToNum: Map<MetricType, number> = new Map<MetricType, number>(
    [
      [MetricType.NUM_RUNS, 1],
      [MetricType.NUM_FAILURES, 2],
      [MetricType.AVG_RUNTIME, 3],
      [MetricType.TOTAL_RUNTIME, 4],
      [MetricType.AVG_CORES, 5],
    ],
);

const mockMetrics: Map<string, Map<MetricType, number>> = new Map<string, Map<MetricType, number>>(
    [
      ['2023-01-01', mockMetricTypeToNum],
      ['2023-01-02', mockMetricTypeToNum],
      ['2023-01-03', mockMetricTypeToNum],
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
  it('should render the TableContainer with snapshot view', () => {
    renderWithContext(<ResourcesTable/>, { data: tests });
    expect(screen.getByTestId('tableBody')).toBeInTheDocument();
    expect(screen.getByText('Test Suite')).toBeInTheDocument();
    expect(screen.getByText('# Runs')).toBeInTheDocument();
    expect(screen.getByText('# Failures')).toBeInTheDocument();
    expect(screen.getByText('Avg Runtime')).toBeInTheDocument();
    expect(screen.getByText('Total Runtime')).toBeInTheDocument();
    expect(screen.getByText('Avg Cores')).toBeInTheDocument();
  });
  it('should render the TableContainer with timeline view', () => {
    const params: Params = {
      page: 0,
      rowsPerPage: 25,
      filter: '',
      date: new Date(),
      period: Period.DAY,
      sort: SortType.SORT_AVG_CORES,
      ascending: true,
      timelineView: true,
    };
    renderWithContext(<ResourcesTable/>, { data: tests, params: params, datesToShow: ['1', '2'] });
    expect(screen.getByTestId('tableBody')).toBeInTheDocument();
    expect(screen.getByText('Test Suite')).toBeInTheDocument();
    expect(screen.getAllByTestId('timelineHeader')).toHaveLength(2);
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
