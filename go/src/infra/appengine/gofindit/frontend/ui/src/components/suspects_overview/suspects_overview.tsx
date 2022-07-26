// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import './suspects_overview.css';

import Paper from '@mui/material/Paper';

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';

import { SuspectSummary } from './../../services/analysis_details';

interface Props {
  suspects: SuspectSummary[];
}

export const SuspectsOverview = ({ suspects }: Props) => {
  return (
    <TableContainer className='suspectsOverview' component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Suspect CL</TableCell>
            <TableCell>Source analysis</TableCell>
            <TableCell>Culprit status</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {suspects.map((suspect) => (
            <TableRow key={suspect.id}>
              <TableCell>
                <a href={suspect.url}>{suspect.title}</a>
              </TableCell>
              <TableCell>{suspect.accuseSource}</TableCell>
              <TableCell>{suspect.culpritStatus}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};
