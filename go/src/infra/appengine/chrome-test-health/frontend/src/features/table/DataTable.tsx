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
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';
import { Button, LinearProgress, SxProps, TablePagination, Theme } from '@mui/material';
import styles from './DataTable.module.css';
import DataTableRow from './DataTableRow';

export interface DataTableProps {
  rows: Row[],
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

export interface Row {
  id: string,
  isExpandable?: boolean,
  onExpand?: (row: Row) => void,
  rows?: Row[],
}

export interface Column {
  name: string,
  // Can return undefined to render no cell, or an optional colSpan
  renderer: (column: Column, row: Row) => string | [string, number] | undefined,
  align: any,
  isSortedBy?: boolean,
  isSortAscending?: boolean,
  onClick?: () => void,
  sx?: SxProps<Theme>,
}

function showSortArrow(col: Column) {
  return (
    col.isSortAscending ?
          <ArrowUpwardIcon sx={{ visibility: col.isSortedBy ? 'visible' : 'hidden', height: '20px', width: '20px' }}/> :
          <ArrowDownwardIcon sx={{ visibility: col.isSortedBy ? 'visible' : 'hidden', height: '20px', width: '20px' }}/>
  );
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
      {
        column.onClick ?
          <Button className={styles.sortButtonText} onClick={column.onClick}>
            {column.name}
            {showSortArrow(column)}
          </Button> :
          <p>{column.name}</p>
      }
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
      <TableContainer sx={{
        maxHeight: 'calc(100vh - ' + (props.showPaginator ? '164' : '214') + 'px)',
      }}>
        <LinearProgress sx={{ visibility: props.isLoading ? 'visible' : 'hidden' }} data-testid='loading-bar'/>
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
