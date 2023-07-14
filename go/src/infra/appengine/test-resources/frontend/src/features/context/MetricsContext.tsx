// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createContext, useEffect, useReducer, useState } from 'react';
import {
  FetchDirectoryMetricsResponse,
  FetchTestMetricsResponse,
  MetricType,
  Period,
  SortType,
} from '../../api/resources';
import { formatDate } from '../../utils/formatUtils';
import { dataReducer, loadDirectoryMetrics, loadTestMetrics } from './LoadMetrics';

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
  onExpand?: (node: Node) => void,
  nodes: Node[]
}

// This node is for a file system path, which may be a directory or file
// A directory may contain multiple files. A file may contain multiple tests.
export interface Path extends Node {
  path: string,
  loaded: boolean,
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
  directoryView: boolean,
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
    updateDirectoryView: (directoryView: boolean) => void,
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
        updateDirectoryView: () => {/**/},
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
        directoryView: false,
      },
      isLoading: false,
    },
);

interface LoadingState {
  count: number,
  isLoading: boolean,
}

type LoadingAction =
 | { type: 'start' }
 | { type: 'end' }

function loadingCountReducer(state: LoadingState, action: LoadingAction): LoadingState {
  const newState = { ...state };
  switch (action.type) {
    case 'start':
      newState.count++;
      break;
    case 'end':
      newState.count--;
      break;
  }
  newState.isLoading = newState.count !== 0;
  return newState;
}

export const MetricsContextProvider = (props : MetricsContextProviderProps) => {
  const [page, setPage] = useState(props.page || 0);
  const [rowsPerPage, setRowsPerPage] = useState(50);
  const [filter, setFilter] = useState('');
  const [date, setDate] = useState(new Date(Date.now() - 86400000));
  const [period, setPeriod] = useState(Period.DAY);
  const [sort, setSort] = useState(SortType.SORT_TOTAL_RUNTIME);
  const [ascending, setAscending] = useState(false);
  const [timelineView, setTimelineView] = useState(props.timelineView || false);
  const [directoryView, setDirectoryView] = useState(false);

  const params: Params = { page, rowsPerPage, filter, date, period, sort, ascending, timelineView, directoryView };

  const [data, dataDispatch] = useReducer(dataReducer, []);
  const [lastPage, setLastPage] = useState(false);
  const [datesToShow, setDatesToShow] = useState<string[]>([formatDate(date)]);
  const [loading, loadingDispatch] = useReducer(loadingCountReducer, { count: 0, isLoading: false });

  function loadFailure(error: any) {
    loadingDispatch({ type: 'end' });
    throw error;
  }

  function loadPathNode(node: Node) {
    if (Object.hasOwn(node, 'loaded') && !(node as Path).loaded) {
      loadingDispatch({ type: 'start' });
      loadDirectoryMetrics(params, node.id,
          (response: FetchDirectoryMetricsResponse) => {
            dataDispatch({
              type: 'merge_dir',
              nodes: response.node,
              parentId: node.id,
              onExpand: loadPathNode,
            });
            loadingDispatch({ type: 'end' });
          },
          loadFailure,
      );
    }
  }

  function load(params: Params) {
    loadingDispatch({ type: 'start' });
    if (params.directoryView) {
      loadDirectoryMetrics(
          params,
          '/',
          (response: FetchDirectoryMetricsResponse, fetchedDates: string[]) => {
            dataDispatch({
              type: 'merge_dir',
              nodes: response.node,
              onExpand: loadPathNode,
            });
            setTimelineView(params.timelineView);
            setDirectoryView(params.directoryView);
            setDatesToShow(fetchedDates);
            loadingDispatch({ type: 'end' });
          },
          loadFailure,
      );
    } else {
      loadTestMetrics(
          params,
          (response: FetchTestMetricsResponse, fetchedDates: string[]) => {
            dataDispatch({ type: 'merge_test', tests: response.tests });
            setLastPage(response.lastPage);
            setTimelineView(params.timelineView);
            setDirectoryView(params.directoryView);
            setDatesToShow(fetchedDates);
            loadingDispatch({ type: 'end' });
          },
          loadFailure,
      );
    }
  }

  useEffect(() => {
    load(params);
  // Adding this because we don't want a dependency on api
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const api: Api = {
    updatePage: (newPage: number) => {
      if (params.page !== newPage) {
        params.page = newPage;
        setPage(newPage);
        load(params);
      }
    },
    updateRowsPerPage: (newRowsPerPage: number) => {
      params.rowsPerPage = newRowsPerPage;
      setRowsPerPage(params.rowsPerPage);
      load(params);
    },
    updateFilter: (newFilter: string) => {
      params.filter = newFilter;
      params.page = 0;
      setFilter(params.filter);
      setPage(params.page);
      load(params);
    },
    updateDate: (newDate: Date) => {
      params.date = newDate;
      params.page = 0;
      setDate(params.date);
      setPage(params.page);
      load(params);
    },
    updatePeriod: (newPeriod: Period) => {
      params.period = newPeriod;
      params.page = 0;
      setPeriod(params.period);
      setPage(params.page);
      load(params);
    },
    updateSort: (newSort: SortType) => {
      params.sort = newSort;
      params.page = 0;
      setSort(params.sort);
      setPage(params.page);
      load(params);
    },
    updateAscending: (newAscending: boolean) => {
      params.ascending = newAscending;
      params.page = 0;
      setAscending(params.ascending);
      setPage(params.page);
      load(params);
    },
    updateTimelineView: (newTimelineView: boolean) => {
      params.timelineView = newTimelineView;
      // Don't set timeline view until the data has been loaded.
      load(params);
    },
    updateDirectoryView: (newDirectoryView: boolean) => {
      params.directoryView = newDirectoryView;
      // Don't set directory view until the data has been loaded.
      load(params);
    },
  };

  return (
    <MetricsContext.Provider value={{ data, datesToShow, lastPage, isLoading: loading.isLoading, api, params }}>
      { props.children }
    </MetricsContext.Provider>
  );
};

export default MetricsContext;
