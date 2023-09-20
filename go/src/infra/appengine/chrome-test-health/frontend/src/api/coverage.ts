// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Auth } from './auth';
import { prpcClient } from './client';

export interface CoverageMetric {
  name: string,
  covered: number,
  total: number,
}

export interface SummaryNode {
  name: string,
  path: string,
  summaries: CoverageMetric[]
}

export interface Summary {
  dirs: SummaryNode[],
  files: SummaryNode[],
  path: string,
  summaries: CoverageMetric[]
}

export interface GetSummaryCoverageRequest {
  gitiles_host: string,
  gitiles_project: string,
  gitiles_ref: string,
  gitiles_revision: string,
  path: string,
  unit_tests_only: boolean,
  data_type: string,
  bucket: string,
  builder: string
}

export interface GetSummaryCoverageResponse {
  summary: Summary
}

export async function getSummaryCoverage(
    auth: Auth,
    request: GetSummaryCoverageRequest,
):
  Promise<GetSummaryCoverageResponse> {
  const resp: GetSummaryCoverageResponse = await prpcClient.call(
      auth,
      'test_resources.Coverage',
      'GetCoverageSummary',
      request,
  );
  return resp;
}
