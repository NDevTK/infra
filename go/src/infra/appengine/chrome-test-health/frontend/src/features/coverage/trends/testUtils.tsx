// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { BrowserRouter } from 'react-router-dom';
import { ReactElement } from 'react';
import { Platform } from '../../../api/coverage';
import { renderWithAuth } from '../../../features/auth/testUtils';
import { Params } from './LoadTrends';
import { TrendsContext, TrendsContextValue } from './TrendsContext';

export interface OptionalContext {
  data?: [],
  api?: {
    updatePlatform?: () => {/**/},
    updateUnitTestsOnly?: () => {/**/},
    updatePaths?: () => {/**/},
    updatePresets?: () => {/**/},
    loadAbsTrends?: () => {/**/},
    loadIncTrends?: () => {/**/},
  },
  params?: OptionalParams,
  isLoading?: boolean,
  isConfigLoaded?: boolean,
  isAbsTrend?: boolean,
}

export interface OptionalParams {
  presets?: string[],
  paths?: string[],
  unitTestsOnly?: boolean,
  platform?: string,
  builder?: string,
  bucket?: string,
  platformList?: Platform[]
}

export function createParams(params? : OptionalParams) : Params {
  return {
    presets: params?.presets || [],
    paths: params?.paths || [],
    unitTestsOnly: params?.unitTestsOnly || false,
    platform: params?.platform || '',
    builder: params?.builder || '',
    bucket: params?.builder || '',
    platformList: params?.platformList || [],
  };
}

const defaultApi = () => {/**/};

export function renderWithContext(
    ui: ReactElement,
    opts: OptionalContext = {},
) {
  const ctx : TrendsContextValue = {
    data: opts.data || [],
    api: {
      updatePlatform: opts.api?.updatePlatform || defaultApi,
      updateUnitTestsOnly: opts.api?.updateUnitTestsOnly || defaultApi,
      updatePaths: opts.api?.updatePaths || defaultApi,
      updatePresets: opts.api?.updatePresets || defaultApi,
      loadAbsTrends: opts.api?.loadAbsTrends || defaultApi,
      loadIncTrends: opts.api?.loadIncTrends || defaultApi,
    },
    params: createParams(opts.params),
    isLoading: (opts.isLoading === undefined ? false : opts.isLoading),
    isConfigLoaded: (opts.isConfigLoaded === undefined ? true : opts.isConfigLoaded),
    isAbsTrend: (opts.isAbsTrend === undefined ? true : opts.isAbsTrend),
  };
  return renderWithAuth(
      <BrowserRouter>
        <TrendsContext.Provider value={ctx}>
          {ui}
        </TrendsContext.Provider>
      </BrowserRouter>,
  );
}
