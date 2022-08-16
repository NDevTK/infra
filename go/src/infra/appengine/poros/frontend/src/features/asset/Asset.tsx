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
  Collapse,
  Dialog,
  DialogTitle,
  Divider,
  FormControl,
  FormControlLabel,
  FormHelperText,
  IconButton,
  InputLabel,
  MenuItem,
  Select,
  Switch,
  Typography,
} from '@mui/material';
import Grid from '@mui/material/Grid';
import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import {
  addMachine,
  clearSelectedRecord,
  createAssetAsync,
  getDefaultResources,
  AssetRecordValidation,
  removeMachine,
  setAlias,
  setAssetType,
  setDescription,
  setName,
  setResourceId,
  updateAssetAsync,
  setNameValidFalse,
  setDescriptionValidFalse,
  setResourceIdValidFalse,
  setAliasNameValidFalse,
  deleteAssetAsync,
  changeShowDefaultMachines,
} from './assetSlice';
import { useAppSelector, useAppDispatch } from '../../app/hooks';
import { AssetResourceModel } from '../../api/asset_resource_service';
import { ResourceModel } from '../../api/resource_service';
import { queryResourceAsync } from '../resource/resourceSlice';
import {
  setRightSideDrawerClose,
  setDeleteAssetDialogClose,
  setDeleteAssetDialogOpen,
} from '../utility/utilitySlice';

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
  const defaultAssetResources: AssetResourceModel[] = useAppSelector(
    (state) => state.asset.defaultAssetResources
  );
  const recordValidation: AssetRecordValidation = useAppSelector(
    (state) => state.asset.recordValidation
  );
  const deleteAssetDialogOpen: boolean = useAppSelector(
    (state) => state.utility.deleteAssetDialogOpen
  );
  const showDefaultMachines: boolean = useAppSelector(
    (state) => state.asset.showDefaultMachines
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
    assetResourcesToDelete: AssetResourceModel[],
    defaultAssetResources: AssetResourceModel[]
  ) => {
    if (!validateInput()) {
      return;
    }
    assetResourcesToSave = [...assetResourcesToSave, ...defaultAssetResources];
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

  const validateInput = () => {
    let valid = true;
    if (name === '') {
      dispatch(setNameValidFalse());
      valid = false;
    }
    if (description === '') {
      dispatch(setDescriptionValidFalse());
      valid = false;
    }
    assetResourcesToSave.forEach((assetResource, index) => {
      if (assetResource.resourceId === '') {
        dispatch(setResourceIdValidFalse({ index: index }));
        valid = false;
      }
      if (assetResource.aliasName === '') {
        dispatch(setAliasNameValidFalse({ index: index }));
        valid = false;
      }
    });
    if (recordValidation.aliasNameUnique.some((unique) => unique === false)) {
      valid = false;
    }

    return valid;
  };

  const handleCancelClick = () => {
    dispatch(clearSelectedRecord());
    dispatch(setRightSideDrawerClose());
  };

  const handleDeleteAssetConfirm = () => {
    dispatch(deleteAssetAsync(assetId));
    dispatch(setDeleteAssetDialogClose());
  };

  const handleDeleteAssetClick = () => {
    if (assetId !== '') {
      dispatch(setDeleteAssetDialogOpen());
    }
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
              defaultValue="active_directory"
              value={assetType}
              onChange={(e) => {
                dispatch(getDefaultResources(e.target.value));
                dispatch(setAssetType(e.target.value));
              }}
              fullWidth
              variant="standard"
              placeholder="Type"
              inputProps={{ 'data-testid': 'type' }}
            >
              <MenuItem value={'active_directory'}>Active Directory</MenuItem>
              <MenuItem value={'active_directory_splunk'}>
                Active Directory with Splunk
              </MenuItem>
            </Select>
          </FormControl>
        </Grid>
      </Grid>
    );
  };

  const renderMenuItem = (name: string, resourceId: string) => {
    return (
      <MenuItem
        value={resourceId}
        data-testid={'resource-option-' + resourceId}
        key={'resource-option-' + resourceId}
      >
        {name}
      </MenuItem>
    );
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
        key={'row-' + index}
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
              id={'resource-' + index}
              key={'resource-' + index}
              onChange={(e) =>
                dispatch(setResourceId({ id: index, value: e.target.value }))
              }
              value={resourceId}
              variant="standard"
              placeholder="Type"
              inputProps={{ 'data-testid': 'resource-' + index }}
            >
              {resources.map((resource) =>
                renderMenuItem(resource.name, resource.resourceId)
              )}
            </Select>
            {!recordValidation.resourceIdValid[index] && (
              <FormHelperText style={{ color: 'red' }}>
                {' '}
                Resource is required
              </FormHelperText>
            )}
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
            inputProps={{ 'data-testid': 'alias-' + index }}
            helperText={
              !recordValidation.aliasNameValid[index]
                ? 'Alias name is required'
                : !recordValidation.aliasNameUnique[index]
                ? 'Alias name must be unique'
                : ''
            }
            FormHelperTextProps={{ style: { color: 'red' } }}
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
            data-testid={'add-button-' + index}
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
            data-testid={'delete-button-' + index}
          >
            <DeleteIcon fontSize="inherit"> Delete </DeleteIcon>
          </IconButton>
        </Grid>
      </Grid>
    );
  };

  const renderDefaultMachines = (aliasName: string, resourceId: string) => {
    return (
      <Grid
        container
        spacing={2}
        padding={1}
        style={{
          display: 'flex',
          justifyContent: 'flex-start',
          alignItems: 'left',
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
              value={resourceId}
              variant="standard"
              placeholder="Type"
              disabled
              hidden
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
            value={aliasName}
            variant="standard"
            fullWidth
            disabled
          />
        </Grid>
      </Grid>
    );
  };

  return (
    <div>
      <Dialog
        onClose={() => dispatch(setDeleteAssetDialogClose())}
        open={deleteAssetDialogOpen}
      >
        <DialogTitle>Do you want to delete this template?</DialogTitle>
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
          <Button
            variant="outlined"
            size="small"
            onClick={() => dispatch(setDeleteAssetDialogClose())}
          >
            Cancel
          </Button>
          <Button
            variant="contained"
            size="small"
            onClick={handleDeleteAssetConfirm}
          >
            Confirm
          </Button>
        </Stack>
      </Dialog>
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
              alignItems: 'left',
            }}
            xs={8}
          >
            <Typography variant="h5" data-testid="form-heading">
              Lab Template
            </Typography>
          </Grid>
          <Grid
            item
            style={{
              display: 'flex',
              justifyContent: 'end',
              alignItems: 'right',
            }}
            xs={4}
          >
            <Button
              variant="outlined"
              onClick={handleDeleteAssetClick}
              endIcon={<DeleteIcon />}
            >
              Delete
            </Button>
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
              variant="standard"
              inputProps={{ 'data-testid': 'name' }}
              helperText={
                !recordValidation.nameValid ? 'Asset name is required' : ''
              }
              FormHelperTextProps={{ style: { color: 'red' } }}
            />
          </Grid>
        </Grid>
        <Grid container spacing={2} padding={1}>
          <Grid item xs={12}>
            <TextField
              id="description"
              label="Description"
              multiline
              rows={2}
              variant="standard"
              onChange={(e) => dispatch(setDescription(e.target.value))}
              value={description}
              fullWidth
              inputProps={{ 'data-testid': 'description' }}
              helperText={
                !recordValidation.descriptionValid
                  ? 'Asset description is required'
                  : ''
              }
              FormHelperTextProps={{ style: { color: 'red' } }}
            />
          </Grid>
        </Grid>
        {renderAssetTypeDropdown()}
        <Divider sx={{ padding: 1 }} />
        <Grid container spacing={2} padding={1} paddingRight={0}>
          <Grid
            item
            style={{
              display: 'flex',
              justifyContent: 'flex-start',
              alignItems: 'center',
            }}
            xs={6}
          >
            <Typography variant="inherit" data-testid="machines-heading">
              Associated Machines
            </Typography>
          </Grid>

          <Grid
            item
            style={{
              display: 'flex',
              justifyContent: 'flex-end',
              alignItems: 'right',
            }}
            xs={6}
          >
            <FormControlLabel
              control={
                <Switch
                  checked={showDefaultMachines}
                  onChange={(e) => {
                    dispatch(changeShowDefaultMachines(e.target.checked));
                  }}
                />
              }
              label={
                <Typography fontSize={12}>Show default machines</Typography>
              }
              // labelPlacement="bottom"
            />
          </Grid>
        </Grid>

        <Collapse in={showDefaultMachines}>
          {defaultAssetResources.map((entity, index) =>
            renderDefaultMachines(entity.aliasName, entity.resourceId)
          )}
        </Collapse>
        {assetResourcesToSave.map((entity, index) =>
          renderRow(index, entity.aliasName, entity.resourceId)
        )}

        <Grid container spacing={2} padding={1}>
          <Grid item xs={12}>
            <TextField
              disabled
              label="Id"
              id="asset-id"
              variant="standard"
              value={assetId}
              fullWidth
              inputProps={{ 'data-testid': 'asset-id' }}
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
                data-testid="cancel-button"
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
                    assetResourcesToDelete,
                    defaultAssetResources
                  );
                }}
                endIcon={<SaveIcon />}
                data-testid="save-button"
              >
                Save
              </Button>
            </Stack>
          </Grid>
        </Grid>
      </Box>
    </div>
  );
};
