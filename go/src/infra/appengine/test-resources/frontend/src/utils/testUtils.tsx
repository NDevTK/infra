// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { ReactElement } from 'react';
import { MetricsContext, Api, Test, MetricsContextValue } from '../features/context/MetricsContext';

export interface OptionalContext {
  tests?: Test[],
  page?: number,
  lastPage?: boolean,
  api?: OptionalApi,
}

export interface OptionalApi {
  // Page navigation
  nextPage?: () => void,
  prevPage?: () => void,
  firstPage?: () => void,
}

const defaultApi : Api = {
  nextPage: () => {
    // do nothing.
  },
  prevPage: () => {
    // do nothing.
  },
  firstPage: () => {
    // do nothing.
  },
};

export function renderWithContext(
    ui: ReactElement,
    opts: OptionalContext,
) {
  const ctx : MetricsContextValue = {
    tests: opts.tests || [],
    page: opts.page || 0,
    lastPage: opts.lastPage || true,
    api: {
      // Page navigation
      nextPage: opts.api?.nextPage || defaultApi.nextPage,
      prevPage: opts.api?.prevPage || defaultApi.prevPage,
      firstPage: opts.api?.firstPage || defaultApi.firstPage,
    },
  };
  render(
      <BrowserRouter>
        <MetricsContext.Provider value= {ctx}>
          {ui}
        </MetricsContext.Provider>
      </BrowserRouter>,
  );
}
