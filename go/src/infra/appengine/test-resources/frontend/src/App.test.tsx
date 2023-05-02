// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render, screen } from '@testing-library/react';
import App from './App';

jest.mock('./features/NavBar', () => {
  const LandingPage = () => <div data-testid="Navbar" />;
  return LandingPage;
});

describe('when rendering the application', () => {
  it('should render the navbar component', () => {
    render(<App />);
    expect(screen.getByTestId('Navbar')).toBeInTheDocument();
  });
});
