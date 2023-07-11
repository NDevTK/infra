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
import { useContext } from 'react';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';
import { Button, LinearProgress, TableFooter, TablePagination } from '@mui/material';
import { SortType } from '../../api/resources';
import { MetricsContext } from '../context/MetricsContext';
import ResourcesRow from './ResourcesRow';
import styles from './ResourcesTable.module.css';

function ResourcesTable() {
  const { tests, lastPage, isLoading, api, params } = useContext(MetricsContext);

  const handleChangePage = (
      _: React.MouseEvent<HTMLButtonElement> | null,
      newPage: number,
  ) => {
    api.updatePage(newPage);
  };
  const handleChangeRowsPerPage = (
      event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    api.updateRowsPerPage(Number(event.target.value));
  };

  const handleButtonClick = (event) => {
    if (Number(params.sort) as SortType === event as SortType) {
      if (params.ascending ) {
        api.updateAscending(false);
      } else {
        api.updateAscending(true);
      }
    } else {
      api.updateSort(event);
      api.updateAscending(true);
    }
  };

  function sortableColumnLabel(sortType: SortType, colName: string) {
    return (
      <Button
        className={styles.filterButtonText}
        onClick={() => {
          handleButtonClick(sortType);
        }
        }
      >
        {colName}
        {
        params.ascending ?
          <ArrowDownwardIcon className={params.sort === sortType ? styles.icon : styles.iconNoShow}/> :
          <ArrowUpwardIcon className={params.sort === sortType ? styles.icon : styles.iconNoShow}/>
        }
      </Button>
    );
  }

  function tableMessageBoard(message: string) {
    return (
      <TableRow>
        <TableCell colSpan={7} align="center" className={styles.tableCellNoData}>
          {message}
        </TableCell>
      </TableRow>
    );
  }

  return (
    <>
      <TableContainer component={Paper}>
        <LinearProgress sx={{ visibility: isLoading ? 'visible' : 'hidden' }} data-testid='loading-bar'/>
        <Table sx={{ minWidth: 650 }} size="small" aria-label="simple table">
          <TableHead>
            <TableRow className={styles.headerRow}>
              <TableCell component="th" align="left">
                {sortableColumnLabel(SortType.SORT_NAME, 'Test')}
              </TableCell>
              <TableCell component="th" align="right">
                Test Suite
              </TableCell>
              <TableCell component="th" align="right">
                {sortableColumnLabel(SortType.SORT_NUM_RUNS, '# Runs')}
              </TableCell>
              <TableCell component="th" align="right">
                {sortableColumnLabel(SortType.SORT_NUM_FAILURES, '# Failures')}
              </TableCell>
              <TableCell component="th" align="right">
                {sortableColumnLabel(SortType.SORT_AVG_RUNTIME, 'Avg Runtime')}
              </TableCell>
              <TableCell component="th" align="right">
                {sortableColumnLabel(SortType.SORT_TOTAL_RUNTIME, 'Total Runtime')}
              </TableCell>
              <TableCell component="th" align="right">
                {sortableColumnLabel(SortType.SORT_AVG_CORES, 'Avg Cores')}
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody data-testid="tableBody">
            {tests.length > 0 ?
             tests.map((row) => (
               <ResourcesRow
                 key={row.testId} {
                   ...{
                     test: row,
                     lastPage: lastPage,
                   }
                 }/>
             )) : tableMessageBoard(isLoading ? 'Loading...' : 'No data available')}
          </TableBody>
          <TableFooter>
            <TableRow>
              <TablePagination
                data-testid="tableRowTest"
                rowsPerPageOptions={[25, 50, 100, 200]}
                count={lastPage ? (params.page * params.rowsPerPage): -1}
                rowsPerPage={params.rowsPerPage}
                page={params.page}
                onPageChange={handleChangePage}
                onRowsPerPageChange={handleChangeRowsPerPage}
              />
            </TableRow>
          </TableFooter>
        </Table>
      </TableContainer>
    </>
  );
}

export default ResourcesTable;
