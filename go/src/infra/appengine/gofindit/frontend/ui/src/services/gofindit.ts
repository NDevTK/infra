// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { AuthorizedPrpcClient } from '../clients/authorized_client';

export const getGoFinditService = () => {
  const client = new AuthorizedPrpcClient();
  return new GoFinditService(client);
};

// A service to handle GoFindit-related pRPC requests.
export class GoFinditService {
  // The name of the pRPC service to connect to.
  private static SERVICE = 'gofindit.GoFinditService';

  client: AuthorizedPrpcClient;

  constructor(client: AuthorizedPrpcClient) {
    this.client = client;
  }

  async queryAnalysis(
    request: QueryAnalysisRequest
  ): Promise<QueryAnalysisResponse> {
    return this.client.call(GoFinditService.SERVICE, 'QueryAnalysis', request);
  }
}

export type AnalysisStatus =
  | 'ANALYSIS_STATUS_UNSPECIFIED'
  | 'CREATED'
  | 'RUNNING'
  | 'FOUND'
  | 'NOTFOUND'
  | 'ERROR';

export type SuspectConfidenceLevel =
  | 'SUSPECT_CONFIDENCE_LEVEL_UNSPECIFIED'
  | 'LOW'
  | 'MEDIUM'
  | 'HIGH';

export type CulpritActionType =
  | 'CULPRIT_ACTION_TYPE_UNSPECIFIED'
  | 'CULPRIT_AUTO_REVERTED'
  | 'REVERT_CL_CREATED'
  | 'CULPRIT_CL_COMMENTED'
  | 'BUG_COMMENTED';

export type BuildFailureType =
  | 'BUILD_FAILURE_TYPE'
  | 'COMPILE'
  | 'TEST'
  | 'INFRA'
  | 'OTHER';

export function isAnalysisComplete(status: AnalysisStatus) {
  const completeStatuses: Array<AnalysisStatus> = [
    'FOUND',
    'NOTFOUND',
    'ERROR',
  ];
  return completeStatuses.includes(status);
}

export interface QueryAnalysisRequest {
  buildFailure: BuildFailure;
}

export interface QueryAnalysisResponse {
  analyses: Analysis[];
}

export interface Analysis {
  analysisId: string;
  status: AnalysisStatus;
  lastPassedBbid: string;
  firstFailedBbid: string;
  heuristicResult?: HeuristicAnalysisResult;
  nthSectionResult?: NthSectionAnalysisResult;
  culprit?: GitilesCommit;
  culpritAction?: CulpritAction[];
  builder: BuilderID;
  buildFailureType: BuildFailureType;
}

export interface BuildFailure {
  bbid: string;
  failedStepName: string;
}

export interface BuilderID {
  project: string;
  bucket: string;
  builder: string;
}

export interface HeuristicAnalysisResult {
  status: AnalysisStatus;
  suspects?: HeuristicSuspect[];
}

export interface HeuristicSuspect {
  gitilesCommit: GitilesCommit;
  reviewUrl: string;
  score: string;
  justification: string;
  confidenceLevel: SuspectConfidenceLevel;
}

export interface NthSectionAnalysisResult {
  status: AnalysisStatus;
  culprit?: GitilesCommit;
  remainingNthSectionRange?: RegressionRange;
}

export interface GitilesCommit {
  host: string;
  project: string;
  id: string;
  ref: string;
  position: string;
}

export interface RegressionRange {
  lastPassed: GitilesCommit;
  firstFailed: GitilesCommit;
}

export interface CulpritAction {
  actionType: CulpritActionType;
  revertClUrl?: string;
  bugUrl: string;
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

export interface PrimeSuspect {
  cl: CL;
  culpritStatus: string;
  accuseSource: string;
}
