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

import { DetailedSuspect } from './../../services/analysis_details';
import { HeuristicAnalysisTableRow } from './heuristic_analysis_table_row/heuristic_analysis_table_row';

interface Props {
  suspects: DetailedSuspect[];
}

export const HeuristicAnalysisTable = ({ suspects }: Props) => {
  return (
    <TableContainer component={Paper} className='heuristicTableContainer'>
      <Table className='heuristicTable' size='small'>
        <TableHead>
          <TableRow>
            <TableCell>Suspect CL</TableCell>
            <TableCell>Confidence</TableCell>
            <TableCell>Score</TableCell>
            <TableCell>Justification</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {suspects.map((suspect) => (
            <HeuristicAnalysisTableRow
              key={suspect.commitID}
              suspect={suspect}
            />
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};
