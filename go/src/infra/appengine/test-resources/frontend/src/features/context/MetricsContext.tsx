// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createContext, useCallback, useEffect, useState } from 'react';
import { MetricType, Period, SortType, TestMetricsArray, fetchTestMetrics } from '../../api/resources';
import { formatDate } from '../../utils/formatUtils';

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
  lastPage: boolean,
  api: Api,
  params: Params,
}

export interface Params {
  page: number,
  rowsPerPage: number,
  filter: string,
  date: string,
  period: Period,
  sort: SortType,
  ascending: boolean,
}

export interface Api {
  // Page navigation
  setPage: (page: number) => void,
  setRowsPerPage: (rowsPerPage: number) => void,

  // Test selection-related APIs
  setFilter: (filter: string) => void,
  setDate: (date: string) => void,
  setPeriod: (period: Period) => void,
  setSort: (sort: SortType) => void,
  setAscending: (ascending: boolean) => void,
}

export const MetricsContext = createContext<MetricsContextValue>(
    {
      tests: [],
      lastPage: true,
      api: {
        setPage: () => {/**/},
        setRowsPerPage: () => {/**/},
        setFilter: () => {/**/},
        setDate: () => {/**/},
        setPeriod: () => {/**/},
        setSort: () => {/**/},
        setAscending: () => {/**/},
      },
      params: {
        page: 0,
        rowsPerPage: 0,
        filter: '',
        date: '',
        period: Period.DAY,
        sort: SortType.SORT_NAME,
        ascending: true,
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
  const [rowsPerPage, setRowsPerPage] = useState(50);
  const [filter, setFilter] = useState('');
  const [date, setDate] = useState(formatDate(new Date()));
  const [period, setPeriod] = useState(Period.DAY);
  const [sort, setSort] = useState(SortType.SORT_NAME);
  const [ascending, setAscending] = useState(true);

  const fetchTestMetricsHelper = useCallback(() => {
    return fetchTestMetrics({
      'component': 'Blink',
      'period': Number(period) as Period,
      'dates': [date],
      'metrics': [
        MetricType.NUM_RUNS,
        MetricType.AVG_RUNTIME,
        MetricType.TOTAL_RUNTIME,
        MetricType.NUM_FAILURES,
        // MetricType.AVG_CORES,
      ],
      'filter': filter,
      'page_offset': page * rowsPerPage,
      'page_size': rowsPerPage,
      'sort': { metric: Number(sort) as SortType, ascending: ascending },
    }).then((resp) => {
      const tests: Test[] = [];
      // Populate Test
      if (resp.tests !== undefined) {
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
      }
      setTests(tests);
      setLastPage(resp.lastPage);
    }).catch((error) => {
      throw error;
    });
  }, [page, rowsPerPage, filter, date, period, sort, ascending]);

  useEffect(() => {
    fetchTestMetricsHelper();
  }, [fetchTestMetricsHelper]);

  const api: Api = {
    setPage,
    setRowsPerPage,
    setFilter,
    setDate,
    setPeriod,
    setSort,
    setAscending,
  };

  return (
    <MetricsContext.Provider value={{ tests, lastPage, api, params: { page, rowsPerPage, filter, date, period, sort, ascending } }}>
      { children }
    </MetricsContext.Provider>
  );
};

export default MetricsContext;
