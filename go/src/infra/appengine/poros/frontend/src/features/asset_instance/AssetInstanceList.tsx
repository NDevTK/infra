// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React, { useEffect, useRef } from 'react';
import {
  DataGrid,
  GridRowsProp,
  GridColDef,
  GridFilterItem,
  GridFilterModel,
  GridToolbarContainer,
  GridToolbarColumnsButton,
  GridToolbarDensitySelector,
  GridCellParams,
  MuiEvent,
  GridValueGetterParams,
  GridRenderCellParams,
  gridDateComparator,
} from '@mui/x-data-grid';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';

import RefreshIcon from '@mui/icons-material/Refresh';
import { Card, CardContent, IconButton, Link, Typography } from '@mui/material';
import Grid from '@mui/material/Grid';
import { onSelectRecord, queryAssetInstanceAsync } from './assetInstanceSlice';
import { useAppSelector, useAppDispatch } from '../../app/hooks';
import {
  fetchLogsAsync,
  setActiveEntity,
  setActiveLogAssetDetails,
  setRightSideDrawerOpen,
} from '../utility/utilitySlice';
import { queryAssetAsync } from '../asset/assetSlice';
import { AssetModel } from '../../api/asset_service';
import EditIcon from '@mui/icons-material/Edit';
import NotesIcon from '@mui/icons-material/Notes';
import { AssetInstanceLogs } from './AssetInstanceLogs';

export function AssetInstanceList() {
  const dispatch = useAppDispatch();

  // Calls only once when the component is loaded.
  React.useEffect(() => {
    dispatch(setActiveEntity('assetInstances'));
  }, []);
  React.useEffect(() => {
    dispatch(queryAssetAsync({ pageSize: 100, pageToken: '' }));
  }, []);
  React.useEffect(() => {
    dispatch(queryAssetInstanceAsync({ pageSize: 100, pageToken: '' }));
  }, []);
  const assets: AssetModel[] = useAppSelector(
    (state) => state.assetInstance.assets
  );
  const rows: GridRowsProp = useAppSelector(
    (state) => state.assetInstance.assetInstances
  );

  const showLogs: boolean = useAppSelector((state) => state.utility.showLogs);

  const ref = useRef<HTMLDivElement>(null);

  const columns: GridColDef[] = [
    { field: 'assetInstanceId', headerName: 'Id', flex: 1, hide: true },
    {
      field: 'assetName',
      headerName: 'Lab Name',
      flex: 1,
      valueGetter: getAssetName,
    },
    { field: 'assetId', headerName: 'AssetId', flex: 1, hide: true },
    { field: 'status', headerName: 'Status', flex: 1 },
    {
      field: 'createdAt',
      headerName: 'Created At',
      flex: 1,
      valueGetter: getLocalTime,
      sortComparator: gridDateComparator,
    },
    {
      field: 'createdBy',
      headerName: 'Created By',
      flex: 1,
    },
    {
      field: 'deleteAt',
      headerName: 'Delete At',
      flex: 1,
      valueGetter: getLocalTime,
    },
    {
      field: 'projectId',
      headerName: 'Project',
      flex: 2,
      renderCell: (cellValues) => {
        return <Link target={'_blank'} href={`${getProject(cellValues)}`}>{getProject(cellValues)}</Link>;
      },
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
      field: 'Logs',
      renderCell: (cellValues) => {
        return (
          <IconButton
            aria-label="delete"
            size="small"
            onClick={() => {
              handleShowLogsClick(cellValues);
            }}
          >
            <NotesIcon fontSize="inherit" />
          </IconButton>
        );
      },
    },
  ];

  const handleRefreshClick = () => {
    dispatch(queryAssetInstanceAsync({ pageSize: 100, pageToken: '' }));
  };

  const handleEditClick = (cellValues: GridRenderCellParams) => {
    const selectedRow = cellValues.row;
    dispatch(setRightSideDrawerOpen());
    dispatch(onSelectRecord({ assetInstanceId: selectedRow.assetInstanceId }));
  };

  const handleShowLogsClick = async (cellValues: GridRenderCellParams) => {
    const selectedRow = cellValues.row;
    dispatch(
      setActiveLogAssetDetails({
        name: getAssetName(cellValues),
        createdAt: selectedRow.createdAt.toLocaleString(),
        status: selectedRow.status,
      })
    );
    await dispatch(
      fetchLogsAsync({ assetInstanceId: selectedRow.assetInstanceId })
    );
    ref.current?.scrollIntoView({
      block: 'center',
      inline: 'center',
    });
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
    if (params.field === 'createdAt') {
      return params.row.createdAt.toLocaleString();
    } else if (
      params.field === 'deleteAt' &&
      params.row.deleteAt !== undefined
    ) {
      return params.row.deleteAt.toLocaleString();
    }
  }

  function getProject(params: GridValueGetterParams) {
    if (params.row.projectId == '') return '';

    return `https://console.cloud.google.com/compute/instances?project=${params.row.projectId}`
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
              <Typography variant="h6">Lab Instances</Typography>
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
              initialState={{
                sorting: {
                  sortModel: [{ field: 'createdAt', sort: 'desc' }],
                },
                filter: {
                  filterModel: {
                    items: [{ columnField: 'status', operatorValue: 'isAnyOf', value: ['STATUS_PENDING', 'STATUS_RUNNING', 'STATUS_COMPLETED'] }],
                  },
                },
              }}
            />
          </div>
        </CardContent>
      </Card>
      <hr style={{ height: 1, visibility: 'hidden' }} />
      <div style={{ maxWidth: '80vw', margin: 'auto' }} ref={ref}>
        {showLogs ? (
          <Card>
            <AssetInstanceLogs></AssetInstanceLogs>
          </Card>
        ) : null}
      </div>
    </div>
  );
}
