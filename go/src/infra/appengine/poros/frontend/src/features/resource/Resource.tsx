// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React from 'react';
import RefreshIcon from '@mui/icons-material/Refresh';
import DeleteIcon from '@mui/icons-material/Delete';
import {
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
} from './resourceSlice';

import { useAppSelector, useAppDispatch } from '../../app/hooks';
import { setRightSideDrawerClose } from '../utility/utilitySlice';

export const Resource = () => {
  const [activeResourceType, setActiveResourceType] = React.useState('machine');
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
  const resourceId: string = useAppSelector(
    (state) => state.resource.record.resourceId
  );
  const resource = useAppSelector((state) => state.resource.record);
  const dispatch = useAppDispatch();

  // Event Handlers
  const handleSaveClick = (
    name: string,
    type: string,
    operatingSystem: string,
    description: string,
    imageProject: string,
    imageFamily: string,
    resourceId: string
  ) => {
    if (resourceId === '') {
      dispatch(
        createResourceAsync({
          name,
          type,
          operatingSystem,
          description,
          imageProject,
          imageFamily,
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

  const handleCancelClick = () => {
    dispatch(setRightSideDrawerClose());
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
              defaultValue="machine"
              value={type}
              onChange={(e) => {
                setActiveResourceType(e.target.value);
                dispatch(setType(e.target.value));
              }}
              fullWidth
              variant="standard"
              placeholder="Type"
            >
              <MenuItem value={'machine'}>Machine</MenuItem>
              {/* <MenuItem value={'domain'}>Domain</MenuItem> */}
            </Select>
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
              <MenuItem value={'chromeos_machine'}>chromeos_machine</MenuItem>
            </Select>
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
            rows={4}
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
          />
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
          <Typography id="form-heading" data-testid="form-heading" variant="h5">
            Resource
          </Typography>
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
            inputProps={{ 'data-testid': 'description' }}
            value={description}
            fullWidth
          />
        </Grid>
      </Grid>

      {renderTypeDropdown()}

      {activeResourceType == 'machine' ? renderOperatingSystemDropdown() : null}

      {activeResourceType == 'machine' ? renderImageProjectInput() : null}

      {activeResourceType == 'machine' ? renderImageFamilyInput() : null}

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
              startIcon={<RefreshIcon />}
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
                  resourceId
                )
              }
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
