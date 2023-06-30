// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
import { useState } from 'react';
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';
import { Button, TableCell, TableRow } from '@mui/material';
import { formatNumber, formatTime } from '../../utils/formatUtils';
import { MetricType } from '../../api/resources';
import { Test } from '../context/MetricsContext';
import styles from './ResourcesRow.module.css';
import VariantRow from './VariantRow';

export interface ResourcesRowProps {
  test: Test,
  lastPage: boolean,
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

function ResourcesRow(resourcesRowParams: ResourcesRowProps) {
  const [isOpen, setIsOpen] = useState(false);
  const rotate = isOpen ? 'rotate(0deg)' : 'rotate(270deg)';
  return (
    <>
      <TableRow
        data-testid="tableRowTest"
        key={resourcesRowParams.test.testId}
        className={styles.tableRow}
      >
        <TableCell scope="row" className={styles.titleCell}>
          {
          resourcesRowParams.test.variants.length != 0 ? (
            <Button
              data-testid="clickButton"
              onClick={() => setIsOpen(!isOpen)}
              style={{ transform: rotate }}
              className={styles.btn}
            >
              <ArrowDropDownIcon/>
            </Button>
          ) : null
          }
          {resourcesRowParams.test.testName}
        </TableCell>
        <TableCell align="right"></TableCell>
        {displayMetrics(resourcesRowParams.test.metrics)}
      </TableRow>
      {
        isOpen ? (
        resourcesRowParams.test.variants.map((variant, idx) =>
          <VariantRow {...{ variant: variant, tableKey: idx }} key={idx}/>,
        ) ) : null
      }
    </>
  );
}

export default ResourcesRow;
