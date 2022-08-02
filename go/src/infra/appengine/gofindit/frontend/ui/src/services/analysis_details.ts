// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

/**
 * Given a Buildbucket build ID, gets the details for a single failure analysis.
 *
 * @return {Promise} a promise that fulfills with the requested data.
 */
export const getAnalysisDetails = async (
  buildID: string
): Promise<AnalysisDetails> => {
  const response = await fetch(
    `/api/analysis/b/${encodeURIComponent(buildID)}`
  );
  return await response.json();
};

export interface AnalysisDetails {
  id: number;
  status: string;
  failureType: string;
  buildID: number;
  builder: string;
  suspectRange: SuspectRange;
  bugs: AssociatedBug[];
  revertChangeList: ChangeListDetails;
  suspects: SuspectSummary[];
  heuristicAnalysis: HeuristicSuspect[];
}

export interface SuspectRange {
  linkText: string;
  url: string;
}

export interface HeuristicSuspect {
  commitID: string;
  reviewURL: string;
  score: number;
  confidence: string;
  justification: string[];
}

export interface ChangeListDetails {
  title: string;
  url: string;
  status: string;
  submitTime: string;
  commitPosition: string;
}

export interface SuspectSummary {
  id: string;
  title: string;
  url: string;
  culpritStatus: string;
  accuseSource: string;
}

export interface AssociatedBug {
  linkText: string;
  url: string;
}
