// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { BrowserRouter } from 'react-router-dom';
import { ReactElement } from 'react';
import { Platform } from '../../../api/coverage';
import { renderWithAuth } from '../../../features/auth/testUtils';
import { SummaryContext, SummaryContextValue } from './SummaryContext';
import { Node, Params } from './LoadSummary';

export interface OptionalContext {
  data?: Node[],
  api?: {
    updatePlatform?: (platform: string) => void,
    updateRevision?: (revision: string) => void,
    updateUnitTestsOnly?: (unitTestOnly: boolean) => void,
    updateSortOrder: (sortAscending: boolean) => void
  },
  params?: OptionalParams,
  isLoading?: boolean,
  isConfigLoaded?: boolean,
  isSorted?: false,
  isSortedAscending?: true,
}

export interface OptionalParams {
  host?: string,
  project?: string,
  ref?: string,
  revision?: string,
  unitTestsOnly?: boolean,
  platform?: string,
  builder?: string,
  bucket?: string,
  platformList?: Platform[]
}

export function createParams(params? : OptionalParams) : Params {
  return {
    host: params?.host || '',
    project: params?.project || '',
    gitilesRef: params?.ref || '',
    revision: params?.revision || '',
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
  const ctx : SummaryContextValue = {
    data: opts.data || [],
    api: {
      updatePlatform: opts.api?.updatePlatform || defaultApi,
      updateUnitTestsOnly: opts.api?.updateUnitTestsOnly || defaultApi,
      updateRevision: opts.api?.updateRevision || defaultApi,
      updateSortOrder: opts.api?.updateSortOrder || defaultApi,
    },
    params: createParams(opts.params),
    isLoading: (opts.isLoading === undefined ? false : opts.isLoading),
    isConfigLoaded: (opts.isConfigLoaded === undefined ? true : opts.isConfigLoaded),
    isSorted: opts.isSorted == undefined ? false : opts.isSorted,
    isSortedAscending: opts.isSortedAscending == undefined ? false : opts.isSortedAscending,
  };
  return renderWithAuth(
      <BrowserRouter>
        <SummaryContext.Provider value={ctx}>
          {ui}
        </SummaryContext.Provider>
      </BrowserRouter>,
  );
}

