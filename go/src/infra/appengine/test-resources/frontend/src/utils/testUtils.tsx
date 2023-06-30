// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { ReactElement } from 'react';
import { MetricsContext, Api, Test, MetricsContextValue, Params } from '../features/context/MetricsContext';
import { Period, SortType } from '../api/resources';

export interface OptionalContext {
  tests?: Test[],
  lastPage?: boolean,
  isLoading?: boolean,
  api?: OptionalApi,
  params?: Params,
}

export interface OptionalApi {
    // Page navigation
    setPage: (page: number) => void,
    setRowsPerPage: (rowsPerPage: number) => void,

    // Filter related Apis
    setFilter: (filter: string) => void,
    setDate: (date: string) => void,
    setPeriod: (period: Period) => void,
    setSort: (sort: SortType) => void,
    setAscending: (ascending: boolean) => void,
}

const defaultApi: Api = {
  setPage: () => {/**/},
  setRowsPerPage: () => {/**/},
  setFilter: () => {/**/},
  setDate: () => {/**/},
  setPeriod: () => {/**/},
  setSort: () => {/**/},
  setAscending: () => {/**/},
};

const defaultParams: Params = {
  page: 0,
  rowsPerPage: 25,
  filter: '',
  date: '2023-05-30',
  period: Period.DAY,
  sort: SortType.SORT_NAME,
  ascending: true,
};

export function renderWithContext(
    ui: ReactElement,
    opts: OptionalContext,
) {
  const ctx : MetricsContextValue = {
    tests: opts.tests || [],
    lastPage: opts.lastPage || true,
    api: {
      // Page navigation
      setPage: opts.api?.setPage || defaultApi.setPage,
      setRowsPerPage: opts.api?.setRowsPerPage || defaultApi.setRowsPerPage,
      setFilter: opts.api?.setFilter || defaultApi.setFilter,
      setDate: opts.api?.setDate || defaultApi.setDate,
      setPeriod: opts.api?.setPeriod || defaultApi.setPeriod,
      setSort: opts.api?.setSort || defaultApi.setSort,
      setAscending: opts.api?.setAscending || defaultApi.setAscending,
    },
    params: {
      page: opts.params?.page || defaultParams.page,
      rowsPerPage: opts.params?.rowsPerPage || defaultParams.rowsPerPage,
      filter: opts.params?.filter || defaultParams.filter,
      date: opts.params?.date || defaultParams.date,
      period: opts.params?.period || defaultParams.period,
      sort: opts.params?.sort || defaultParams.sort,
      ascending: opts.params?.ascending || defaultParams.ascending,
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
