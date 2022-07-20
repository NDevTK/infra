// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import TableCell from '@mui/material/TableCell';

import { SuspectJustificationItem } from './../../../../services/analysis_details';

interface Props {
  justification: SuspectJustificationItem;
}

const HeuristicJustificationTableCells = ({ justification }: Props) => {
  const { Score, FilePath, Reason} = justification;
  return (
    <>
      <TableCell align='right' >
        {Score}
      </TableCell>
      <TableCell>
        {FilePath}
      </TableCell>
      <TableCell>
        {Reason}
      </TableCell>
    </>
  );
}

export default HeuristicJustificationTableCells;
