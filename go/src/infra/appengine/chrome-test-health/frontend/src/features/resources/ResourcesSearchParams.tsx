// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { useCallback, useContext, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { MetricsContext } from '../context/MetricsContext';
import { formatDate } from '../../utils/formatUtils';

export const PAGE = 'page';
export const ROWS_PER_PAGE = 'rows';
export const FILTER = 'filter';
export const DATE = 'date';
export const PERIOD = 'period';
export const SORT_BY = 'sort';
export const ASCENDING = 'asc';
export const TIMELINE_VIEW = 'tl';
export const TIMELINE_VIEW_METRIC = 'tlm';
export const DIRECTORY_VIEW = 'dir';
export const SORT_INDEX = 'ind';

function ResourcesParamControls() {
  const { params } = useContext(MetricsContext);

  const [, setSearchParams] = useSearchParams();

  const updateParams = useCallback((search: URLSearchParams) => {
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
  }, [params]);

  useEffect(() => {
    setSearchParams(updateParams);
  }, [setSearchParams, updateParams]);

  return (<></>);
}

export default ResourcesParamControls;
