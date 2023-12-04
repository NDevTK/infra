// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {
  GetAbsoluteTrendsRequest,
  GetAbsoluteTrendsResponse,
  GetIncrementalTrendsRequest,
  GetIncrementalTrendsResponse,
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
    successCallback: (
    response: GetAbsoluteTrendsResponse,
  ) => void,
    failureCallback: (error: any) => void,
) {
  const request: GetAbsoluteTrendsRequest = {
    bucket: params.bucket,
    builder: params.builder,
    unit_tests_only: params.unitTestsOnly,
    presets: params.presets,
    paths: params.paths,
    components,
  };

  getAbsoluteCoverageTrends(auth, request).then((response) => {
    successCallback(response);
  }).catch(failureCallback);
}

export function loadIncrementalCoverageTrends(
    auth: Auth,
    params: Params,
    components: string[],
    successCallback: (
    response: GetIncrementalTrendsResponse,
  ) => void,
    failureCallback: (error: any) => void,
) {
  const request: GetIncrementalTrendsRequest = {
    presets: params.presets,
    paths: params.paths,
    components,
  };

  getIncrementalCoverageTrends(auth, request).then((response) => {
    successCallback(response);
  }).catch(failureCallback);
}
