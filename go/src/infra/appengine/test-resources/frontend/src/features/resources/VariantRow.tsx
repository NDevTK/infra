// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { TableCell } from '@mui/material';
import { TestVariantData } from '../../api/resources';
import { AggregatedMetrics, aggregateMetrics } from './ResourcesRow';

function VariantRow(variant: TestVariantData) {
  const aggregatedMetrics: AggregatedMetrics = aggregateMetrics(
      new Map(Object.entries(variant.metrics)),
  );
  return (
    <>
      <TableCell
        component="th"
        scope="row"
        data-testid="variantRowCellTest"
      >
        {variant.builder}
      </TableCell>
      <TableCell align="right">{variant.suite}</TableCell>
      <TableCell align="right">{aggregatedMetrics.numRuns}</TableCell>
      <TableCell align="right">{aggregatedMetrics.numFailures}</TableCell>
      <TableCell align="right">{aggregatedMetrics.avgRuntime}s</TableCell>
      <TableCell align="right">{aggregatedMetrics.totalRuntime}s</TableCell>
      <TableCell align="right">{aggregatedMetrics.avgCores}</TableCell>
    </>
  );
}

export default VariantRow;
