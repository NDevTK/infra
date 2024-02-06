// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { IconButton, TableCell, TableRow, styled } from '@mui/material';
import { useState } from 'react';
import { ArrowDropDownIcon } from '@mui/x-date-pickers';
import { Column, Row } from './DataTable';

export interface DataTableRowProps {
  row: Row<any>,
  depth: number,
  columns: Column[],
  expandedRowIds?: string[],
  onTrigger?: (boolean) => void,
}

const StyledTableRow = styled(TableRow)(({ theme }) => ({
  '&:nth-of-type(odd)': {
    backgroundColor: theme.palette.action.hover,
  },
}));

function DataTableRow(props: DataTableRowProps) {
  const [isOpen, setIsOpen] = useState(props.expandedRowIds?.includes(props.row.id));
  const rotate = isOpen ? 'rotate(0deg)' : 'rotate(270deg)';
  const row = props.row;

  function handleOpenToggle() {
    if (!isOpen && props.row.onExpand !== undefined) {
      props.row.onExpand(props.row);
    }
    if (props.onTrigger !== undefined) {
      props.onTrigger(!isOpen);
    }
    setIsOpen(!isOpen);
  }
  return (
    <>
      <StyledTableRow
        data-testid={'tablerow-' + row.id}
        data-depth={props.depth}
      >
        {
          props.columns.map((column, index) => {
            const cell = column.renderer(column, row);
            const contents = cell?.value;
            const colSpan = cell?.colSpan ? cell.colSpan : undefined;
            const cellSx = cell?.sxProps ? cell.sxProps : {};
            if (contents === undefined) {
              return;
            }
            return (
              <TableCell
                key={index}
                data-testid={'tableCell'}
                align={column.align}
                colSpan={colSpan}
                sx={{ ...cellSx, paddingLeft: props.depth * 2 + 2, whiteSpace: 'nowrap' }}
              >
                {
                  index === 0 && props.row.isExpandable ? (
                    <IconButton
                      data-testid={'clickButton-' + row.id}
                      color="primary"
                      size="small"
                      onClick={handleOpenToggle}
                      style={{ transform: rotate }}
                      sx={{ margin: 0, padding: 0, ml: -2 }}
                    >
                      <ArrowDropDownIcon/>
                    </IconButton>
                  ) : null
                }
                {contents}
              </TableCell>
            );
          })
        }
      </StyledTableRow>
      {
      isOpen && row.rows !== undefined && row.rows.length > 0 ? (
        <>
          {row.rows.map((row) => (
            <DataTableRow
              key={row.id}
              row={row}
              depth={props.depth + 1}
              columns={props.columns}
              expandedRowIds={props.expandedRowIds}
              onTrigger={props.onTrigger}
            />
          ))}
          {row.footer ?
          <StyledTableRow
          >
            <TableCell
              colSpan={props.columns.length}
              sx={{ paddingLeft: props.depth * 2 + 2, whiteSpace: 'nowrap' }}
            >
              {row.footer}
            </TableCell>
          </StyledTableRow> : null}
        </>
      ) : null
      }
    </>
  );
}

export default DataTableRow;
