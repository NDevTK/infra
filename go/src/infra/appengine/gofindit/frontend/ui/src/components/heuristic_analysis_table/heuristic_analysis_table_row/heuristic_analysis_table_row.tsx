// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import './heuristic_analysis_table_row.css';

import { nanoid } from '@reduxjs/toolkit';

import Link from '@mui/material/Link';
import TableCell from '@mui/material/TableCell';
import TableRow from '@mui/material/TableRow';

import { getCommitShortHash } from '../../../tools/link_constructors';
import { HeuristicSuspect } from '../../../services/luci_bisection';

interface Props {
  suspect: HeuristicSuspect;
}

export const HeuristicAnalysisTableRow = ({ suspect }: Props) => {
  const {
    gitilesCommit,
    reviewUrl,
    reviewTitle,
    justification,
    score,
    confidenceLevel,
  } = suspect;

  const reasons = justification.split('\n');
  const reasonCount = reasons.length;

  let suspectDescription = getCommitShortHash(gitilesCommit.id);
  if (reviewTitle) {
    suspectDescription += `: ${reviewTitle}`;
  }

  return (
    <>
      <TableRow data-testid='heuristic_analysis_table_row'>
        <TableCell rowSpan={reasonCount} className='overviewCell'>
          <Link
            href={reviewUrl}
            target='_blank'
            rel='noreferrer'
            underline='always'
          >
            {suspectDescription}
          </Link>
        </TableCell>
        <TableCell rowSpan={reasonCount} className='overviewCell'>
          {confidenceLevel}
        </TableCell>
        <TableCell rowSpan={reasonCount} className='overviewCell' align='right'>
          {score}
        </TableCell>
        {reasonCount > 0 && <TableCell>{reasons[0]}</TableCell>}
      </TableRow>
      {reasons.slice(1).map((reason) => (
        <TableRow key={nanoid()}>
          <TableCell>{reason}</TableCell>
        </TableRow>
      ))}
    </>
  );
};
