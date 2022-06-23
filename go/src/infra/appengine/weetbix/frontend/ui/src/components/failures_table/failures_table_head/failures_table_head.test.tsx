// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';

import {
  render,
  screen,
} from '@testing-library/react';

import { identityFunction } from '../../../testing_tools/functions';
import FailuresTableHead from './failures_table_head';

describe('Test FailureTableHead', () => {
  it('should display sortable table head', async () => {
    render(
        <table>
          <FailuresTableHead
            isAscending={false}
            toggleSort={identityFunction}
            sortMetric={'latestFailureTime'}/>
        </table>,
    );

    await (screen.findByTestId('failure_table_head'));

    expect(screen.getByText('User Cls Failed Presubmit')).toBeInTheDocument();
    expect(screen.getByText('Builds Failed')).toBeInTheDocument();
    expect(screen.getByText('Presubmit-Blocking Failures Exonerated')).toBeInTheDocument();
    expect(screen.getByText('Total Failures')).toBeInTheDocument();
    expect(screen.getByText('Latest Failure Time')).toBeInTheDocument();
  });
});
