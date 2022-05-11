/* eslint-disable jest/no-conditional-expect */
// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.


import { impactFilterNamed, newMockFailure, newMockGroup } from '../testing_tools/mocks/failures_mock';
import {
  ClusterFailure,
  FailureGroup,
  groupFailures,
  rejectedIngestedInvocationIdsExtractor,
  rejectedPresubmitRunIdsExtractor,
  rejectedTestRunIdsExtractor,
  sortFailureGroups,
  treeDistinctValues,
} from './failures_tools';

interface ExtractorTestCase {
    failure: ClusterFailure;
    filter: string;
    shouldExtractTestRunId: boolean;
    shouldExtractIngestedInvocationId: boolean;
}

describe.each<ExtractorTestCase>([{
  failure: newMockFailure().build(),
  filter: 'Actual Impact',
  shouldExtractTestRunId: false,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().testRunBlocked().build(),
  filter: 'Actual Impact',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().ingestedInvocationBlocked().build(),
  filter: 'Actual Impact',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}, {
  failure: newMockFailure().exonerateOccursOnOtherCLs().build(),
  filter: 'Actual Impact',
  shouldExtractTestRunId: false,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().testRunBlocked().exonerateOccursOnOtherCLs().build(),
  filter: 'Actual Impact',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().ingestedInvocationBlocked().exonerateOccursOnOtherCLs().build(),
  filter: 'Actual Impact',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().exonerateNotCritical().build(),
  filter: 'Actual Impact',
  shouldExtractTestRunId: false,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().testRunBlocked().exonerateNotCritical().build(),
  filter: 'Actual Impact',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().ingestedInvocationBlocked().exonerateNotCritical().build(),
  filter: 'Actual Impact',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().build(),
  filter: 'Without Weetbix Exoneration',
  shouldExtractTestRunId: false,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().ingestedInvocationBlocked().build(),
  filter: 'Without Weetbix Exoneration',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}, {
  failure: newMockFailure().exonerateOccursOnOtherCLs().build(),
  filter: 'Without Weetbix Exoneration',
  shouldExtractTestRunId: false,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().ingestedInvocationBlocked().exonerateOccursOnOtherCLs().build(),
  filter: 'Without Weetbix Exoneration',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}, {
  failure: newMockFailure().exonerateNotCritical().build(),
  filter: 'Without Weetbix Exoneration',
  shouldExtractTestRunId: false,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().ingestedInvocationBlocked().exonerateNotCritical().build(),
  filter: 'Without Weetbix Exoneration',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().build(),
  filter: 'Without All Exoneration',
  shouldExtractTestRunId: false,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().ingestedInvocationBlocked().build(),
  filter: 'Without All Exoneration',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}, {
  failure: newMockFailure().exonerateOccursOnOtherCLs().build(),
  filter: 'Without All Exoneration',
  shouldExtractTestRunId: false,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().ingestedInvocationBlocked().exonerateOccursOnOtherCLs().build(),
  filter: 'Without All Exoneration',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}, {
  failure: newMockFailure().exonerateNotCritical().build(),
  filter: 'Without All Exoneration',
  shouldExtractTestRunId: false,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().ingestedInvocationBlocked().exonerateNotCritical().build(),
  filter: 'Without All Exoneration',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}, {
  failure: newMockFailure().build(),
  filter: 'Without Retrying Test Runs',
  shouldExtractTestRunId: false,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().testRunBlocked().build(),
  filter: 'Without Retrying Test Runs',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}, {
  failure: newMockFailure().ingestedInvocationBlocked().build(),
  filter: 'Without Retrying Test Runs',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}, {
  failure: newMockFailure().exonerateOccursOnOtherCLs().build(),
  filter: 'Without Retrying Test Runs',
  shouldExtractTestRunId: false,
  shouldExtractIngestedInvocationId: false,
}, {
  failure: newMockFailure().testRunBlocked().exonerateOccursOnOtherCLs().build(),
  filter: 'Without Retrying Test Runs',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}, {
  failure: newMockFailure().ingestedInvocationBlocked().exonerateOccursOnOtherCLs().build(),
  filter: 'Without Retrying Test Runs',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}, {
  failure: newMockFailure().build(),
  filter: 'Without Any Retries',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}, {
  failure: newMockFailure().testRunBlocked().build(),
  filter: 'Without Any Retries',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}, {
  failure: newMockFailure().ingestedInvocationBlocked().build(),
  filter: 'Without Any Retries',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}, {
  failure: newMockFailure().exonerateOccursOnOtherCLs().build(),
  filter: 'Without Any Retries',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}, {
  failure: newMockFailure().testRunBlocked().exonerateOccursOnOtherCLs().build(),
  filter: 'Without Any Retries',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}, {
  failure: newMockFailure().ingestedInvocationBlocked().exonerateOccursOnOtherCLs().build(),
  filter: 'Without Any Retries',
  shouldExtractTestRunId: true,
  shouldExtractIngestedInvocationId: true,
}])('Extractors with %j', (tc: ExtractorTestCase) => {
  it('should return ids in only the cases expected by failure type and impact filter.', () => {
    const testRunIds = rejectedTestRunIdsExtractor(impactFilterNamed(tc.filter))(tc.failure);
    if (tc.shouldExtractTestRunId) {
      expect(testRunIds.size).toBeGreaterThan(0);
    } else {
      expect(testRunIds.size).toBe(0);
    }
    const ingestedInvocationIds = rejectedIngestedInvocationIdsExtractor(impactFilterNamed(tc.filter))(tc.failure);
    if (tc.shouldExtractIngestedInvocationId) {
      expect(ingestedInvocationIds.size).toBeGreaterThan(0);
    } else {
      expect(ingestedInvocationIds.size).toBe(0);
    }
    const presubmitRunIds = rejectedPresubmitRunIdsExtractor(impactFilterNamed(tc.filter))(tc.failure);
    // presubmitRunId is extracted under exactly the same conditions as ingestedInvocationId.
    if (tc.shouldExtractIngestedInvocationId) {
      expect(presubmitRunIds.size).toBeGreaterThan(0);
    } else {
      expect(presubmitRunIds.size).toBe(0);
    }
  });
});

