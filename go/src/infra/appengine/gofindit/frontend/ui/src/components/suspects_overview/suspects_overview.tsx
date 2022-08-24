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

import { PrimeSuspect } from '../../services/gofindit';

interface Props {
  suspects: PrimeSuspect[];
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
            <TableRow key={suspect.cl.commitID}>
              <TableCell>
                <a href={suspect.cl.reviewURL}>{suspect.cl.title}</a>
              </TableCell>
              <TableCell>{suspect.accuseSource}</TableCell>
              <TableCell>{suspect.culpritStatus}</TableCell>
            </TableRow>
          ))}
          {suspects.length === 0 && (
            <TableRow>
              <TableCell colSpan={3} className='dataPlaceholder'>
                No suspects to display
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </TableContainer>
  );
};
