// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import Paper from '@mui/material/Paper';
import { useContext } from 'react';
import { MetricType, SortType } from '../../../api/resources';
import { formatNumber, formatTime } from '../../../utils/formatUtils';
import DataTable, { Column, PaginatorProps, Row } from '../../../components/table/DataTable';
import { Node, TestMetricsContext } from './TestMetricsContext';

function TestMetricsTable() {
  const { data, lastPage, isLoading, api, params, datesToShow } = useContext(TestMetricsContext);

  function constructColumns() {
    const cols: Column[] = [{
      name: 'Test',
      renderer: (_: Column, row: Row<Node>) => {
        const node = row as Node;
        if (node.subname) {
          return node.name;
        } else {
          return [node.name, 2];
        }
      },
      align: 'left',
      isSortedBy: params.sort === SortType.SORT_NAME,
      isSortAscending: params.sort === SortType.SORT_NAME ? params.ascending : undefined,
      sx: { width: '30%' },
      onClick: () => {
        if (params.sort === SortType.SORT_NAME) {
          api.updateAscending(!params.ascending);
        } else {
          api.updateSort(SortType.SORT_NAME);
        }
      },
    }, {
      name: 'Test Suite',
      renderer: (_: Column, row: Row<Node>) => {
        const node = row as Node;
        return node.subname ? node.subname : undefined;
      },
      align: 'left',
      sx: { width: '20%' },
    }];
    if (params.timelineView) {
      datesToShow.map((date, index) => {
        cols.push({
          name: date,
          renderer: (col: Column, row: Row<Node>) => {
            const node = row as Node;
            return formatNumber(Number(node.metrics.get(col.name)?.get(params.timelineMetric)));
          },
          isSortedBy: params.sortIndex === index,
          isSortAscending: params.sortIndex === index ? params.ascending : undefined,
          align: 'right',
          sx: { whiteSpace: 'nowrap', width: '8%', minWidth: '100px', maxWidth: '140px' },
          onClick: () => {
            if (index === params.sortIndex) {
              api.updateAscending(!params.ascending);
            } else {
              api.updateSortIndex(index);
            }
          },
        });
      });
    } else {
      const columns: [SortType, MetricType, string, (val:any) => string, string][] = [
        [SortType.SORT_NUM_RUNS, MetricType.NUM_RUNS, '# Runs', formatNumber, 'How many times a test was run, including all in-process, build-level, and attempt-level retries.'],
        [SortType.SORT_NUM_FAILURES, MetricType.NUM_FAILURES, '# Failures', formatNumber, 'How many times the test failed, counting failures that succeeded on retry.'],
        [SortType.SORT_AVG_RUNTIME, MetricType.AVG_RUNTIME, 'Avg Runtime', formatTime, 'Average runtime for a single run of a test or sum of average runtimes of tests in the file/directory.'],
        [SortType.SORT_TOTAL_RUNTIME, MetricType.TOTAL_RUNTIME, 'Total Runtime', formatTime, 'Total time spent running this test in given period.'],
        [SortType.SORT_AVG_CORES, MetricType.AVG_CORES, 'Avg Cores', formatNumber, 'Average number of cores spent running this test.'],
      ];
      columns.map(([sortType, metricType, name, format, description]) => {
        cols.push({
          name: name,
          renderer: (_: Column, row: Row<Node>) => {
            const node = row as Node;
            return format(node.metrics.get(datesToShow[0])?.get(metricType));
          },
          align: 'right',
          isSortedBy: params.sort == sortType,
          isSortAscending: params.sort === sortType ? params.ascending : undefined,
          sx: { whiteSpace: 'nowrap', width: '8%', minWidth: '100px', maxWidth: '140px' },
          onClick: () => {
            if (params.sort === sortType) {
              api.updateAscending(!params.ascending);
            } else {
              api.updateSort(sortType);
            }
          },
          description: description,
        });
      });
    }
    return cols;
  }

  const paginatorProps: PaginatorProps = {
    rowsPerPageOptions: [25, 50, 100, 200],
    count: lastPage ? (params.page * params.rowsPerPage): -1,
    rowsPerPage: params.rowsPerPage,
    page: params.page,
    onPageChange: (
        _: React.MouseEvent<HTMLButtonElement> | null,
        newPage: number,
    ) => {
      api.updatePage(newPage);
    },
    onChangeRowsPerPage: (
        event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
    ) => {
      api.updateRowsPerPage(Number(event.target.value));
    },
  };

  return (
    <Paper>
      <DataTable isLoading={isLoading} rows={data} columns={constructColumns()} showPaginator={!params.directoryView} paginatorProps={paginatorProps}/>
    </Paper>
  );
}

export default TestMetricsTable;
