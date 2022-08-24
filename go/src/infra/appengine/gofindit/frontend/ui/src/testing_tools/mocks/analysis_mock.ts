// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Analysis } from '../../services/gofindit';

export const getMockAnalysis = (id: string): Analysis => {
  return {
    analysisId: id,
    status: 'FOUND',
    lastPassedBbid: '0',
    firstFailedBbid: id,
    builder: {
      project: 'chromium/test',
      bucket: 'ci',
      builder: 'mock-builder-cc64',
    },
    failureType: 'Compile',
    culpritAction: [
      {
        actionType: 'BUG_COMMENTED',
        bugUrl: 'https://crbug.com/testProject/11223344',
      },
      {
        actionType: 'BUG_COMMENTED',
        bugUrl: 'https://buganizer.corp.google.com/99887766',
      },
    ],
    heuristicResult: {
      status: 'NOTFOUND',
    },
    nthSectionResult: {
      status: 'RUNNING',
      remainingNthSectionRange: {
        lastPassed: {
          host: 'testHost',
          project: 'testProject',
          ref: 'test/ref/dev',
          id: 'abc123abc123',
          position: '102',
        },
        firstFailed: {
          host: 'testHost',
          project: 'testProject',
          ref: 'test/ref/dev',
          id: 'def456def456',
          position: '103',
        },
      },
    },
  };
};
