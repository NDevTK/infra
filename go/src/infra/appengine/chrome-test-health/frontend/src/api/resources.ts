// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

export interface TestDateMetricData {
  testId: string,
  testName: string,
  fileName: string,
  metrics: MetricsDateMap,
  variants: TestVariantData[],
}

// Note that string represent date in "YYYY-MM-DD format"
export type MetricsDateMap = {[key: string]: TestMetricsArray}

export interface TestMetricsArray {
  data: TestMetricsData[]
}

export interface TestMetricsData {
  metricType: MetricType,
  metricValue: number
}

export enum MetricType {
  UNKNOWN_METRIC = 'UNKNOWN_METRIC',
  NUM_RUNS = 'NUM_RUNS',
  NUM_FAILURES = 'NUM_FAILURES',
  TOTAL_RUNTIME = 'TOTAL_RUNTIME',
  AVG_CORES = 'AVG_CORES',
  AVG_RUNTIME = 'AVG_RUNTIME',
  P50_RUNTIME = 'P50_RUNTIME',
  P90_RUNTIME = 'P90_RUNTIME',
}

export interface TestVariantData {
  suite: string,
  builder: string,
  project: string,
  bucket: string,
  metrics: MetricsDateMap,
}

export enum Period {
  UNKNOWN_PERIOD = 0,
  DAY = 1,
  WEEK = 2,
  MONTH = 3,
}

export interface FetchTestMetricsResponse {
  tests: TestDateMetricData[],
  lastPage: boolean,
}

export enum SortType {
  UNKNOWN_SORTTYPE = 0,
  SORT_NAME = 1,
  SORT_NUM_RUNS = 2,
  SORT_NUM_FAILURES = 3,
  SORT_TOTAL_RUNTIME = 4,
  SORT_AVG_CORES = 5,
  SORT_AVG_RUNTIME = 6,
  SORT_P50_RUNTIME = 7,
  SORT_P90_RUNTIME = 8,
 }

export interface SortBy {
  metric: SortType,
  ascending: boolean,
  sort_date: string,
}

export interface FetchTestMetricsRequest {
  components: string[],
  period: Period,
  dates: string[],
  metrics: MetricType[],
  filter?: string,
  file_names?: string[],
  page_offset: number,
  page_size: number,
  sort: SortBy,
}

export interface ListComponentsResponse {
  components: string[],
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
      return JSON.parse(text.substring(4));
    } else {
      throw text;
    }
  },
};

export async function listComponents(): Promise<ListComponentsResponse> {
  const resp: ListComponentsResponse = await prpcClient.call(
      'test_resources.Stats',
      'ListComponents',
      {},
  );
  return resp;
}

export function isTestMetricsResponse(
    object: any,
): object is FetchTestMetricsResponse {
  return 'lastPage' in object;
}

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

export interface FetchDirectoryMetricsRequest {
  components: string[],
  period: Period,
  dates: string[],
  parent_ids: string[],
  metrics: MetricType[],
  filter?: string,
  sort: SortBy,
}

export enum DirectoryNodeType {
  UNKNOWN_NODE_TYPE = 'UNKNOWN',
  DIRECTORY = 'DIRECTORY',
  FILENAME = 'FILENAME',
}

export interface DirectoryNode {
  id: string,
  type: DirectoryNodeType,
  name: string,
  metrics: MetricsDateMap,
}

export interface FetchDirectoryMetricsResponse {
  nodes: DirectoryNode[],
}

export async function fetchDirectoryMetrics(
    request: FetchDirectoryMetricsRequest,
): Promise<FetchDirectoryMetricsResponse> {
  const resp: FetchDirectoryMetricsResponse = await prpcClient.call(
      'test_resources.Stats',
      'FetchDirectoryMetrics',
      request,
  );
  return resp;
}
