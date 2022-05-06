// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React from 'react';

import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import FormControl from '@mui/material/FormControl';
import Grid from '@mui/material/Grid';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import OutlinedInput from '@mui/material/OutlinedInput';
import Select, { SelectChangeEvent } from '@mui/material/Select';

import {
    FailureFilter,
    FailureFilters,
    VariantGroup,
    ImpactFilter,
    ImpactFilters
} from '../../../tools/failures_tools';

interface Props {
    failureFilter: FailureFilter,
    onFailureFilterChanged: (event: SelectChangeEvent) => void,
    impactFilter: ImpactFilter,
    onImpactFilterChanged: (event: SelectChangeEvent) => void,
    variantGroups: VariantGroup[],
    selectedVariantGroups: string[],
    handleVariantGroupsChange: (event: SelectChangeEvent<string[]>) => void,
}

const FailuresTableFilter = ({
    failureFilter,
    onFailureFilterChanged,
    impactFilter,
    onImpactFilterChanged,
    variantGroups,
    selectedVariantGroups,
    handleVariantGroupsChange,
}: Props) => {
    return (
        <>
            <Grid container item xs={12} columnGap={2} data-testid="failure_table_filter">
                <Grid item xs={2}>
                    <FormControl fullWidth data-testid="failure_filter">
                        <InputLabel id="failure_filter_label">Failure filter</InputLabel>
                        <Select
                            labelId="failure_filter_label"
                            id="impact_filter"
                            value={failureFilter}
                            label="Failure filter"
                            onChange={onFailureFilterChanged}
                            inputProps={{ 'data-testid': 'failure_filter_input' }}
                        >
                            {
                                FailureFilters.map((filter) => (
                                    <MenuItem key={filter} value={filter}>{filter}</MenuItem>
                                ))
                            }
                        </Select>
                    </FormControl>
                </Grid>
                <Grid item xs={2}>
                    <FormControl fullWidth data-testid="impact_filter">
                        <InputLabel id="impact_filter_label">Impact filter</InputLabel>
                        <Select
                            labelId="impact_filter_label"
                            id="impact_filter"
                            value={impactFilter.name}
                            label="Impact filter"
                            onChange={onImpactFilterChanged}
                            inputProps={{ 'data-testid': 'impact_filter_input' }}
                        >
                            {
                                ImpactFilters.map((filter) => (
                                    <MenuItem key={filter.name} value={filter.name}>{filter.name}</MenuItem>
                                ))
                            }
                        </Select>
                    </FormControl>
                </Grid>
                <Grid item xs={6}>
                    <FormControl fullWidth data-testid="group_by">
                        <InputLabel id="group_by_label">Group by</InputLabel>
                        <Select
                            labelId="group_by_label"
                            id="group_by"
                            multiple
                            value={selectedVariantGroups}
                            onChange={handleVariantGroupsChange}
                            input={<OutlinedInput id="group_by_select" label="Group by" />}
                            renderValue={(selected) => (
                                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                                    {selected.map((value) => (
                                        <Chip key={value} label={value} />
                                    ))}
                                </Box>
                            )}
                            inputProps={{ 'data-testid': 'group_by_input' }}
                        >
                            {variantGroups.map((variantGroup) => (
                                <MenuItem
                                    key={variantGroup.key}
                                    value={variantGroup.key}
                                >
                                    {variantGroup.key} ({variantGroup.values.length})
                                </MenuItem>
                            ))}
                        </Select>
                    </FormControl>
                </Grid>
            </Grid>
        </>
    );
};

export default FailuresTableFilter;