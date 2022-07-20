// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

/**
 * Given a Buildbucket build ID, gets the details for a single failure analysis.
 *
 * @return {Promise} a promise that fulfills with the requested data.
 */
export const getAnalysisDetails = async (
    buildbucketId: string
): Promise<AnalysisDetails> => {
  const response = await fetch(`/api/analysis/b/${encodeURIComponent(buildbucketId)}`);
  return await response.json();
};

export interface AnalysisDetails {
  id: number;
  status: string;
  failureType: string;
  buildbucketId: number;
  builder: string;
  suspectRange: string[];
  bugs: string[];
  revertChangeList: ChangeListDetails;
  suspects: SuspectSummary[];
  heuristicAnalysis: HeuristicAnalysisResult;
}

export interface HeuristicAnalysisResult {
  Items: HeuristicAnalysisResultItem[];
}

export interface HeuristicAnalysisResultItem {
  Commit: string;
  ReviewUrl: string;
  Justification: SuspectJustification;
}

export interface SuspectJustification {
  IsNonBlamable: boolean;
  Items: SuspectJustificationItem[];
}

export interface SuspectJustificationItem {
  Score: number;
  FilePath: string;
  Reason: string;
}

export interface ChangeListDetails {
  title: string;
  url: string;
  status: string;
  submitTime: string;
  commitPosition: string;
}

export interface SuspectSummary {
  id: String;
  title: string;
  url: string;
  culpritStatus: string;
  accuseSource: string;
}
