// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { PrimeSuspect } from '../../services/luci_bisection';

export const getMockPrimeSuspect = (commitID: string): PrimeSuspect => {
  return {
    cl: {
      commitID: commitID,
      title: `Mock suspect CL ${commitID}`,
      reviewURL: `https://chromium-review.googlesource.com/mock_suspect_summary/${commitID}`,
    },
    culpritStatus: 'VERIFYING',
    accuseSource: 'Heuristic',
  };
};
