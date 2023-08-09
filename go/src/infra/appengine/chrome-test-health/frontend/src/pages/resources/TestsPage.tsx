// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Box } from '@mui/material';
import TestMetricsTable from '../../features/resources/tests/TestMetricsTable';
import { TestMetricsContextProvider } from '../../features/resources/tests/TestMetricsContext';
import TestMetricsToolbar from '../../features/resources/tests/TestMetricsToolbar';
import TestMetricsSearchParams, {
  ASCENDING,
  DATE,
  DIRECTORY_VIEW,
  FILTER,
  PAGE,
  PERIOD,
  ROWS_PER_PAGE,
  SORT_BY,
  SORT_INDEX,
  TIMELINE_VIEW,
  TIMELINE_VIEW_METRIC,
} from '../../features/resources/tests/TestMetricsSearchParams';
import { MetricType, Period, SortType } from '../../api/resources';

function TestsPage() {
  const params = new URLSearchParams(window.location.search);
  const props = {
    page: Number(params.get(PAGE)) || 0,
    rowsPerPage: Number(
        params.get(ROWS_PER_PAGE) || localStorage.getItem(ROWS_PER_PAGE),
    ) || 50,
    filter: params.get(FILTER) || '',
    date: new Date(
      params.has(DATE) ? params.get(DATE) + 'T00:00:00' : (Date.now()),
    ),
    period: params.has(PERIOD) ? Number(params.get(PERIOD)) as Period : Period.WEEK,
    sort: params.has(SORT_BY) ? Number(params.get(SORT_BY)) as SortType : SortType.SORT_AVG_CORES,
    ascending: params.get(ASCENDING) === 'true',
    sortIndex: Number(params.get(SORT_INDEX)) || 0,
    timelineMetric: params.has(TIMELINE_VIEW_METRIC) ?
      String(params.get(TIMELINE_VIEW_METRIC)) as MetricType : MetricType.AVG_CORES,
    timelineView: params.get(TIMELINE_VIEW) === 'true',
    directoryView: params.get(DIRECTORY_VIEW) === 'true',
  };
  return (
    <TestMetricsContextProvider {...props}>
      <TestMetricsToolbar/>
      <Box sx={{ margin: '10px 20px' }}>
        <TestMetricsTable/>
      </Box>
      <TestMetricsSearchParams/>
    </TestMetricsContextProvider>
  );
}

export default TestsPage;
