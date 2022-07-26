// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import './analysis_overview.css';

import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableRow from '@mui/material/TableRow';

import { PlainTable } from './../plain_table';

interface AnalysisSummary {
  id: number;
  status: string;
  failureType: string;
  buildID: number;
  builder: string;
  suspectRange: string[];
  bugs: string[];
}

interface Props {
  analysis: AnalysisSummary;
}

function processSuspectRange(suspects: string[]) {
  let suspectRangeText = '';
  let suspectRangeUrl = '';

  var suspectRange = [];

  const suspectCount = suspects.length;
  if (suspectCount > 0) {
    suspectRange.push(suspects[0]);
  }
  if (suspectCount > 1) {
    suspectRange.push(suspects[suspectCount - 1]);
  }

  if (suspectRange.length > 0) {
    suspectRangeText = suspectRange.join(' ... ');
    suspectRangeUrl = `/placeholder/url?earliest=${suspectRange[0]}&latest=${
      suspectRange[suspectRange.length - 1]
    }`;
  }

  return [suspectRangeText, suspectRangeUrl];
}

export const AnalysisOverview = ({ analysis }: Props) => {
  const [suspectRangeText, suspectRangeUrl] = processSuspectRange(
    analysis.suspectRange
  );

  return (
    <TableContainer>
      <PlainTable>
        <colgroup>
          <col style={{ width: '15%' }} />
          <col style={{ width: '35%' }} />
          <col style={{ width: '15%' }} />
          <col style={{ width: '35%' }} />
        </colgroup>
        <TableBody>
          <TableRow>
            <TableCell variant='head'>Analysis ID</TableCell>
            <TableCell>{analysis.id}</TableCell>
            <TableCell variant='head'>Buildbucket ID</TableCell>
            <TableCell>
              <a href={`${analysis.buildID}`}>{analysis.buildID}</a>
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
              <a href={suspectRangeUrl}>{suspectRangeText}</a>
            </TableCell>
            <TableCell variant='head'>Failure Type</TableCell>
            <TableCell>{analysis.failureType}</TableCell>
          </TableRow>
          <TableRow>
            <TableCell>
              <br />
            </TableCell>
          </TableRow>
          <TableRow>
            <TableCell variant='head'>Related bugs</TableCell>
            <TableCell colSpan={3}>
              {analysis.bugs.map((bugUrl) => (
                <span className='bugLink' key={bugUrl}>
                  <a href={bugUrl}>{bugUrl}</a>
                </span>
              ))}
            </TableCell>
          </TableRow>
        </TableBody>
      </PlainTable>
    </TableContainer>
  );
};
