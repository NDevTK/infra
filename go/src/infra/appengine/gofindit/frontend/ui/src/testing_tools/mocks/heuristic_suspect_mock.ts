// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { HeuristicSuspect } from '../../services/luci_bisection';

export const getMockHeuristicSuspect = (commitID: string): HeuristicSuspect => {
  return {
    gitilesCommit: {
      host: 'not.a.real.host',
      project: 'chromium',
      id: commitID,
      ref: 'ref/main',
      position: '1',
    },
    reviewUrl: `https://chromium-review.googlesource.com/placeholder/+${commitID}`,
    reviewTitle: '[MyApp] Added new functionality to improve my app',
    score: '15',
    justification:
      'The file "dir/a/b/x.cc" was added and it was in the failure log.\n' +
      'The file "content/util.c" was modified. It was related to the file obj/content/util.o which was in the failure log.',
    confidenceLevel: 'HIGH',
  };
};
