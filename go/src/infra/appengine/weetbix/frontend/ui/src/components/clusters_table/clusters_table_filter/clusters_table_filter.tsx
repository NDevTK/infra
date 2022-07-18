// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {
  useState,
} from 'react';

import IconButton from '@mui/material/IconButton';
import FormControl from '@mui/material/FormControl';
import Grid from '@mui/material/Grid';
import TextField from '@mui/material/TextField';
import Popover from '@mui/material/Popover';
import InputAdornment from '@mui/material/InputAdornment';
import HelpOutline from '@mui/icons-material/HelpOutline';
import Search from '@mui/icons-material/Search';
import Typography from '@mui/material/Typography';

interface Props {
    failureFilter: string,
    setFailureFilter: (filter: string) => void
}

const FilterHelp = () => {
  // TODO: more styling on this.
  return <Typography sx={{ p: 2, maxWidth: '800px' }}>
    <p>Searching will display clusters and cluster impact based only on test failures that match your search.</p>
    <p>Searching supports a subset of <a href="https://google.aip.dev/160">AIP-160 filtering</a>.</p>
    <p>A bare value is searched for in the columns test_id and failure_reason.  Values are case-sensitive. E.g. <b>ninja</b> or <b>&ldquo;test failed&rdquo;</b>.</p>
    <p>You can use AND, OR and NOT (case sensitive) logical operators, along with grouping. &lsquo;-&rsquo; is equivalent to NOT. Multiple bare values are considered to be AND separated.  These are equivalent: <b>hello world</b> and <b>hello AND world</b>.
      More examples: <b>a OR b</b> or <b>a AND NOT(b or -c)</b>.</p>
    <p>You can search particular columns with &lsquo;=&rsquo;, &lsquo;!=&rsquo; and &lsquo;:&rsquo; (has) operators. The right hand side of the operator must be a simple value. E.g. <b>test_id:telemetry</b>, <b>-failure_reason:Timeout</b> or <b>ingested_invocation_id=&ldquo;build-8822963500388678513&rdquo;</b>.</p>
    <p>Supported columns to search on:
      <ul>
        <li>test_id</li>
        <li>failure_reason</li>
        <li>realm</li>
        <li>ingested_invocation_id</li>
        <li>cluster_algorithm</li>
        <li>cluster_id</li>
        <li>variant_hash</li>
        <li>test_run_id</li>
        <li>presubmit_run_owner</li>
      </ul>
    </p>
  </Typography>;
};

const ClustersTableFilter = ({
  failureFilter,
  setFailureFilter,
}: Props) => {
  const [filter, setFilter] = useState<string>(failureFilter);
  const [filterHelpAnchorEl, setFilterHelpAnchorEl] = useState<HTMLButtonElement | null>(null);

  return (
    <Grid container item xs={12} columnGap={2} data-testid="clusters_table_filter">
      <Grid item xs={12}>
        <FormControl fullWidth data-testid="failure_filter">
          <TextField
            id="failure_filter"
            value={filter}
            variant='outlined'
            label='Filter failures'
            placeholder='Filter test failures used in clusters'
            onChange={(e) => setFilter(e.target.value)}
            onKeyUp={(e) => {
              if (e.key == 'Enter') {
                setFailureFilter(filter);
              }
            }}
            onBlur={()=> setFailureFilter(filter)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <Search />
                </InputAdornment>),
              endAdornment: (
                <InputAdornment position="end">
                  <IconButton
                    aria-label="toggle search help"
                    edge="end"
                    onClick={(e) => setFilterHelpAnchorEl(e.currentTarget)}
                  >
                    {<HelpOutline />}
                  </IconButton>
                </InputAdornment>),
            }}
            inputProps={{
              'data-testid': 'failure_filter_input',
            }}>
          </TextField>
        </FormControl>
        <Popover open={Boolean(filterHelpAnchorEl)} anchorEl={filterHelpAnchorEl} onClose={() => setFilterHelpAnchorEl(null)} anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'right',
        }} transformOrigin={{
          vertical: 'top',
          horizontal: 'right',
        }}>
          <FilterHelp />
        </Popover>
      </Grid>
    </Grid>
  );
};

export default ClustersTableFilter;
