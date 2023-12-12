// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {
  Box,
  Divider,
  FormControl,
  IconButton,
  InputAdornment,
  InputLabel,
  MenuItem,
  Select,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Toolbar,
} from '@mui/material';
import { useContext, useState } from 'react';
import AddCircleOutlineIcon from '@mui/icons-material/AddCircleOutline';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';
import { Platform } from '../../../api/coverage';
import { TrendsContext } from './TrendsContext';

const PRESET_LIST = ['Blink'];

function TrendsToolbar() {
  const { api, params, isConfigLoaded, isAbsTrend } = useContext(TrendsContext);
  const [path, setPath] = useState('');

  const handlePlatformChange = (event) => {
    api.updatePlatform(event.target.value);
  };

  const handleUnitTestsOnlyToggle = (_, isUnitTestsOnly: boolean) => {
    api.updateUnitTestsOnly(isUnitTestsOnly);
  };

  const handleAddPath = () => {
    api.updatePaths([path, ...params.paths]);
    setPath('');
  };

  const handleRemovePath = (remPath) => {
    const remainingPaths = params.paths.filter((p) => p !== remPath);
    api.updatePaths([...remainingPaths]);
  };

  const handlePathChange = (event) => {
    setPath(event.target.value);
  };

  const handlePresetSelect = (event) => {
    api.updatePresets([event.target.innerText, ...params.presets]);
  };

  const handlePresetUnselect = (event) => {
    const remainingPresets = params.presets.filter((p) => p !== event.target.innerText);
    api.updatePresets([...remainingPresets]);
  };

  const loadTrends = () => {
    isAbsTrend ? api.loadAbsTrends(): api.loadIncTrends();
  };

  const renderToolbar = () => {
    return (
      <>
        <Toolbar sx={{ mt: 2, mb: 2 }}>
          <Box sx={{ flexDirection: 'row', flexGrow: 1 }}>
            <Box sx={{ mb: 2, display: 'flex' }}>
              {
                isAbsTrend?
                <FormControl data-testid="platformTest" variant="standard" sx={{ mr: 3, flexGrow: 2 }}>
                  <InputLabel shrink={true}>Platform</InputLabel>
                  <Select
                    value={params.platform}
                    label="Platform"
                    onChange={handlePlatformChange}
                  >
                    {
                      params.platformList.map((platform: Platform) => {
                        return (<MenuItem key={platform.uiName} value={platform.platform}>{platform.uiName}</MenuItem>);
                      })
                    }
                  </Select>
                </FormControl>:
                null
              }

              <FormControl sx={{ mt: 1, alignSelf: 'center', mr: 3 }}>
                <ToggleButtonGroup
                  size="small"
                  color="primary"
                  value={params.unitTestsOnly}
                  exclusive
                  onChange={handleUnitTestsOnlyToggle}
                  data-testid="unitTestsOnlyToggleTest"
                  sx={{ paddingTop: 0.5, flex: 1 }}
                >
                  <ToggleButton value={true}>
                    Unit Tests
                  </ToggleButton>
                  <ToggleButton value={false}>
                    All Tests
                  </ToggleButton>
                </ToggleButtonGroup>
              </FormControl>

              <FormControl sx={{ mr: 3, flex: 3 }}>
                <TextField
                  id="path-input"
                  data-testid="pathTest"
                  label="Path"
                  variant="standard"
                  onChange={handlePathChange}
                  value={path}
                  sx={{ flexGrow: 1 }}
                  InputProps={{
                    endAdornment:
                      <InputAdornment position="start">
                        <IconButton
                          aria-label="add path"
                          onClick={handleAddPath}
                          edge="end"
                        >
                          <AddCircleOutlineIcon />
                        </IconButton>
                      </InputAdornment>,
                  }}
                />
              </FormControl>
            </Box>

            <Box sx={{ mb: 2 }}>
              {
                PRESET_LIST.map((preset, i) => {
                  return (
                    params.presets.includes(preset) ?
                      <Chip key={`chip-${i}`} onClick={handlePresetUnselect} label={preset} color="primary" /> :
                      <Chip key={`chip-${i}`} onClick={handlePresetSelect} label={preset} color="primary" variant="outlined" />
                  );
                })
              }
            </Box>

            <Box sx={{ mb: 2 }}>
              {
                params.paths.map((path, i) => {
                  return (
                    <Chip sx={{ mr: 1, mb: 1 }} key={i} label={path} variant="outlined" onDelete={() => handleRemovePath(path)} />
                  );
                })
              }
            </Box>

            <Box sx={{ display: 'flex' }}>
              <Button
                onClick={loadTrends}
                sx={{ margin: 'auto' }}
                variant="contained"
              >
                Load Trend
              </Button>
            </Box>
          </Box>
        </Toolbar>
        <Divider />
      </>
    );
  };

  return (
    isConfigLoaded ? renderToolbar() : <></>
  );
}

export default TrendsToolbar;
