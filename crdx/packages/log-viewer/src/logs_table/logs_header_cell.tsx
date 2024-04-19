// Copyright 2024 The Chromium Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {
  Box,
  Grid,
  SxProps,
  TableCell,
  TableSortLabel,
  Theme,
} from '@mui/material';
import { visuallyHidden } from '@mui/utils';
import { ReactNode } from 'react';

import { SortOrder } from '@/constants/table_constants';
import { LogsTableEntry } from '@/types/table';

interface Props {
  title?: string;
  label: string;
  sortable?: boolean;
  sortOrder?: SortOrder;
  sortId?: keyof LogsTableEntry;
  onHeaderSort?: (
    event: React.MouseEvent<unknown>,
    property: keyof LogsTableEntry,
  ) => void;
  width?: string;
  sx?: SxProps<Theme>;
  children?: ReactNode;
}

export function LogsHeaderCell({
  title,
  label,
  sortable,
  sortOrder,
  sortId,
  onHeaderSort,
  width,
  sx,
  children,
}: Props) {
  function createSortHandler(property: keyof LogsTableEntry | undefined) {
    if (!property) {
      return () => {};
    }
    return (event: React.MouseEvent<unknown>) => {
      onHeaderSort?.(event, property);
    };
  }

  return (
    <TableCell
      variant="head"
      title={title}
      align="left"
      size="small"
      sortDirection={sortable ? sortOrder : false}
      sx={{
        fontSize: '11px',
        height: '1rem',
        textAlign: 'left',
        pl: 0,
        pb: 0,
        width,
        ...sx,
      }}
    >
      <Grid container item direction="row" rowSpacing={2}>
        <TableSortLabel
          data-testid={`header-${sortId}`}
          active={sortable}
          direction={sortable ? sortOrder : 'asc'}
          onClick={createSortHandler(sortId)}
          disabled={!sortable}
          sx={{
            width,
          }}
        >
          {label}
          {sortable && (
            <Box component="span" sx={visuallyHidden}>
              {sortOrder === 'desc' ? 'sorted descending' : 'sorted ascending'}
            </Box>
          )}
        </TableSortLabel>
        {children}
      </Grid>
    </TableCell>
  );
}
