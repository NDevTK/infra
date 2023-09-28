// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Auth } from './auth';
import { prpcClient } from './client';
import {
  GetSummaryByComponentRequest,
  GetSummaryCoverageRequest,
  GetTeamsResponse,
  SummaryNode,
  getSummaryCoverage,
  getSummaryCoverageByComponent,
  getTeams,
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
    const expected: SummaryNode[] = [
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
        'isDir': true,
        'children': [] as SummaryNode[],
      },
    ];
    const resp = await getSummaryCoverage(auth, dummyRequest);

    expect(mockCall.mock.calls.length).toBe(1);
    expect(mockCall.mock.calls[0].length).toBe(4);
    expect(mockCall.mock.calls[0][3]).toEqual(dummyRequest);
    expect(resp).toEqual(expected);
  });
});

describe('getSummaryCoverageByComponent', () => {
  const dummyRequest: GetSummaryByComponentRequest = {
    gitiles_host: 'chromium.googlesource.com',
    gitiles_project: 'chromium/src',
    gitiles_ref: 'refs/heads/main',
    gitiles_revision: '03d4e64771cbc97f3ca5e4bbe85490d7cf909a0a',
    components: [
      'Blink>CSS',
    ],
    unit_tests_only: true,
    bucket: 'ci',
    builder: 'linux-code-coverage',
  };
  it('returns coverage summary for each component', async () => {
    const mockCall = jest.spyOn(prpcClient, 'call').mockResolvedValue({
      'summary': [
        {
          'dirs': [
            {
              'name': 'css-parser/',
              'path': '//third_party/blink/web_tests/css-parser/',
              'summaries': [
                {
                  'covered': 0,
                  'name': 'branch',
                  'total': 0,
                },
                {
                  'covered': 0,
                  'name': 'function',
                  'total': 0,
                },
                {
                  'covered': 0,
                  'name': 'instantiation',
                  'total': 0,
                },
                {
                  'covered': 0,
                  'name': 'line',
                  'total': 0,
                },
                {
                  'covered': 0,
                  'name': 'region',
                  'total': 0,
                },
              ],
            },
            {
              'name': 'basic/',
              'path': '//third_party/blink/web_tests/css1/basic/',
              'summaries': [
                {
                  'covered': 0,
                  'name': 'branch',
                  'total': 0,
                },
                {
                  'covered': 0,
                  'name': 'function',
                  'total': 0,
                },
                {
                  'covered': 0,
                  'name': 'instantiation',
                  'total': 0,
                },
                {
                  'covered': 30,
                  'name': 'line',
                  'total': 100,
                },
                {
                  'covered': 0,
                  'name': 'region',
                  'total': 0,
                },
              ],
            },
          ],
          'files': [
            {
              'name': 'test_file.cc',
              'path': '//third_party/test_file.cc',
              'summaries': [
                {
                  'covered': 0,
                  'name': 'branch',
                  'total': 0,
                },
                {
                  'covered': 0,
                  'name': 'function',
                  'total': 0,
                },
                {
                  'covered': 0,
                  'name': 'instantiation',
                  'total': 0,
                },
                {
                  'covered': 50,
                  'name': 'line',
                  'total': 100,
                },
                {
                  'covered': 0,
                  'name': 'region',
                  'total': 0,
                },
              ],
            },
          ],
          'path': 'Blink>CSS',
          'summaries': [
            {
              'covered': 0,
              'name': 'branch',
              'total': 0,
            },
            {
              'covered': 0,
              'name': 'function',
              'total': 0,
            },
            {
              'covered': 0,
              'name': 'instantiation',
              'total': 0,
            },
            {
              'covered': 0,
              'name': 'line',
              'total': 0,
            },
            {
              'covered': 0,
              'name': 'region',
              'total': 0,
            },
          ],
        },
      ],
    });
    const expected: SummaryNode[] = [
      {
        'name': 'third_party/',
        'path': '//third_party/',
        'isDir': true,
        'children': [
          {
            'name': 'blink/',
            'path': '//third_party/blink/',
            'isDir': true,
            'children': [
              {
                'name': 'web_tests/',
                'path': '//third_party/blink/web_tests/',
                'isDir': true,
                'children': [
                  {
                    'name': 'css-parser/',
                    'path': '//third_party/blink/web_tests/css-parser/',
                    'isDir': true,
                    'children': [],
                    'summaries': [
                      {
                        'name': 'line',
                        'covered': 0,
                        'total': 0,
                      },
                    ],
                  },
                  {
                    'name': 'css1/',
                    'path': '//third_party/blink/web_tests/css1/',
                    'isDir': true,
                    'children': [
                      {
                        'name': 'basic/',
                        'path': '//third_party/blink/web_tests/css1/basic/',
                        'isDir': true,
                        'children': [],
                        'summaries': [
                          {
                            'name': 'line',
                            'covered': 30,
                            'total': 100,
                          },
                        ],
                      },
                    ],
                    'summaries': [
                      {
                        'name': 'line',
                        'covered': 30,
                        'total': 100,
                      },
                    ],
                  },
                ],
                'summaries': [
                  {
                    'name': 'line',
                    'covered': 30,
                    'total': 100,
                  },
                ],
              },
            ],
            'summaries': [
              {
                'name': 'line',
                'covered': 30,
                'total': 100,
              },
            ],
          },
          {
            'name': 'test_file.cc',
            'path': '//third_party/test_file.cc',
            'isDir': false,
            'children': [],
            'summaries': [
              {
                'name': 'line',
                'covered': 50,
                'total': 100,
              },
            ],
          },
        ],
        'summaries': [
          {
            'name': 'line',
            'covered': 80,
            'total': 200,
          },
        ],
      },
    ];
    const resp = await getSummaryCoverageByComponent(auth, dummyRequest);
    expect(mockCall.mock.calls.length).toBe(1);
    expect(mockCall.mock.calls[0].length).toBe(4);
    expect(mockCall.mock.calls[0][3]).toEqual(dummyRequest);
    expect(resp).toEqual(expected);
  });
});

describe('getTeams', () => {
  it('returns list of teams', async () => {
    const mockCall = jest.spyOn(prpcClient, 'call').mockResolvedValue(
        {
          teams: [
            {
              'id': '1346050280',
              'name': '1346050280',
              'components': [
                'Internals>Instrumentation',
                'Internals>Instrumentation>Memory',
              ],
            },
            {
              'id': '1346052032',
              'name': '1346052032',
              'components': [
                'Infra>OmahaProxy',
              ],
            },
          ],
        });
    const expected: GetTeamsResponse = {
      teams: [
        {
          id: '1346050280',
          name: '1346050280',
          components: [
            'Internals>Instrumentation',
            'Internals>Instrumentation>Memory',
          ],
        },
        {
          id: '1346052032',
          name: '1346052032',
          components: [
            'Infra>OmahaProxy',
          ],
        },
      ],
    };
    const resp = await getTeams(auth);

    expect(mockCall.mock.calls.length).toBe(1);
    expect(mockCall.mock.calls[0].length).toBe(4);
    expect(resp).toEqual(expected);
  });
});
