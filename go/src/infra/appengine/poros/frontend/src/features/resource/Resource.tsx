// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React from 'react';
import CancelIcon from '@mui/icons-material/Cancel';
import SaveIcon from '@mui/icons-material/Save';
import DeleteIcon from '@mui/icons-material/Delete';
import {
  Dialog,
  DialogTitle,
  Select,
  TextField,
  Grid,
  Box,
  Typography,
  MenuItem,
  Button,
  Stack,
  InputLabel,
  FormControl,
  FormHelperText,
} from '@mui/material';
import {
  createResourceAsync,
  setName,
  setType,
  setOperatingSystem,
  setDescription,
  updateResourceAsync,
  setImageProject,
  setImageFamily,
  setImageSource,
  ResourceRecordValidation,
  setOperatingSystemValidFalse,
  setDescriptionValidFalse,
  setNameValidFalse,
  setImageProjectValidFalse,
  setImageFamilyValidFalse,
  setImageSourceValidFalse,
  setTypeValidFalse,
  deleteResourceAsync,
} from './resourceSlice';

import { useAppSelector, useAppDispatch } from '../../app/hooks';
import {
  setRightSideDrawerClose,
  setDeleteResourceDialogClose,
  setDeleteResourceDialogOpen,
} from '../utility/utilitySlice';

