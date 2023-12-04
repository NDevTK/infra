// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { prpcClient } from '../../../api/client';
import { Auth } from '../../../api/auth';
import { CoverageTrend, GetAbsoluteTrendsResponse, GetIncrementalTrendsResponse } from '../../../api/coverage';
import { Params, loadAbsoluteCoverageTrends, loadIncrementalCoverageTrends } from './LoadTrends';

const auth = new Auth('', new Date('3000-01-01'));

describe('loadAbsoluteCoverageTrends', () => {
  const params: Params = {
    presets: ['P1'],
    paths: ['//a/b/'],
    unitTestsOnly: true,
    bucket: 'ci',
    builder: 'linux-code-coverage',
    platform: 'linux',
    platformList: [
      {
        platform: 'linux',
        bucket: 'ci',
        builder: 'linux-code-coverage',
        uiName: 'Linux',
        latestRevision: '12345',
      },
    ],
  };
  const components: string[] = [
    'Blink>CSS',
  ];

  it('loads absolute line coverage trends', async () => {
    jest.spyOn(prpcClient, 'call').mockResolvedValue(
        {
          data: [
            {
              'date': '2023-06-11',
              'covered': 78,
              'total': 100,
            },
            {
              'date': '2023-06-12',
              'covered': 81,
              'total': 100,
            },
          ],
        });
    let data: CoverageTrend[] = [];
    loadAbsoluteCoverageTrends(
        auth,
        params,
        components,
        (resp: GetAbsoluteTrendsResponse) => {
          data = resp.data;
          expect(data.length).toEqual(2);
        },
        () => {/**/},
    );
  });
});

describe('loadIncrementalCoverageTrends', () => {
  const params: Params = {
    presets: ['P1'],
    paths: ['//a/b/'],
    unitTestsOnly: true,
    bucket: '',
    builder: '',
    platform: '',
    platformList: [],
  };
  const components: string[] = [
    'Blink>CSS',
  ];

  it('loads incremental line coverage trends', async () => {
    jest.spyOn(prpcClient, 'call').mockResolvedValue(
        {
          data: [
            {
              'date': '2023-06-11',
              'covered': 78,
              'total': 100,
            },
            {
              'date': '2023-06-12',
              'covered': 81,
              'total': 100,
            },
          ],
        });
    let data: CoverageTrend[] = [];
    loadIncrementalCoverageTrends(
        auth,
        params,
        components,
        (resp: GetIncrementalTrendsResponse) => {
          data = resp.data;
          expect(data.length).toEqual(2);
        },
        () => {/**/},
    );
  });
});
