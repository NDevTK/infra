// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';

import { ChangeListDetails } from '../../services/analysis_details';
import { ChangeListOverview } from './change_list_overview';
import { getMockChangeListDetails } from '../../testing_tools/mocks/change_list_details_mock';

describe('Test ChangeListOverview component', () => {
  test('if all change list details are displayed', async () => {
    const mockChangeList = getMockChangeListDetails('12835');

    render(<ChangeListOverview changeList={mockChangeList} />);

    await screen.findByTestId('change_list_overview_table_body');

    const expectedStaticFields = [
      ['status', 'status'],
      ['submitted time', 'submitTime'],
      ['commit position', 'commitPosition'],
    ];

    // check static field labels and values are displayed
    expectedStaticFields.forEach(([label, property]) => {
      const fieldLabel = screen.getByText(new RegExp(`^(${label})$`, 'i'));
      expect(fieldLabel).toBeInTheDocument();
      expect(fieldLabel.nextSibling?.textContent).toBe(
        `${mockChangeList[property as keyof ChangeListDetails]}`
      );
    });

    // check the link to the code review is displayed
    expect(screen.getByText(mockChangeList.title).getAttribute('href')).toBe(
      mockChangeList.url
    );
  });
});
