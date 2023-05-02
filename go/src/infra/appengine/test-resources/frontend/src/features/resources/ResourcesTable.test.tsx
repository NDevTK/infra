// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render, screen } from '@testing-library/react';
import ResourcesTable from './ResourcesTable';

describe('when rendering the ResourcesTable', () => {
  it('should render the TableContainer', () => {
    render(<ResourcesTable />);
    expect(screen.getByText('Test Suite')).toBeInTheDocument();
    expect(screen.getByText('# Runs')).toBeInTheDocument();
    expect(screen.getByText('# Failures')).toBeInTheDocument();
    expect(screen.getByText('Avg Runtime')).toBeInTheDocument();
    expect(screen.getByText('Total Runtime')).toBeInTheDocument();
    expect(screen.getByText('Avg Cores')).toBeInTheDocument();
  });
});
