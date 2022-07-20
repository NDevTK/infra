// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';

import {
  screen,
} from '@testing-library/react';

import {
  getMockRuleClusterSummary,
  getMockSuggestedClusterSummary,
} from '../../../testing_tools/mocks/cluster_mock';
import { renderWithRouterAndClient } from '../../../testing_tools/libs/mock_router';
import ClustersTableRow from './clusters_table_row';

describe('Test ClustersTableRow component', () => {
  it('given a rule cluster', async () => {
    const mockCluster = getMockRuleClusterSummary('abcdef1234567890abcdef1234567890');
    renderWithRouterAndClient(
        <table>
          <tbody>
            <ClustersTableRow
              project='testproject'
              cluster={mockCluster}/>
          </tbody>
        </table>,
    );

    await screen.findByText(mockCluster.title);

    expect(screen.getByText(mockCluster.bug?.linkText || '')).toBeInTheDocument();
    expect(screen.getByText(mockCluster.presubmitRejects || '0')).toBeInTheDocument();
    expect(screen.getByText(mockCluster.criticalFailuresExonerated || '0')).toBeInTheDocument();
    expect(screen.getByText(mockCluster.failures || '0')).toBeInTheDocument();
  });

  it('given a suggested cluster', async () => {
    const mockCluster = getMockSuggestedClusterSummary('abcdef1234567890abcdef1234567890');
    renderWithRouterAndClient(
        <table>
          <tbody>
            <ClustersTableRow
              project='testproject'
              cluster={mockCluster}/>
          </tbody>
        </table>,
    );

    await screen.findByText(mockCluster.title);

    expect(screen.getByText(mockCluster.presubmitRejects || '0')).toBeInTheDocument();
    expect(screen.getByText(mockCluster.criticalFailuresExonerated || '0')).toBeInTheDocument();
    expect(screen.getByText(mockCluster.failures || '0')).toBeInTheDocument();
  });
});
