// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import TableCell from '@mui/material/TableCell';
import TableRow from '@mui/material/TableRow';

import HeuristicJustificationTableCells from './heuristic_justification_table_cells/heuristic_justification_table_cells';
import { HeuristicAnalysisResultItem } from './../../../services/analysis_details';

interface Props {
  resultData: HeuristicAnalysisResultItem;
}

const HeuristicAnalysisTableRow = ({resultData}: Props) => {
  const {Commit, ReviewUrl, Justification} = resultData;

  const justificationItems = Justification.Items
  const fileCount = justificationItems.length;

  var totalScore = 0;
  justificationItems.forEach(item => {
    totalScore += item.Score;
  })

  var firstRowCells, justificationRows;

  if (fileCount < 1) {
    firstRowCells = (
      <>
        <TableCell />
        <TableCell />
        <TableCell />
      </>
    );
    justificationRows = (
      <></>
    );
  } else {
    firstRowCells = (
      <HeuristicJustificationTableCells justification={justificationItems[0]} />
    );
    justificationRows = (
      <>
        {
          justificationItems.slice(1).map((justification) => (
            <TableRow key={`${justification.FilePath}|${justification.Reason}`}>
              <HeuristicJustificationTableCells justification={justification} />
            </TableRow>
          ))
        }
      </>
    );
  }

  return (
    <>
      <TableRow>
        <TableCell rowSpan={fileCount} >
          <a href={`${ReviewUrl}`}>
            {Commit}: [TODO: Get title of commit]
          </a>
        </TableCell>
        <TableCell rowSpan={fileCount} >
          [TODO: Culprit status]
        </TableCell>
        <TableCell
          rowSpan={fileCount}
          align='right'
        >
          {totalScore}
        </TableCell>
        {firstRowCells}
      </TableRow>
      {justificationRows}
    </>
  );
}

export default HeuristicAnalysisTableRow;
