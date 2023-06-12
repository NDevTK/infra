// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React, { useEffect } from 'react';
import CloseIcon from '@mui/icons-material/Close';
import Button from '@mui/material/Button';

import RefreshIcon from '@mui/icons-material/Refresh';
import { Box, Typography } from '@mui/material';
import Grid from '@mui/material/Grid';
import { useAppSelector, useAppDispatch } from '../../app/hooks';
import { fetchLogsAsync, hideLogs } from '../utility/utilitySlice';

export function AssetInstanceLogs() {
  let interval: any;
  const logs: string = useAppSelector((state) => state.utility.logs);

  const assetDetails: any = useAppSelector(
    (state) => state.utility.activeLogsAssetDetails
  );

  const assetInstanceId: string = useAppSelector(
    (state) => state.utility.activeLogsAssetInstanceId
  );

  const createInterval = () => {
    clearInterval(interval);
    interval = setInterval(async () => {
      await dispatch(fetchLogsAsync({ assetInstanceId: assetInstanceId }));
    }, 30000);
  };

  useEffect(() => {
    createInterval();
    return () => clearInterval(interval);
  }, [assetInstanceId]);

  const dispatch = useAppDispatch();

  const handleCloseClick = () => {
    clearInterval(interval);
    dispatch(hideLogs());
  };

  const handleRefreshClick = async () => {
    await dispatch(fetchLogsAsync({ assetInstanceId: assetInstanceId }));
    createInterval();
  };

  return (
    <Box
      sx={{
        padding: 4,
      }}
    >
      <Grid container>
        <Grid
          item
          style={{
            display: 'flex',
            justifyContent: 'flex-start',
            alignItems: 'left',
          }}
          xs={10}
        >
          <Typography id="form-heading" data-testid="form-heading" variant="h5">
            Deployment Logs for Asset: {assetDetails.name}
          </Typography>
        </Grid>

        <Grid
          item
          style={{
            display: 'flex',
            justifyContent: 'end',
            alignItems: 'right',
          }}
          xs={2}
        >
          <Button
            variant="outlined"
            onClick={handleRefreshClick}
            endIcon={<RefreshIcon />}
            style={{ marginRight: 6 }}
          >
            Refresh
          </Button>
          <Button
            variant="outlined"
            onClick={handleCloseClick}
            endIcon={<CloseIcon />}
          >
            Close
          </Button>
        </Grid>
        <Grid
          item
          style={{
            display: 'flex',
            justifyContent: 'flex-start',
            alignItems: 'left',
          }}
          xs={12}
        >
          <Typography
            id="form-heading"
            data-testid="form-heading"
            variant="subtitle1"
          >
            Created At: {assetDetails.createdAt}
          </Typography>
        </Grid>
        <Grid
          item
          style={{
            display: 'flex',
            justifyContent: 'flex-start',
            alignItems: 'left',
          }}
          xs={12}
        >
          <Typography
            id="form-heading"
            data-testid="form-heading"
            variant="subtitle1"
          >
            Status: {assetDetails.status}
          </Typography>
        </Grid>

        <Grid container spacing={2} padding={1} paddingTop={5}>
          <Grid item xs={12}>
            <Typography
              paragraph={true}
              padding={2}
              style={{
                height: 500,
                overflow: 'scroll',
                whiteSpace: 'pre',
                textAlign: 'left',
                display: 'flex',
                flexDirection: 'column-reverse',
              }}
            >
              {logs}
            </Typography>
          </Grid>
        </Grid>
      </Grid>
    </Box>
  );
}
