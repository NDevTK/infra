// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {
  CoverageTrend,
  GetAbsoluteTrendsRequest,
  GetIncrementalTrendsRequest,
  Platform,
  getAbsoluteCoverageTrends,
  getIncrementalCoverageTrends,
} from '../../../api/coverage';
import { Auth } from '../../../api/auth';

export interface Params {
  presets: string[],
  paths: string[],
  unitTestsOnly: boolean,
  platform: string,
  builder: string,
  bucket: string,
  platformList: Platform[]
}

export function loadAbsoluteCoverageTrends(
    auth: Auth,
    params: Params,
    components: string[],
    successCallback: (response: CoverageTrend[]) => void,
    failureCallback: (error: any) => void,
) {
  const request: GetAbsoluteTrendsRequest = {
    bucket: params.bucket,
    builder: params.builder,
    unit_tests_only: params.unitTestsOnly,
    paths: params.paths,
    components,
  };

  getAbsoluteCoverageTrends(auth, request).then((response) => {
    let trends = [] as CoverageTrend[];
    response.reports.forEach((report) => {
      trends = [...trends, {
        date: report.date,
        covered: report.linesCovered,
        total: report.totalLines,
      }];
    });
    trends = sortTrends(trends);
    successCallback(trends);
  }).catch(failureCallback);
}

export function loadIncrementalCoverageTrends(
    auth: Auth,
    params: Params,
    successCallback: (response: CoverageTrend[]) => void,
    failureCallback: (error: any) => void,
) {
  const request: GetIncrementalTrendsRequest = {
    paths: params.paths,
    unit_tests_only: params.unitTestsOnly,
  };

  getIncrementalCoverageTrends(auth, request).then((response) => {
    let trends = [] as CoverageTrend[];
    response.reports.forEach((report) => {
      trends = [...trends, {
        date: report.date,
        covered: report.fileChangesCovered,
        total: report.totalFileChanges,
      }];
    });
    trends = sortTrends(trends);
    successCallback(trends);
  }).catch(failureCallback);
}

function sortTrends(trends: CoverageTrend[]): CoverageTrend[] {
  const sorted = trends.sort((a, b) => {
    const d1 = new Date(b.date);
    const d2 = new Date(a.date);
    return d1.valueOf() - d2.valueOf();
  });
  return sorted;
}
