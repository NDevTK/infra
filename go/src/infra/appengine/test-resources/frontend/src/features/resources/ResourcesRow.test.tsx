// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { fireEvent, render } from '@testing-library/react';
import { MetricType, TestMetricsArray, TestVariantData } from '../../api/resources';
import MockMetrics from '../../utils/MockMetrics.json';
import ResourcesRow, { AggregatedMetrics, aggregateMetrics } from './ResourcesRow';

const mockMetricsWithData: Map<string, TestMetricsArray> =
  new Map<string, TestMetricsArray>(
      Object.entries(
          {
            '01-02-2012': {
              'data': [
                {
                  metric_type: MetricType.NUM_RUNS,
                  metric_value: 2,
                },
                {
                  metric_type: MetricType.AVG_CORES,
                  metric_value: 2,
                },
                {
                  metric_type: MetricType.AVG_RUNTIME,
                  metric_value: 2,
                },
                {
                  metric_type: MetricType.NUM_FAILURES,
                  metric_value: 2,
                },
                {
                  metric_type: MetricType.TOTAL_RUNTIME,
                  metric_value: 2,
                },
              ],
            },
            '01-03-2012': {
              'data': [
                {
                  metric_type: MetricType.NUM_RUNS,
                  metric_value: 4,
                },
                {
                  metric_type: MetricType.AVG_CORES,
                  metric_value: 4,
                },
                {
                  metric_type: MetricType.AVG_RUNTIME,
                  metric_value: 20,
                },
                {
                  metric_type: MetricType.NUM_FAILURES,
                  metric_value: 2,
                },
                {
                  metric_type: MetricType.TOTAL_RUNTIME,
                  metric_value: 2,
                },
              ],
            },
          },
      ),
  );

const aggregatedResults: AggregatedMetrics = {
  avgCores: 3,
  avgRuntime: 11,
  numFailures: 4,
  numRuns: 6,
  totalRuntime: 4,
};

const mockVariant: TestVariantData[] = [
  {
    suite: 'suite',
    builder: 'builder',
    metrics: mockMetricsWithData,
  },
];

const testDateMetricData = {
  test_id: MockMetrics[0].test_id,
  test_name: MockMetrics[0].test_name,
  file_name: MockMetrics[0].file_name,
  metrics: mockMetricsWithData,
  variants: mockVariant,
};


describe('when rendering the ResourcesRow', () => {
  it('should render the TableRow and VariantRow', () => {
    const { getByTestId } = render(
        <table>
          <tbody>
            <ResourcesRow {...testDateMetricData}/>
          </tbody>
        </table>,
    );
    const button = getByTestId('clickButton');
    fireEvent.click(button);
    const tableRow = getByTestId('tableRowTest');
    const variantRow = getByTestId('variantRowTest');
    expect(variantRow).toBeInTheDocument();
    expect(tableRow).toBeInTheDocument();
  });
});

describe('when calling aggregateMetrics', () => {
  it('should return an object with correct aggregation', () => {
    expect(aggregateMetrics(mockMetricsWithData)).toEqual(aggregatedResults);
  });
});
