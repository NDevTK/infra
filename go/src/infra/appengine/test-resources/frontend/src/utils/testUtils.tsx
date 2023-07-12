// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { ReactElement } from 'react';
import { MetricsContext, Api, Node, MetricsContextValue, Params } from '../features/context/MetricsContext';
import { Period, SortType } from '../api/resources';

export interface OptionalContext {
  data?: Node[],
  datesToShow?: string[],
  lastPage?: boolean,
  isLoading?: boolean,
  api?: OptionalApi,
  params?: Params,
  isTimelineView?: boolean,
}

export interface OptionalApi {
  // Page navigation
  updatePage: (page: number) => void,
  updateRowsPerPage: (rowsPerPage: number) => void,

  // Filter related Apis
  updateFilter: (filter: string) => void,
  updateDate: (date: Date) => void,
  updatePeriod: (period: Period) => void,
  updateSort: (sort: SortType) => void,
  updateAscending: (ascending: boolean) => void,
  updateTimelineView: (timelineView: boolean) => void,
}

const defaultApi: Api = {
  updatePage: () => {/**/},
  updateRowsPerPage: () => {/**/},
  updateFilter: () => {/**/},
  updateDate: () => {/**/},
  updatePeriod: () => {/**/},
  updateSort: () => {/**/},
  updateAscending: () => {/**/},
  updateTimelineView: () => {/**/},
};

const defaultParams: Params = {
  page: 0,
  rowsPerPage: 25,
  filter: '',
  date: new Date(),
  period: Period.DAY,
  sort: SortType.SORT_NAME,
  ascending: true,
  timelineView: false,
};

export function renderWithContext(
    ui: ReactElement,
    opts: OptionalContext,
) {
  const ctx : MetricsContextValue = {
    data: opts.data || [],
    datesToShow: opts.datesToShow || [],
    lastPage: opts.lastPage || true,
    api: {
      updatePage: opts.api?.updatePage || defaultApi.updatePage,
      updateRowsPerPage: opts.api?.updateRowsPerPage || defaultApi.updateRowsPerPage,
      updateFilter: opts.api?.updateFilter || defaultApi.updateFilter,
      updateDate: opts.api?.updateDate || defaultApi.updateDate,
      updatePeriod: opts.api?.updatePeriod || defaultApi.updatePeriod,
      updateSort: opts.api?.updateSort || defaultApi.updateSort,
      updateAscending: opts.api?.updateAscending || defaultApi.updateAscending,
      updateTimelineView: opts.api?.updateTimelineView || defaultApi.updateTimelineView,
    },
    params: {
      page: opts.params?.page || defaultParams.page,
      rowsPerPage: opts.params?.rowsPerPage || defaultParams.rowsPerPage,
      filter: opts.params?.filter || defaultParams.filter,
      date: opts.params?.date || defaultParams.date,
      period: opts.params?.period || defaultParams.period,
      sort: opts.params?.sort || defaultParams.sort,
      ascending: opts.params?.ascending || defaultParams.ascending,
      timelineView: opts.params?.timelineView || defaultParams.timelineView,
    },
    isLoading: opts.isLoading || true,
  };
  render(
      <BrowserRouter>
        <MetricsContext.Provider value= {ctx}>
          {ui}
        </MetricsContext.Provider>
      </BrowserRouter>,
  );
}
