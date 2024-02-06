// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import Paper from '@mui/material/Paper';
import { useContext, useState } from 'react';
import { Alert, Button, Snackbar, Tooltip } from '@mui/material';
import LinkIcon from '@mui/icons-material/Link';
import { ComponentContext } from '../../../features/components/ComponentContext';
import { DirectoryNodeType, MetricType, SortType } from '../../../api/resources';
import { formatNumber, formatTime } from '../../../utils/formatUtils';
import DataTable, { Column, PaginatorProps, Row } from '../../../components/table/DataTable';
import { Node, Path, Test, TestMetricsContext } from './TestMetricsContext';
import styles from './TestMetricsTable.module.css';
import { createSearchParams } from './TestMetricsSearchParams';

export interface TestMetricsTableProps {
  expandRowId: string[],
}

export function getFormatter(metricType: MetricType) {
  return metricType === MetricType.TOTAL_RUNTIME || metricType === MetricType.AVG_RUNTIME ? formatTime : formatNumber;
}

function TestMetricsTable(props: TestMetricsTableProps) {
  const { data, lastPage, isLoading, api, params, datesToShow } = useContext(TestMetricsContext);
  const { components } = useContext(ComponentContext);
  const [openAlert, setOpenAlert] = useState(false);

  const handleClipboardButtonClick = (expDir: string, expTest: string) => {
    setOpenAlert(true);
    const searchParams = createSearchParams(components, {
      ...params,
      date: params.date,
    }, expDir, expTest);
    navigator.clipboard.writeText(window.location.host + window.location.pathname + '?' + decodeURIComponent(searchParams.toString()));
  };

  const handleClose = () => {
    setOpenAlert(false);
  };

  const createCopyLink = (name: string, parentIds: string, fileName: string) => {
    return <>
      {name}
      <Button
        size="small"
        className={styles.clipboard}
        onClick={() => handleClipboardButtonClick(parentIds, fileName)}
        style ={{ padding: 0, marginLeft: 15, minWidth: 0 }}
      >
        <Tooltip title="Copy link of view to clipboard">
          <LinkIcon/>
        </Tooltip>
      </Button>
    </>;
  };
  function constructColumns() {
    const cols: Column[] = [{
      name: 'Test',
      renderer: (_: Column, row: Row<Node>) => {
        const node = row as Path;
        if (node.subname) {
          return { value: node.name };
        } else {
          if (node.type === DirectoryNodeType.DIRECTORY) {
            return { value: createCopyLink(node.name, node.id, ''), colSpan: 2 };
          }
          if (node.type === DirectoryNodeType.FILENAME) {
            return { value: createCopyLink(node.name, node.id, node.id), colSpan: 2 };
          }
          const fileName = (row as Test).fileName;
          return { value: createCopyLink(node.name, node.id, fileName), colSpan: 2 };
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
        return node.subname ? { value: node.subname } : undefined;
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
            const value = Number(node.metrics.get(col.name)?.get(params.timelineMetric));
            return { value: getFormatter(params.timelineMetric)(value) };
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
      const columns: [SortType, MetricType, string, string][] = [
        [SortType.SORT_NUM_RUNS, MetricType.NUM_RUNS, '# Runs', 'How many times a test was run, including all in-process, build-level, and attempt-level retries.'],
        [SortType.SORT_NUM_FAILURES, MetricType.NUM_FAILURES, '# Failures', 'How many times the test failed, counting failures that succeeded on retry.'],
        [SortType.SORT_AVG_RUNTIME, MetricType.AVG_RUNTIME, 'Avg Runtime', 'Average runtime for a single run of a test or weighted (by number of runs) sum of average runtimes of tests in the file/directory.'],
        [SortType.SORT_TOTAL_RUNTIME, MetricType.TOTAL_RUNTIME, 'Total Runtime', 'Total time spent running this test in given period.'],
        [SortType.SORT_AVG_CORES, MetricType.AVG_CORES, 'Avg Cores', 'Average number of cores spent running this test or the file/directory.'],
      ];
      columns.map(([sortType, metricType, name, description]) => {
        cols.push({
          name: name,
          renderer: (_: Column, row: Row<Node>) => {
            const node = row as Node;
            const value = Number(node.metrics.get(datesToShow[0])?.get(metricType));
            return { value: getFormatter(metricType)(value) };
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
      <DataTable
        isLoading={isLoading}
        rows={data}
        columns={constructColumns()}
        showPaginator={!params.directoryView}
        paginatorProps={paginatorProps}
        initialExpandRowIds={props.expandRowId}
      />
      <Snackbar
        open={openAlert}
        autoHideDuration={5000}
        onClose={handleClose}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert onClose={handleClose} variant='standard' severity="info" sx={{ width: '100%' }}>
              Link copied to clipboard
        </Alert>
      </Snackbar>
    </Paper>
  );
}

export default TestMetricsTable;
