// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { fireEvent, screen } from '@testing-library/react';
import { act } from 'react-dom/test-utils';
import { MetricType } from '../../../api/resources';
import { Node, Test } from './TestMetricsContext';
import { renderWithContext } from './testUtils';
import TestMetricsTable, { getFormatter } from './TestMetricsTable';

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
  isExpandable: true,
  rows: [
    {
      id: 'v1',
      name: 'suite',
      subname: 'builder',
      metrics: mockMetrics,
      isExpandable: false,
      rows: [],
    },
    {
      id: 'v1',
      name: 'suite',
      subname: 'builder',
      metrics: mockMetrics,
      isExpandable: false,
      rows: [],
    },
  ],
}];

describe('when rendering the ResourcesTable', () => {
  it('snapshot view', () => {
    renderWithContext(<TestMetricsTable expandRowId={[]}/>, { data: tests });
    expect(screen.getByTestId('tableBody')).toBeInTheDocument();
    expect(screen.getByText('Test Suite')).toBeInTheDocument();
    expect(screen.getByText('# Runs')).toBeInTheDocument();
    expect(screen.getByText('# Failures')).toBeInTheDocument();
    expect(screen.getByText('Avg Runtime')).toBeInTheDocument();
    expect(screen.getByText('Total Runtime')).toBeInTheDocument();
    expect(screen.getByText('Avg Cores')).toBeInTheDocument();
  });
  it('timeline view', () => {
    renderWithContext(<TestMetricsTable expandRowId={[]}/>, { data: tests, params: { timelineView: true }, datesToShow: ['1', '2'] });
    expect(screen.getByTestId('tableBody')).toBeInTheDocument();
    expect(screen.getByText('Test Suite')).toBeInTheDocument();
  });
  it('directory view', () => {
    const nodes: Node[] = [{
      id: '/',
      name: 'src',
      metrics: mockMetrics,
      isExpandable: true,
      rows: [],
    }];
    renderWithContext(<TestMetricsTable expandRowId={[]}/>, { data: nodes, params: { directoryView: true }, datesToShow: ['1', '2'] });
    expect(screen.getByTestId('tableBody')).toBeInTheDocument();
    expect(screen.getByText('src')).toBeInTheDocument();
    expect(screen.queryByTestId('tablePagination')).toBeNull();
  });
});

describe('when rendering the ResourcesTable', () => {
  it('should render loading screen', () => {
    renderWithContext(<TestMetricsTable expandRowId={[]}/>, { data: [], isLoading: true });
    expect(screen.getByText('Loading...')).toBeInTheDocument();
    expect(screen.getByTestId('loading-bar')).not.toHaveClass(
        'hidden',
    );
  });
});

describe('when clicking copylink in ResourcesTable', () => {
  beforeEach(() => {
    let clipboardData = '';
    const mockClipboard = {
      writeText: jest.fn(
          (data) => {
            clipboardData = data;
          },
      ),
      readText: jest.fn(
          () => {
            return clipboardData;
          },
      ),
    };
    Object.assign(navigator, {
      clipboard: mockClipboard,
    });
  });
  it('should copy link to clipboard', async () => {
    const nodes: Node[] = [{
      id: 'pathName/',
      name: 'src',
      metrics: mockMetrics,
      isExpandable: true,
      rows: [
        {
          id: 'v1',
          name: 'suite',
          subname: 'builder',
          metrics: mockMetrics,
          isExpandable: false,
          rows: [],
        },
        {
          id: 'v1',
          name: 'suite',
          subname: 'builder',
          metrics: mockMetrics,
          isExpandable: false,
          rows: [],
        },
      ],
    }];

    renderWithContext(<TestMetricsTable expandRowId={[]}/>, { data: nodes, params: { directoryView: true } });
    await act(async () => {
      fireEvent.click(screen.getByTestId('LinkIcon'));
    });
    expect(navigator.clipboard.readText()).toContain('&expp=pathName/');
  });
});

// Test getFormatter
test.each([
  [MetricType.TOTAL_RUNTIME, 10000, '2h 46m'],
  [MetricType.AVG_RUNTIME, 500, '8m 20s'],
  [MetricType.AVG_CORES, 1000, '1,000'],
  [MetricType.NUM_FAILURES, 1, '1'],
])('.getFormatter(%p)(%p)', (metricType, metricValue, expected) => {
  expect(getFormatter(metricType)(metricValue)).toBe(expected);
});
