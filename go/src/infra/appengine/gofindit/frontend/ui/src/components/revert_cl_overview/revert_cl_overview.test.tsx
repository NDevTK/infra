// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';

import { RevertCL } from '../../services/luci_bisection';
import { RevertCLOverview } from './revert_cl_overview';
import { getMockRevertCL } from '../../testing_tools/mocks/revert_cl_mock';

describe('Test ChangeListOverview component', () => {
  test('if all change list details are displayed', async () => {
    const mockRevertCL = getMockRevertCL('12835');

    render(<RevertCLOverview revertCL={mockRevertCL} />);

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
        `${mockRevertCL[property as keyof RevertCL]}`
      );
    });

    // check the link to the code review is displayed
    expect(screen.getByText(mockRevertCL.cl.title).getAttribute('href')).toBe(
      mockRevertCL.cl.reviewURL
    );
  });
});
