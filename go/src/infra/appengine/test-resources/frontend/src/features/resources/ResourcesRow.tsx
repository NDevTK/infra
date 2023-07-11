// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
import { useState } from 'react';
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';
import { IconButton, TableCell, TableRow } from '@mui/material';
import { formatNumber, formatTime } from '../../utils/formatUtils';
import { MetricType } from '../../api/resources';
import { Node } from '../context/MetricsContext';

export interface ResourcesRowProps {
  data: Node,
  depth: number,
}

// Display the metrics in TableCell Format
export function displayMetrics(metrics: Map<MetricType, number>) {
  return (
    <>
      <TableCell data-testid="tableCell" align="right">{formatNumber(metrics.get(MetricType.NUM_RUNS) || 0)}</TableCell>
      <TableCell data-testid="tableCell" align="right">{formatNumber(metrics.get(MetricType.NUM_FAILURES) || 0)}</TableCell>
      <TableCell data-testid="tableCell" align="right">{formatTime(metrics.get(MetricType.AVG_RUNTIME) || 0)}</TableCell>
      <TableCell data-testid="tableCell" align="right">{formatTime(metrics.get(MetricType.TOTAL_RUNTIME) || 0)}</TableCell>
      <TableCell data-testid="tableCell" align="right">{formatNumber(metrics.get(MetricType.AVG_CORES) || 0)}</TableCell>
    </>
  );
}

function ResourcesRow(props: ResourcesRowProps) {
  const [isOpen, setIsOpen] = useState(false);
  const rotate = isOpen ? 'rotate(0deg)' : 'rotate(270deg)';
  return (
    <>
      <TableRow
        data-testid={'tablerow-' + props.data.id}
        data-depth={props.depth}
        key={props.data.id}
      >
        <TableCell
          colSpan={props.data.subname === undefined ? 2 : 1}
          sx={{ paddingLeft: props.depth * 2 + 2, whiteSpace: 'nowrap' }}>
          {
            props.data.isLeaf ? null : (
              <IconButton
                data-testid={'clickButton-' + props.data.id}
                color="primary"
                size="small"
                onClick={() => setIsOpen(!isOpen)}
                style={{ transform: rotate }}
                sx={{ margin: 0, padding: 0, ml: -2 }}
              >
                <ArrowDropDownIcon/>
              </IconButton>
            )
          }
          {props.data.name}
        </TableCell>
        {props.data.subname === undefined ? null : (
          <TableCell sx={{ whiteSpace: 'nowrap' }}>{props.data.subname}</TableCell>
        )}
        {displayMetrics(props.data.metrics)}
      </TableRow>
      {
        isOpen && props.data.nodes.length > 0 ? (
          props.data.nodes.map(
              (row) => <ResourcesRow key={row.id} data={row} depth={props.depth + 1} />,
          )
        ) : null
      }
    </>
  );
}

export default ResourcesRow;
