// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
import { useContext, useState } from 'react';
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';
import { IconButton, TableCell, TableRow } from '@mui/material';
import { formatNumber, formatTime } from '../../utils/formatUtils';
import { MetricType } from '../../api/resources';
import { Node, MetricsContext } from '../context/MetricsContext';

export interface ResourcesRowProps {
  data: Node,
  depth: number,
}

function ResourcesRow(props: ResourcesRowProps) {
  const { params, datesToShow } = useContext(MetricsContext);
  const [isOpen, setIsOpen] = useState(false);
  const rotate = isOpen ? 'rotate(0deg)' : 'rotate(270deg)';

  function handleOpenToggle() {
    if (!isOpen && props.data.onExpand !== undefined) {
      props.data.onExpand(props.data);
    }
    setIsOpen(!isOpen);
  }

  function displayMetrics() {
    if (params.timelineView) {
      const bodyArr = [] as JSX.Element[];
      datesToShow.forEach((date) => {
        bodyArr.push(
            <TableCell key={date} component="th" align="right" data-testid="timelineTest">
              {formatNumber(Number(props.data.metrics.get(date)?.get(MetricType.AVG_RUNTIME)))}
            </TableCell>,
        );
      });
      return (
        <>
          { bodyArr }
        </>
      );
    }
    return (
      <>
        <TableCell data-testid="tableCell" align="right">{formatNumber(props.data.metrics.get(datesToShow[0])?.get(MetricType.NUM_RUNS) || 0)}</TableCell>
        <TableCell data-testid="tableCell" align="right">{formatNumber(props.data.metrics.get(datesToShow[0])?.get(MetricType.NUM_FAILURES) || 0)}</TableCell>
        <TableCell data-testid="tableCell" align="right">{formatTime(props.data.metrics.get(datesToShow[0])?.get(MetricType.AVG_RUNTIME) || 0)}</TableCell>
        <TableCell data-testid="tableCell" align="right">{formatTime(props.data.metrics.get(datesToShow[0])?.get(MetricType.TOTAL_RUNTIME) || 0)}</TableCell>
        <TableCell data-testid="tableCell" align="right">{formatNumber(props.data.metrics.get(datesToShow[0])?.get(MetricType.AVG_CORES) || 0)}</TableCell>
      </>
    );
  }
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
                onClick={handleOpenToggle}
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
        {displayMetrics()}
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
