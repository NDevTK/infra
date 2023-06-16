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

interface MetricsContextValue {
  tests: Test[],
  lastPage: boolean,
  api: Api
}

export interface Api {
  // Page navigation
  nextPage: () => void,
  prevPage: () => void,
  firstPage: () => void,
  updateRowsPerPage: (rowsPerPage: number) => void,

  // Filter related Apis
  updateFilter: (filter: string) => void,
  updateDate: (date: string) => void,
  updatePeriod: (period: Period) => void,
  updateComponent: (component: string) => void,
}

export const MetricsContext = createContext<MetricsContextValue>(
    {
      tests: [],
      lastPage: true,
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
        updateFilter: () => {
          // Do nothing
        },
        updateDate: () => {
          // Do nothing
        },
        updatePeriod: () => {
          // Do nothing
        },
        updateRowsPerPage: () => {
          // Do nothing
        },
        updateComponent: () => {
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
  const [page, setPage] = useState(0);
  const [filter, setFilter] = useState('');
  const [period, setPeriod] = useState(Period.DAY);
  const [rowsPerPage, setRowsPerPage] = useState(5);
  const [date, setDate] = useState('2023-05-30');
  const [component, setComponent] = useState('Blink');

  useEffect(() => {
    // Initialize MetricContextValue on mount
    fetchTestMetricsHelper();
  }, []);

  function fetchTestMetricsHelper() {
    return fetchTestMetrics({
      'component': component,
      'period': period,
      'dates': [date],
      'metrics': [
        MetricType.NUM_RUNS,
        MetricType.AVG_RUNTIME,
        MetricType.TOTAL_RUNTIME,
        MetricType.NUM_FAILURES,
        // MetricType.AVG_CORES,
      ],
      'filter': filter,
      'page_offset': page,
      'page_size': rowsPerPage,
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
      setPage(page + 1);
      fetchTestMetricsHelper();
    },
    prevPage: () => {
      setPage(page - 1);
      fetchTestMetricsHelper();
    },
    firstPage: () => {
      setPage(0);
      fetchTestMetricsHelper();
    },
    updateFilter: (newFilter: string) => {
      setFilter(newFilter);
      setPage(0);
      fetchTestMetricsHelper();
    },
    updateDate: (newDate: string) => {
      setDate(newDate);
      setPage(0);
      fetchTestMetricsHelper();
    },
    updatePeriod: (newPeriod: Period) => {
      setPeriod(newPeriod);
      setPage(0);
      fetchTestMetricsHelper();
    },
    updateRowsPerPage: (newRowsPerPage: number) => {
      setRowsPerPage(newRowsPerPage),
      fetchTestMetricsHelper();
    },
    updateComponent: (newComponent: string) => {
      setComponent(newComponent),
      fetchTestMetricsHelper();
    },
  };

  return (
    <MetricsContext.Provider value={{ tests, lastPage, api }}>
      { children }
    </MetricsContext.Provider>
  );
};

export default MetricsContext;
