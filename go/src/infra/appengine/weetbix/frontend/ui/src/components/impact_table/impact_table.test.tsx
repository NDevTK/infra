// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';

import {
  render,
  screen,
} from '@testing-library/react';

import { getMockCluster } from '../../testing_tools/mocks/cluster_mock';
import ImpactTable from './impact_table';

describe('Test ImpactTable component', () => {
  it('given a cluster, should display it', async () => {
    const cluster = getMockCluster('1234567890abcdef1234567890abcdef');
    render(<ImpactTable cluster={cluster} />);

    await screen.findByText('User Cls Failed Presubmit');
    // Check for 7d unexpected failures total.
    expect(screen.getByText('15800')).toBeInTheDocument();

    // Check for 7d critical failures exonerated.
    expect(screen.getByText('13800')).toBeInTheDocument();
  });
});
