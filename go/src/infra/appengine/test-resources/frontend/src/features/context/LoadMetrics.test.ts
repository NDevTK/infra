// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {
  MetricType,
  Period,
  SortType,
  TestDateMetricData,
  TestMetricsDateMap } from '../../api/resources';
import { computeDates, dataReducer } from './LoadMetrics';

function metricsMap(
    metrics: {[date: string]: [MetricType, number][]},
): TestMetricsDateMap {
  const ret: TestMetricsDateMap = {};
  for (const date in metrics) {
    if (Object.hasOwn(metrics, date)) {
      ret[date] = {
        data: metrics[date].map(
            (tuple) => ({ metricType: tuple[0], metricValue: tuple[1] }),
        ),
      };
    }
  }
  return ret;
}

describe('computeDates', () => {
  const table = [
    [Period.DAY, false, new Date('2023-07-11T00:00:00'), ['2023-07-11']],
  ];
  it.each(table)(
      'period:%p timeline:%p date:%p',
      (period, timeline, date, expected) => {
        expect(computeDates({
          page: 0,
          rowsPerPage: 0,
          filter: '',
          date: date as Date,
          period: period as Period,
          sort: SortType.SORT_NAME,
          ascending: true,
          timelineView: timeline as boolean,
        })).toEqual(expected);
      });
});

describe('Merge TestMetrics', () => {
  it('populate tests with a single variant correctly', () => {
    const metrics = metricsMap({
      '2012-01-02': [
        [MetricType.NUM_RUNS, 1],
        [MetricType.NUM_FAILURES, 2],
      ],
    });
    const tests: TestDateMetricData[] = [{
      testId: '12',
      testName: 'name',
      fileName: 'file',
      metrics: metrics,
      variants: [
        {
          suite: 'suite',
          builder: 'builder',
          metrics: metrics,
        },
      ],
    }];
    const merged = dataReducer([], { type: 'merge_test', tests });
    expect(merged).toHaveLength(1);
    expect(merged[0].id).toEqual(tests[0].testId);
    expect(merged[0].name).toEqual(tests[0].testName);
    expect(merged[0].metrics.size).toEqual(1);
    expect(merged[0].metrics.get('2012-01-02')?.get(MetricType.NUM_RUNS))
        .toEqual(1);
    expect(merged[0].metrics.get('2012-01-02')?.get(MetricType.NUM_FAILURES))
        .toEqual(2);

    expect(merged[0].nodes).toHaveLength(1);
    const v = merged[0].nodes[0];
    expect(v.name).toEqual(tests[0].variants[0].builder);
    expect(v.subname).toEqual(tests[0].variants[0].suite);
    expect(v.metrics.size).toEqual(1);
    expect(v.metrics.get('2012-01-02')?.get(MetricType.NUM_RUNS)).toEqual(1);
    expect(v.metrics.get('2012-01-02')?.get(MetricType.NUM_FAILURES))
        .toEqual(2);
  });

  it('populate tests with a multiple dates correctly', () => {
    const metrics = metricsMap({
      '2012-01-02': [
        [MetricType.NUM_RUNS, 1],
        [MetricType.NUM_FAILURES, 2],
      ],
      '2012-01-03': [
        [MetricType.NUM_RUNS, 3],
        [MetricType.NUM_FAILURES, 4],
      ],
    });
    const tests: TestDateMetricData[] = [{
      testId: '12',
      testName: 'name',
      fileName: 'file',
      metrics: metrics,
      variants: [],
    }];
    const merged = dataReducer([], { type: 'merge_test', tests });
    expect(merged).toHaveLength(1);
    expect(merged[0].id).toEqual(tests[0].testId);
    expect(merged[0].name).toEqual(tests[0].testName);
    expect(merged[0].metrics.size).toEqual(2);
    expect(merged[0].metrics.get('2012-01-02')?.get(MetricType.NUM_RUNS))
        .toEqual(1);
    expect(merged[0].metrics.get('2012-01-02')?.get(MetricType.NUM_FAILURES))
        .toEqual(2);
    expect(merged[0].metrics.get('2012-01-03')?.get(MetricType.NUM_RUNS))
        .toEqual(3);
    expect(merged[0].metrics.get('2012-01-03')?.get(MetricType.NUM_FAILURES))
        .toEqual(4);
  });
});
