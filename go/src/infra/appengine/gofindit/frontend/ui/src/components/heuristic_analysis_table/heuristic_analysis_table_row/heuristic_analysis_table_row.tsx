// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import TableCell from '@mui/material/TableCell';
import TableRow from '@mui/material/TableRow';

import { DetailedSuspect } from '../../../services/analysis_details';

interface Props {
  suspect: DetailedSuspect;
}

export const HeuristicAnalysisTableRow = ({ suspect }: Props) => {
  const { commitID, reviewURL, score, confidence, justification } = suspect;
  const reasonCount = justification.length;

  return (
    <>
      <TableRow>
        <TableCell rowSpan={reasonCount}>
          <a href={reviewURL}>{commitID}: [TODO: Get title of commit]</a>
        </TableCell>
        <TableCell rowSpan={reasonCount}>{confidence}</TableCell>
        <TableCell rowSpan={reasonCount} align='right'>
          {score}
        </TableCell>
        {reasonCount > 0 && <TableCell>{justification[0]}</TableCell>}
      </TableRow>
      {justification.slice(1).map((reason) => (
        <TableRow key={reason}>
          <TableCell>{reason}</TableCell>
        </TableRow>
      ))}
    </>
  );
};
