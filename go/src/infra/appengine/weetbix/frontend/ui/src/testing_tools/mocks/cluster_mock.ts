// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import fetchMock from 'fetch-mock-jest';

import {
  Cluster,
  ClusterSummary,
  QueryClusterSummariesRequest,
  QueryClusterSummariesResponse,
  QueryClusterFailuresRequest,
  QueryClusterFailuresResponse,
  DistinctClusterFailure,
} from '../../services/cluster';

export const getMockCluster = (id: string): Cluster => {
  return {
    'name': `projects/testproject/clusters/rules-v2/${id}`,
    'hasExample': true,
    'title': '',
    'userClsFailedPresubmit': {
      'oneDay': { 'nominal': '98' },
      'threeDay': { 'nominal': '158' },
      'sevenDay': { 'nominal': '167' },
    },
    'criticalFailuresExonerated': {
      'oneDay': { 'nominal': '5625' },
      'threeDay': { 'nominal': '14052' },
      'sevenDay': { 'nominal': '13800' },
    },
    'failures': {
      'oneDay': { 'nominal': '7625' },
      'threeDay': { 'nominal': '16052' },
      'sevenDay': { 'nominal': '15800' },
    },
    'equivalentFailureAssociationRule': '',
  };
};

export const getMockRuleClusterSummary = (id: string): ClusterSummary => {
  return {
    'clusterId': {
      'algorithm': 'rules-v2',
      'id': id,
    },
    'title': 'reason LIKE "blah%"',
    'bug': {
      'system': 'buganizer',
      'id': '123456789',
      'linkText': 'b/123456789',
      'url': 'https://buganizer/123456789',
    },
    'presubmitRejects': '27',
    'criticalFailuresExonerated': '918',
    'failures': '1871',
  };
};

export const getMockSuggestedClusterSummary = (id: string): ClusterSummary => {
  return {
    'clusterId': {
      'algorithm': 'reason-v3',
      'id': id,
    },
    'bug': undefined,
    'title': 'reason LIKE "blah%"',
    'presubmitRejects': '29',
    'criticalFailuresExonerated': '919',
    'failures': '1872',
  };
};

export const mockQueryClusterSummaries = (request: QueryClusterSummariesRequest, response: QueryClusterSummariesResponse) => {
  fetchMock.post({
    url: 'http://localhost/prpc/weetbix.v1.Clusters/QueryClusterSummaries',
    body: request,
  }, {
    headers: {
      'X-Prpc-Grpc-Code': '0',
    },
    body: ')]}\'' + JSON.stringify(response),
  }, { overwriteRoutes: true });
};

export const mockQueryClusterFailures = (parent: string, failures: DistinctClusterFailure[] | undefined) => {
  const request: QueryClusterFailuresRequest = {
    parent: parent,
  };
  const response: QueryClusterFailuresResponse = {
    failures: failures,
  };
  fetchMock.post({
    url: 'http://localhost/prpc/weetbix.v1.Clusters/QueryClusterFailures',
    body: request,
  }, {
    headers: {
      'X-Prpc-Grpc-Code': '0',
    },
    body: ')]}\'' + JSON.stringify(response),
  }, { overwriteRoutes: true });
};
