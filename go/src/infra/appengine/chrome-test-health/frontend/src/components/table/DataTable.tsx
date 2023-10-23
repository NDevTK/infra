// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import { LinearProgress, SxProps, TablePagination, TableSortLabel, Theme, Tooltip } from '@mui/material';
import styles from './DataTable.module.css';
import DataTableRow from './DataTableRow';

export interface DataTableProps {
  rows: Row<any>[],
  columns: Column[],
  isLoading?: boolean,
  showPaginator?: boolean,
  paginatorProps?: PaginatorProps,
}

export interface PaginatorProps {
  rowsPerPageOptions: number[],
  count: number,
  rowsPerPage: number,
  page: number,
  onPageChange: (_: React.MouseEvent<HTMLButtonElement> | null, newPage: number) => void,
  onChangeRowsPerPage: (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => void,
}

export interface Row<T extends Row<T>> {
  id: string,
  isExpandable?: boolean,
  onExpand?: (row: T) => void,
  rows?: T[],
  footer?: JSX.Element,
}

export interface Column {
  name: string,
  // Can return undefined to render no cell, or an optional colSpan
  renderer: <T extends Row<any>>(column: Column, row: T) => string | [string, number] | [string, number|undefined, SxProps<Theme>] | undefined,
  align: any,
  isSortedBy?: boolean,
  isSortAscending?: boolean,
  onClick?: () => void,
  sx?: SxProps<Theme>,
  description?: string,
}

function columnHeader(column: Column): JSX.Element {
  return (
    <TableCell
      component="th"
      data-testid='columnHeaderTest'
      key={column.name}
      align={column.align}
      sx={column.sx}
    >
      <Tooltip title={column.description}>
        {column.onClick ? (
          <TableSortLabel
            active={column.isSortedBy}
            direction={column.isSortAscending ? 'asc' : 'desc'}
            onClick={column.onClick}
          >{column.name}</TableSortLabel>
        ) : (
          <span>{column.name}</span>
        )}
      </Tooltip>
    </TableCell>
  );
}

function messageRow(colSpan: number, message: string): JSX.Element {
  return (
    <TableRow>
      <TableCell colSpan={colSpan} align="center" className={styles.tableCellNoData}>
        {message}
      </TableCell>
    </TableRow>
  );
}

function DataTable(props: DataTableProps) {
  return (
    <Paper>
      <LinearProgress sx={{ visibility: props.isLoading ? 'visible' : 'hidden' }} data-testid='loading-bar'/>
      <TableContainer sx={{
        maxHeight: 'calc(100vh - ' + (props.showPaginator ? '214' : '164') + 'px)',
      }}>
        <Table stickyHeader size="small" aria-label="simple table">
          <TableHead>
            <TableRow className={styles.headerRow}>
              {props.columns.map((column) => columnHeader(column))}
            </TableRow>
          </TableHead>
          <TableBody data-testid="tableBody">
            {
              props.rows.length > 0 ?
              props.rows.map(
                  (row) => <DataTableRow key={row.id} row={row} depth={0} columns={props.columns}/>,
              ) : messageRow(props.columns.length, props.isLoading ? 'Loading...' : 'No data available' )
            }
          </TableBody>
        </Table>
      </TableContainer>
      {props.showPaginator && props.paginatorProps !== undefined ? (
          <TablePagination
            data-testid="tablePagination"
            component="div"
            rowsPerPageOptions={props.paginatorProps.rowsPerPageOptions || []}
            count={props.paginatorProps.count || 0}
            rowsPerPage={props.paginatorProps.rowsPerPage || 0}
            page={props.paginatorProps.page || 0}
            onPageChange={props.paginatorProps.onPageChange}
            onRowsPerPageChange={props.paginatorProps.onChangeRowsPerPage}
            showFirstButton
            sx={{ borderTop: 1, borderColor: 'grey.300' }}
          />
      ) : null}
    </Paper>
  );
}

export default DataTable;
