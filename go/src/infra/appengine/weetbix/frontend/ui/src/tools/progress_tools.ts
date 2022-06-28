// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import dayjs from 'dayjs';
import isSameOrAfter from 'dayjs/plugin/isSameOrAfter';
import { getClustersService, GetReclusteringProgressRequest, ReclusteringProgress, ClusteringVersion } from '../services/cluster';

dayjs.extend(isSameOrAfter);

export const fetchProgress = async (project: string): Promise<ReclusteringProgress> => {
  const clustersService = getClustersService();
  const request: GetReclusteringProgressRequest = {
    name: `projects/${encodeURIComponent(project)}/reclusteringProgress`,
  };
  const response = await clustersService.getReclusteringProgress(request);
  return response;
};

export const progressNotYetStarted = -1;
export const noProgressToShow = -2;

export const progressToLatestAlgorithms = (progress: ReclusteringProgress): number => {
  return progressTo(progress, (target: ClusteringVersion) => {
    return target.algorithmsVersion >= progress.next.algorithmsVersion;
  });
};

export const progressToLatestConfig = (progress: ReclusteringProgress): number => {
  const targetConfigVersion = dayjs(progress.next.configVersion);
  return progressTo(progress, (target: ClusteringVersion) => {
    return dayjs(target.configVersion).isSameOrAfter(targetConfigVersion);
  });
};

export const progressToRulesVersion = (progress: ReclusteringProgress, rulesVersion: string): number => {
  const ruleDate = dayjs(rulesVersion);
  return progressTo(progress, (target: ClusteringVersion) => {
    return dayjs(target.rulesVersion).isSameOrAfter(ruleDate);
  });
};

// progressTo returns the progress to completing a re-clustering run
// satisfying the given re-clustering target, expressed as a predicate.
// If re-clustering to a goal that would satisfy the target has started,
// the returned value is value from 0 to 1000. If the run is pending,
// the value -1 is returned.
const progressTo = (progress: ReclusteringProgress, predicate: (target: ClusteringVersion) => boolean): number => {
  if (predicate(progress.last)) {
    // Completed
    return 1000;
  }
  if (predicate(progress.next)) {
    return progress.progressPerMille || 0;
  }
  // Run not yet started (e.g. because we are still finishing a previous
  // re-clustering).
  return progressNotYetStarted;
};
