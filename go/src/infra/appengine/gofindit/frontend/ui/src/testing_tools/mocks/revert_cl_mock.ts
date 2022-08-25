// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { RevertCL } from '../../services/gofindit';

export const getMockRevertCL = (commitID: string): RevertCL => {
  return {
    cl: {
      commitID: commitID,
      title: 'Title of this mock change list',
      reviewURL: `https://chromium-review.googlesource.com/placeholder/cl?commit=${commitID}`,
    },
    status: 'MERGED',
    submitTime: '2022-02-02 01:23:45',
    commitPosition: '96542',
  };
};
