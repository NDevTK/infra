// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {
  FetchTestMetricsRequest,
  FetchTestMetricsResponse,
  MetricType,
  Period,
  TestDateMetricData,
  TestMetricsDateMap,
  fetchTestMetrics } from '../../api/resources';
import { formatDate } from '../../utils/formatUtils';
import { Node, Params } from './MetricsContext';

export function computeDates(params: Params): string[] {
  const computedDates: string[] = [];
  const datesBefore = params.timelineView ? 4 : 0;
  for (let x = datesBefore; x >= 0; x--) {
    const newDate = new Date(params.date);
    newDate.setDate(
        params.date.getDate() - (x * (params.period === Period.DAY ? 1 : 7)),
    );
    computedDates.push(formatDate(newDate));
  }
  return computedDates;
}

export function loadTestMetrics(
    params: Params,
    successCallback: (response: FetchTestMetricsResponse, fetchedDates: string[]) => void,
    failureCallback: (erorr: any) => void,
) {
  const datesToFetch = computeDates(params);
  const request: FetchTestMetricsRequest = {
    component: 'Blink',
    period: params.period,
    dates: datesToFetch,
    metrics: [
      MetricType.NUM_RUNS,
      MetricType.AVG_RUNTIME,
      MetricType.TOTAL_RUNTIME,
      MetricType.NUM_FAILURES,
      MetricType.AVG_CORES,
    ],
    filter: params.filter,
    page_offset: params.page * params.rowsPerPage,
    page_size: params.rowsPerPage,
    sort: {
      metric: params.sort,
      ascending: params.ascending,
    },
  };
  fetchTestMetrics(request).then((response) => {
    successCallback(response, datesToFetch);
  }).catch(failureCallback);
}

type DataAction =
 | { type: 'merge_test', tests: TestDateMetricData[] }

export function dataReducer(state: Node[], action: DataAction): Node[] {
  switch (action.type) {
    case 'merge_test':
      return action.tests.map((test) => ({
        id: test.testId,
        name: test.testName,
        fileName: test.fileName,
        metrics: createMetricsMap(test.metrics),
        isLeaf: false,
        nodes: test.variants.map((variant) => ({
          id: `${test.testId}:${variant.builder}:${variant.suite}`,
          name: variant.builder,
          subname: variant.suite,
          metrics: createMetricsMap(test.metrics),
          isLeaf: true,
          nodes: [],
        })),
      }));
  }
  return state;
}

function createMetricsMap(
    metrics: TestMetricsDateMap,
): Map<string, Map<MetricType, number>> {
  const ret = new Map<string, Map<MetricType, number>>();
  for (const date in metrics) {
    if (Object.hasOwn(metrics, date)) {
      const metricMap = new Map<MetricType, number>();
      metrics[date].data.forEach((metric) => {
        metricMap.set(metric.metricType, metric.metricValue);
      });
      ret.set(date, metricMap);
    }
  }
  return ret;
}

