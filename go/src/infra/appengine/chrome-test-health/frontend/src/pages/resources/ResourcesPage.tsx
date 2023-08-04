// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Box } from '@mui/material';
import ResourcesTable from '../../features/resources/ResourcesTable';
import { MetricsContextProvider } from '../../features/context/MetricsContext';
import ResourcesToolbar from '../../features/resources/ResourcesToolbar';
import ResourcesSearchParams, {
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
} from '../../features/resources/ResourcesSearchParams';
import ComponentParams from '../../features/components/ComponentParams';
import { MetricType, Period, SortType } from '../../api/resources';

function ResourcesPage() {
  const params = new URLSearchParams(window.location.search);
  const props = {
    page: Number(params.get(PAGE) || 0),
    rowsPerPage: Number(
        params.get(ROWS_PER_PAGE) ||
      localStorage.getItem(ROWS_PER_PAGE) ||
      50),
    filter: params.get(FILTER) || '',
    date: new Date(
      params.has(DATE) ? params.get(DATE) + 'T00:00:00' : (Date.now()),
    ),
    period: params.has(PERIOD) ? Number(params.get(PERIOD)) as Period : null || Period.WEEK,
    sort: params.has(SORT_BY) ? Number(params.get(SORT_BY)) as SortType : null || SortType.SORT_AVG_CORES,
    ascending: params.get(ASCENDING) === 'true',
    sortIndex: Number(params.get(SORT_INDEX) || 0),
    timelineMetric: params.has(TIMELINE_VIEW_METRIC) ?
      String(params.get(TIMELINE_VIEW_METRIC)) as MetricType : null || MetricType.AVG_CORES,
    timelineView: params.get(TIMELINE_VIEW) === 'true',
    directoryView: params.get(DIRECTORY_VIEW) === 'true',
  };
  return (
    <MetricsContextProvider {...props}>
      <ResourcesToolbar/>
      <Box sx={{ margin: '10px 20px' }}>
        <ResourcesTable/>
      </Box>
      <ResourcesSearchParams/>
      <ComponentParams/>
    </MetricsContextProvider>
  );
}

export default ResourcesPage;
