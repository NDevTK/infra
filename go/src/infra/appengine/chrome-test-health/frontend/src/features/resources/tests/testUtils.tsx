// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { BrowserRouter } from 'react-router-dom';
import { ReactElement } from 'react';
import { MetricType, Period, SortType } from '../../../api/resources';
import { renderWithAuth } from '../../auth/testUtils';
import {
  Node,
  Params,
  TestMetricsContext,
  TestMetricsContextValue,
} from './TestMetricsContext';

export interface OptionalContext {
  data?: Node[],
  datesToShow?: string[],
  lastPage?: boolean,
  isLoading?: boolean,
  api?: {
    // Page navigation
    updatePage?: (page: number) => void,
    updateRowsPerPage?: (rowsPerPage: number) => void,

    // Filter related Apis
    updateFilter?: (filter: string) => void,
    updateDate?: (date: Date) => void,
    updatePeriod?: (period: Period) => void,
    updateSort?: (sort: SortType) => void,
    updateAscending?: (ascending: boolean) => void,
    updateSortDate?: (date: string) => void,
    updateSortIndex?: (index: number) => void,
    updateTimelineMetric?: (metric: MetricType) => void,
    updateTimelineView?: (timelineView: boolean) => void,
    updateDirectoryView?: (directoryView: boolean) => void,
  },
  params?: OptionalParams,
  isTimelineView?: boolean,
  isDirectoryView?: boolean,
}

type OptionalParams = {
  page?: number,
  rowsPerPage?: number,
  filter?: string,
  date?: Date,
  period?: Period,
  sort?: SortType,
  ascending?: boolean,
  sortIndex?: number,
  timelineMetric?: MetricType,
  timelineView?: boolean,
  directoryView?: boolean,
}

export function createParams(params? : OptionalParams) : Params {
  return {
    page: params?.page || 0,
    rowsPerPage: params?.rowsPerPage || 50,
    filter: params?.filter || '',
    date: params?.date || new Date('2023-01-02'),
    period: params?.period || Period.WEEK,
    sort: params?.sort || SortType.SORT_NAME,
    ascending: (params?.ascending === undefined ? true : params.ascending),
    sortIndex: params?.sortIndex || 0,
    timelineMetric: params?.timelineMetric || MetricType.AVG_CORES,
    timelineView: params?.timelineView || false,
    directoryView: params?.directoryView || false,
  };
}

const defaultApi = () => {/**/};

export function renderWithContext(
    ui: ReactElement,
    opts: OptionalContext = {},
) {
  const ctx : TestMetricsContextValue = {
    data: opts.data || [],
    datesToShow: opts.datesToShow || [],
    lastPage: (opts.lastPage === undefined ? true : opts.lastPage),
    api: {
      updatePage: opts.api?.updatePage || defaultApi,
      updateRowsPerPage: opts.api?.updateRowsPerPage || defaultApi,
      updateFilter: opts.api?.updateFilter || defaultApi,
      updateDate: opts.api?.updateDate || defaultApi,
      updatePeriod: opts.api?.updatePeriod || defaultApi,
      updateSort: opts.api?.updateSort || defaultApi,
      updateAscending: opts.api?.updateAscending || defaultApi,
      updateSortIndex: opts.api?.updateSortIndex || defaultApi,
      updateTimelineMetric: opts.api?.updateTimelineMetric || defaultApi,
      updateTimelineView: opts.api?.updateTimelineView || defaultApi,
      updateDirectoryView: opts.api?.updateDirectoryView || defaultApi,
    },
    params: createParams(opts.params),
    isLoading: (opts.isLoading === undefined ? false : opts.isLoading),
  };
  return renderWithAuth(
      <BrowserRouter>
        <TestMetricsContext.Provider value= {ctx}>
          {ui}
        </TestMetricsContext.Provider>
      </BrowserRouter>,
  );
}

