// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { ReactElement } from 'react';
import { MetricsContext, Api, Test } from '../features/context/MetricsContext';

export function renderWithContext(
    ui: ReactElement,
    tests: Test[],
    lastPage: boolean,
    mockApi: Api,
) {
  render(
      <BrowserRouter>
        <MetricsContext.Provider value= {{ tests: tests, lastPage: lastPage, api: mockApi }}>
          {ui}
        </MetricsContext.Provider>
      </BrowserRouter>,
  );
}
