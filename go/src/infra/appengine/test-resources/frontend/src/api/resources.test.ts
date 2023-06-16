// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { FetchTestMetricsResponse, MetricType, SortType, TestMetricsArray, fetchTestMetrics, prpcClient } from './resources';

const mockMetricsWithData: Map<string, TestMetricsArray> =
  new Map<string, TestMetricsArray>(
      Object.entries(
          {
            '2012-01-02': {
              'data': [
                {
                  metricType: 'NUM_RUNS' as MetricType,
                  metricValue: 2,
                },
                {
                  metricType: 'NUM_FAILURES' as MetricType,
                  metricValue: 2,
                },
                {
                  metricType: 'AVG_RUNTIME' as MetricType,
                  metricValue: 2,
                },
                {
                  metricType: 'TOTAL_RUNTIME' as MetricType,
                  metricValue: 2,
                },
                {
                  metricType: 'AVG_CORES' as MetricType,
                  metricValue: 2,
                },
              ],
            },
          },
      ),
  );

describe('fetchTestMetrics', () => {
  it('returns metrics', async () => {
    const mockCall = jest.spyOn(prpcClient, 'call').mockResolvedValue({
      tests: [
        {
          testId: '1',
          testName: 'A',
          fileName: 'A',
          metrics: mockMetricsWithData,
          variants: [
            {
              suite: 'suite',
              builder: 'builder',
              metrics: mockMetricsWithData,
            },
          ],
        },
      ],
      last_page: false,
    });
    const expected: FetchTestMetricsResponse = {
      tests: [
        {
          testId: '1',
          testName: 'A',
          fileName: 'A',
          metrics: mockMetricsWithData,
          variants: [
            {
              suite: 'suite',
              builder: 'builder',
              metrics: mockMetricsWithData,
            },
          ],
        },
      ],
      last_page: false,
    };
    const resp = await fetchTestMetrics(
        {
          'component': 'component',
          'period': 'period',
          'dates': ['date'],
          'metrics': [
            MetricType.NUM_RUNS,
            MetricType.AVG_RUNTIME,
            MetricType.TOTAL_RUNTIME,
            MetricType.NUM_FAILURES,
            // AVG_CORES is currently unsupported.
            // MetricType.AVG_CORES,
          ],
          'filter': 'filter',
          'page_offset': 'page',
          'page_size': 'rowsPerPage',
          'sort': { metric: SortType.SORT_NAME, ascending: true },
        },
    );

    expect(mockCall.mock.calls.length).toBe(1);
    expect(mockCall.mock.calls[0].length).toBe(3);
    expect(mockCall.mock.calls[0][0]).toBe('test_resources.Stats');
    expect(mockCall.mock.calls[0][1]).toBe('FetchTestMetrics');
    expect(mockCall.mock.calls[0][2]).toEqual({
      component: 'component',
      period: 'period',
      dates: ['date'],
      metrics: ['NUM_RUNS', 'AVG_RUNTIME', 'TOTAL_RUNTIME', 'NUM_FAILURES'],
      filter: 'filter',
      page_offset: 'page',
      page_size: 'rowsPerPage',
      sort: { metric: 0, ascending: true },
    });
    expect(resp).toEqual(expected);
  });
});
