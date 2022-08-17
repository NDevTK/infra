// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';

import dayjs from 'dayjs';

import {
  render,
  screen,
} from '@testing-library/react';

import {
  createMockVariantGroups,
  newMockFailure,
  newMockGroup,
} from '../../../../testing_tools/mocks/failures_mock';
import FailuresTableRows from './failures_table_rows';

describe('Test FailureTableRows component', () => {
  it('given a group without children', async () => {
    const mockGroup = newMockGroup({ type: 'leaf', value: 'testgroup' })
        .withFailures(2)
        .withPresubmitRejects(3)
        .withInvocationFailures(4)
        .withCriticalFailuresExonerated(5)
        .build();
    render(
        <table>
          <tbody>
            <FailuresTableRows
              project='testproject'
              group={mockGroup}
              variantGroups={createMockVariantGroups()}/>
          </tbody>
        </table>,
    );

    await screen.findByText(mockGroup.key.value);

    expect(screen.getByText(mockGroup.presubmitRejects)).toBeInTheDocument();
    expect(screen.getByText(mockGroup.invocationFailures)).toBeInTheDocument();
    expect(screen.getByText(mockGroup.criticalFailuresExonerated)).toBeInTheDocument();
    expect(screen.getByText(mockGroup.failures)).toBeInTheDocument();
    expect(screen.getByText(dayjs(mockGroup.latestFailureTime).fromNow())).toBeInTheDocument();
  });

  it('given a group with a failure then should display links and variants', async () => {
    const mockGroup = newMockGroup({ type: 'leaf', value: 'testgroup' })
        .withFailure(newMockFailure().build())
        .withFailures(2)
        .withPresubmitRejects(3)
        .withInvocationFailures(4)
        .withCriticalFailuresExonerated(5)
        .build();
    render(
        <table>
          <tbody>
            <FailuresTableRows
              project='testproject'
              group={mockGroup}
              variantGroups={createMockVariantGroups()}/>
          </tbody>
        </table>,
    );

    await screen.findByLabelText('Failure invocation id');

    expect(screen.getByTestId('ungrouped_variants')).toBeInTheDocument();
    expect(screen.getByLabelText('Presubmit rejects link')).toBeInTheDocument();
  });
});
