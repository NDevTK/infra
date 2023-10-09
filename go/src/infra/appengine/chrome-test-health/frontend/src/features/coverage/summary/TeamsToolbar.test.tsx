// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { screen } from '@testing-library/react';
import { renderWithContext } from './testUtils';
import TeamsToolbar from './TeamsToolbar';

describe('when rendering the TeamsToolbar', () => {
  it('should render toolbar elements', () => {
    renderWithContext(<TeamsToolbar />);
    expect(screen.getByTestId('teamsAutocompleteTest')).toBeInTheDocument();
  });
});
