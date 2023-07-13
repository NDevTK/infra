// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

export interface TestDateMetricData {
  testId: string,
  testName: string,
  fileName: string,
  // Note that string represent date in "YYYY-MM-DD format"
  metrics: TestMetricsDateMap,
  variants: TestVariantData[],
}

export type TestMetricsDateMap = {[key: string]: TestMetricsArray}

export interface TestMetricsArray {
  data: TestMetricsData[]
}

export interface TestMetricsData {
  metricType: MetricType,
  metricValue: number
}

export enum MetricType {
  UNKNOWN_METRIC = 'UNKOWN_METRIC',
  NUM_RUNS = 'NUM_RUNS',
  NUM_FAILURES = 'NUM_FAILURES',
  AVG_RUNTIME = 'AVG_RUNTIME',
  TOTAL_RUNTIME = 'TOTAL_RUNTIME',
  AVG_CORES = 'AVG_CORES',
}

export interface TestVariantData {
  suite: string,
  builder: string,
  metrics: TestMetricsDateMap,
}

export enum Period {
  DAY = 0,
  WEEK = 1,
  MONTH = 2,
}

export interface FetchTestMetricsResponse {
  tests: TestDateMetricData[],
  lastPage: boolean,
}

export enum SortType {
  SORT_NAME = 0,
  SORT_NUM_RUNS = 1,
  SORT_NUM_FAILURES = 2,
  SORT_AVG_RUNTIME = 3,
  SORT_TOTAL_RUNTIME = 4,
  SORT_AVG_CORES = 5,
 }

export interface SortBy {
  metric: SortType,
  ascending: boolean,
}

export interface FetchTestMetricsRequest {
  component: string,
  period: Period,
  dates: string[],
  metrics: MetricType[],
  filter?: string,
  page_offset: number,
  page_size: number,
  sort: SortBy,
}

export const prpcClient = {
  call: async function <Type>(
      service: string,
      method: string,
      message: unknown,
  ): Promise<Type> {
    const url = `/prpc/${service}/${method}`;
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      body: JSON.stringify(message),
    });
    const text = await response.text();
    if (text.startsWith(')]}\'')) {
      return JSON.parse(text.substr(4));
    } else {
      return JSON.parse(text);
    }
  },
};

export async function fetchTestMetrics(
    fetchTestMetricsRequest: FetchTestMetricsRequest,
): Promise<FetchTestMetricsResponse> {
  const resp: FetchTestMetricsResponse = await prpcClient.call(
      'test_resources.Stats',
      'FetchTestMetrics',
      fetchTestMetricsRequest,
  );
  return resp;
}
