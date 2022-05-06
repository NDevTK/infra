// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';

import React from 'react';

import {
    render,
    screen
} from '@testing-library/react';

import { identityFunction } from '../../../testing_tools/functions';
import { createMockVariantGroups } from '../../../testing_tools/mocks/failures_mock';
import { defaultImpactFilter } from '../../../tools/failures_tools';
import FailuresTableFilter from './failures_table_filter';

describe('Test FailureTableFilter component', () => {

    it('should display 3 filters.', async () => {
        render(
            <FailuresTableFilter
                failureFilter="All Failures"
                onFailureFilterChanged={identityFunction}
                impactFilter={defaultImpactFilter}
                onImpactFilterChanged={identityFunction}
                variantGroups={createMockVariantGroups()}
                selectedVariantGroups={[]}
                handleVariantGroupsChange={identityFunction}
            />
        );

        await screen.findByTestId('failure_table_filter');

        expect(screen.getByTestId('failure_filter')).toBeInTheDocument();
        expect(screen.getByTestId('impact_filter')).toBeInTheDocument();
        expect(screen.getByTestId('group_by')).toBeInTheDocument();
    });

    it('given non default selected values then should display them', async () => {
        render(
            <FailuresTableFilter
                failureFilter="All Failures"
                onFailureFilterChanged={identityFunction}
                impactFilter={defaultImpactFilter}
                onImpactFilterChanged={identityFunction}
                variantGroups={createMockVariantGroups()}
                selectedVariantGroups={['v1', 'v2']}
                handleVariantGroupsChange={identityFunction}
            />
        );

        await screen.findByTestId('failure_table_filter');

        expect(screen.getByTestId('failure_filter_input')).toHaveValue('All Failures');
        expect(screen.getByTestId('impact_filter_input')).toHaveValue(defaultImpactFilter.name);
        expect(screen.getByTestId('group_by_input')).toHaveValue('v1,v2');
    });
});