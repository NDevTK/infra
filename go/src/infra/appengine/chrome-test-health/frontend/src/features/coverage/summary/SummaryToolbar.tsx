// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {
  Divider,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  ToggleButton,
  ToggleButtonGroup,
  Toolbar,
} from '@mui/material';
import { useContext } from 'react';
import { Platform } from '../../../api/coverage';
import { SummaryContext } from './SummaryContext';

function SummaryToolbar() {
  const { api, params, isConfigLoaded } = useContext(SummaryContext);

  const handlePlatformChange = (event) => {
    api.updatePlatform(event.target.value);
  };

  const handleUnitTestsOnlyToggle = (_, isUnitTestsOnly: boolean) => {
    api.updateUnitTestsOnly(isUnitTestsOnly);
  };

  const renderToolbar = () => {
    return (
      <>
        <Toolbar sx={{ mt: 0.5 }}>
          <FormControl data-testid="platformTest" variant="standard" sx={{ mr: 3, width: 400 }}>
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
          </FormControl>

          <FormControl>
            <ToggleButtonGroup
              size="small"
              color="primary"
              value={params.unitTestsOnly}
              exclusive
              onChange={handleUnitTestsOnlyToggle}
              data-testid="unitTestsOnlyToggleTest"
              sx={{ paddingTop: 0.5, mr: 3 }}
            >
              <ToggleButton value={true}>
                Unit Tests
              </ToggleButton>
              <ToggleButton value={false}>
                All Tests
              </ToggleButton>
            </ToggleButtonGroup>
          </FormControl>
        </Toolbar>
        <Divider />
      </>
    );
  };

  return (
    isConfigLoaded ? renderToolbar() : <></>
  );
}

export default SummaryToolbar;
