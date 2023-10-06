// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { screen } from '@testing-library/react';
import { renderWithContext } from './testUtils';
import SummaryToolbar from './SummaryToolbar';

describe('when rendering the SummaryToolbar', () => {
  it('should render toolbar elements', () => {
    renderWithContext(<SummaryToolbar />);
    expect(screen.getByTestId('platformTest')).toBeInTheDocument();
    expect(screen.getByTestId('unitTestsOnlyToggleTest')).toBeInTheDocument();
  });
  it('should render not toolbar if config not loaded', () => {
    renderWithContext(<SummaryToolbar />, { isConfigLoaded: false });
    expect(screen.queryAllByTestId('platformTest')).toHaveLength(0);
    expect(screen.queryAllByTestId('unitTestsOnlyToggleTest')).toHaveLength(0);
  });
});
