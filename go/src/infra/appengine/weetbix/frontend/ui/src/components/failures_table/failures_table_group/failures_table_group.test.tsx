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
  GroupKey,
} from '../../../tools/failures_tools';

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
              project='testproject'
              group={mockGroup}
              variantGroups={createMockVariantGroups()}/>
          </tbody>
        </table>,
    );

    await (screen.findByText(mockGroup.key.value));

    expect(screen.getByText(mockGroup.key.value)).toBeInTheDocument();
  });

  it('given a group with children then should display just the group when not expanded', async () => {
    const mockGroup = createDefaultMockFailureGroupWithChildren();
    render(
        <table>
          <tbody>
            <FailuresTableGroup
              project='testproject'
              group={mockGroup}
              variantGroups={createMockVariantGroups()}/>
          </tbody>
        </table>,
    );

    await screen.findByText(mockGroup.key.value);

    expect(screen.getAllByRole('row')).toHaveLength(1);
  });

  it('given a test name group it should show a test history link', async () => {
    const mockGroup = createDefaultMockFailureGroupWithChildren();
    mockGroup.key = { type: 'test', value: 'ninja://package/sometest.Blah?a=1' };
    const parentKeys : GroupKey[] = [{
      type: 'variant',
      key: 'k1',
      value: 'v1',
    }, {
      // Consider a variant with special characters.
      type: 'variant',
      key: 'key %+',
      value: 'value %+',
    }];

    render(
        <table>
          <tbody>
            <FailuresTableGroup
              project='testproject'
              parentKeys={parentKeys}
              group={mockGroup}
              variantGroups={createMockVariantGroups()}/>
          </tbody>
        </table>,
    );

    await screen.findByText(mockGroup.key.value);

    expect(screen.getByLabelText('Test history link')).toBeInTheDocument();
    expect(screen.getByLabelText('Test history link')).toHaveAttribute('href',
        'https://ci.chromium.org/ui/test/testproject/ninja%3A%2F%2Fpackage%2Fsometest.Blah%3Fa%3D1?q=V%3Ak1%3Dv1%20V%3Akey%2520%2525%252B%3Dvalue%2520%2525%252B');
    expect(screen.getAllByRole('row')).toHaveLength(1);
  });

  it('given a group with children then should display all when expanded', async () => {
    const mockGroup = createDefaultMockFailureGroupWithChildren();
    render(
        <table>
          <tbody>
            <FailuresTableGroup
              project='testproject'
              group={mockGroup}
              variantGroups={createMockVariantGroups()}/>
          </tbody>
        </table>,
    );

    await screen.findByText(mockGroup.key.value);

    fireEvent.click(screen.getByLabelText('Expand group'));

    await (screen.findByText(mockGroup.children[2].failures));

    expect(screen.getAllByRole('row')).toHaveLength(4);
  });
});
