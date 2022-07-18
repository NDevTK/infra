// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableSortLabel from '@mui/material/TableSortLabel';
import { SortableMetricName } from '../../../services/cluster';

interface Props {
    toggleSort: (metric: SortableMetricName) => void,
    sortMetric: SortableMetricName,
    isAscending: boolean,
}

const ClustersTableHead = ({
  toggleSort,
  sortMetric,
  isAscending,
}: Props) => {
  return (
    <TableHead data-testid="clusters_table_head">
      <TableRow>
        <TableCell>Cluster</TableCell>
        <TableCell sx={{ width: '130px' }}>Bug</TableCell>
        <TableCell
          sortDirection={sortMetric === 'presubmit_rejects' ? (isAscending ? 'asc' : 'desc') : false}
          sx={{ cursor: 'pointer', width: '100px' }}>
          <TableSortLabel
            aria-label="Sort by User CLs failed Presubmit"
            active={sortMetric === 'presubmit_rejects'}
            direction={isAscending ? 'asc' : 'desc'}
            onClick={() => toggleSort('presubmit_rejects')}>
              User Cls Failed Presubmit
          </TableSortLabel>
        </TableCell>
        <TableCell
          sortDirection={sortMetric === 'critical_failures_exonerated' ? (isAscending ? 'asc' : 'desc') : false}
          sx={{ cursor: 'pointer', width: '100px' }}>
          <TableSortLabel
            active={sortMetric === 'critical_failures_exonerated'}
            direction={isAscending ? 'asc' : 'desc'}
            onClick={() => toggleSort('critical_failures_exonerated')}>
              Presubmit-Blocking Failures Exonerated
          </TableSortLabel>
        </TableCell>
        <TableCell
          sortDirection={sortMetric === 'failures' ? (isAscending ? 'asc' : 'desc') : false}
          sx={{ cursor: 'pointer', width: '100px' }}>
          <TableSortLabel
            active={sortMetric === 'failures'}
            direction={isAscending ? 'asc' : 'desc'}
            onClick={() => toggleSort('failures')}>
              Total Failures
          </TableSortLabel>
        </TableCell>
      </TableRow>
    </TableHead>
  );
};

export default ClustersTableHead;
