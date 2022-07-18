// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import Link from '@mui/material/Link';
import TableCell from '@mui/material/TableCell';
import TableRow from '@mui/material/TableRow';

import { linkToCluster } from '../../../tools/urlHandling/links';
import { ClusterSummary } from '../../../services/cluster';

interface Props {
  project: string,
  cluster: ClusterSummary,
}

const ClustersTableRow = ({
  project,
  cluster,
}: Props) => {
  return (
    <TableRow>
      <TableCell data-testid="clusters_table_title"><Link href={linkToCluster(project, cluster.clusterId)} underline="hover">{cluster.title}</Link></TableCell>
      <TableCell data-testid="clusters_table_bug">
        {
          cluster.bug &&
            <Link href={cluster.bug.url} underline="hover">{cluster.bug.linkText}</Link>
        }
      </TableCell>
      <TableCell className="number">{cluster.presubmitRejects || '0'}</TableCell>
      <TableCell className="number">{cluster.criticalFailuresExonerated || '0'}</TableCell>
      <TableCell className="number">{cluster.failures || '0'}</TableCell>
    </TableRow>
  );
};

export default ClustersTableRow;
