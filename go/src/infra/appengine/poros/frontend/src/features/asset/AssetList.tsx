// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React from 'react';
import {
  DataGrid,
  GridRowsProp,
  GridColDef,
  GridComparatorFn,
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
import LaunchIcon from '@mui/icons-material/Launch';
import {
  Card,
  CardContent,
  Dialog,
  DialogTitle,
  Typography,
} from '@mui/material';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import {
  clearSelectedRecord,
  onSelectRecord,
  queryAssetAsync,
  queryAssetResourceAsync,
  createAssetInstanceAsync,
  setAssetSpinRecord,
  getDefaultResources,
} from './assetSlice';
import { useAppSelector, useAppDispatch } from '../../app/hooks';
import {
  setRightSideDrawerOpen,
  setActiveEntity,
  setSpinDialogClose,
  setSpinDialogOpen,
  setRightSideDrawerClose,
} from '../utility/utilitySlice';
import { Link, Navigate, useNavigate } from 'react-router-dom';

export function AssetList() {
  const dispatch = useAppDispatch();

  // Calls only once when the component is loaded.
  React.useEffect(() => {
    dispatch(setActiveEntity('assets'));
  }, []);
  React.useEffect(() => {
    dispatch(queryAssetAsync({ pageSize: 100, pageToken: '' }));
  }, []);

  const spinDialogOpen = useAppSelector(
    (state) => state.utility.spinDialogOpen
  );
  const assetSpinRecord = useAppSelector(
    (state) => state.asset.assetSpinRecord
  );
  const rows: GridRowsProp = useAppSelector((state) => state.asset.assets);

  const dateComparator: GridComparatorFn<Date> = (v1, v2) =>
    new Date(v1).valueOf() - new Date(v2).valueOf();

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
      type: 'date',
      valueGetter: getLocalTime,
      sortComparator: dateComparator,
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
    {
      field: 'Deploy',
      renderCell: (cellValues) => {
        return (
          <IconButton
            aria-label="refresh"
            size="small"
            onClick={() => {
              handleSpinClick(cellValues);
            }}
          >
            <LaunchIcon fontSize="inherit" />
          </IconButton>
        );
      },
    },
  ];
  const handleEditClick = (cellValues: GridRenderCellParams) => {
    const selectedRow = cellValues.row;
    handleRightSideDrawerOpen();
    dispatch(clearSelectedRecord());
    dispatch(onSelectRecord({ assetId: selectedRow.assetId }));
    dispatch(queryAssetResourceAsync());
  };

  const handleCreateClick = () => {
    dispatch(clearSelectedRecord());
    dispatch(getDefaultResources('active_directory'));
    handleRightSideDrawerOpen();
  };

  const handleRightSideDrawerOpen = () => {
    dispatch(setRightSideDrawerOpen());
  };

  const handleRefreshClick = () => {
    dispatch(queryAssetAsync({ pageSize: 100, pageToken: '' }));
  };

  const handleSpinClick = (celleValues: GridRenderCellParams) => {
    dispatch(setSpinDialogOpen());
    dispatch(setAssetSpinRecord(celleValues.row.assetId));
  };

  const handleSpinClose = () => {
    dispatch(setSpinDialogClose());
  };

  const handleSpinConfirm = () => {
    dispatch(createAssetInstanceAsync(assetSpinRecord));
    dispatch(setSpinDialogClose());
    dispatch(setRightSideDrawerClose());
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
    return params.row.createdAt.toLocaleString();
  }

  return (
    <div>
      <Dialog onClose={handleSpinClose} open={spinDialogOpen}>
        <DialogTitle>Do you want to deploy this template?</DialogTitle>
        <Stack
          direction="row"
          spacing={6}
          sx={{
            padding: 1,
            margin: 1,
            display: 'flex',
            marginLeft: 'auto',
            marginRight: 'auto',
          }}
        >
          <Button variant="outlined" size="small" onClick={handleSpinClose}>
            Cancel
          </Button>
          <Button
            variant="contained"
            size="small"
            onClick={handleSpinConfirm}
            component={Link}
            to="/assetInstances"
          >
            Confirm
          </Button>
        </Stack>
      </Dialog>
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
              <Typography variant="h6">Lab templates</Typography>
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
              initialState={{
                sorting: {
                  sortModel: [{ field: 'createdAt', sort: 'asc' }],
                },
              }}
            />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
