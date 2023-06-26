// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createContext, useEffect, useState } from 'react';
import { MetricType, Period, SortType, TestMetricsArray, fetchTestMetrics } from '../../api/resources';

type MetricsContextProviderProps = {
  children: React.ReactNode
}

export interface Test {
  testId: string,
  testName: string,
  fileName: string,
  metrics: Map<MetricType, number>,
  variants: TestVariant[]
}

export interface TestVariant {
  suite: string,
  builder: string,
  metrics: Map<MetricType, number>
}

export interface MetricsContextValue {
  tests: Test[],
  page: number,
  lastPage: boolean,
  api: Api
}

export interface Api {
  // Page navigation
  nextPage: () => void,
  prevPage: () => void,
  firstPage: () => void,
}

export const MetricsContext = createContext<MetricsContextValue>(
    {
      tests: [],
      lastPage: true,
      page: 0,
      api: {
        nextPage: () => {
          // Do nothing
        },
        prevPage: () => {
          // Do nothing
        },
        firstPage: () => {
          // Do nothing
        },
      },
    },
);

export function createMetricsMap(metrics: Map<string, TestMetricsArray>): Map<MetricType, number> {
  let numRuns = 0;
  let numFailures = 0;
  let avgRuntime = 0;
  let totalRuntime = 0;
  let avgCores = 0;
  let fixedMetricsMap = metrics;

  // This is done because for testing, Object.entries on the map gives us an empty array
  // While the counterpart returned from the backend does not give us an empty array
  // despite both arguments being the same type. I will update this if I ever
  // find out the root cause of it. For now, adding this bandaid fix.
  if (new Map<string, TestMetricsArray>(Object.entries(metrics)).size !== 0) {
    fixedMetricsMap = new Map<string, TestMetricsArray>(Object.entries(metrics));
  }
  // We are just accessing the singular object in the map. But because it's a map
  // we "loop" anyways.
  fixedMetricsMap.forEach((data) => {
    data.data.forEach((metric) => {
      let metricType: MetricType = metric.metricType;
      let metricValue: number = metric.metricValue;
      if (metricValue === undefined) {
        metricValue = 0;
      }
      if (metricType === undefined) {
        metricType = 'NUM_RUNS' as MetricType;
      }
      switch (metricType) {
        case MetricType.NUM_RUNS:
          numRuns += metricValue;
          break;
        case MetricType.NUM_FAILURES:
          numFailures += metricValue;
          break;
        case MetricType.AVG_RUNTIME:
          avgRuntime += metricValue;
          break;
        case MetricType.TOTAL_RUNTIME:
          totalRuntime += metricValue;
          break;
        case MetricType.AVG_CORES:
          avgCores += metricValue;
          break;
        default:
          throw new Error('No metric type found for data - ' + String(metricType));
      }
    });
  });
  return new Map<MetricType, number>(
      [
        [MetricType.NUM_RUNS, numRuns],
        [MetricType.NUM_FAILURES, numFailures],
        [MetricType.AVG_RUNTIME, avgRuntime],
        [MetricType.TOTAL_RUNTIME, totalRuntime],
        [MetricType.AVG_CORES, avgCores],
      ],
  );
}

export const MetricsContextProvider = ({ children } : MetricsContextProviderProps) => {
  const [tests, setTests] = useState<Test[]>([]);
  const [lastPage, setLastPage] = useState(false);
  let [page, setPage] = useState(0);

  useEffect(() => {
    // Initialize MetricContextValue on mount
    fetchTestMetricsHelper();
    // eslint-disable-next-line react-hooks/exhaustive-deps
    // TODO: Figure out how to fix the lint error
  }, []);

  function fetchTestMetricsHelper() {
    return fetchTestMetrics({
      'component': 'Blink',
      'period': Period.DAY,
      'dates': ['2023-05-30'],
      'metrics': [
        MetricType.NUM_RUNS,
        MetricType.AVG_RUNTIME,
        MetricType.TOTAL_RUNTIME,
        MetricType.NUM_FAILURES,
        // MetricType.AVG_CORES,
      ],
      'page_offset': page * 25,
      'page_size': 25,
      'sort': { metric: SortType.SORT_NAME, ascending: true },
    }).then((resp) => {
      const tests: Test[] = [];
      // Populate Test
      for (const testDateMetricData of resp.tests) {
        const metrics = testDateMetricData.metrics;
        const testVariants: TestVariant[] = [];
        // Construct variants
        for (const testVariant of testDateMetricData.variants) {
          testVariants.push({
            suite: testVariant.suite,
            builder: testVariant.builder,
            metrics: createMetricsMap(testVariant.metrics),
          });
        }
        const newTest: Test = {
          testId: testDateMetricData.testId,
          testName: testDateMetricData.testName,
          fileName: testDateMetricData.fileName,
          metrics: createMetricsMap(metrics),
          variants: testVariants,
        };
        tests.push(newTest);
      }
      setTests(tests);
      setLastPage(resp.last_page);
    }).catch((error) => {
      throw error;
    });
  }

  const api: Api = {
    nextPage: () => {
      page++;
      setPage(page);
      fetchTestMetricsHelper();
    },
    prevPage: () => {
      if (page > 0) {
        page--;
        setPage(page);
        fetchTestMetricsHelper();
      }
    },
    firstPage: () => {
      page = 0;
      setPage(0);
      fetchTestMetricsHelper();
    },
  };

  return (
    <MetricsContext.Provider value={{ tests, page, lastPage, api }}>
      { children }
    </MetricsContext.Provider>
  );
};

export default MetricsContext;
