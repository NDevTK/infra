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
export const DIRECTORY_VIEW = 'dir';
export const SORT_INDEX = 'ind';

function ResourcesParamControls() {
  const { params } = useContext(MetricsContext);

  const [, setSearchParams] = useSearchParams();

  const updateParams = useCallback(() => {
    const newSearchParams = new URLSearchParams();
    if (params.page > 0 && !params.directoryView) {
      newSearchParams.set(PAGE, String(params.page));
    }
    if (!params.directoryView) {
      newSearchParams.set(ROWS_PER_PAGE, String(params.rowsPerPage));
    }
    if (params.filter !== '') {
      newSearchParams.set(FILTER, params.filter);
    }
    newSearchParams.set(DATE, formatDate(params.date));
    newSearchParams.set(PERIOD, String(params.period));
    newSearchParams.set(SORT_BY, String(params.sort));
    newSearchParams.set(ASCENDING, String(params.ascending));
    newSearchParams.set(TIMELINE_VIEW, String(params.timelineView));
    newSearchParams.set(DIRECTORY_VIEW, String(params.directoryView));
    if (params.timelineView) {
      newSearchParams.set(SORT_INDEX, String(params.sortIndex));
    }
    setSearchParams(newSearchParams);
    // We don't want a dependency on searchParams
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params]);

  useEffect(() => {
    updateParams();
  }, [updateParams]);

  return (<></>);
}

export default ResourcesParamControls;
