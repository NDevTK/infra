// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React from 'react';
import CancelIcon from '@mui/icons-material/Cancel';
import SaveIcon from '@mui/icons-material/Save';
import { TextField, Grid, Box, Typography, Button, Stack } from '@mui/material';

import { useAppSelector, useAppDispatch } from '../../app/hooks';
import { setRightSideDrawerClose } from '../utility/utilitySlice';
import { setDeleteTime, updateAssetInstanceAsync } from './assetInstanceSlice';

export const AssetInstance = () => {
  const assetInstanceId: string = useAppSelector(
    (state) => state.assetInstance.record.assetInstanceId
  );
  const assetId: string = useAppSelector(
    (state) => state.assetInstance.record.assetId
  );
  const status: string = useAppSelector(
    (state) => state.assetInstance.record.status
  );
  const deleteAtBuffer: string = useAppSelector(
    (state) => state.assetInstance.deleteAtBuffer
  );
  const assetInstance = useAppSelector((state) => state.assetInstance.record);
  const dispatch = useAppDispatch();

  // Event Handlers
  const handleSaveClick = (id: string) => {
    dispatch(
      updateAssetInstanceAsync({
        assetInstance,
        updateMask: ['delete_at'],
      })
    );
  };

  const handleCancelClick = () => {
    dispatch(setRightSideDrawerClose());
  };

  return (
    <Box
      sx={{
        width: 465,
        maxWidth: '100%',
        padding: 1,
      }}
    >
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
          <Typography id="form-heading" data-testid="form-heading" variant="h5">
            Lab Instance
          </Typography>
        </Grid>
      </Grid>

      <Grid container spacing={2} padding={1} paddingTop={1}>
        <Grid item xs={12}>
          <TextField
            disabled
            label="Id"
            id="assetInstanceId"
            variant="standard"
            value={assetInstanceId}
            fullWidth
            inputProps={{ 'data-testid': 'asset-instance-id' }}
          />
        </Grid>
      </Grid>

      <Grid container spacing={2} padding={1} paddingTop={1}>
        <Grid item xs={12}>
          <TextField
            disabled
            label="Associated AssetId"
            id="assetId"
            variant="standard"
            value={assetId}
            fullWidth
            inputProps={{ 'data-testid': 'asset-id' }}
          />
        </Grid>
      </Grid>

      <Grid container spacing={2} padding={1} paddingTop={1}>
        <Grid item xs={12}>
          <TextField
            disabled
            label="Status"
            id="status"
            variant="standard"
            value={status}
            fullWidth
            inputProps={{ 'data-testid': 'status' }}
          />
        </Grid>
      </Grid>

      <Grid container spacing={2} padding={1} paddingTop={1}>
        <Grid item xs={12}>
          <TextField
            id="deleteTime"
            label="Change the delete time for the machine"
            type="datetime-local"
            value={deleteAtBuffer}
            variant="standard"
            onChange={(e) => dispatch(setDeleteTime(e.target.value))}
            fullWidth
            InputLabelProps={{
              shrink: true,
            }}
            inputProps={{ 'data-testid': 'delete-time' }}
          />
        </Grid>
      </Grid>

      <Grid container spacing={2} padding={1}>
        <Grid
          item
          style={{
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'right',
          }}
          xs={12}
        >
          <Stack direction="row" spacing={2}>
            <Button
              variant="outlined"
              onClick={handleCancelClick}
              endIcon={<CancelIcon />}
              id="cancel-button"
              data-testid="cancel-button"
            >
              Cancel
            </Button>
            <Button
              variant="contained"
              onClick={() => handleSaveClick(assetInstanceId)}
              endIcon={<SaveIcon />}
              id="save-button"
              data-testid="save-button"
            >
              Save
            </Button>
          </Stack>
        </Grid>
      </Grid>
    </Box>
  );
};
