// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import * as React from 'react';
import TableCell from '@material-ui/core/TableCell';

import { format } from '../../utils/formatUtils';
import styles from './MetricsTableDataCell.module.css';
import { DataPoint } from './metricsSlice';
import {
  MetricOption,
  MetricOptionColorType,
} from '../dataSources/dataSourcesSlice';

interface Props {
  data?: DataPoint;
  metric: MetricOption;
}

function calculateColor(
  metric: MetricOption,
  data?: DataPoint
): string | undefined {
  if (metric.color === undefined) {
    return;
  }
  let currValue;
  let prevValue;
  if (data !== undefined) {
    currValue = data.value;
  } else if (metric.color.emptyValue !== undefined) {
    currValue = metric.color.emptyValue;
  }
  if (data?.previous !== undefined) {
    prevValue = data.previous.value;
  } else if (metric.color.emptyValue !== undefined) {
    prevValue = metric.color.emptyValue;
  }
  if (currValue === undefined || prevValue === undefined) {
    return;
  }
  let delta = currValue - prevValue;
  if (metric.color.type === MetricOptionColorType.DeltaPercentage) {
    delta = delta / prevValue;
  }
  let color: string | undefined = undefined;
  metric.color.breakpoints.forEach((breakpoint) => {
    if (breakpoint[0] < 0 && delta <= breakpoint[0]) {
      color = breakpoint[1];
    } else if (breakpoint[0] > 0 && delta >= breakpoint[0]) {
      color = breakpoint[1];
    }
  });
  return color;
}

export const testables = {
  calculateColor: calculateColor,
};

/*
  Renders a data cell
*/
const MetricsTableDataCell: React.FunctionComponent<Props> = (props: Props) => {
  const style: React.CSSProperties = {};
  const color = calculateColor(props.metric, props.data);
  if (color !== undefined) {
    style.color = color;
  }
  return (
    <TableCell
      align="right"
      className={styles.data}
      data-testid="mt-row-data"
      style={style}
    >
      {props.data === undefined
        ? '-'
        : format(props.data.value, props.metric.unit)}
    </TableCell>
  );
};

export default MetricsTableDataCell;
