// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { AuthorizedPrpcClient } from '../clients/authorized_client';

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
