// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableRow from '@mui/material/TableRow';

import { RevertCL } from '../../services/analysis_details';
import { PlainTable } from '../plain_table/plain_table';

interface Props {
  revertCL: RevertCL;
}

export const RevertCLOverview = ({ revertCL }: Props) => {
  return (
    <TableContainer>
      <PlainTable>
        <colgroup>
          <col style={{ width: '15%' }} />
          <col style={{ width: '85%' }} />
        </colgroup>
        <TableBody data-testid='change_list_overview_table_body'>
          <TableRow>
            <TableCell variant='head' colSpan={2}>
              <a href={revertCL.cl.reviewURL}>{revertCL.cl.title}</a>
            </TableCell>
          </TableRow>
          <TableRow>
            <TableCell variant='head'>Status</TableCell>
            <TableCell>{revertCL.status}</TableCell>
          </TableRow>
          <TableRow>
            <TableCell variant='head'>Submitted time</TableCell>
            <TableCell>{revertCL.submitTime}</TableCell>
          </TableRow>
          <TableRow>
            <TableCell variant='head'>Commit position</TableCell>
            <TableCell>{revertCL.commitPosition}</TableCell>
          </TableRow>
        </TableBody>
      </PlainTable>
    </TableContainer>
  );
};
