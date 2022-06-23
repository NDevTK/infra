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
  GridValueGetterParams,
} from '@mui/x-data-grid';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';

import RefreshIcon from '@mui/icons-material/Refresh';
import { Card, CardContent, Typography } from '@mui/material';
import Grid from '@mui/material/Grid';
import { queryAssetInstanceAsync } from './assetInstanceSlice';
import { useAppSelector, useAppDispatch } from '../../app/hooks';
import { setActiveEntity } from '../utility/utilitySlice';
import { queryAssetAsync } from '../asset/assetSlice';
import { AssetModel } from '../../api/asset_service';

export function AssetInstanceList() {
  const dispatch = useAppDispatch();

  // Calls only once when the component is loaded.
  React.useEffect(() => {
    dispatch(setActiveEntity('assetInstances'));
  }, []);
  React.useEffect(() => {
    dispatch(queryAssetAsync({ pageSize: 100, pageToken: '' }));
  }, []);
  const assets: AssetModel[] = useAppSelector(
    (state) => state.assetInstance.assets
  );
  const rows: GridRowsProp = useAppSelector(
    (state) => state.assetInstance.assetInstances
  );
  const columns: GridColDef[] = [
    { field: 'assetInstanceId', headerName: 'Id', flex: 1, hide: true },
    {
      field: 'assetName',
      headerName: 'Asset Name',
      flex: 1,
      valueGetter: getAssetName,
    },
    { field: 'assetId', headerName: 'AssetId', flex: 1 },
    { field: 'status', headerName: 'Status', flex: 1 },
    {
      field: 'createdAt',
      headerName: 'Created At',
      flex: 0.5,
      valueGetter: getLocalTime,
    },
  ];

  const handleRefreshClick = () => {
    dispatch(queryAssetInstanceAsync({ pageSize: 100, pageToken: '' }));
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

  function getAssetName(params: GridValueGetterParams) {
    const assetId = params.row.assetId;
    return assets.filter((asset) => asset.assetId == assetId)[0].name;
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
              <Typography variant="h6">AssetInstances</Typography>
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
              getRowId={(r) => r.assetInstanceId}
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
