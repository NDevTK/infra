// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
import { useState } from 'react';
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';
import { Button, TableCell, TableRow } from '@mui/material';
import { TestDateMetricData, MetricType, TestMetricsArray } from '../../api/resources';
import styles from './ResourcesRow.module.css';
import VariantRow from './VariantRow';

export type AggregatedMetrics = {
  numRuns: number,
  numFailures: number,
  avgRuntime: number,
  totalRuntime: number,
  avgCores: number,
}

export function aggregateMetrics(metric : Map<string, TestMetricsArray>) : AggregatedMetrics {
  let numRuns = 0;
  let numFailures = 0;
  let avgRuntime = 0;
  let totalRuntime = 0;
  let avgCores = 0;
  let runTimeCount = 0;
  let coresCount = 0;
  metric.forEach((data) => {
    data.data.forEach((metric) => {
      switch (metric.metric_type) {
        case MetricType.NUM_RUNS:
          numRuns += metric.metric_value;
          break;
        case MetricType.NUM_FAILURES:
          numFailures += metric.metric_value;
          break;
        case MetricType.AVG_RUNTIME:
          avgRuntime += metric.metric_value;
          runTimeCount ++;
          break;
        case MetricType.TOTAL_RUNTIME:
          totalRuntime += metric.metric_value;
          break;
        case MetricType.AVG_CORES:
          avgCores += metric.metric_value;
          coresCount ++;
          break;
      }
    });
  });
  avgRuntime = avgRuntime / runTimeCount;
  avgCores = avgCores / coresCount;
  return {
    numRuns: numRuns,
    numFailures: numFailures,
    avgRuntime: avgRuntime,
    totalRuntime: totalRuntime,
    avgCores: avgCores,
  };
}

function ResourcesRow(testDateMetricData : TestDateMetricData) {
  const [isOpen, setIsOpen] = useState(false);
  const aggregatedMetrics: AggregatedMetrics = aggregateMetrics(testDateMetricData.metrics);
  const rotate = isOpen ? 'rotate(0deg)' : 'rotate(270deg)';
  return (
    <>
      <TableRow
        className={styles.slimRows}
        data-testid="tableRowTest"
        key={testDateMetricData.test_id}
      >
        <>
          <TableCell component="th" scope="row" className={styles.noPadding}>
            {
            testDateMetricData.variants.length != 0 ? (
              <Button
                data-testid="clickButton"
                onClick={() => setIsOpen(!isOpen)}
                className={styles.noPadding}
                style={{ transform: rotate }}
              >
                <ArrowDropDownIcon></ArrowDropDownIcon>
              </Button>
            ) : null
            }
            {testDateMetricData.test_name}
          </TableCell>
          <TableCell align="right"></TableCell>
          <TableCell align="right">{aggregatedMetrics.numRuns}</TableCell>
          <TableCell align="right">{aggregatedMetrics.numFailures}</TableCell>
          <TableCell align="right">{aggregatedMetrics.avgRuntime}s</TableCell>
          <TableCell align="right">{aggregatedMetrics.totalRuntime}s</TableCell>
          <TableCell align="right">{aggregatedMetrics.avgCores}</TableCell>
        </>
      </TableRow>
      {
        isOpen ? (
        testDateMetricData.variants.map((variant, idx) =>
          <TableRow
            key={idx}
            data-testid="variantRowTest"
          >
            <VariantRow {...variant}/>
          </TableRow>) ) : null
      }
    </>
  );
}

export default ResourcesRow;
