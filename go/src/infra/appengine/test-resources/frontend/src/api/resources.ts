// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import MockMetrics from '../utils/MockMetrics.json';

export interface TestDateMetricData {
  test_id: string,
  test_name: string,
  file_name: string,
  // Note that string represent date in "XX-XX-XXXX format"
  metrics: Map<string, TestMetricsArray>,
  variants: TestVariantData[],
}

export interface TestMetricsArray {
  data: TestMetricsData[]
}

interface TestMetricsData {
  metric_type: MetricType,
  metric_value: number
}

export enum MetricType {
  NUM_RUNS = 'NUM_RUNS',
  NUM_FAILURES = 'NUM_FAILURES',
  AVG_RUNTIME = 'AVG_RUNTIME',
  TOTAL_RUNTIME = 'TOTAL_RUNTIME',
  AVG_CORES = 'AVG_CORES',
}

export interface TestVariantData {
  suite: string,
  builder: string,
  metrics: Map<string, TestMetricsArray>
}

export async function fetchTestDateMetricData(
): Promise<TestDateMetricData[]> {
  const mockDataArray: TestDateMetricData[] = [];
  MockMetrics.forEach((metric) => {
    const dateToTestMetricsArrayMap = new Map<string, TestMetricsArray>(
        Object.entries(metric.metrics),
    );
    const fixedMap = new Map<string, TestMetricsArray>();
    dateToTestMetricsArrayMap.forEach((data, date)=> {
      const testMetricsDataArray: TestMetricsData[] = [];
      data.data.forEach((testMetricsData) => {
        testMetricsDataArray.push(
            {
              metric_type: testMetricsData.metric_type as MetricType,
              metric_value: testMetricsData.metric_value,
            },
        );
      });
      fixedMap.set(date, { data: testMetricsDataArray });
    });
    mockDataArray.push({
      test_id: metric.test_id,
      test_name: metric.test_name,
      file_name: metric.file_name,
      metrics: fixedMap,
      variants: metric.variants ? metric.variants : new Array(0),
    });
  });
  return mockDataArray;
}
