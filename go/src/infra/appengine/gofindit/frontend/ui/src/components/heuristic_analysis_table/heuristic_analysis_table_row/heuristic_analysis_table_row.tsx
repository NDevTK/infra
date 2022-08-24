// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import './heuristic_analysis_table_row.css';

import { nanoid } from '@reduxjs/toolkit';

import TableCell from '@mui/material/TableCell';
import TableRow from '@mui/material/TableRow';

import { getCommitShortHash } from '../../../tools/link_constructors';
import { HeuristicSuspect } from '../../../services/gofindit';

interface Props {
  suspect: HeuristicSuspect;
}

export const HeuristicAnalysisTableRow = ({ suspect }: Props) => {
  const { gitilesCommit, reviewUrl, justification, score, confidenceLevel } =
    suspect;

  const reasons = justification.split('\n');
  const reasonCount = reasons.length;

  // TODO: use the title of the suspect commit for link its code review
  const commitTitle = '';

  return (
    <>
      <TableRow data-testid='heuristic_analysis_table_row'>
        <TableCell rowSpan={reasonCount} className='overviewCell'>
          <a href={reviewUrl}>
            {getCommitShortHash(gitilesCommit.id)}
            {commitTitle}
          </a>
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
