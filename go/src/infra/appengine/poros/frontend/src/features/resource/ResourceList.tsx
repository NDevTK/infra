// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React from 'react';
import {
  DataGrid,
  GridRowsProp,
  GridColDef,
  GridToolbarContainer,
  GridToolbarColumnsButton,
  GridToolbarDensitySelector,
  GridCellParams,
  MuiEvent,
  GridRenderCellParams,
} from '@mui/x-data-grid';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';

import RefreshIcon from '@mui/icons-material/Refresh';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import { Typography } from '@mui/material';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import {
  clearSelectedRecord,
  onSelectRecord,
  queryResourceAsync,
} from './resourceSlice';
import { useAppSelector, useAppDispatch } from '../../app/hooks';
import {
  setRightSideDrawerOpen,
  setActiveEntity,
} from '../utility/utilitySlice';

export function ResourceList() {
  const dispatch = useAppDispatch();

  // Calls only once when the component is loaded.
  React.useEffect(() => {
    dispatch(setActiveEntity('resources'));
  }, []);

  const rows: GridRowsProp = useAppSelector(
    (state) => state.resource.resources
  );
  const columns: GridColDef[] = [
    { field: 'resourceId', headerName: 'Id', width: 150 },
    { field: 'name', headerName: 'Name', width: 150 },
    { field: 'type', headerName: 'Type', width: 150 },
    { field: 'description', headerName: 'Description', width: 150 },
    { field: 'machineInfo', headerName: 'Machine Info', width: 150 },
    { field: 'domainInfo', headerName: 'Domain Info', width: 150 },
    { field: 'createdBy', headerName: 'Created By', width: 150 },
    { field: 'createdAt', headerName: 'Created At', width: 150 },
    {
      field: 'Edit',
      renderCell: (cellValues) => {
        return (
          <IconButton
            aria-label="delete"
            size="small"
            onClick={() => {
              handleEditClick(cellValues);
            }}
          >
            <EditIcon fontSize="inherit" />
          </IconButton>
        );
      },
    },
  ];
  const handleEditClick = (cellValues: GridRenderCellParams) => {
    const selectedRow = cellValues.row;
    handleRightSideDrawerOpen();
    dispatch(onSelectRecord({ resourceId: selectedRow.resourceId }));
  };

  const handleCreateClick = () => {
    dispatch(clearSelectedRecord());
    handleRightSideDrawerOpen();
  };

  const handleRightSideDrawerOpen = () => {
    dispatch(setRightSideDrawerOpen());
  };

  const handleRefreshClick = () => {
    dispatch(queryResourceAsync({ pageSize: 100, pageToken: '' }));
  };

  function CustomToolbar() {
    return (
      <GridToolbarContainer>
        <GridToolbarColumnsButton />
        <GridToolbarDensitySelector />
      </GridToolbarContainer>
    );
  }

  return (
    <div>
      <Grid container spacing={2} padding={1}>
        <Grid
          item
          style={{
            display: 'flex',
            justifyContent: 'flex-start',
            alignItems: 'center',
          }}
          xs={8}
        >
          <Typography variant="h6">Resources</Typography>
        </Grid>
        <Grid
          item
          style={{
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center',
          }}
          xs={4}
        >
          <Stack direction="row" spacing={2}>
            <Button
              variant="outlined"
              startIcon={<RefreshIcon />}
              onClick={handleRefreshClick}
            >
              Refresh
            </Button>
            <Button
              variant="contained"
              onClick={handleCreateClick}
              endIcon={<DeleteIcon />}
            >
              Create
            </Button>
          </Stack>
        </Grid>
      </Grid>

      <div style={{ width: '100%' }}>
        <DataGrid
          autoHeight
          getRowId={(r) => r.resourceId}
          rows={rows}
          columns={columns}
          components={{
            Toolbar: CustomToolbar,
          }}
          onCellClick={(
            params: GridCellParams,
            event: MuiEvent<React.MouseEvent>
          ) => {
            event.defaultMuiPrevented = true;
          }}
        />
      </div>
    </div>
  );
}
