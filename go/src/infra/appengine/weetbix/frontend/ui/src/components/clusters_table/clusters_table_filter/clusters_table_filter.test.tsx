// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';

import {
  render,
  screen,
} from '@testing-library/react';

import { identityFunction } from '../../../testing_tools/functions';
import ClustersTableFilter from './clusters_table_filter';

describe('Test ClustersTableFilter component', () => {
  it('should display the failures filter', async () => {
    render(
        <ClustersTableFilter
          failureFilter=""
          setFailureFilter={identityFunction}/>,
    );

    await screen.findByTestId('clusters_table_filter');

    expect(screen.getByTestId('failure_filter')).toBeInTheDocument();
  });

  it('given an existing filter, the filter should be pre-populated', async () => {
    render(
        <ClustersTableFilter
          failureFilter="some restriction"
          setFailureFilter={identityFunction}/>,
    );

    await screen.findByTestId('clusters_table_filter');

    expect(screen.getByTestId('failure_filter_input')).toHaveValue('some restriction');
  });
});
