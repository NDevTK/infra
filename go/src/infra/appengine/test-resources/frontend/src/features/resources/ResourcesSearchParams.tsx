// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { useCallback, useContext, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { MetricsContext } from '../context/MetricsContext';

export const PAGE = 'page';
export const ROWS_PER_PAGE = 'rows';
export const FILTER = 'filter';
export const DATE = 'date';
export const PERIOD = 'period';
export const SORT_BY = 'sort';
export const ASCENDING = 'asc';

function ResourcesParamControls() {
  const { params } = useContext(MetricsContext);

  const [search, setSearchParams] = useSearchParams();

  const updateParams = useCallback(() => {
    search.set(PAGE, String(params.page));
    search.set(ROWS_PER_PAGE, String(params.rowsPerPage));
    search.set(FILTER, params.filter);
    search.set(DATE, params.date);
    search.set(PERIOD, String(params.period));
    search.set(SORT_BY, String(params.sort));
    search.set(ASCENDING, String(params.ascending));
    setSearchParams(search);
  }, [search, setSearchParams, params]);

  useEffect(() => {
    updateParams();
  }, [updateParams]);

  return (<></>);
}

export default ResourcesParamControls;
