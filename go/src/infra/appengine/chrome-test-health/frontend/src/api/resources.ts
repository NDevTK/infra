// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Auth } from './auth';
import { prpcClient } from './client';

export interface ListComponentsResponse {
  components: string[],
}

export async function listComponents(auth: Auth):
  Promise<ListComponentsResponse> {
  const resp: ListComponentsResponse = await prpcClient.call(
      auth,
      'test_resources.Stats',
      'ListComponents',
      {},
  );
  if (resp.components === undefined) {
    resp.components = [];
  }
  return resp;
}

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
}

export interface TestVariantData {
  suite: string,
  builder: string,
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

export function isTestMetricsResponse(
    object: any,
): object is FetchTestMetricsResponse {
  return 'lastPage' in object;
}

// Protobuf json transport drops certain fields/values if they are the default.
// This function replaces the defaults.
function fixFetchTestMetricsResponse(resp: FetchTestMetricsResponse) {
  if (resp.tests === undefined) {
    resp.tests = [];
  }
  resp.tests.forEach((test) => {
    fixMetricsDateMap(test.metrics);
    if (test.variants === undefined) {
      test.variants = [];
    } else {
      test.variants.forEach((variant) => fixMetricsDateMap(variant.metrics));
    }
  });
}

function fixMetricsDateMap(map: MetricsDateMap) {
  Object.keys(map).forEach((date) => {
    map[date].data.forEach((data) => {
      if (data.metricValue === undefined) {
        data.metricValue = 0;
      }
    });
  });
}

export async function fetchTestMetrics(
    auth: Auth,
    fetchTestMetricsRequest: FetchTestMetricsRequest,
): Promise<FetchTestMetricsResponse> {
  const resp: FetchTestMetricsResponse = await prpcClient.call(
      auth,
      'test_resources.Stats',
      'FetchTestMetrics',
      fetchTestMetricsRequest,
  );
  fixFetchTestMetricsResponse(resp);
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

function fixFetchDirectoryMetricsResponse(resp: FetchDirectoryMetricsResponse) {
  if (resp.nodes === undefined) {
    resp.nodes = [];
  } else {
    resp.nodes.forEach((node) => fixMetricsDateMap(node.metrics));
  }
}

export async function fetchDirectoryMetrics(
    auth: Auth,
    request: FetchDirectoryMetricsRequest,
): Promise<FetchDirectoryMetricsResponse> {
  const resp: FetchDirectoryMetricsResponse = await prpcClient.call(
      auth,
      'test_resources.Stats',
      'FetchDirectoryMetrics',
      request,
  );
  fixFetchDirectoryMetricsResponse(resp);
  return resp;
}
