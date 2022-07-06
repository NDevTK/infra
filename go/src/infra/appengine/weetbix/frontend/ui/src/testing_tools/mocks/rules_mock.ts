// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import fetchMock from 'fetch-mock-jest';

import { ListRulesResponse } from '../../services/rules';
import { createDefaultMockRule } from './rule_mock';

export const createDefaultMockListRulesResponse = (): ListRulesResponse => {
  const rule1 = createDefaultMockRule();
  rule1.name = 'projects/chromium/rules/ce83f8395178a0f2edad59fc1a160001';
  rule1.ruleId = 'ce83f8395178a0f2edad59fc1a160001';
  rule1.bug = {
    system: 'monorail',
    id: 'chromium/90001',
    linkText: 'crbug.com/90001',
    url: 'https://monorail-staging.appspot.com/p/chromium/issues/detail?id=90001',
  };
  rule1.ruleDefinition = 'test LIKE "rule1%"';
  const rule2 = createDefaultMockRule();
  rule2.name = 'projects/chromium/rules/ce83f8395178a0f2edad59fc1a160002';
  rule2.ruleId = 'ce83f8395178a0f2edad59fc1a160002';
  rule2.bug = {
    system: 'monorail',
    id: 'chromium/90002',
    linkText: 'crbug.com/90002',
    url: 'https://monorail-staging.appspot.com/p/chromium/issues/detail?id=90002',
  };
  rule2.ruleDefinition = 'reason LIKE "rule2%"';
  return {
    rules: [rule1, rule2],
  };
};

export const mockFetchRules = () => {
  fetchMock.post('http://localhost/prpc/weetbix.v1.Rules/List', {
    headers: {
      'X-Prpc-Grpc-Code': '0',
    },
    body: ')]}\'' + JSON.stringify(createDefaultMockListRulesResponse()),
  });
};
