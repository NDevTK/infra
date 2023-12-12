// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { screen } from '@testing-library/react';
import TrendsToolbar from './TrendsToolbar';
import { renderWithContext } from './testUtils';

describe('when rendering the TrendsToolbar on Absolute Coverage Trends Page', () => {
  it('should render toolbar elements', () => {
    renderWithContext(<TrendsToolbar />, { isAbsTrend: true });
    expect(screen.getByTestId('platformTest')).toBeInTheDocument();
    expect(screen.getByTestId('unitTestsOnlyToggleTest')).toBeInTheDocument();
    expect(screen.getByTestId('pathTest')).toBeInTheDocument();
  });
});

describe('when rendering the TrendsToolbar on Incremental Coverage Trends Page', () => {
  it('should render toolbar elements', () => {
    renderWithContext(<TrendsToolbar />, { isAbsTrend: false });
    expect(screen.getByTestId('unitTestsOnlyToggleTest')).toBeInTheDocument();
    expect(screen.getByTestId('pathTest')).toBeInTheDocument();
  });
});
