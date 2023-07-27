// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { screen } from '@testing-library/react';
import { renderWithContext } from '../../utils/testUtils';
import ResourcesToolbar from './ResourcesToolbar';

describe('when rendering the ResourcesTableToolbar', () => {
  it('should render toolbar elements', () => {
    renderWithContext(<ResourcesToolbar/>);
    expect(screen.getByTestId('CalendarIcon')).toBeInTheDocument();
    expect(screen.getByRole('textbox', {
      name: /date/i,
    })).toBeInTheDocument();
    expect(screen.getByTestId('formControlTest')).toBeInTheDocument();
    expect(screen.getByTestId('textFieldTest')).toBeInTheDocument();
    expect(screen.getByTestId('timelineViewToggle')).toBeInTheDocument();
  });
});
