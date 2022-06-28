// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Cluster } from '../../services/cluster';

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
