// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';

import {
  render,
  screen,
} from '@testing-library/react';

import { identityFunction } from '../../../testing_tools/functions';
import ClustersTableHead from './clusters_table_head';

describe('Test ClustersTableHead', () => {
  it('should display sortable table head', async () => {
    render(
        <table>
          <ClustersTableHead
            isAscending={false}
            toggleSort={identityFunction}
            sortMetric={'critical_failures_exonerated'}/>
        </table>,
    );

    await (screen.findByTestId('clusters_table_head'));

    expect(screen.getByText('User Cls Failed Presubmit')).toBeInTheDocument();
    expect(screen.getByText('Presubmit-Blocking Failures Exonerated')).toBeInTheDocument();
    expect(screen.getByText('Total Failures')).toBeInTheDocument();
  });
});
