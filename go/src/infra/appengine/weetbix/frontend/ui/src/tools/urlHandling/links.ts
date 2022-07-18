// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { ClusterId } from '../../services/shared_models';
import { Changelist, ClusterFailure } from '../failures_tools';

export const linkToCluster = (project: string, c: ClusterId): string => {
  if (c.algorithm.startsWith('rules-') || c.algorithm == 'rules') {
    return linkToRule(project, c.id);
  } else {
    const projectEncoded = encodeURIComponent(project);
    const algorithmEncoded = encodeURIComponent(c.algorithm);
    const idEncoded = encodeURIComponent(c.id);
    return `/p/${projectEncoded}/clusters/${algorithmEncoded}/${idEncoded}`;
  }
};

export const linkToRule = (project: string, ruleId: string): string => {
  const projectEncoded = encodeURIComponent(project);
  const ruleIdEncoded = encodeURIComponent(ruleId);
  return `/p/${projectEncoded}/rules/${ruleIdEncoded}`;
};

export const failureLink = (failure: ClusterFailure) => {
  const query = `ID:${failure.testId} `;
  if (failure.ingestedInvocationId?.startsWith('build-')) {
    return `https://ci.chromium.org/ui/b/${failure.ingestedInvocationId.replace('build-', '')}/test-results?q=${encodeURIComponent(query)}`;
  }
  return `https://ci.chromium.org/ui/inv/${failure.ingestedInvocationId}/test-results?q=${encodeURIComponent(query)}`;
};
export const clLink = (cl: Changelist) => {
  return `https://${cl.host}/c/${cl.change}/${cl.patchset}`;
};
export const clName = (cl: Changelist) => {
  const host = cl.host.replace('-review.googlesource.com', '');
  return `${host}/${cl.change}/${cl.patchset}`;
};
