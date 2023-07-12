// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createContext, useEffect, useState } from 'react';
import { MetricType, Period, SortType, TestMetricsArray, fetchTestMetrics } from '../../api/resources';
import { formatDate } from '../../utils/formatUtils';

type MetricsContextProviderProps = {
  page?: number,
  timelineView?: boolean,
  children: React.ReactNode,
}

export interface Node {
  id: string,
  name: string,
  subname?: string,
  metrics: Map<string, Map<MetricType, number>>,
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
  datesToShow: string[],
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
  timelineView: boolean,
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
    updateTimelineView: (timelineView: boolean) => void,
}

export const MetricsContext = createContext<MetricsContextValue>(
    {
      data: [],
      datesToShow: [] as string[],
      lastPage: true,
      api: {
        updatePage: () => {/**/},
        updateRowsPerPage: () => {/**/},
        updateFilter: () => {/**/},
        updateDate: () => {/**/},
        updatePeriod: () => {/**/},
        updateSort: () => {/**/},
        updateAscending: () => {/**/},
        updateTimelineView: () => {/**/},
      },
      params: {
        page: 0,
        rowsPerPage: 0,
        filter: '',
        date: new Date(),
        period: Period.DAY,
        sort: SortType.SORT_NAME,
        ascending: true,
        timelineView: false,
      },
      isLoading: false,
    },
);

export function createTimelineViewMetricsMap(metrics: Map<string, TestMetricsArray>): Map<string, number> {
  const computedMetricsMap = new Map<string, number>();
  let fixedMetricsMap = metrics;

  if (new Map<string, TestMetricsArray>(Object.entries(metrics)).size !== 0) {
    fixedMetricsMap = new Map<string, TestMetricsArray>(Object.entries(metrics));
  }
  fixedMetricsMap.forEach((data, date) => {
    let avgRuntime = 0;
    // Obtain avg_runtime metric
    data.data.forEach((metric) => {
      if (metric.metricType === MetricType.AVG_RUNTIME) {
        avgRuntime = metric.metricValue;
      }
    });
    computedMetricsMap.set(date, avgRuntime);
  });
  return computedMetricsMap;
}

export function createMetricsMap(metrics: Map<string, TestMetricsArray>): Map<string, Map<MetricType, number>> {
  let fixedMetricsMap = metrics;
  // This is done because for testing, Object.entries on the map gives us an empty array
  // While the counterpart returned from the backend does not give us an empty array
  // despite both arguments being the same type. I will update this if I ever
  // find out the root cause of it. For now, adding this bandaid fix.
  if (new Map<string, TestMetricsArray>(Object.entries(metrics)).size !== 0) {
    fixedMetricsMap = new Map<string, TestMetricsArray>(Object.entries(metrics));
  }

  const metricsMap = new Map<string, Map<MetricType, number>>();
  fixedMetricsMap.forEach((data, date) => {
    const metricToVal = new Map<MetricType, number>();
    data.data.forEach((metric) => {
      metricToVal.set(metric.metricType, metric.metricValue);
    });
    metricsMap.set(date, metricToVal);
  });
  return metricsMap;
}

function computeDates(date: Date, period: Period, datesBefore: number): string[] {
  const computedDates = [] as string[];
  for (let x = datesBefore; x >= 0; x--) {
    const newDate = formatDate(new Date(new Date().setDate(new Date(date).getDate() - (x * (period === Period.DAY ? 1 : 7)))));
    computedDates.push(newDate);
  }
  return computedDates;
}

export const MetricsContextProvider = (props : MetricsContextProviderProps) => {
  const [data, setData] = useState<Node[]>([]);
  const [lastPage, setLastPage] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [datesToShow, setDatesToShow] = useState<string[]>([formatDate(new Date(Date.now() - 86400000))]);
  let [page, setPage] = useState(props.page || 0);
  let [rowsPerPage, setRowsPerPage] = useState(50);
  let [filter, setFilter] = useState('');
  let [date, setDate] = useState(new Date(Date.now() - 86400000));
  let [period, setPeriod] = useState(Period.DAY);
  let [sort, setSort] = useState(SortType.SORT_NAME);
  let [ascending, setAscending] = useState(true);
  let [timelineView, setTimelineView] = useState(props.timelineView || false);

  let loadingCount = 0;

  function fetchTestMetricsHelper() {
    setIsLoading(true);
    loadingCount++;
    const datesToFetch = computeDates(date, period, timelineView ? 4 : 0);
    return fetchTestMetrics({
      'component': 'Blink',
      'period': Number(period) as Period,
      'dates': datesToFetch,
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
      setDatesToShow(datesToFetch);
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
    updateTimelineView: (newTimelineView: boolean) => {
      timelineView = newTimelineView;
      fetchTestMetricsHelper();
      setTimelineView(newTimelineView);
    },
  };

  const params: Params = { page, rowsPerPage, filter, date, period, sort, ascending, timelineView };

  return (
    <MetricsContext.Provider value={{ data, lastPage, isLoading, api, params, datesToShow }}>
      { props.children }
    </MetricsContext.Provider>
  );
};

export default MetricsContext;
