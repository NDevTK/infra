// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';

import {
  fireEvent,
  render,
  screen,
} from '@testing-library/react';

import {
  createDefaultMockFailureGroup,
  createDefaultMockFailureGroupWithChildren,
  createMockVariantGroups,
} from '../../../testing_tools/mocks/failures_mock';
import FailuresTableGroup from './failures_table_group';

describe('Test FailureTableGroup component', () => {
  it('given a group without children then should display 1 row', async () => {
    const mockGroup = createDefaultMockFailureGroup();
    render(
        <table>
          <tbody>
            <FailuresTableGroup
              group={mockGroup}
              variantGroups={createMockVariantGroups()}/>
          </tbody>
        </table>,
    );

    await (screen.findByText(mockGroup.name));

    expect(screen.getByText(mockGroup.name)).toBeInTheDocument();
  });

  it('given a group with children then should display just the group when not expanded', async () => {
    const mockGroup = createDefaultMockFailureGroupWithChildren();
    render(
        <table>
          <tbody>
            <FailuresTableGroup
              group={mockGroup}
              variantGroups={createMockVariantGroups()}/>
          </tbody>
        </table>,
    );

    await screen.findByText(mockGroup.name);

    expect(screen.getAllByRole('row')).toHaveLength(1);
  });

  it('given a group with children then should display all when expanded', async () => {
    const mockGroup = createDefaultMockFailureGroupWithChildren();
    render(
        <table>
          <tbody>
            <FailuresTableGroup
              group={mockGroup}
              variantGroups={createMockVariantGroups()}/>
          </tbody>
        </table>,
    );

    await screen.findByText(mockGroup.name);

    fireEvent.click(screen.getByLabelText('Expand group'));

    await (screen.findByText(mockGroup.children[2].failures));

    expect(screen.getAllByRole('row')).toHaveLength(4);
  });
});
