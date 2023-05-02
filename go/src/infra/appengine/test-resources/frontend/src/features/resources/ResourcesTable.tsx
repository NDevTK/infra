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


function createData(
    testName: string,
    testSuite: string,
    numRuns: number,
    numFailures: number,
    avgRuntime: number,
    totalRuntime: number,
    avgCores: number,
) {
  return {
    testName, testSuite, numRuns, numFailures, avgRuntime, totalRuntime, avgCores,
  };
}

const rows = [
  createData('Fake Test Name', 'Fake test suite', 24, 4.0, 2, 2, 2),
];

function ResourcesTable() {
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
        <TableBody>
          {rows.map((row) => (
            <TableRow
              key={row.testName}
            >
              <TableCell component="th" scope="row">
                {row.testName}
              </TableCell>
              <TableCell align="right">{row.testSuite}</TableCell>
              <TableCell align="right">{row.numRuns}</TableCell>
              <TableCell align="right">{row.numFailures}</TableCell>
              <TableCell align="right">{row.avgRuntime}</TableCell>
              <TableCell align="right">{row.numRuns}</TableCell>
              <TableCell align="right">{row.avgCores}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

export default ResourcesTable;
