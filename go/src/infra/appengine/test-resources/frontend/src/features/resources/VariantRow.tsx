// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { TableCell, TableRow } from '@mui/material';
import { TestVariant } from '../context/MetricsContext';
import { displayMetrics } from './ResourcesRow';
import styles from './VariantRow.module.css';

export interface VariantProps {
  variant: TestVariant,
  tableKey: number,
}

function VariantRow(variantProps: VariantProps) {
  return (
    <TableRow
      key={variantProps.tableKey}
      data-testid="variantRowTest"
      className={styles.tableRow}
    >
      <TableCell
        component="td"
        scope="row"
        data-testid="variantRowCellTest"
        className={styles.variantRow}
      >
        {variantProps.variant.builder}
      </TableCell>
      <TableCell component="td" align="right">{variantProps.variant.suite}</TableCell>
      {displayMetrics(variantProps.variant.metrics)}
    </TableRow>
  );
}

export default VariantRow;
