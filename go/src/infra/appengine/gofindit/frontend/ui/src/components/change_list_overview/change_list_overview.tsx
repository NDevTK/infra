// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableRow from '@mui/material/TableRow';

import { ChangeListDetails } from './../../services/analysis_details';
import { PlainTable } from '../plain_table/plain_table';

interface Props {
  changeList: ChangeListDetails;
}

export const ChangeListOverview = ({ changeList }: Props) => {
  return (
    <TableContainer>
      <PlainTable>
        <colgroup>
          <col style={{ width: '15%' }} />
          <col style={{ width: '85%' }} />
        </colgroup>
        <TableBody>
          <TableRow>
            <TableCell variant='head' colSpan={2}>
              <a href={changeList.url}>{changeList.title}</a>
            </TableCell>
          </TableRow>
          <TableRow>
            <TableCell variant='head'>Status</TableCell>
            <TableCell>{changeList.status}</TableCell>
          </TableRow>
          <TableRow>
            <TableCell variant='head'>Submitted time</TableCell>
            <TableCell>{changeList.submitTime}</TableCell>
          </TableRow>
          <TableRow>
            <TableCell variant='head'>Commit position</TableCell>
            <TableCell>{changeList.commitPosition}</TableCell>
          </TableRow>
        </TableBody>
      </PlainTable>
    </TableContainer>
  );
};
