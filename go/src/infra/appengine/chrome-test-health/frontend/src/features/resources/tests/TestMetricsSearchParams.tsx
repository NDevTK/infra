// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { useContext, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { formatDate } from '../../../utils/formatUtils';
import { ComponentContext, updateComponentsUrl } from '../../components/ComponentContext';
import { Params, TestMetricsContext } from './TestMetricsContext';

export const PAGE = 'p';
export const ROWS_PER_PAGE = 'rows';
export const FILTER = 'filter';
export const DATE = 'date';
export const PERIOD = 'period';
export const SORT_BY = 'sort';
export const ASCENDING = 'asc';
export const TIMELINE_VIEW = 'tl';
export const TIMELINE_VIEW_METRIC = 'tlm';
export const DIRECTORY_VIEW = 'dir';
export const SORT_INDEX = 'sidx';

export function createSearchParams(components: string[], params: Params) {
  const search = new URLSearchParams();
  // Unfortunately, having two search params objects in the dom tree seems to
  // create a race condition as they overwrite each other's parameters, even
  // with functional updates.
  updateComponentsUrl(components, search);
  if (params.page > 0 && !params.directoryView) {
    search.set(PAGE, String(params.page));
  } else {
    search.delete(PAGE);
  }
  if (!params.directoryView) {
    search.set(ROWS_PER_PAGE, String(params.rowsPerPage));
    localStorage.setItem(ROWS_PER_PAGE, String(params.rowsPerPage));
  } else {
    search.delete(ROWS_PER_PAGE);
  }
  if (params.filter !== '') {
    search.set(FILTER, params.filter);
  } else {
    search.delete(FILTER);
  }
  search.set(DATE, formatDate(params.date));
  search.set(PERIOD, String(params.period));
  search.set(SORT_BY, String(params.sort));
  search.set(ASCENDING, String(params.ascending));
  search.set(TIMELINE_VIEW, String(params.timelineView));
  search.set(DIRECTORY_VIEW, String(params.directoryView));
  if (params.timelineView) {
    search.set(SORT_INDEX, String(params.sortIndex));
    search.set(TIMELINE_VIEW_METRIC, String(params.timelineMetric));
  } else {
    search.delete(TIMELINE_VIEW_METRIC);
    search.delete(SORT_INDEX);
  }
  return search;
}

function TestMetricsSearchParams() {
  const { params } = useContext(TestMetricsContext);
  const { components } = useContext(ComponentContext);

  const [, setSearchParams] = useSearchParams();

  useEffect(() => {
    setSearchParams(createSearchParams(components, params));
  }, [setSearchParams, params, components]);

  return (<></>);
}

export default TestMetricsSearchParams;
