// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { SuspectSummary } from '../../services/analysis_details';

export const getMockSuspectSummary = (id: string): SuspectSummary => {
  return {
    id: id,
    title: `Mock suspect CL ${id}`,
    url: `https://chromium-review.googlesource.com/mock_suspect_summary/${id}`,
    culpritStatus: 'VERIFYING',
    accuseSource: 'Heuristic',
  };
};