describe('groupFailures', () => {
  it('should put each failure in a separate group when given unique grouping keys', () => {
    const failures = [
      newMockFailure().build(),
      newMockFailure().build(),
      newMockFailure().build(),
    ];
    let unique = 0;
    const groups: FailureGroup[] = groupFailures(failures, () => ['' + unique++]);
    expect(groups.length).toBe(3);
    expect(groups[0].children.length).toBe(1);
  });
  it('should put each failure in a single group when given a single grouping key', () => {
    const failures = [
      newMockFailure().build(),
      newMockFailure().build(),
      newMockFailure().build(),
    ];
    const groups: FailureGroup[] = groupFailures(failures, () => ['group1']);
    expect(groups.length).toBe(1);
    expect(groups[0].children.length).toBe(3);
  });
  it('should put group failures into multiple levels', () => {
    const failures = [
      newMockFailure().withVariantGroups('v1', 'a').withVariantGroups('v2', 'a').build(),
      newMockFailure().withVariantGroups('v1', 'a').withVariantGroups('v2', 'b').build(),
      newMockFailure().withVariantGroups('v1', 'b').withVariantGroups('v2', 'a').build(),
      newMockFailure().withVariantGroups('v1', 'b').withVariantGroups('v2', 'b').build(),
    ];
    const groups: FailureGroup[] = groupFailures(failures, (f) => f.variant.map((v) => v.value || ''));
    expect(groups.length).toBe(2);
    expect(groups[0].children.length).toBe(2);
    expect(groups[1].children.length).toBe(2);
    expect(groups[0].children[0].children.length).toBe(1);
  });
});

