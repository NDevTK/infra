// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Box, Paper, SxProps, Theme, Typography } from '@mui/material';
import { useContext } from 'react';
import DataTable, { Column, Row } from '../../../components/table/DataTable';
import { SummaryContext } from './SummaryContext';
import { MetricType, Node } from './LoadSummary';

const colorPallete = (percentage: number): string => {
  if (percentage < 50) {
    return '#FC8D7E'; // Red;
  } else if (percentage < 70) {
    return '#FAC687'; // Orange;
  } else if (percentage < 90) {
    return '#F5F57D'; // Yellow;
  } else {
    return '#A6F5A6'; // Green;
  }
};

function SummaryTable() {
  const { data, isLoading } = useContext(SummaryContext);

  function constructColumns() {
    const cols: Column[] = [{
      name: 'Directories/Files',
      renderer: (_: Column, row: Row<Node>) => {
        const node = row as Node;
        return node.name;
      },
      align: 'left',
      sx: { width: '30%' },
    },
    ];

    const columns: [MetricType, string][] = [
      [MetricType.LINE, 'Line Coverage'],
    ];
    columns.map(([metricType, name]) => {
      cols.push({
        name: name,
        renderer: (_: Column, row: Row<Node>) => {
          const node = row as Node;
          const metricData = node.metrics.get(metricType);

          if (metricData === undefined) return '--';

          const perc = metricData.percentageCovered.toFixed(2);
          const covered = metricData.covered;
          const total = metricData.total;

          const value = `${perc}% (${covered}/${total})`;
          const sxProps: SxProps<Theme> = { backgroundColor: colorPallete(metricData.percentageCovered) } as SxProps<Theme>;
          return [value, undefined, sxProps];
        },
        align: 'left',
        sx: { whiteSpace: 'nowrap', width: '20%', minWidth: '100px', maxWidth: '140px' },
      });
    });
    return cols;
  }

  return (
    <Box>
      <Box data-testid="legend" style={{ display: 'flex', margin: '20px 0' }}>
        <Typography sx={{ marginRight: '10px' }}>Legend:</Typography>
        <Box sx={{ display: 'flex', width: '600px', justifyContent: 'space-evenly' }}>
          <Box sx={{ backgroundColor: colorPallete(100), flexGrow: 1, textAlign: 'center' }}>&gt; 90%</Box>
          <Box sx={{ backgroundColor: colorPallete(80), flexGrow: 1, textAlign: 'center' }}>70% - 90%</Box>
          <Box sx={{ backgroundColor: colorPallete(60), flexGrow: 1, textAlign: 'center' }}>50%-70%</Box>
          <Box sx={{ backgroundColor: colorPallete(40), flexGrow: 1, textAlign: 'center' }}> &lt; 50% </Box>
        </Box>
      </Box>
      <Paper>
        <DataTable isLoading={isLoading} rows={data} columns={constructColumns()} showPaginator={false} />
      </Paper>
    </Box>
  );
}

export default SummaryTable;
