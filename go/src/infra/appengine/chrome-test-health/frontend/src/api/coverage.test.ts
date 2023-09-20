// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Auth } from './auth';
import { prpcClient } from './client';
import {
  GetSummaryCoverageRequest,
  GetSummaryCoverageResponse,
  getSummaryCoverage,
} from './coverage';

const auth = new Auth('', new Date('3000-01-01'));

describe('getSummaryCoverage', () => {
  const dummyRequest: GetSummaryCoverageRequest = {
    gitiles_host: 'chromium.googlesource.com',
    gitiles_project: 'chromium/src',
    gitiles_ref: 'refs/heads/main',
    gitiles_revision: '03d4e64771cbc97f3ca5e4bbe85490d7cf909a0a',
    path: '//',
    data_type: 'dirs',
    unit_tests_only: true,
    bucket: 'ci',
    builder: 'linux-code-coverage',
  };
  it('returns coverage summary', async () => {
    const mockCall = jest.spyOn(prpcClient, 'call').mockResolvedValue({
      summary: {
        dirs: [
          {
            'name': 'apps/',
            'path': '//apps/',
            'summaries': [
              {
                'covered': 75,
                'name': 'branch',
                'total': 286,
              },
              {
                'covered': 451,
                'name': 'line',
                'total': 1028,
              },
            ],
          },
        ],
        files: [],
        path: '//',
        summaries: [
          {
            'covered': 795625,
            'name': 'branch',
            'total': 1791268,
          },
          {
            'covered': 2519159,
            'name': 'line',
            'total': 4567235,
          },
        ],
      },
    });
    const expected: GetSummaryCoverageResponse = {
      summary: {
        dirs: [
          {
            'name': 'apps/',
            'path': '//apps/',
            'summaries': [
              {
                'covered': 75,
                'name': 'branch',
                'total': 286,
              },
              {
                'covered': 451,
                'name': 'line',
                'total': 1028,
              },
            ],
          },
        ],
        files: [],
        path: '//',
        summaries: [
          {
            'covered': 795625,
            'name': 'branch',
            'total': 1791268,
          },
          {
            'covered': 2519159,
            'name': 'line',
            'total': 4567235,
          },
        ],
      },
    };
    const resp = await getSummaryCoverage(auth, dummyRequest);

    expect(mockCall.mock.calls.length).toBe(1);
    expect(mockCall.mock.calls[0].length).toBe(4);
    expect(mockCall.mock.calls[0][3]).toEqual(dummyRequest);
    expect(resp).toEqual(expected);
  });
});
