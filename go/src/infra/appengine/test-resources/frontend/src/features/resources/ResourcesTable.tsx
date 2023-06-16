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
import { useContext } from 'react';
import { MetricsContext } from '../context/MetricsContext';
import ResourceRow from './ResourcesRow';

function ResourcesTable() {
  const { tests, lastPage, api } = useContext(MetricsContext);

  return (
    <TableContainer component={Paper}>
      <Table sx={{ minWidth: 650 }} size="small" aria-label="simple table">
        <TableHead>
          <TableRow>
            <TableCell>Test</TableCell>
            <TableCell component="th" align="right">Test Suite</TableCell>
            <TableCell component="th" align="right"># Runs</TableCell>
            <TableCell component="th" align="right"># Failures</TableCell>
            <TableCell component="th" align="right">Avg Runtime</TableCell>
            <TableCell component="th" align="right">Total Runtime</TableCell>
            <TableCell component="th" align="right">Avg Cores</TableCell>
          </TableRow>
        </TableHead>
        {tests ?
        <TableBody data-testid="tableBody">
          {tests.map((row) => (
            <ResourceRow
              key={row.testId} {
                ...{
                  test: row,
                  lastPage: lastPage,
                  api: api,
                }
              }/>
          ))}
        </TableBody> : null
        }
      </Table>
    </TableContainer>
  );
}

export default ResourcesTable;
