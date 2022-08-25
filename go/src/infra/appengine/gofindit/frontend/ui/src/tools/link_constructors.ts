// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { GitilesCommit } from '../services/gofindit';

export interface ExternalLink {
  linkText: string;
  url: string;
}

export const linkToBuild = (buildID: string) => {
  return {
    linkText: buildID,
    url: `https://ci.chromium.org/b/${buildID}`,
  };
};

export const linkToCommit = (commit: GitilesCommit) => {
  return {
    linkText: getCommitShortHash(commit.id),
    url: `https://${commit.host}/${commit.project}/+log/${commit.id}$`,
  };
};

export const linkToCommitRange = (
  lastPassed: GitilesCommit,
  firstFailed: GitilesCommit
) => {
  const host = lastPassed.host;
  const project = lastPassed.project;
  const lastPassedShortHash = getCommitShortHash(lastPassed.id);
  const firstFailedShortHash = getCommitShortHash(firstFailed.id);
  return {
    linkText: `${lastPassedShortHash} ... ${firstFailedShortHash}`,
    url: `https://${host}/${project}/+log/${lastPassedShortHash}..${firstFailedShortHash}`,
  };
};

export const getCommitShortHash = (commitID: string) => {
  return commitID.substring(0, 7);
};
