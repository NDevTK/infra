// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { AnalysisSummary } from '../../components/analysis_overview/analysis_overview';

export const getMockAnalysisSummary = (id: string): AnalysisSummary => {
  return {
    analysisID: id,
    status: 'FOUND',
    failureType: 'Compile failure',
    buildID: id,
    builder: 'mock-builder-cc64',
    suspectRange: {
      linkText: '123abc ... 123abe',
      url: 'https://chromium.googlesource.com/placeholder/123abc..123abe',
    },
    bugs: [
      {
        linkText: 'crbug.com/1200',
        url: 'https://bugs.chromium.org/placeholder/chromium/issues/detail?id=1200',
      },
      {
        linkText: 'crbug.com/1209',
        url: 'https://bugs.chromium.org/placeholder/chromium/issues/detail?id=1209',
      },
    ],
  };
};
