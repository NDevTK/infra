// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createContext, useCallback, useContext, useEffect, useMemo, useReducer, useState } from 'react';
import {
  DirectoryNode,
  DirectoryNodeType,
  FetchDirectoryMetricsResponse,
  FetchTestMetricsResponse,
  MetricType,
  Period,
  SortType,
  TestDateMetricData,
  isTestMetricsResponse,
} from '../../api/resources';
import { formatDate } from '../../utils/formatUtils';
import { ComponentContext } from '../components/ComponentContext';
import { dataReducer, getLoadedParentIds, loadDirectoryMetrics, loadTestMetrics } from './LoadMetrics';

type MetricsContextProviderProps = {
  page: number,
  rowsPerPage: number,
  filter: string,
  date: Date,
  period: Period,
  sort: SortType,
  ascending: boolean,
  sortIndex: number,
  timelineView: boolean,
  directoryView: boolean,
  children?: React.ReactNode,
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
  type: DirectoryNodeType,
  loaded: boolean,
}

export function isPath(object: any): object is Path {
  return 'path' in object;
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
  sortIndex: number,
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
    updateSortIndex: (index: number) => void,

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
        updateSortIndex: () => {/**/},
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
        sortIndex: 0,
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

export function convertToSortIndex(datesToShow: string[], date: Date ) {
  return datesToShow.findIndex((c) => {
    return c === formatDate(date);
  });
}

function snapToPeriod(date: Date) {
  const ret = new Date(date);
  ret.setDate(ret.getDate() - ret.getDay());
  return ret;
}
export const MetricsContextProvider = (props : MetricsContextProviderProps) => {
  const { components } = useContext(ComponentContext);
  const [page, setPage] = useState(props.page);
  const [rowsPerPage, setRowsPerPage] = useState(props.rowsPerPage);
  const [filter, setFilter] = useState(props.filter);
  const [date, setDate] = useState(props.period === Period.WEEK ? snapToPeriod(props.date) : props.date);
  const [period, setPeriod] = useState(props.period);
  const [sort, setSort] = useState(props.sort);
  const [ascending, setAscending] = useState(props.ascending);
  const [sortIndex, setSortIndex] = useState(props.sortIndex);
  const [timelineView, setTimelineView] = useState(props.timelineView);
  const [directoryView, setDirectoryView] = useState(props.directoryView);

  const params: Params = useMemo(() => ({
    page, rowsPerPage, filter, date, period, sort, ascending, sortIndex,
    timelineView, directoryView,
  }), [
    page, rowsPerPage, filter, date, period, sort, ascending, sortIndex,
    timelineView, directoryView,
  ]);

  const [data, dataDispatch] = useReducer(dataReducer, []);
  const [lastPage, setLastPage] = useState(false);
  const [datesToShow, setDatesToShow] = useState<string[]>([formatDate(date)]);
  const [loading, loadingDispatch] = useReducer(loadingCountReducer, { count: 0, isLoading: false });

  const loadFailure = useCallback((error: any) => {
    loadingDispatch({ type: 'end' });
    throw error;
  }, [loadingDispatch]);

  const loadPathNode = useCallback((node: Node) => {
    if (isPath(node) && !node.loaded) {
      loadingDispatch({ type: 'start' });
      if (node.type === DirectoryNodeType.FILENAME) {
        loadTestMetrics(components, params,
            (response: FetchTestMetricsResponse) => {
              dataDispatch({
                type: 'merge_test',
                tests: response.tests,
                parentId: node.id,
              });
              loadingDispatch({ type: 'end' });
            },
            loadFailure,
            [node.path],
        );
      } else {
        loadDirectoryMetrics(components, params, [node.id],
            (response: FetchDirectoryMetricsResponse) => {
              dataDispatch({
                type: 'merge_dir',
                nodes: response.nodes,
                parentId: node.id,
                onExpand: loadPathNode,
              });
              loadingDispatch({ type: 'end' });
            },
            loadFailure,
        );
      }
    }
  }, [loadingDispatch, dataDispatch, loadFailure, components, params]);

  const load = useCallback((_from: string, components: string[], params: Params) => {
    loadingDispatch({ type: 'start' });
    if (params.directoryView) {
      // If we're not switching to directory view, we will need to reload
      // the tree with the current loaded/expanded state.
      if (directoryView) {
        const [directories, filenames] = getLoadedParentIds(data);

        // The rebuildState callback allows us to dispatch both RPC requests
        // at the same time and merge the data once we get both responses back,
        // as opposed to chaining promises, which would lead to sequential reqs.
        let directoryNodes: DirectoryNode[] | undefined;
        let tests: TestDateMetricData[] | undefined;
        const rebuildState = (
            response: FetchDirectoryMetricsResponse | FetchTestMetricsResponse,
            fetchedDates: string[],
        ) => {
          if (isTestMetricsResponse(response)) {
            tests = response.tests;
          } else {
            // A empty directory response will have no nodes field
            directoryNodes = response.nodes || [];
          }
          if (directoryNodes === undefined || tests === undefined) {
            return;
          }
          loadingDispatch({ type: 'end' });
          dataDispatch({
            type: 'rebuild_state',
            nodes: directoryNodes,
            tests: tests,
            onExpand: loadPathNode,
          });
          setTimelineView(params.timelineView);
          setDatesToShow(fetchedDates);
        };

        loadDirectoryMetrics(
            components, params, ['/', ...directories],
            rebuildState, loadFailure,
        );
        if (filenames.length > 0) {
          loadTestMetrics(components, {
            ...params,
            page: 0,
            rowsPerPage: 1000,
          }, rebuildState, loadFailure, filenames);
        } else {
          tests = [];
        }
      } else {
        loadDirectoryMetrics(
            components,
            params,
            ['/'],
            (response: FetchDirectoryMetricsResponse, fetchedDates: string[]) => {
              loadingDispatch({ type: 'end' });
              dataDispatch({
                type: 'merge_dir',
                nodes: response.nodes,
                onExpand: loadPathNode,
              });
              setTimelineView(params.timelineView);
              setDirectoryView(params.directoryView);
              setDatesToShow(fetchedDates);
            },
            loadFailure,
        );
      }
    } else {
      loadTestMetrics(
          components,
          params,
          (response: FetchTestMetricsResponse, fetchedDates: string[]) => {
            loadingDispatch({ type: 'end' });
            dataDispatch({ type: 'merge_test', tests: response.tests });
            setLastPage(response.lastPage);
            setTimelineView(params.timelineView);
            setDirectoryView(params.directoryView);
            setDatesToShow(fetchedDates);
          },
          loadFailure,
      );
    }
  }, [
    data, directoryView,
    loadPathNode, loadingDispatch, dataDispatch, loadFailure,
    setTimelineView, setDirectoryView, setDatesToShow, setLastPage,
  ]);

  useEffect(() => {
    load('useEffect components', components, params);
    // We don't want to run this effect every time load or params change
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [components]);

  const api: Api = {
    updatePage: (newPage: number) => {
      if (page !== newPage) {
        params.page = newPage;
        setPage(newPage);
        load('updatePage', components, params);
      }
    },
    updateRowsPerPage: (newRowsPerPage: number) => {
      if (rowsPerPage !== newRowsPerPage) {
        params.rowsPerPage = newRowsPerPage;
        setRowsPerPage(params.rowsPerPage);
        load('updateRowsPerPage', components, params);
      }
    },
    updateFilter: (newFilter: string) => {
      if (filter !== newFilter) {
        params.filter = newFilter;
        params.page = 0;
        setFilter(params.filter);
        setPage(params.page);
        load('updateFilter', components, params);
      }
    },
    updateDate: (newDate: Date) => {
      if (date.getTime() !== newDate.getTime()) {
        params.date = newDate;
        params.page = 0;
        setDate(params.date);
        setPage(params.page);
        params.sortIndex = params.timelineView ? 4 : 0;
        setSortIndex(params.sortIndex);
        load('updateDate', components, params);
      }
    },
    updatePeriod: (newPeriod: Period) => {
      if (period !== newPeriod) {
        params.period = newPeriod;
        params.page = 0;
        // Snap to valid date for weekly view
        if (newPeriod === Period.WEEK) {
          params.date = (snapToPeriod(params.date));
          setDate(params.date);
        }
        setPeriod(params.period);
        setPage(params.page);
        load('updatePeriod', components, params);
      }
    },
    updateSort: (newSort: SortType) => {
      if (sort !== newSort) {
        params.sort = newSort;
        params.page = 0;
        setSort(params.sort);
        setPage(params.page);
        load('updateSort', components, params);
      }
    },
    updateAscending: (newAscending: boolean) => {
      if (ascending !== newAscending) {
        params.ascending = newAscending;
        params.page = 0;
        setAscending(params.ascending);
        setPage(params.page);
        load('updateAscending', components, params);
      }
    },
    updateSortIndex: (newSortIndex: number) => {
      if (sortIndex !== newSortIndex) {
        params.sortIndex = newSortIndex;
        setSortIndex(params.sortIndex);
        load('updateSortIndex', components, params);
      }
    },
    updateTimelineView: (newTimelineView: boolean) => {
      if (timelineView !== newTimelineView) {
        params.timelineView = newTimelineView;
        params.sortIndex = params.timelineView ? 4 : 0;
        setSortIndex(params.sortIndex);
        // Don't set timeline view until the data has been loaded.
        load('updateTimelineView', components, params);
      }
    },
    updateDirectoryView: (newDirectoryView: boolean) => {
      if (directoryView !== newDirectoryView) {
        params.directoryView = newDirectoryView;
        // Don't set directory view until the data has been loaded.
        load('updateDirectoryView', components, params);
      }
    },
  };

  return (
    <MetricsContext.Provider value={{ data, datesToShow, lastPage, isLoading: loading.isLoading, api, params }}>
      { props.children }
    </MetricsContext.Provider>
  );
};

export default MetricsContext;
