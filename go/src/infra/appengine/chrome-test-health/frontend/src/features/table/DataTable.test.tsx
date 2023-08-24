// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render, screen } from '@testing-library/react';
import DataTable, { Column, PaginatorProps, Row } from './DataTable';

const tests: Row<any>[] = [
  {
    id: 'testId',
    isExpandable: true,
    rows: [
      {
        id: 'v1',
        isExpandable: false,
        rows: [],
      },
      {
        id: 'v1',
        isExpandable: false,
        rows: [],
      },
    ],
  },
  {
    id: 'testId1',
    isExpandable: true,
    rows: [
      {
        id: 'v1',
        isExpandable: false,
        rows: [],
      },
      {
        id: 'v1',
        isExpandable: false,
        rows: [],
      },
    ],
  },
];

const columns: Column[] = [
  {
    name: 'Test',
    align: 'left',
    renderer: (_: Column, _1: Row<any>) => {
      return '';
    },
  },
  {
    name: 'Test2',
    align: 'left',
    renderer: (_: Column, _1: Row<any>) => {
      return '';
    },
  },
];

describe('when rendering the DataTable', () => {
  it('show no data screen if no data', () => {
    render(
        <DataTable rows={[]} columns={[]}/>,
    );
    expect(screen.getByText('No data available')).toBeInTheDocument();
    expect(screen.queryByTestId('tablePagination')).toBeNull();
  });
  it('show is loading', () => {
    render(
        <DataTable rows={[]} columns={[]} isLoading={true}/>,
    );
    expect(screen.getByText('Loading...')).toBeInTheDocument();
    expect(screen.queryByTestId('loading-bar')).toBeInTheDocument();
  });
  it('display table correctly', () => {
    const paginatorProps: PaginatorProps = {
      rowsPerPage: 5,
      rowsPerPageOptions: [5, 10],
      count: 5,
      page: 0,
      onChangeRowsPerPage(_event) {
        return;
      },
      onPageChange(_, _newPage) {
        return;
      },
    };
    render(
        <DataTable rows={tests} columns={columns} showPaginator={true} paginatorProps={paginatorProps}/>,
    );
    expect(screen.getByTestId('tableBody')).toBeInTheDocument();
    expect(screen.getByTestId('tablerow-testId')).toBeInTheDocument();
    expect(screen.getByTestId('tablerow-testId1')).toBeInTheDocument();
    expect(screen.getAllByTestId('columnHeaderTest')).toHaveLength(2);
    expect(screen.getByTestId('tablePagination')).toBeInTheDocument();
  });
});
