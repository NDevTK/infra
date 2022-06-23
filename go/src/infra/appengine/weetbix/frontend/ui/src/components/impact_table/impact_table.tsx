// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import Box from '@mui/material/Box';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';

import { Cluster, Counts } from '../../services/cluster';
import HelpTooltip from '../help_tooltip/help_tooltip';


const userClsFailedPresubmitTooltipText = 'The number of distinct developer changelists that failed at least one presubmit (CQ) run because of failure(s) in this cluster.';
const criticalFailuresExoneratedTooltipText = 'The number of failures on test variants which were configured to be presubmit-blocking, which were exonerated (i.e. did not actually block presubmit) because infrastructure determined the test variant to be failing or too flaky at tip-of-tree. If this number is non-zero, it means a test variant which was configured to be presubmit-blocking is not stable enough to do so, and should be fixed or made non-blocking.';
const totalFailuresTooltipText = 'The total number of test results in this cluster. Weetbix only clusters test results which are unexpected and have a status of crash, abort or fail.';

interface Props {
    cluster: Cluster;
}

const ImpactTable = ({ cluster }: Props) => {
  const metric = (counts: Counts): number => {
    return counts.nominal;
  };

  return (
    <TableContainer component={Box}>
      <Table data-testid="impact-table" size="small" sx={{ maxWidth: 500 }}>
        <TableHead>
          <TableRow>
            <TableCell></TableCell>
            <TableCell align="right">1 day</TableCell>
            <TableCell align="right">3 days</TableCell>
            <TableCell align="right">7 days</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          <TableRow>
            <TableCell>User Cls Failed Presubmit <HelpTooltip text={userClsFailedPresubmitTooltipText} /></TableCell>
            <TableCell align="right">{metric(cluster.presubmitRejects1d)}</TableCell>
            <TableCell align="right">{metric(cluster.presubmitRejects3d)}</TableCell>
            <TableCell align="right">{metric(cluster.presubmitRejects7d)}</TableCell>
          </TableRow>
          <TableRow>
            <TableCell>Presubmit-Blocking Failures Exonerated <HelpTooltip text={criticalFailuresExoneratedTooltipText} /></TableCell>
            <TableCell align="right">{metric(cluster.criticalFailuresExonerated1d)}</TableCell>
            <TableCell align="right">{metric(cluster.criticalFailuresExonerated3d)}</TableCell>
            <TableCell align="right">{metric(cluster.criticalFailuresExonerated7d)}</TableCell>
          </TableRow>
          <TableRow>
            <TableCell>Total Failures <HelpTooltip text={totalFailuresTooltipText} /></TableCell>
            <TableCell align="right">{metric(cluster.failures1d)}</TableCell>
            <TableCell align="right">{metric(cluster.failures3d)}</TableCell>
            <TableCell align="right">{metric(cluster.failures7d)}</TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default ImpactTable;
