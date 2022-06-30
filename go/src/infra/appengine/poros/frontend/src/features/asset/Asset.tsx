// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React from 'react';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';

import CancelIcon from '@mui/icons-material/Cancel';
import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import SaveIcon from '@mui/icons-material/Save';
import {
  FormControl,
  IconButton,
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
  createAssetAsync,
  removeMachine,
  setAlias,
  setAssetType,
  setDescription,
  setName,
  setResourceId,
  updateAssetAsync,
} from './assetSlice';
import { useAppSelector, useAppDispatch } from '../../app/hooks';
import { AssetResourceModel } from '../../api/asset_resource_service';
import { ResourceModel } from '../../api/resource_service';
import { queryResourceAsync } from '../resource/resourceSlice';
import { setRightSideDrawerClose } from '../utility/utilitySlice';

export const Asset = () => {
  const name: string = useAppSelector((state) => state.asset.record.name);
  const description: string = useAppSelector(
    (state) => state.asset.record.description
  );
  const assetType: string = useAppSelector(
    (state) => state.asset.record.assetType
  );
  const assetId: string = useAppSelector((state) => state.asset.record.assetId);
  const asset = useAppSelector((state) => state.asset.record);
  const assetResourcesToSave: AssetResourceModel[] = useAppSelector(
    (state) => state.asset.assetResourcesToSave
  );
  const assetResourcesToDelete: AssetResourceModel[] = useAppSelector(
    (state) => state.asset.assetResourcesToDelete
  );
  const resources: ResourceModel[] = useAppSelector(
    (state) => state.asset.resources
  );
  const dispatch = useAppDispatch();
  React.useEffect(() => {
    dispatch(queryResourceAsync({ pageSize: 100, pageToken: '' }));
  }, []);

  const handleSaveClick = (
    name: string,
    description: string,
    assetType: string,
    assetId: string,
    assetResourcesToSave: AssetResourceModel[],
    assetResourcesToDelete: AssetResourceModel[]
  ) => {
    if (assetId === '') {
      dispatch(
        createAssetAsync({
          name,
          description,
          assetType,
          assetResourcesToSave,
        })
      );
    } else {
      dispatch(
        updateAssetAsync({
          asset,
          assetUpdateMask: ['name', 'description', 'asset_type'],
          assetResourceUpdateMask: ['resource_id', 'alias_name'],
          assetResourcesToSave,
          assetResourcesToDelete,
        })
      );
    }
  };

  const handleCancelClick = () => {
    dispatch(setRightSideDrawerClose());
  };

  const renderAssetTypeDropdown = () => {
    return (
      <Grid container spacing={2} padding={1} paddingTop={3}>
        <Grid item xs={12}>
          <FormControl variant="standard" fullWidth>
            <InputLabel>Type</InputLabel>
            <Select
              label="AssetType"
              id="assettype"
              value={assetType}
              onChange={(e) => {
                setAssetType(e.target.value);
                dispatch(setAssetType(e.target.value));
              }}
              fullWidth
              inputProps={{ fullWidth: true }}
              variant="standard"
              placeholder="Type"
            >
              <MenuItem value={'active_directory'}>Active Directory</MenuItem>
            </Select>
          </FormControl>
        </Grid>
      </Grid>
    );
  };

  const renderMenuItem = (name: string, resourceId: string) => {
    return <MenuItem value={resourceId}> {name} </MenuItem>;
  };

  const renderRow = (index: number, aliasName: string, resourceId: string) => {
    return (
      <Grid
        container
        spacing={2}
        padding={1}
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
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
            <InputLabel>Resource</InputLabel>
            <Select
              id={'image-' + index}
              onChange={(e) =>
                dispatch(setResourceId({ id: index, value: e.target.value }))
              }
              value={resourceId}
              variant="standard"
              placeholder="Type"
            >
              {resources.map((resource) =>
                renderMenuItem(resource.name, resource.resourceId)
              )}
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
            variant="standard"
            fullWidth
          />
        </Grid>

        <Grid
          item
          xs={1}
          style={{
            display: 'bottom',
            justifyContent: 'flex-end',
            alignItems: 'bottom',
          }}
        >
          <IconButton
            aria-label="add"
            size="small"
            onClick={() => {
              dispatch(addMachine());
            }}
          >
            <AddIcon fontSize="inherit" />
          </IconButton>
        </Grid>
        <Grid
          item
          xs={1}
          style={{
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center',
          }}
        >
          <IconButton
            aria-label="delete"
            size="small"
            onClick={() => {
              dispatch(removeMachine(index));
            }}
          >
            <DeleteIcon fontSize="inherit"> Delete </DeleteIcon>
          </IconButton>
        </Grid>
      </Grid>
    );
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
      {renderAssetTypeDropdown()}
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
          <Typography variant="inherit">Associated Machines</Typography>
        </Grid>
      </Grid>

      {assetResourcesToSave.map((entity, index) =>
        renderRow(index, entity.aliasName, entity.resourceId)
      )}

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
            >
              Cancel
            </Button>
            <Button
              variant="contained"
              onClick={() => {
                handleSaveClick(
                  name,
                  description,
                  assetType,
                  assetId,
                  assetResourcesToSave,
                  assetResourcesToDelete
                );
              }}
              endIcon={<SaveIcon />}
            >
              Save
            </Button>
          </Stack>
        </Grid>
      </Grid>
    </Box>
  );
};
