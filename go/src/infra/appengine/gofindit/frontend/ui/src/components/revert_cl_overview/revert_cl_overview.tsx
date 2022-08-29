// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import Link from '@mui/material/Link';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableRow from '@mui/material/TableRow';

import { RevertCL } from '../../services/luci_bisection';
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
              <Link
                href={revertCL.cl.reviewURL}
                target='_blank'
                rel='noreferrer'
                underline='always'
              >
                {revertCL.cl.title}
              </Link>
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
