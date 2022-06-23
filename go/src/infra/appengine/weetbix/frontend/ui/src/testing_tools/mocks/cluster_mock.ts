// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

export const getMockCluster = (id: string) => {
  return {
    'clusterId': {
      'algorithm': 'rules-v2',
      'id': id,
    },
    'criticalFailuresExonerated1d': {
      'nominal': 5625,
      'residual': 5319,
    },
    'criticalFailuresExonerated3d': {
      'nominal': 14052,
      'residual': 13221,
    },
    'criticalFailuresExonerated7d': {
      'nominal': 13800,
      'residual': 13780,
    },
    'presubmitRejects1d': {
      'nominal': 98,
      'residual': 97,
    },
    'presubmitRejects3d': {
      'nominal': 158,
      'residual': 157,
    },
    'presubmitRejects7d': {
      'nominal': 167,
      'residual': 163,
    },
    'testRunFailures1d': {
      'nominal': 2427,
      'residual': 2425,
    },
    'testRunFailures3d': {
      'nominal': 4716,
      'residual': 4494,
    },
    'testRunFailures7d': {
      'nominal': 4938,
      'residual': 4662,
    },
    'failures1d': {
      'nominal': 7625,
      'residual': 7319,
    },
    'failures3d': {
      'nominal': 16052,
      'residual': 15221,
    },
    'failures7d': {
      'nominal': 15800,
      'residual': 15780,
    },
    'title': '',
    'failureAssociationRule': '',
  };
};
