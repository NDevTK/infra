// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import { useContext, useEffect } from 'react';
import MetricContext from '../../context/MetricContext';
import { fetchTestDateMetricData } from '../../api/resources';
import ComponentRow from './ResourcesRow';

function ResourcesTable() {
  const metricsCtx = useContext(MetricContext);
  useEffect(() => {
    fetchTestDateMetricData().then((fetchedMetric) => {
      metricsCtx?.setMetrics(fetchedMetric);
    });
  }, [JSON.stringify(metricsCtx)]);
  return (
    <TableContainer component={Paper}>
      <Table sx={{ minWidth: 650 }} aria-label="simple table">
        <TableHead>
          <TableRow>
            <TableCell>Test</TableCell>
            <TableCell align="right">Test Suite</TableCell>
            <TableCell align="right"># Runs</TableCell>
            <TableCell align="right"># Failures</TableCell>
            <TableCell align="right">Avg Runtime</TableCell>
            <TableCell align="right">Total Runtime</TableCell>
            <TableCell align="right">Avg Cores</TableCell>
          </TableRow>
        </TableHead>
        {metricsCtx?.testDateMetricData ?
        <TableBody>
          {metricsCtx.testDateMetricData.map((row) => (
            <ComponentRow key={row.test_id} {...row}/>
          ))}
        </TableBody> : null
        }
      </Table>
    </TableContainer>
  );
}

export default ResourcesTable;
