// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { AuthorizedPrpcClient } from '../clients/authorized_client';
import { AssociatedBug, ClusterId } from './shared_models';

export const getClustersService = () => {
  const client = new AuthorizedPrpcClient();
  return new ClustersService(client);
};

// A service to handle cluster-related gRPC requests.
export class ClustersService {
  private static SERVICE = 'weetbix.v1.Clusters';

  client: AuthorizedPrpcClient;

  constructor(client: AuthorizedPrpcClient) {
    this.client = client;
  }

  async batchGet(request: BatchGetClustersRequest): Promise<BatchGetClustersResponse> {
    return this.client.call(ClustersService.SERVICE, 'BatchGet', request);
  }

  async getReclusteringProgress(request: GetReclusteringProgressRequest): Promise<ReclusteringProgress> {
    return this.client.call(ClustersService.SERVICE, 'GetReclusteringProgress', request);
  }

  async queryClusterSummaries(request: QueryClusterSummariesRequest): Promise<QueryClusterSummariesResponse> {
    return this.client.call(ClustersService.SERVICE, 'QueryClusterSummaries', request);
  }
}

export interface BatchGetClustersRequest {
  // The LUCI project shared by all clusters to retrieve.
  // Required.
  // Format: projects/{project}.
  parent: string;

  // The resource name of the clusters retrieve.
  // Format: projects/{project}/clusters/{cluster_algorithm}/{cluster_id}.
  // At most 1,000 clusters may be requested at a time.
  names: string[];
}

export interface BatchGetClustersResponse {
  clusters: Cluster[] | undefined;
}

export interface Cluster {
  // The resource name of the cluster.
  // Format: projects/{project}/clusters/{cluster_algorithm}/{cluster_id}.
  name: string;
  // Whether there is a recent example in the cluster.
  hasExample: boolean | undefined;
  // A human-readable name for the cluster.
  // Only populated for suggested clusters where has_example = true.
  title: string | undefined;
  // The total number of user changelists which failed presubmit.
  userClsFailedPresubmit: MetricValues;
  // The total number of failures in the cluster that occurred on tryjobs
  // that were critical (presubmit-blocking) and were exonerated for a
  // reason other than NOT_CRITICAL or UNEXPECTED_PASS.
  criticalFailuresExonerated: MetricValues;
  // The total number of failures in the cluster.
  failures: MetricValues;
  // The failure association rule equivalent to the cluster. Populated only
  // for suggested clusters where has_example = true; for rule-based
  // clusters, lookup the rule instead. Used to facilitate creating a new
  // rule based on this cluster.
  equivalentFailureAssociationRule: string | undefined;
}

export interface MetricValues {
  // The impact for the last day.
  oneDay: Counts;
  // The impact for the last three days.
  threeDay: Counts;
  // The impact for the last week.
  sevenDay: Counts;
}

export interface Counts {
  // The value of the metric (summed over all failures).
  // 64-bit integer serialized as a string.
  nominal: string | undefined;
}

export interface GetReclusteringProgressRequest {
  // The name of the reclustering progress resource.
  // Format: projects/{project}/reclusteringProgress.
  name: string;
}

// ReclusteringProgress captures the progress re-clustering a
// given LUCI project's test results with a specific rules
// version and/or algorithms version.
export interface ReclusteringProgress {
  // ProgressPerMille is the progress of the current re-clustering run,
  // measured in thousandths (per mille).
  progressPerMille: number | undefined;
  // Last is the goal of the last completed re-clustering run.
  last: ClusteringVersion;
  // Next is the goal of the current re-clustering run. (For which
  // ProgressPerMille is specified.)
  // It may be the same as the goal of the last completed reclustering run.
  next: ClusteringVersion;
}

// ClusteringVersion captures the rules and algorithms a re-clustering run
// is re-clustering to.
export interface ClusteringVersion {
  rulesVersion: string; // RFC 3339 encoded date/time.
  configVersion: string; // RFC 3339 encoded date/time.
  algorithmsVersion: number;
}

export interface QueryClusterSummariesRequest {
  // The LUCI project.
  project: string;

  // An AIP-160 style filter on the failures that are used as input to
  // clustering.
  failureFilter: string;

  // An AIP-132 style order_by clause, which specifies the sort order
  // of the result.
  orderBy: string;
}

export type SortableMetricName = 'presubmit_rejects' | 'critical_failures_exonerated' | 'failures';

export interface QueryClusterSummariesResponse {
  clusterSummaries: ClusterSummary[] | null;
}

export interface ClusterSummary {
  // The identity of the cluster.
  clusterId: ClusterId;
  // A one-line description of the cluster.
  title: string;
  // The bug associated with the cluster. This is only present for
  // clusters defined by failure association rules.
  bug: AssociatedBug | undefined;
  // The number of distinct user CLs rejected by the cluster.
  // 64-bit integer serialized as a string.
  presubmitRejects: string | undefined;
  presubmitRejectsByDay: string[] | undefined;
  // The number of failures that were critical (on builders critical
  // to CQ succeeding and not exonerated for non-criticality)
  // and exonerated.
  // 64-bit integer serialized as a string.
  criticalFailuresExonerated: string | undefined;
  criticalFailuresExoneratedByDay: string[] | undefined;
  // The total number of test results in the cluster.
  // 64-bit integer serialized as a string.
  failures: string | undefined;
  failuresByDay: string[] | undefined;
}
