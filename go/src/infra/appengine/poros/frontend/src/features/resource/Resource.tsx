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
  setImage,
  updateResourceAsync,
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
  const image: string = useAppSelector((state) => state.resource.record.image);
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
    image: string,
    resourceId: string
  ) => {
    if (resourceId === '') {
      dispatch(
        createResourceAsync({
          name,
          type,
          operatingSystem,
          description,
          image,
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
            'image',
          ],
        })
      );
    }
  };

  const handleCancelClick = () => {
    dispatch(setRightSideDrawerClose());
  };

  // Render functions

  // This function will be used once we give user the ability to select type of Resource
  const renderTypeDropdown = () => {
    return (
      <Grid container spacing={2} padding={1} paddingTop={3}>
        <Grid item xs={12}>
          <FormControl variant="standard" fullWidth>
            <InputLabel>Type</InputLabel>
            <Select
              label="Type"
              id="type"
              inputProps={{ "data-testid": "type" }}
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
              <MenuItem value={'domain'}>Domain</MenuItem>
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
              inputProps={{ "data-testid": "operating-system" }}
              defaultValue="windows_machine"
              value={operatingSystem}
              onChange={(e) => {
                dispatch(setOperatingSystem(e.target.value));
              }}
              fullWidth
              variant="standard"
              placeholder="Type"
            >
              <MenuItem id="os-option" data-testid="os-option" value={'windows_machine'}>windows_machine</MenuItem>
              <MenuItem value={'linux_machine'}>linux_machine</MenuItem>
              <MenuItem value={'chromeos_machine'}>chromeos_machine</MenuItem>
            </Select>
          </FormControl>
        </Grid>
      </Grid>
    );
  };

  const renderMachineMetaDropdown = () => {
    return (
      <Grid container spacing={2} padding={1} paddingTop={6}>
        <Grid item xs={12}>
          <FormControl variant="standard" fullWidth>
            <InputLabel>VM Images</InputLabel>
            <Select
              id="image"
              inputProps={{ "data-testid": "image" }}
              value={image}
              onChange={(e) => dispatch(setImage(e.target.value))}
              fullWidth
              variant="standard"
              placeholder="Type"
            >
              <MenuItem data-testid="image-option" value={'image-1'}>Image 1</MenuItem>
              <MenuItem value={'image-2'}>Image 2</MenuItem>
              <MenuItem value={'image-3'}>Image 3</MenuItem>
              <MenuItem value={'image-4'}>Image 4</MenuItem>
              <MenuItem value={'image-5'}>Image 5</MenuItem>
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
            inputProps={{ "data-testid": "domain-info" }}
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
          <Typography id="form-heading" data-testid="form-heading" variant="h5">Resource</Typography>
        </Grid>
      </Grid>
      <Grid container spacing={2} padding={1}>
        <Grid item xs={12}>
          <TextField
            label="Name"
            id="name"
            inputProps={{ "data-testid": "name" }}
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
            inputProps={{ "data-testid": "description" }}
            value={description}
            fullWidth
          />
        </Grid>
      </Grid>

      {activeResourceType == 'machine'
        ? renderMachineMetaDropdown()
        : renderDomainMetaInput()}

      {activeResourceType == 'machine' ? renderOperatingSystemDropdown() : null}

      <Grid container spacing={2} padding={1} paddingTop={6}>
        <Grid item xs={12}>
          <TextField
            disabled
            label="Id"
            id="resource-id"
            variant="standard"
            inputProps={{ "data-testid": "resource-id" }}
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
                  image,
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
