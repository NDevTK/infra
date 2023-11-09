// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {
  Divider,
  FormControl,
  InputAdornment,
  InputLabel,
  MenuItem,
  Select,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Toolbar,
  Tooltip,
} from '@mui/material';
import { DatePicker, LocalizationProvider } from '@mui/x-date-pickers';
import dayjs from 'dayjs';
import HelpOutlineIcon from '@mui/icons-material/HelpOutline';
import CameraAltIcon from '@mui/icons-material/CameraAlt';
import HistoryIcon from '@mui/icons-material/History';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import { useContext, useEffect, useState } from 'react';
import { MetricType, Period } from '../../../api/resources';
import { TestMetricsContext } from './TestMetricsContext';

function TestMetricsToolbar() {
  const { api, params } = useContext(TestMetricsContext);
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
    if (String(params.period) === String(Period.WEEK)) {
      return date.day() !== 0;
    }
    return false;
  };

  // Adding null in method signature to account for selecting selected toggle
  // in which we want to ignore
  const handleTimelineToggle = (_, isTimeline: boolean | null) => {
    if (isTimeline !== null) {
      api.updateTimelineView(isTimeline);
    }
  };

  // Adding null in method signature to account for selecting selected toggle
  // in which we want to ignore
  const handleDirectoryToggle = (_, isDirectory: boolean | null) => {
    if (isDirectory !== null) {
      api.updateDirectoryView(isDirectory);
    }
  };

  const handleMetricChange = (event) => {
    api.updateTimelineMetric(event.target.value);
  };

  return (
    <>
      <Toolbar sx={{ mt: 0.5 }}>

        <TextField
          data-testid="textFieldTest"
          label="Filter"
          variant="standard"
          onChange={handleFilterChange}
          value={filter}
          autoComplete='off'
          InputProps={{
            endAdornment: (
              <InputAdornment position="end">
                <Tooltip title="Filter by id, name, file, bucket, builder, or
                  test_suite. These can also be specified as colon separated
                  pairs (e.g. builder:linux-rel). Regexes are supported.">
                  <HelpOutlineIcon />
                </Tooltip>
              </InputAdornment>
            ),
          }}
          sx={{ mr: 3, minWidth: '300px' }}
        />

        <FormControl data-testid="formControlTest" variant="standard" sx={{ mr: 3 }}>
          <InputLabel shrink={true}>Period</InputLabel>
          <Select
            value={Number(params.period) as Period}
            label="Period"
            onChange={handlePeriodChange}
          >
            <MenuItem value={Period.DAY}>Day</MenuItem>
            <MenuItem value={Period.WEEK}>Week</MenuItem>
          </Select>
        </FormControl>

        <LocalizationProvider dateAdapter={AdapterDayjs}>
          <DatePicker
            label="Date"
            disableFuture
            onChange={handleDateChange}
            format="YYYY-MM-DD"
            value={dayjs(params.date)}
            slotProps={{ textField: { variant: 'standard' } }}
            shouldDisableDate={handleShouldDisableDate}
            sx={{ mr: 3 }}
          />
        </LocalizationProvider>

        <ToggleButtonGroup
          size="small"
          color="primary"
          value={params.timelineView}
          exclusive
          onChange={handleTimelineToggle}
          aria-label="timeline view"
          data-testid="timelineViewToggle"
          sx={{ paddingTop: 0.5, mr: 3 }}
        >
          <ToggleButton value={false} aria-label="snapshot">
            <CameraAltIcon />
          </ToggleButton>
          <ToggleButton value={true} aria-label="timeline">
            <HistoryIcon />
          </ToggleButton>
        </ToggleButtonGroup>

        {params.timelineView ?
                  <FormControl data-testid="formControlMetricTest" variant="standard" sx={{ mr: 3 }}>
                    <InputLabel shrink={true}>Metric</InputLabel>
                    <Select
                      value={params.timelineMetric}
                      label="Show Metric"
                      onChange={handleMetricChange}
                    >
                      <MenuItem value={MetricType.NUM_RUNS}># Runs</MenuItem>
                      <MenuItem value={MetricType.NUM_FAILURES}># Failures</MenuItem>
                      <MenuItem value={MetricType.TOTAL_RUNTIME}>Total Runtime</MenuItem>
                      <MenuItem value={MetricType.AVG_CORES}>Avg Cores</MenuItem>
                      <MenuItem value={MetricType.AVG_RUNTIME}>Avg Runtime</MenuItem>
                    </Select>
                  </FormControl> :
                null}

        <ToggleButtonGroup
          size="small"
          color="primary"
          value={params.directoryView}
          exclusive
          onChange={handleDirectoryToggle}
          aria-label="directory view"
          data-testid="directoryViewToggle"
          sx={{ paddingTop: 0.5, mr: 3 }}
        >
          <ToggleButton value={false}>By Test</ToggleButton>
          <ToggleButton value={true}>By Directory</ToggleButton>
        </ToggleButtonGroup>
      </Toolbar>
      <Divider />
    </>
  );
}

export default TestMetricsToolbar;
