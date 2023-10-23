// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { act, fireEvent, render, screen } from '@testing-library/react';
import DataTableRow from './DataTableRow';
import { Column, Row } from './DataTable';

const columns: Column[] = [
  {
    name: 'Test',
    align: 'left',
    renderer: (_: Column, _row: Row<any>) => {
      return '';
    },
  },
  {
    name: 'Test2',
    align: 'left',
    renderer: (_: Column, _row: Row<any>) => {
      return '';
    },
  },
];


describe('when rendering the DataTableRow', () => {
  it('should render a single row', () => {
    const test: Row<any> = {
      id: 'testId',
      isExpandable: false,
      rows: [],
    };
    render(
        <table>
          <tbody>
            <DataTableRow row={test} depth={0} columns={columns}/>
          </tbody>
        </table>,
    );
    const tableRow = screen.getByTestId('tablerow-testId');
    expect(tableRow).toBeInTheDocument();
  });
  it('should render expandable rows', () => {
    const test: Row<any> = {
      id: 'testId',
      isExpandable: true,
      rows: [
        {
          id: 'v1',
          isExpandable: false,
          rows: [],
        },
        {
          id: 'v2',
          isExpandable: false,
          rows: [],
        },
      ],
    };

    render(
        <table>
          <tbody>
            <DataTableRow row={test} depth={0} columns={columns}/>
          </tbody>
        </table>,
    );
    const testRow = screen.getByTestId('tablerow-testId');
    expect(testRow).toBeInTheDocument();
    expect(testRow.getAttribute('data-depth')).toEqual('0');

    const button = screen.getByTestId('clickButton-testId');
    fireEvent.click(button);

    const v1Row = screen.getByTestId('tablerow-v1');
    expect(v1Row).toBeInTheDocument();
    expect(v1Row.getAttribute('data-depth')).toEqual('1');

    expect(screen.getByTestId('tablerow-v2')).toBeInTheDocument();
  });
  it('should render correct number of columns', async () => {
    const test: Row<any> = {
      id: 'testId',
      isExpandable: true,
      rows: [
        {
          id: 'v1',
          isExpandable: false,
          rows: [],
        },
        {
          id: 'v2',
          isExpandable: false,
          rows: [],
        },
      ],
    };
    render(
        <table>
          <tbody>
            <DataTableRow row={test} depth={0} columns={columns}/>
          </tbody>
        </table>,
    );
    expect(screen.getAllByTestId('tableCell')).toHaveLength(2);
  });
  it('should render footer', async () => {
    const test: Row<any> = {
      id: 'testId',
      isExpandable: true,
      rows: [
        {
          id: 'v1',
          isExpandable: false,
          rows: [],
        },
      ],
      footer: <div data-testid="footerTestId">Test</div>,
    };
    render(
        <table>
          <tbody>
            <DataTableRow row={test} depth={0} columns={columns}/>
          </tbody>
        </table>,
    );
    await act(async () => {
      fireEvent.click(screen.getByTestId('clickButton-testId'));
    });
    expect(screen.getByTestId('footerTestId')).toBeInTheDocument();
  });
});
