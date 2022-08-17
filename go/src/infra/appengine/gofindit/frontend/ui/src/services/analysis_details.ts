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
  analysisID: string;
  status: string;
  buildID: string;
  failureType: string;
  builder: string;
  suspectRange: SuspectRange;
  bugs: AssociatedBug[];
  revertCL: RevertCL;
  primeSuspects: PrimeSuspect[];
  heuristicResults: HeuristicDetails;
}

export interface SuspectRange {
  linkText: string;
  url: string;
}

export interface HeuristicDetails {
  isComplete: boolean;
  suspects: HeuristicSuspect[];
}

export interface CL {
  commitID: string;
  title: string;
  reviewURL: string;
}

export interface RevertCL {
  cl: CL;
  status: string;
  submitTime: string;
  commitPosition: string;
}

export interface HeuristicSuspect {
  cl: CL;
  score: string;
  confidence: string;
  justification: string[];
}

export interface PrimeSuspect {
  cl: CL;
  culpritStatus: string;
  accuseSource: string;
}

export interface AssociatedBug {
  linkText: string;
  url: string;
}
