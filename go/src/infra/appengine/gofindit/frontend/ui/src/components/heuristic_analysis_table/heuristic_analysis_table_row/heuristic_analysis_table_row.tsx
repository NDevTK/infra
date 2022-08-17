// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import './heuristic_analysis_table_row.css';

import { nanoid } from '@reduxjs/toolkit';

import TableCell from '@mui/material/TableCell';
import TableRow from '@mui/material/TableRow';

import { HeuristicSuspect } from '../../../services/analysis_details';

interface Props {
  suspect: HeuristicSuspect;
}

const getCommitShortHash = (commitID: string) => {
  return commitID.substring(0, 7);
};

export const HeuristicAnalysisTableRow = ({ suspect }: Props) => {
  const { cl, score, confidence, justification } = suspect;
  const reasonCount = justification.length;

  let clDescription = '';
  if (cl.title) {
    clDescription = `: ${cl.title}`;
  }

  return (
    <>
      <TableRow data-testid='heuristic_analysis_table_row'>
        <TableCell rowSpan={reasonCount} className='overviewCell'>
          <a href={cl.reviewURL}>
            {getCommitShortHash(cl.commitID)}
            {clDescription}
          </a>
        </TableCell>
        <TableCell rowSpan={reasonCount} className='overviewCell'>
          {confidence}
        </TableCell>
        <TableCell rowSpan={reasonCount} className='overviewCell' align='right'>
          {score}
        </TableCell>
        {reasonCount > 0 && <TableCell>{justification[0]}</TableCell>}
      </TableRow>
      {justification.slice(1).map((reason) => (
        <TableRow key={nanoid()}>
          <TableCell>{reason}</TableCell>
        </TableRow>
      ))}
    </>
  );
};
