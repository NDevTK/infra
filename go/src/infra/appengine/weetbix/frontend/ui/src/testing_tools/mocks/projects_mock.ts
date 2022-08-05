// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import fetchMock from 'fetch-mock-jest';

import { ProjectConfig } from '../../services/project';

export const createMockProjectConfig = (): ProjectConfig => {
  return {
    name: 'projects/chromium/config',
    monorail: {
      project: 'chromium',
      displayPrefix: 'crbug.com',
    },
  };
};

export const mockFetchProjectConfig = () => {
  fetchMock.post('http://localhost/prpc/weetbix.v1.Projects/GetConfig', {
    headers: {
      'X-Prpc-Grpc-Code': '0',
    },
    body: ')]}\'' + JSON.stringify(createMockProjectConfig()),
  });
};
