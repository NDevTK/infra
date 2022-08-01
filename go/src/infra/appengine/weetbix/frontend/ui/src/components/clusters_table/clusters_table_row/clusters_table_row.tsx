// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Link as RouterLink } from 'react-router-dom';
import Link from '@mui/material/Link';
import TableCell from '@mui/material/TableCell';
import TableRow from '@mui/material/TableRow';

import { linkToCluster } from '../../../tools/urlHandling/links';
import { ClusterSummary } from '../../../services/cluster';
import { Bar } from 'react-chartjs-2';
import Tooltip from '@mui/material/Tooltip';

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
      <TableCell data-testid="clusters_table_title">
        <Link component={RouterLink} to={linkToCluster(project, cluster.clusterId)} underline="hover">{cluster.title}</Link>
      </TableCell>
      <TableCell data-testid="clusters_table_bug">
        {
          cluster.bug &&
          <Link href={cluster.bug.url} underline="hover">{cluster.bug.linkText}</Link>
        }
      </TableCell>
      <TableCell className="number">{cluster.presubmitRejects || '0'} <DailyChart data={cluster.presubmitRejectsByDay} /></TableCell>
      <TableCell className="number">{cluster.criticalFailuresExonerated || '0'} <DailyChart data={cluster.criticalFailuresExoneratedByDay} /></TableCell>
      <TableCell className="number">{cluster.failures || '0'} <DailyChart data={cluster.failuresByDay} /></TableCell>
    </TableRow>
  );
};

interface DailyChartProps {
  data: string[] | undefined;
}
const DailyChart = ({ data }: DailyChartProps) => {
  if (!data) return <></>;
  const values = data.map(v => parseInt(v, 10));
  const max = values.reduce((max, value) => value > max ? value : max, 0) || 1;
  return <Tooltip title={<>
    Today: {data[6]}<br />
    Yesterday: {data[5]}<br />
    2 days ago: {data[4]}<br />
    3 days ago: {data[3]}<br />
    4 days ago: {data[2]}<br />
    5 days ago: {data[1]}<br />
    6 days ago: {data[0]}<br />
    </>}>
    <svg viewBox="0 0 130 50" xmlns="http://www.w3.org/2000/svg">
      {values.map((value, i) => {
        let height = value / max * 48;
        return <g key={i}>
          <rect x={i * 20} y={48 - height} width="10" height={height} fill="#0072e5" />
          <rect x={i * 20} y="48" width="10" height="2" fill="#000" />
        </g>;
      })}
    </svg>
  </Tooltip>;
}

export default ClustersTableRow;