export const Resource = () => {
  const [activeResourceType, setActiveResourceType] = React.useState('ad_joined_machine');
  const name: string = useAppSelector((state) => state.resource.record.name);
  const type: string = useAppSelector((state) => state.resource.record.type);
  const operatingSystem: string = useAppSelector(
    (state) => state.resource.record.operatingSystem
  );
  const description: string = useAppSelector(
    (state) => state.resource.record.description
  );
  const imageProject: string = useAppSelector(
    (state) => state.resource.record.imageProject
  );
  const imageFamily: string = useAppSelector(
    (state) => state.resource.record.imageFamily
  );
  const imageSource: string = useAppSelector(
    (state) => state.resource.record.imageSource
  );
  const resourceId: string = useAppSelector(
    (state) => state.resource.record.resourceId
  );
  const recordValidation: ResourceRecordValidation = useAppSelector(
    (state) => state.resource.recordValidation
  );
  const resource = useAppSelector((state) => state.resource.record);
  const deleteResourceDialogOpen: boolean = useAppSelector(
    (state) => state.utility.deleteResourceDialogOpen
  );
  const dispatch = useAppDispatch();

  // Event Handlers
  const handleSaveClick = (
    name: string,
    type: string,
    operatingSystem: string,
    description: string,
    imageProject: string,
    imageFamily: string,
    imageSource: string,
    resourceId: string
  ) => {
    if (!validateInput()) {
      return;
    }
    if (resourceId === '') {
      dispatch(
        createResourceAsync({
          name,
          type,
          operatingSystem,
          description,
          imageProject,
          imageFamily,
          imageSource,
        })
      );
    } else {
      dispatch(
        updateResourceAsync({
          resource,
          updateMask: [
            'name',
            'description',
            'type',
            'operating_system',
            'image_project',
            'image_family',
          ],
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
    if ((type === 'ad_joined_machine' || type === 'machine') && operatingSystem === '') {
      dispatch(setOperatingSystemValidFalse());
      valid = false;
    }
    if ((type === 'ad_joined_machine' || type === 'machine') && imageProject === '') {
      dispatch(setImageProjectValidFalse());
      valid = false;
    }
    if ((type === 'ad_joined_machine' || type === 'machine') && imageFamily === '') {
      dispatch(setImageFamilyValidFalse());
      valid = false;
    }
    if (type == 'custom_image_machine' && imageSource === '') {
      dispatch(setImageSourceValidFalse());
      valid = false;
    }
    if (type === '') {
      dispatch(setTypeValidFalse());
      valid = false;
    }
    if (!recordValidation.nameUnique) {
      valid = false;
    }

    return valid;
  };

  const handleCancelClick = () => {
    dispatch(setRightSideDrawerClose());
  };

  const handleDeleteResourceConfirm = () => {
    dispatch(deleteResourceAsync(resourceId));
    dispatch(setDeleteResourceDialogClose());
  };

  const handleDeleteResourceClick = () => {
    if (resourceId !== '') {
      dispatch(setDeleteResourceDialogOpen());
    }
  };

  // Render functions
  const renderTypeDropdown = () => {
    return (
      <Grid container spacing={2} padding={1} paddingTop={3}>
        <Grid item xs={12}>
          <FormControl variant="standard" fullWidth>
            <InputLabel>Type</InputLabel>
            <Select
              label="Type"
              id="type"
              inputProps={{ 'data-testid': 'type' }}
              defaultValue="ad_joined_machine"
              value={type}
              onChange={(e) => {
                setActiveResourceType(e.target.value);
                dispatch(setType(e.target.value));
              }}
              fullWidth
              variant="standard"
              placeholder="Type"
            >
              <MenuItem value={'ad_joined_machine'}>AD Joined Machine</MenuItem>
              <MenuItem value={'machine'}>Machine</MenuItem>
              <MenuItem value={'custom_image_machine'}>Custom Image Machine</MenuItem>
              {/* <MenuItem value={'domain'}>Domain</MenuItem> */}
            </Select>
            {!recordValidation.typeValid && (
              <FormHelperText style={{ color: 'red' }}>
                {' '}
                Resource type is required
              </FormHelperText>
            )}
          </FormControl>
        </Grid>
      </Grid>
    );
  };

  const renderOperatingSystemDropdown = () => {
    return (
      <Grid container spacing={2} padding={1} paddingTop={6}>
        <Grid item xs={12}>
          <FormControl variant="standard" fullWidth>
            <InputLabel>Operating System</InputLabel>
            <Select
              label="OperatingSystem"
              id="operating-system"
              inputProps={{ 'data-testid': 'operating-system' }}
              defaultValue="windows_machine"
              value={operatingSystem}
              onChange={(e) => {
                dispatch(setOperatingSystem(e.target.value));
              }}
              fullWidth
              variant="standard"
              placeholder="Type"
            >
              <MenuItem
                id="os-option"
                data-testid="os-option"
                value={'windows_machine'}
              >
                windows_machine
              </MenuItem>
              <MenuItem value={'linux_machine'}>linux_machine</MenuItem>
            </Select>
            {!recordValidation.operatingSystemValid && (
              <FormHelperText style={{ color: 'red' }}>
                {' '}
                Operating system is required
              </FormHelperText>
            )}
          </FormControl>
        </Grid>
      </Grid>
    );
  };

  const renderDomainMetaInput = () => {
    return (
      <Grid container spacing={2} padding={1} paddingTop={6}>
        <Grid item xs={12}>
          <TextField
            id="domain-info"
            inputProps={{ 'data-testid': 'domain-info' }}
            label="Domain Information"
            multiline
            rows={2}
            variant="standard"
            // TODO: when we allow user to select type domain, edit the onchange and value
            onChange={(e) => {
              // dispatch(setDomainInfo(e.target.value))
            }}
            value={'some domain info'}
            fullWidth
          />
        </Grid>
      </Grid>
    );
  };

  const renderImageProjectInput = () => {
    return (
      <Grid container spacing={2} padding={1}>
        <Grid item xs={12}>
          <TextField
            label="Image Project"
            id="image-project"
            value={imageProject}
            onChange={(e) => dispatch(setImageProject(e.target.value))}
            fullWidth
            inputProps={{ 'data-testid': 'image-project' }}
            variant="standard"
            helperText={
              !recordValidation.imageProjectValid
                ? 'Image project is required'
                : ''
            }
            FormHelperTextProps={{ style: { color: 'red' } }}
          />
        </Grid>
      </Grid>
    );
  };

  const renderImageFamilyInput = () => {
    return (
      <Grid container spacing={2} padding={1}>
        <Grid item xs={12}>
          <TextField
            label="Image Family"
            id="image-family"
            value={imageFamily}
            onChange={(e) => dispatch(setImageFamily(e.target.value))}
            fullWidth
            inputProps={{ 'data-testid': 'image-family' }}
            variant="standard"
            helperText={
              !recordValidation.imageFamilyValid
                ? 'Image family is required'
                : ''
            }
            FormHelperTextProps={{ style: { color: 'red' } }}
          />
        </Grid>
      </Grid>
    );
  };

  const renderImageSourceInput = () => {
    return (
      <Grid container spacing={2} padding={1}>
        <Grid item xs={12}>
          <TextField
            label="Image Source"
            id="image-source"
            value={imageSource}
            onChange={(e) => dispatch(setImageSource(e.target.value))}
            fullWidth
            inputProps={{ 'data-testid': 'image-source' }}
            variant="standard"
            helperText={
              !recordValidation.imageSourceValid
                ? 'Source Image path is required'
                : ''
            }
            FormHelperTextProps={{ style: { color: 'red' } }}
          />
        </Grid>
      </Grid>
    );
  };

  return (
    <div>
      <Dialog
        onClose={() => dispatch(setDeleteResourceDialogClose())}
        open={deleteResourceDialogOpen}
      >
        <DialogTitle>Do you want to delete this resource?</DialogTitle>
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
            onClick={() => dispatch(setDeleteResourceDialogClose())}
          >
            Cancel
          </Button>
          <Button
            variant="contained"
            size="small"
            onClick={handleDeleteResourceConfirm}
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
            <Typography
              id="form-heading"
              data-testid="form-heading"
              variant="h5"
            >
              Resource
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
              onClick={handleDeleteResourceClick}
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
              inputProps={{ 'data-testid': 'name' }}
              value={name}
              onChange={(e) => dispatch(setName(e.target.value))}
              fullWidth
              variant="standard"
              helperText={
                !recordValidation.nameValid
                  ? 'Resource name is required'
                  : !recordValidation.nameUnique
                  ? 'Resource name must be unique'
                  : ''
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
              inputProps={{ 'data-testid': 'description' }}
              value={description}
              fullWidth
              helperText={
                !recordValidation.descriptionValid
                  ? 'Resource description is required'
                  : ''
              }
              FormHelperTextProps={{ style: { color: 'red' } }}
            />
          </Grid>
        </Grid>

        {renderTypeDropdown()}

        {(activeResourceType == 'ad_joined_machine' || activeResourceType == 'machine' || activeResourceType == 'custom_image_machine')
          ? renderOperatingSystemDropdown()
          : null}

        {(activeResourceType == 'ad_joined_machine' || activeResourceType == 'machine') ? renderImageProjectInput() : null}

        {(activeResourceType == 'ad_joined_machine' || activeResourceType == 'machine') ? renderImageFamilyInput() : null}

        {(activeResourceType == 'custom_image_machine') ? renderImageSourceInput() : null}

        <Grid container spacing={2} padding={1} paddingTop={6}>
          <Grid item xs={12}>
            <TextField
              disabled
              label="Id"
              id="resource-id"
              variant="standard"
              inputProps={{ 'data-testid': 'resource-id' }}
              value={resourceId}
              fullWidth
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
                id="cancel-button"
                data-testid="cancel-button"
                variant="outlined"
                onClick={handleCancelClick}
                endIcon={<CancelIcon />}
              >
                Cancel
              </Button>
              <Button
                id="save-button"
                data-testid="save-button"
                variant="contained"
                onClick={() =>
                  handleSaveClick(
                    name,
                    type,
                    operatingSystem,
                    description,
                    imageProject,
                    imageFamily,
                    imageSource,
                    resourceId
                  )
                }
                endIcon={<SaveIcon />}
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
