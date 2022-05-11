// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { ClusterFailure } from '../tools/failures_tools';

export const getFailures = async (
    project: string,
    clusterAlgorithm: string,
    clusterID: string,
): Promise<ClusterFailure[]> => {
  const response = await fetch(`/api/projects/${encodeURIComponent(project)}/clusters/${encodeURIComponent(clusterAlgorithm)}/${encodeURIComponent(clusterID)}/failures`);
  return await response.json();
};
