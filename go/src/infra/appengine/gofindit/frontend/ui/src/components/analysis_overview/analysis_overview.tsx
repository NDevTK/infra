// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import './analysis_overview.css';

import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableRow from '@mui/material/TableRow';

import { Analysis } from '../../services/gofindit';
import { PlainTable } from '../plain_table/plain_table';

import {
  ExternalLink,
  linkToBuild,
  linkToCommit,
  linkToCommitRange,
} from '../../tools/link_constructors';

interface Props {
  analysis: Analysis;
}

function getSuspectRange(analysis: Analysis): ExternalLink {
  if (analysis.culprit) {
    return linkToCommit(analysis.culprit);
  }

  if (analysis.nthSectionResult) {
    const result = analysis.nthSectionResult;

    if (result.culprit) {
      return linkToCommit(result.culprit);
    }

    if (result.remainingNthSectionRange) {
      return linkToCommitRange(
        result.remainingNthSectionRange.lastPassed,
        result.remainingNthSectionRange.firstFailed
      );
    }
  }

  return {
    linkText: '',
    url: '',
  };
}

function getBugLinks(analysis: Analysis): ExternalLink[] {
  let bugLinks: ExternalLink[] = [];

  if (analysis.culpritAction) {
    analysis.culpritAction.forEach((action) => {
      if (action.actionType === 'BUG_COMMENTED' && action.bugUrl) {
        // TODO: construct short link text for bug
        bugLinks.push({
          linkText: action.bugUrl,
          url: action.bugUrl,
        });
      }
    });
  }

  return bugLinks;
}

export const AnalysisOverview = ({ analysis }: Props) => {
  const buildLink = linkToBuild(analysis.firstFailedBbid);
  const suspectRange = getSuspectRange(analysis);
  const bugLinks = getBugLinks(analysis);
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
            <TableCell>{analysis.analysisId}</TableCell>
            <TableCell variant='head'>Buildbucket ID</TableCell>
            <TableCell>
              <a href={buildLink.url}>{buildLink.linkText}</a>
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
              <a
                data-testid='analysis_overview_suspect_range'
                href={suspectRange.url}
              >
                {suspectRange.linkText}
              </a>
            </TableCell>
            <TableCell variant='head'>Failure type</TableCell>
            <TableCell>{analysis.failureType}</TableCell>
          </TableRow>
          {bugLinks.length > 0 && (
            <>
              <TableRow>
                <TableCell>
                  <br />
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell variant='head'>Related bugs</TableCell>
                <TableCell colSpan={3}>
                  {bugLinks.map((bugLink) => (
                    <span className='bugLink' key={bugLink.url}>
                      <a href={bugLink.url}>{bugLink.linkText}</a>
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
