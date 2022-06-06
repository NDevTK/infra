// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React from 'react';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';

import RefreshIcon from '@mui/icons-material/Refresh';
import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import {
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Typography,
} from '@mui/material';
import Grid from '@mui/material/Grid';
import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import {
  addMachine,
  clearSelectedRecord,
  createAssetAsync,
  removeMachine,
  setAlias,
  setDescription,
  setName,
  setResourceId,
  updateAssetAsync,
} from './assetSlice';
import { useAppSelector, useAppDispatch } from '../../app/hooks';
import { AssetResourceModel } from '../../api/asset_resource_service';

export const Asset = () => {
  const name: string = useAppSelector((state) => state.asset.record.name);
  const description: string = useAppSelector(
    (state) => state.asset.record.description
  );
  const assetId: string = useAppSelector((state) => state.asset.record.assetId);
  const asset = useAppSelector((state) => state.asset.record);
  const assetResources: AssetResourceModel[] = useAppSelector(
    (state) => state.asset.assetResources
  );
  const dispatch = useAppDispatch();

  const handleSaveClick = (
    name: string,
    description: string,
    assetId: string
  ) => {
    if (assetId === '') {
      dispatch(createAssetAsync({ name, description, assetResources }));
    } else {
      dispatch(
        updateAssetAsync({ asset, updateMask: ['name', 'description'] })
      );
    }
  };

  const handleCancelClick = () => {
    dispatch(clearSelectedRecord());
  };

  const renderMenuItem = (name: string, resourceId: string) => {
    return <MenuItem value={resourceId}> {name} </MenuItem>;
  };

  const renderRow = (index: number, aliasName: string) => {
    return (
      <Grid container spacing={2} padding={1}>
        <Grid
          item
          xs={5}
          style={{
            display: 'flex',
            justifyContent: 'flex-start',
            alignItems: 'center',
          }}
        >
          <FormControl variant="standard" fullWidth>
            <InputLabel>Image</InputLabel>
            <Select
              id={'image-' + index}
              onChange={(e) =>
                dispatch(setResourceId({ id: index, value: e.target.value }))
              }
              variant="outlined"
              placeholder="Type"
            >
              <MenuItem value={'dummy1'}> Dummy1 </MenuItem>
              <MenuItem value={'dummy2'}> Dummy2 </MenuItem>
            </Select>
          </FormControl>
        </Grid>
        <Grid
          item
          xs={5}
          style={{
            display: 'flex',
            justifyContent: 'flex-start',
            alignItems: 'center',
          }}
        >
          <TextField
            label="AliasName"
            id={'alias-' + index}
            value={aliasName}
            onChange={(e) =>
              dispatch(setAlias({ id: index, value: e.target.value }))
            }
            variant="outlined"
          />
        </Grid>
        <Grid
          item
          xs={1}
          style={{
            display: 'flex',
            justifyContent: 'flex-start',
            alignItems: 'center',
          }}
        >
          <Button
            variant="outlined"
            onClick={() => {
              dispatch(removeMachine(index));
            }}
            endIcon={<DeleteIcon />}
          ></Button>
        </Grid>
      </Grid>
    );
  };

  return (
    <Box
      sx={{
        width: 720,
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
          <Typography variant="h5">Asset</Typography>
        </Grid>
      </Grid>
      <Grid container spacing={2} padding={1}>
        <Grid item xs={12}>
          <TextField
            label="Name"
            id="name"
            value={name}
            onChange={(e) => dispatch(setName(e.target.value))}
            fullWidth
            InputProps={{ fullWidth: true }}
            variant="standard"
          />
        </Grid>
      </Grid>
      <Grid container spacing={2} padding={1}>
        <Grid item xs={12}>
          <TextField
            id="description"
            label="Description"
            multiline
            rows={4}
            variant="standard"
            onChange={(e) => dispatch(setDescription(e.target.value))}
            value={description}
            fullWidth
            InputProps={{ fullWidth: true }}
          />
        </Grid>
      </Grid>
      <Grid container spacing={2} padding={1}>
        <Grid item xs={12}>
          <TextField
            disabled
            label="Id"
            id="assetid"
            variant="standard"
            value={assetId}
            fullWidth
            InputProps={{ fullWidth: true }}
          />
        </Grid>
      </Grid>
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
          <Typography variant="h6">Associated Machines</Typography>
        </Grid>
      </Grid>

      {assetResources.map((entity, index) =>
        renderRow(index, entity.aliasName)
      )}
      <Button
        variant="outlined"
        onClick={() => dispatch(addMachine())}
        startIcon={<AddIcon />}
      >
        Add Machine
      </Button>

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
              startIcon={<RefreshIcon />}
            >
              Cancel
            </Button>
            <Button
              variant="contained"
              onClick={() => handleSaveClick(name, description, assetId)}
              endIcon={<DeleteIcon />}
            >
              Save
            </Button>
          </Stack>
        </Grid>
      </Grid>
    </Box>
  );
};
