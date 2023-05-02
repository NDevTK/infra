// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render, screen } from '@testing-library/react';
import NavBar from './NavBar';

const mockedUsedNavigate = jest.fn();

jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockedUsedNavigate,
}));

describe('when rendering the navbar', () => {
  it('should render the ', () => {
    render(<NavBar />);
    expect(screen.getByText('COVERAGE')).toBeInTheDocument();
    expect(screen.getByText('RESOURCES')).toBeInTheDocument();
    expect(screen.getByText('FLAKINESS')).toBeInTheDocument();
  });
});
