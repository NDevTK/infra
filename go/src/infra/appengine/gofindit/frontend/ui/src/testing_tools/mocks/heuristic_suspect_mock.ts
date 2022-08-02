// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { HeuristicSuspect } from '../../services/analysis_details';

export const getMockHeuristicSuspect = (commitID: string): HeuristicSuspect => {
  return {
    commitID: commitID,
    reviewURL: 'https://chromium-review.googlesource.com/placeholder/+12341223',
    score: 15,
    confidence: 'HIGH',
    justification: [
      'The file "dir/a/b/x.cc" was added and it was in the failure log.',
      'The file "content/util.c" was modified. It was related to the file obj/content/util.o which was in the failure log.',
    ],
  };
};
