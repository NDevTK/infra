// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { TableCell, TableRow } from '@mui/material';
import { TestDateMetricData, MetricType } from '../../api/resources';

function ResourcesRow(testDateMetricData : TestDateMetricData) {
  let numRuns = 0;
  let numFailures = 0;
  let avgRuntime = 0;
  let totalRuntime = 0;
  let avgCores = 0;
  let runTimeCount = 0;
  const metrics = testDateMetricData.metrics;
  metrics.forEach((data, _) => {
    data.data.forEach((metric) => {
      switch (metric.metric_type) {
        case MetricType.NUM_RUNS:
          numRuns += metric.metric_value;
          break;
        case MetricType.NUM_FAILURES:
          numFailures += metric.metric_value;
          break;
        case MetricType.AVG_RUNTIME:
          avgRuntime += metric.metric_value;
          runTimeCount ++;
          break;
        case MetricType.TOTAL_RUNTIME:
          totalRuntime += metric.metric_value;
          break;
        case MetricType.AVG_CORES:
          avgCores += metric.metric_value;
          break;
      }
    });
  });
  avgRuntime = avgRuntime / runTimeCount;
  return (
    <TableRow
      data-testid="tableRowTest"
      key={testDateMetricData.test_id}
    >
      <TableCell component="th" scope="row">
        {testDateMetricData.test_name}
      </TableCell>
      <TableCell align="right"></TableCell>
      <TableCell align="right">{numRuns}</TableCell>
      <TableCell align="right">{numFailures}</TableCell>
      <TableCell align="right">{avgRuntime}s</TableCell>
      <TableCell align="right">{totalRuntime}s</TableCell>
      <TableCell align="right">{avgCores}</TableCell>
    </TableRow>
  );
}

export default ResourcesRow;
