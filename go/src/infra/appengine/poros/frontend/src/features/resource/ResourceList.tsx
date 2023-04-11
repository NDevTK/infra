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
import { Card, CardContent, Typography } from '@mui/material';
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
  React.useEffect(() => {
    dispatch(queryResourceAsync({ pageSize: 100, pageToken: '' }));
  }, []);

  const rows: GridRowsProp = useAppSelector(
    (state) => state.resource.resources
  );

  const dateComparator: GridComparatorFn<Date> = (v1, v2) =>
    new Date(v1).valueOf() - new Date(v2).valueOf();

  const columns: GridColDef[] = [
    { field: 'resourceId', headerName: 'Id', flex: 0.4, hide: true },
    { field: 'name', headerName: 'Name', flex: 0.4 },
    { field: 'type', headerName: 'Type', flex: 0.4 },
    { field: 'operatingSystem', headerName: 'Operating System', flex: 0.3 },
    { field: 'description', headerName: 'Description', flex: 0.6 },
    { field: 'imageSource', headerName: 'Image Source', flex: 0.7 },
    { field: 'createdBy', headerName: 'Created By', flex: 0.4 },
    {
      field: 'createdAt',
      headerName: 'Created At',
      flex: 0.4,
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
  ];
  const handleEditClick = (cellValues: GridRenderCellParams) => {
    const selectedRow = cellValues.row;
    handleRightSideDrawerOpen();
    dispatch(clearSelectedRecord());
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

  function getLocalTime(params: GridValueGetterParams) {
    return params.row.createdAt.toLocaleString();
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
              <Typography variant="h6">Resources(Admin Only)</Typography>
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
                  data-testid="refresh-button"
                  variant="outlined"
                  endIcon={<RefreshIcon />}
                  onClick={handleRefreshClick}
                >
                  Refresh
                </Button>
                <Button
                  data-testid="create-button"
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
              initialState={{
                sorting: {
                  sortModel: [{ field: 'createdAt', sort: 'asc' }],
                },
              }}
              disableVirtualization
            />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
