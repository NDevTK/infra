// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createContext, useEffect, useState } from 'react';
import { MetricType, Period, SortType, TestMetricsArray, fetchTestMetrics } from '../../api/resources';
import { formatDate } from '../../utils/formatUtils';

type MetricsContextProviderProps = {
  page?: number,
  children: React.ReactNode,
}

export interface Node {
  id: string,
  name: string,
  subname?: string,
  metrics: Map<MetricType, number>,
  isLeaf: boolean,
  nodes: Node[]
}

// This node is for a single test, which may have multiple variants
export interface Test extends Node {
  fileName: string,
}

// This node is for a single variant, which is a test run in a particular
// configuration (builder, suite)
export type TestVariant = Node

export interface MetricsContextValue {
  data: Node[],
  lastPage: boolean,
  isLoading: boolean,
  api: Api,
  params: Params,
}

export interface Params {
  page: number,
  rowsPerPage: number,
  filter: string,
  date: Date,
  period: Period,
  sort: SortType,
  ascending: boolean,
}

export interface Api {
    // Page navigation
    updatePage: (page: number) => void,
    updateRowsPerPage: (rowsPerPage: number) => void,

    // Test selection-related APIs
    updateFilter: (filter: string) => void,
    updateDate: (date: Date) => void,
    updatePeriod: (period: Period) => void,
    updateSort: (sort: SortType) => void,
    updateAscending: (ascending: boolean) => void,
}

export const MetricsContext = createContext<MetricsContextValue>(
    {
      data: [],
      lastPage: true,
      api: {
        updatePage: () => {/**/},
        updateRowsPerPage: () => {/**/},
        updateFilter: () => {/**/},
        updateDate: () => {/**/},
        updatePeriod: () => {/**/},
        updateSort: () => {/**/},
        updateAscending: () => {/**/},
      },
      params: {
        page: 0,
        rowsPerPage: 0,
        filter: '',
        date: new Date(),
        period: Period.DAY,
        sort: SortType.SORT_NAME,
        ascending: true,
      },
      isLoading: false,
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

export const MetricsContextProvider = (props : MetricsContextProviderProps) => {
  const [data, setData] = useState<Node[]>([]);
  const [lastPage, setLastPage] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  let [page, setPage] = useState(props.page || 0);
  let [rowsPerPage, setRowsPerPage] = useState(50);
  let [filter, setFilter] = useState('');
  let [date, setDate] = useState(new Date(Date.now() - 86400000));
  let [period, setPeriod] = useState(Period.DAY);
  let [sort, setSort] = useState(SortType.SORT_NAME);
  let [ascending, setAscending] = useState(true);

  let loadingCount = 0;

  function fetchTestMetricsHelper() {
    setIsLoading(true);
    loadingCount++;
    return fetchTestMetrics({
      'component': 'Blink',
      'period': Number(period) as Period,
      'dates': [formatDate(date)],
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
        for (const test of resp.tests) {
          const metrics = test.metrics;
          const newTest: Test = {
            id: test.testId,
            name: test.testName,
            fileName: test.fileName,
            metrics: createMetricsMap(metrics),
            isLeaf: false,
            nodes: [],
          };
          // Construct variants
          for (const variant of test.variants) {
            newTest.nodes.push({
              id: newTest.id + ':' + variant.builder + ':' + variant.suite,
              name: variant.builder,
              subname: variant.suite,
              metrics: createMetricsMap(variant.metrics),
              isLeaf: true,
              nodes: [],
            });
          }
          tests.push(newTest);
        }
      }
      setData(tests);
      setLastPage(resp.lastPage);
      loadingCount--;
      setIsLoading(loadingCount !== 0);
    }).catch((error) => {
      loadingCount--;
      setIsLoading(loadingCount !== 0);
      throw error;
    });
  }

  useEffect(() => {
    fetchTestMetricsHelper();
  // Adding this because we don't want a dependency on api
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const api: Api = {
    updatePage: (newPage: number) => {
      page = newPage;
      fetchTestMetricsHelper();
      setPage(newPage);
    },
    updateRowsPerPage: (newRowsPerPage: number) => {
      rowsPerPage = newRowsPerPage;
      fetchTestMetricsHelper();
      setRowsPerPage(newRowsPerPage);
    },
    updateFilter: (newFilter: string) => {
      page = 0;
      filter = newFilter;
      fetchTestMetricsHelper();
      setFilter(newFilter);
      setPage(0);
    },
    updateDate: (newDate: Date) => {
      date = newDate;
      page = 0;
      fetchTestMetricsHelper();
      setDate(newDate);
      setPage(0);
    },
    updatePeriod: (newPeriod: Period) => {
      period = newPeriod;
      page = 0;
      fetchTestMetricsHelper();
      setPeriod(newPeriod);
      setPage(0);
    },
    updateSort: (newSort: SortType) => {
      sort = newSort;
      page = 0;
      fetchTestMetricsHelper();
      setSort(newSort);
      setPage(0);
    },
    updateAscending: (newAscending: boolean) => {
      ascending = newAscending;
      page = 0;
      fetchTestMetricsHelper();
      setAscending(newAscending);
      setPage(0);
    },
  };

  const params: Params = { page, rowsPerPage, filter, date, period, sort, ascending };

  return (
    <MetricsContext.Provider value={{ data, lastPage, isLoading, api, params }}>
      { props.children }
    </MetricsContext.Provider>
  );
};

export default MetricsContext;
