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
import { Box, Button, LinearProgress, TableFooter, TablePagination } from '@mui/material';
import { SortType } from '../../api/resources';
import { MetricsContext, convertToSortIndex } from '../context/MetricsContext';
import ResourcesRow from './ResourcesRow';
import styles from './ResourcesTable.module.css';

function ResourcesTable() {
  const { data, lastPage, isLoading, api, params, datesToShow } = useContext(MetricsContext);

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

  const handleSortType = (event) => {
    if (params.sort === event as SortType) {
      api.updateAscending(!params.ascending);
    } else {
      api.updateSort(event);
    }
  };

  const handleSortDate = (date) => {
    if (date === datesToShow[params.sortIndex]) {
      api.updateAscending(!params.ascending);
    } else {
      api.updateSortIndex(convertToSortIndex(datesToShow, date));
    }
  };

  function sortableColumnLabel(sortType: SortType, colName: string) {
    return (
      <Button
        className={styles.filterButtonText}
        onClick={() => {
          handleSortType(sortType);
        }
        }
      >
        {colName}
        {
        params.ascending ?
          <ArrowUpwardIcon className={params.sort === sortType ? styles.icon : styles.iconNoShow}/> :
          <ArrowDownwardIcon className={params.sort === sortType ? styles.icon : styles.iconNoShow}/>
        }
      </Button>
    );
  }

  function sortableDateColumn(date: string) {
    return (
      <Button
        className={styles.filterButtonText}
        onClick={() => {
          handleSortDate(date);
        }
        }
      >
        {date}
        {
        params.sort !== SortType.SORT_NAME ? (params.ascending ?
        <ArrowUpwardIcon className={datesToShow[params.sortIndex] === date ? styles.icon : styles.iconNoShow}/> :
        <ArrowDownwardIcon className={datesToShow[params.sortIndex] === date ? styles.icon : styles.iconNoShow}/>) :
        null}
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

  function displayHeader() {
    if (params.timelineView) {
      return (
        <>
          {datesToShow.map((date) => (
            <TableCell key={date} component="th" align="right" data-testid="timelineHeader"
              sx={{ whiteSpace: 'nowrap', width: '8%', minWidth: '100px', maxWidth: '140px' }}>
              {sortableDateColumn(date)}
            </TableCell>
          ))}
        </>
      );
    }
    const columns: [SortType, string][] = [
      [SortType.SORT_NUM_RUNS, '# Runs'],
      [SortType.SORT_NUM_FAILURES, '# Failures'],
      [SortType.SORT_AVG_RUNTIME, 'Avg Runtime'],
      [SortType.SORT_TOTAL_RUNTIME, 'Total Runtime'],
      [SortType.SORT_AVG_CORES, 'Avg Cores'],
    ];
    return (
      <>
        {columns.map(([sortType, name]) => (
          <TableCell key={name} component="th" align="right"
            sx={{ whiteSpace: 'nowrap', width: '8%', minWidth: '100px', maxWidth: '140px' }}>
            {sortableColumnLabel(sortType, name)}
          </TableCell>
        ))}
      </>
    );
  }

  return (
    <>
      <TableContainer component={Paper} sx={{ position: 'relative' }}>
        <LinearProgress sx={{ visibility: isLoading ? 'visible' : 'hidden' }} data-testid='loading-bar'/>
        <Box className={styles.loadingDimmer} sx={{ visibility: isLoading? 'visible' : 'hidden' }}></Box>
        <Table sx={{ minWidth: 650 }} size="small" aria-label="simple table">
          <TableHead>
            <TableRow className={styles.headerRow}>
              <TableCell component="th" align="left" sx={{ width: '30%' }}>
                {sortableColumnLabel(SortType.SORT_NAME, 'Test')}
              </TableCell>
              <TableCell component="th" align="left" sx={{ width: '20%' }}>
                Test Suite
              </TableCell>
              {displayHeader()}
            </TableRow>
          </TableHead>
          <TableBody data-testid="tableBody">
            {data.length > 0 ?
             data.map(
                 (row) => <ResourcesRow key={row.id} data={row} depth={0}/>,
             ) : tableMessageBoard(isLoading ? 'Loading...' : 'No data available')}
          </TableBody>
          {params.directoryView ? null : (
          <TableFooter>
            <TableRow>
              <TablePagination
                data-testid="tablePagination"
                rowsPerPageOptions={[25, 50, 100, 200]}
                count={lastPage ? (params.page * params.rowsPerPage): -1}
                rowsPerPage={params.rowsPerPage}
                page={params.page}
                onPageChange={handleChangePage}
                onRowsPerPageChange={handleChangeRowsPerPage}
                showFirstButton
              />
            </TableRow>
          </TableFooter>
          )}
        </Table>
      </TableContainer>
    </>
  );
}

export default ResourcesTable;
