// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {
  FetchTestMetricsResponse,
  MetricType,
  Period,
  SortType,
  MetricsDateMap,
  fetchTestMetrics,
  prpcClient,
  DirectoryNodeType,
  fetchDirectoryMetrics,
  FetchDirectoryMetricsResponse,
  FetchTestMetricsRequest,
  FetchDirectoryMetricsRequest,
} from './resources';

const mockMetricsWithData: MetricsDateMap = {
  '2012-01-02': {
    data: [
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
};

describe('fetchTestMetrics', () => {
  const dummyRequest: FetchTestMetricsRequest = {
    'components': ['component'],
    'period': Period.DAY,
    'dates': ['date'],
    'metrics': [
      MetricType.NUM_RUNS,
      MetricType.AVG_RUNTIME,
      MetricType.TOTAL_RUNTIME,
      MetricType.NUM_FAILURES,
    ],
    'filter': 'filter',
    'page_offset': 0,
    'page_size': 0,
    'sort': {
      metric: SortType.SORT_NAME,
      ascending: true,
      sort_date: '2012-01-02',
    },
  };
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
              bucket: 'try',
              metrics: mockMetricsWithData,
            },
          ],
        },
      ],
      lastPage: false,
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
              bucket: 'try',
              metrics: mockMetricsWithData,
            },
          ],
        },
      ],
      lastPage: false,
    };
    const resp = await fetchTestMetrics(dummyRequest);

    expect(mockCall.mock.calls.length).toBe(1);
    expect(mockCall.mock.calls[0].length).toBe(3);
    expect(mockCall.mock.calls[0][0]).toBe('test_resources.Stats');
    expect(mockCall.mock.calls[0][1]).toBe('FetchTestMetrics');
    expect(mockCall.mock.calls[0][2]).toEqual(dummyRequest);
    expect(resp).toEqual(expected);
  });

  it('returns a response with tests', async () => {
    jest.spyOn(prpcClient, 'call').mockResolvedValue({
      lastPage: false,
    });
    const resp = await fetchTestMetrics(dummyRequest);
    expect(resp.tests).toHaveLength(0);
  });

  it('returns a response with metricValues with 0', async () => {
    jest.spyOn(prpcClient, 'call').mockResolvedValue({
      tests: [
        {
          testId: '1',
          testName: 'A',
          fileName: 'A',
          metrics: {
            '2012-01-02': {
              data: [{
                metricType: 'NUM_RUNS' as MetricType,
              }],
            },
          },
        },
      ],
      lastPage: false,
    });
    const resp = await fetchTestMetrics(dummyRequest);
    expect(resp.tests[0]?.metrics['2012-01-02']?.data[0].metricValue).toBe(0);
  });
});

describe('fetchDirectoryMetrics', () => {
  const dummyRequest: FetchDirectoryMetricsRequest = {
    components: ['component'],
    period: Period.DAY,
    dates: ['2012-01-02'],
    parent_ids: ['/'],
    metrics: [
      MetricType.NUM_RUNS,
      MetricType.AVG_RUNTIME,
      MetricType.TOTAL_RUNTIME,
      MetricType.NUM_FAILURES,
    ],
    filter: 'filter',
    sort: {
      metric: SortType.SORT_NAME,
      ascending: true,
      sort_date: '2012-01-02',
    },
  };
  it('returns metrics', async () => {
    const data: FetchDirectoryMetricsResponse = {
      nodes: [
        {
          id: '//a',
          type: DirectoryNodeType.DIRECTORY,
          name: 'a',
          metrics: mockMetricsWithData,
        },
        {
          id: '//b',
          type: DirectoryNodeType.FILENAME,
          name: 'b',
          metrics: mockMetricsWithData,
        },
      ],
    };
    jest.spyOn(prpcClient, 'call').mockResolvedValue(data);
    const resp = await fetchDirectoryMetrics(dummyRequest);
    expect(resp).toEqual(data);
  });

  it('returns a response with nodes', async () => {
    jest.spyOn(prpcClient, 'call').mockResolvedValue({});
    const resp = await fetchDirectoryMetrics(dummyRequest);
    expect(resp.nodes).toHaveLength(0);
  });
});
