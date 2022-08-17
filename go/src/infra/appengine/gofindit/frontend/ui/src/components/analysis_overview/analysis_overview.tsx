// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import './analysis_overview.css';

import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableRow from '@mui/material/TableRow';

import { AssociatedBug, SuspectRange } from '../../services/analysis_details';
import { PlainTable } from '../plain_table/plain_table';

import { linkToBuild } from '../../tools/link_constructors';

export interface AnalysisSummary {
  analysisID: string;
  status: string;
  failureType: string;
  buildID: string;
  builder: string;
  suspectRange: SuspectRange;
  bugs: AssociatedBug[];
}

interface Props {
  analysis: AnalysisSummary;
}

export const AnalysisOverview = ({ analysis }: Props) => {
  return (
    <TableContainer>
      <PlainTable>
        <colgroup>
          <col style={{ width: '15%' }} />
          <col style={{ width: '35%' }} />
          <col style={{ width: '15%' }} />
          <col style={{ width: '35%' }} />
        </colgroup>
        <TableBody data-testid='analysis_overview_table_body'>
          <TableRow>
            <TableCell variant='head'>Analysis ID</TableCell>
            <TableCell>{analysis.analysisID}</TableCell>
            <TableCell variant='head'>Buildbucket ID</TableCell>
            <TableCell>
              <a href={linkToBuild(analysis.buildID)}>{analysis.buildID}</a>
            </TableCell>
          </TableRow>
          <TableRow>
            <TableCell variant='head'>Status</TableCell>
            <TableCell>{analysis.status}</TableCell>
            <TableCell variant='head'>Builder</TableCell>
            <TableCell>{analysis.builder}</TableCell>
          </TableRow>
          <TableRow>
            <TableCell variant='head'>Suspect range</TableCell>
            <TableCell>
              <a href={analysis.suspectRange.url}>
                {analysis.suspectRange.linkText}
              </a>
            </TableCell>
            <TableCell variant='head'>Failure type</TableCell>
            <TableCell>{analysis.failureType}</TableCell>
          </TableRow>
          {analysis.bugs.length > 0 && (
            <>
              <TableRow>
                <TableCell>
                  <br />
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell variant='head'>Related bugs</TableCell>
                <TableCell colSpan={3}>
                  {analysis.bugs.map((bug) => (
                    <span className='bugLink' key={bug.url}>
                      <a href={bug.url}>{bug.linkText}</a>
                    </span>
                  ))}
                </TableCell>
              </TableRow>
            </>
          )}
        </TableBody>
      </PlainTable>
    </TableContainer>
  );
};
