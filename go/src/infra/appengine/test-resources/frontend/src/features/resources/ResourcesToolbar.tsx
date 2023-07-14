// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Divider, FormControl, Grid, InputLabel, MenuItem, Select, Stack, TextField, ToggleButton, ToggleButtonGroup, Toolbar } from '@mui/material';
import { DatePicker, LocalizationProvider } from '@mui/x-date-pickers';
import dayjs from 'dayjs';
import CameraAltIcon from '@mui/icons-material/CameraAlt';
import TimelineIcon from '@mui/icons-material/Timeline';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import { useContext, useEffect, useState } from 'react';
import { Period } from '../../api/resources';
import { MetricsContext } from '../context/MetricsContext';


function ResourcesToolbar() {
  const { api, params } = useContext(MetricsContext);
  const [filter, setFilter] = useState(params.filter);

  const handleFilterChange = (event) => {
    setFilter(event.target.value);
  };
  const handleDateChange = (event) => {
    api.updateDate(new Date(event));
  };
  const handlePeriodChange = (event) => {
    api.updatePeriod(event.target.value);
  };

  useEffect(() => {
    const timer = setTimeout(() => {
      api.updateFilter(filter);
    }, 500);
    return () => clearTimeout(timer);
  // Adding this because we don't want a dependency on api
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filter]);

  // If we have week selected as the period, disable everything but Sundays
  const handleShouldDisableDate = (date) => {
    if (String(params.period) === '1') {
      return date.day() !== 0;
    }
    return false;
  };

  const handleTimelineToggle = (_, isTimeline: boolean) => {
    api.updateTimelineView(isTimeline);
  };

  const handleDirectoryToggle = (_, isDirectory: boolean) => {
    api.updateDirectoryView(isDirectory);
  };

  return (
    <>
      <Toolbar>
        <Grid container gap={3}>

          <Grid item xs={3}>
            <TextField
              data-testid="textFieldTest"
              fullWidth
              label="Filter"
              variant="standard"
              onChange={handleFilterChange}
              value={filter}
            />
          </Grid>

          <Grid item xs={1}>
            <FormControl data-testid="formControlTest" fullWidth variant="standard">
              <InputLabel shrink={true}>Period</InputLabel>
              <Select
                value={Number(params.period) as Period}
                label="Period"
                onChange={handlePeriodChange}
              >
                <MenuItem value={0}>Day</MenuItem>
                <MenuItem value={1}>Week</MenuItem>
              </Select>
            </FormControl>
          </Grid>

          <Grid item xs={1}>
            <LocalizationProvider dateAdapter={AdapterDayjs}>
              <DatePicker
                label="Date"
                disableFuture
                onChange={handleDateChange}
                format="YYYY-MM-DD"
                defaultValue={dayjs(params.date)}
                slotProps={{ textField: { variant: 'standard' } }}
                shouldDisableDate={handleShouldDisableDate}
              />
            </LocalizationProvider>
          </Grid>

          <Grid item xs={3}>
            <Stack direction="row" spacing={4}>
              <ToggleButtonGroup
                size="small"
                color="primary"
                value={params.timelineView}
                exclusive
                onChange={handleTimelineToggle}
                aria-label="timeline view"
                data-testid="timelineViewToggle"
                sx={{ paddingTop: 0.5 }}
              >
                <ToggleButton value={false} aria-label="snapshot">
                  <CameraAltIcon />
                </ToggleButton>
                <ToggleButton value={true} aria-label="timeline">
                  <TimelineIcon />
                </ToggleButton>
              </ToggleButtonGroup>

              <ToggleButtonGroup
                size="small"
                color="primary"
                value={params.directoryView}
                exclusive
                onChange={handleDirectoryToggle}
                aria-label="directory view"
                data-testid="directoryViewToggle"
                sx={{ paddingTop: 0.5 }}
              >
                <ToggleButton value={false}>By Test</ToggleButton>
                <ToggleButton value={true}>By Directory</ToggleButton>
              </ToggleButtonGroup>
            </Stack>
          </Grid>

        </Grid>
      </Toolbar>
      <Divider />
    </>
  );
}

export default ResourcesToolbar;
