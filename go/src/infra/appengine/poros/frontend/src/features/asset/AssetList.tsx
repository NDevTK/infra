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
  GridValueGetterParams,
} from '@mui/x-data-grid';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';

import RefreshIcon from '@mui/icons-material/Refresh';
import CreateIcon from '@mui/icons-material/Create';
import EditIcon from '@mui/icons-material/Edit';
import { Card, CardContent, Typography } from '@mui/material';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import {
  clearSelectedRecord,
  onSelectRecord,
  queryAssetAsync,
  queryAssetResourceAsync,
} from './assetSlice';
import { useAppSelector, useAppDispatch } from '../../app/hooks';
import {
  setRightSideDrawerOpen,
  setActiveEntity,
} from '../utility/utilitySlice';

export function AssetList() {
  const dispatch = useAppDispatch();

  // Calls only once when the component is loaded.
  React.useEffect(() => {
    dispatch(setActiveEntity('assets'));
  }, []);

  const rows: GridRowsProp = useAppSelector((state) => state.asset.assets);
  const columns: GridColDef[] = [
    { field: 'assetId', headerName: 'Id', flex: 0.5, hide: true },
    { field: 'name', headerName: 'Name', flex: 0.5 },
    { field: 'description', headerName: 'Description', flex: 1 },
    { field: 'assetType', headerName: 'Type', flex: 0.5 },
    { field: 'createdBy', headerName: 'Created By', flex: 0.5 },
    {
      field: 'createdAt',
      headerName: 'Created At',
      flex: 0.5,
      valueGetter: getLocalTime,
    },
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
    dispatch(onSelectRecord({ assetId: selectedRow.assetId }));
    dispatch(queryAssetResourceAsync());
    console.log(cellValues);
  };

  const handleCreateClick = () => {
    dispatch(clearSelectedRecord());
    handleRightSideDrawerOpen();
  };

  const handleRightSideDrawerOpen = () => {
    dispatch(setRightSideDrawerOpen());
  };

  const handleRefreshClick = () => {
    dispatch(queryAssetAsync({ pageSize: 100, pageToken: '' }));
  };

  function CustomToolbar() {
    return (
      <GridToolbarContainer>
        <GridToolbarColumnsButton />
        <GridToolbarDensitySelector />
      </GridToolbarContainer>
    );
  }

  function getLocalTime(params: GridValueGetterParams) {
    return (
      params.row.createdAt.toLocaleString('en-US', { timeZone: 'US/Pacific' }) +
      ' PT'
    );
  }

  return (
    <div>
      <Card>
        <CardContent>
          <Grid container spacing={2} padding={0}>
            <Grid
              item
              style={{
                display: 'flex',
                justifyContent: 'flex-start',
                alignItems: 'center',
              }}
              xs={8}
            >
              <Typography variant="h6">Assets</Typography>
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
                  endIcon={<RefreshIcon />}
                  onClick={handleRefreshClick}
                >
                  Refresh
                </Button>
                <Button
                  variant="contained"
                  onClick={handleCreateClick}
                  endIcon={<CreateIcon />}
                >
                  Create
                </Button>
              </Stack>
            </Grid>
          </Grid>
        </CardContent>
      </Card>
      <hr style={{ height: 1, visibility: 'hidden' }} />
      <Card>
        <CardContent>
          <div style={{ width: '100%' }}>
            <DataGrid
              autoHeight
              density="compact"
              disableDensitySelector
              getRowId={(r) => r.assetId}
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
        </CardContent>
      </Card>
    </div>
  );
}
