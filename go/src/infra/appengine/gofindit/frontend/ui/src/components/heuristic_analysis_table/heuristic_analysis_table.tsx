// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import './heuristic_analysis_table.css';

import Paper from '@mui/material/Paper';

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';

import HeuristicAnalysisTableRow from './heuristic_analysis_table_row/heuristic_analysis_table_row';
import { HeuristicAnalysisResult } from './../../services/analysis_details';

interface Props {
  results: HeuristicAnalysisResult;
}

const HeuristicAnalysisTable = ({results}: Props) => {
  return (
    <TableContainer
      component={Paper}
      className='heuristicTableContainer'
    >
      <Table
        className='heuristicTable'
        size='small'
      >
        <TableHead>
          <TableRow>
            <TableCell rowSpan={2} >
              Suspect CL
            </TableCell>
            <TableCell rowSpan={2} >
              Culprit Status
            </TableCell>
            <TableCell rowSpan={2} >
              Suspect Total Score
            </TableCell>
            <TableCell colSpan={3} >
              Justification
            </TableCell>
          </TableRow>
          <TableRow>
            <TableCell>
              Score
            </TableCell>
            <TableCell>
              File Path
            </TableCell>
            <TableCell>
              Reason
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {
            results.Items.map((result) => (
              <HeuristicAnalysisTableRow
                key={result.Commit}
                resultData={result}
              />
            ))
          }
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default HeuristicAnalysisTable;