describe('treeDistinctValues', () => {
  // A helper to just store the counts to the failures field.
  const setFailures = (g: FailureGroup, values: Set<string>) => {
    g.failures = values.size;
  };
  it('should have count of 1 for a valid feature', () => {
    const groups = groupFailures([newMockFailure().build()], () => ['group']);

    treeDistinctValues(groups[0], () => new Set(['a']), setFailures);

    expect(groups[0].failures).toBe(1);
  });
  it('should have count of 0 for an invalid feature', () => {
    const groups = groupFailures([newMockFailure().build()], () => ['group']);

    treeDistinctValues(groups[0], () => new Set(), setFailures);

    expect(groups[0].failures).toBe(0);
  });

  it('should have count of 1 for two identical features', () => {
    const groups = groupFailures([
      newMockFailure().build(),
      newMockFailure().build(),
    ], () => ['group']);

    treeDistinctValues(groups[0], () => new Set(['a']), setFailures);

    expect(groups[0].failures).toBe(1);
  });
  it('should have count of 2 for two different features', () => {
    const groups = groupFailures([
      newMockFailure().withTestId('a').build(),
      newMockFailure().withTestId('b').build(),
    ], () => ['group']);

    treeDistinctValues(groups[0], (f) => f.testId ? new Set([f.testId]) : new Set(), setFailures);

    expect(groups[0].failures).toBe(2);
  });
  it('should have count of 1 for two identical features in different subgroups', () => {
    const groups = groupFailures([
      newMockFailure().withTestId('a').withVariantGroups('group', 'a').build(),
      newMockFailure().withTestId('a').withVariantGroups('group', 'b').build(),
    ], (f) => ['top', ...f.variant.map((v) => v.value || '')]);

    treeDistinctValues(groups[0], (f) => f.testId ? new Set([f.testId]) : new Set(), setFailures);

    expect(groups[0].failures).toBe(1);
    expect(groups[0].children[0].failures).toBe(1);
    expect(groups[0].children[1].failures).toBe(1);
  });
  it('should have count of 2 for two different features in different subgroups', () => {
    const groups = groupFailures([
      newMockFailure().withTestId('a').withVariantGroups('group', 'a').build(),
      newMockFailure().withTestId('b').withVariantGroups('group', 'b').build(),
    ], (f) => ['top', ...f.variant.map((v) => v.value || '')]);

    treeDistinctValues(groups[0], (f) => f.testId ? new Set([f.testId]) : new Set(), setFailures);

    expect(groups[0].failures).toBe(2);
    expect(groups[0].children[0].failures).toBe(1);
    expect(groups[0].children[1].failures).toBe(1);
  });
});

describe('sortFailureGroups', () => {
  it('sorts top level groups ascending', () => {
    let groups: FailureGroup[] = [
      newMockGroup('c').withFailures(3).build(),
      newMockGroup('a').withFailures(1).build(),
      newMockGroup('b').withFailures(2).build(),
    ];

    groups = sortFailureGroups(groups, 'failures', true);

    expect(groups.map((g) => g.name)).toEqual(['a', 'b', 'c']);
  });
  it('sorts top level groups descending', () => {
    let groups: FailureGroup[] = [
      newMockGroup('c').withFailures(3).build(),
      newMockGroup('a').withFailures(1).build(),
      newMockGroup('b').withFailures(2).build(),
    ];

    groups = sortFailureGroups(groups, 'failures', false);

    expect(groups.map((g) => g.name)).toEqual(['c', 'b', 'a']);
  });
  it('sorts child groups', () => {
    let groups: FailureGroup[] = [
      newMockGroup('c').withFailures(3).build(),
      newMockGroup('a').withFailures(1).withChildren([
        newMockGroup('a3').withFailures(3).build(),
        newMockGroup('a2').withFailures(2).build(),
        newMockGroup('a1').withFailures(1).build(),
      ]).build(),
      newMockGroup('b').withFailures(2).build(),
    ];

    groups = sortFailureGroups(groups, 'failures', true);

    expect(groups.map((g) => g.name)).toEqual(['a', 'b', 'c']);
    expect(groups[0].children.map((g) => g.name)).toEqual(['a1', 'a2', 'a3']);
  });
  it('sorts on an alternate metric', () => {
    let groups: FailureGroup[] = [
      newMockGroup('c').withPresubmitRejects(3).build(),
      newMockGroup('a').withPresubmitRejects(1).build(),
      newMockGroup('b').withPresubmitRejects(2).build(),
    ];

    groups = sortFailureGroups(groups, 'presubmitRejects', true);

    expect(groups.map((g) => g.name)).toEqual(['a', 'b', 'c']);
  });
});
