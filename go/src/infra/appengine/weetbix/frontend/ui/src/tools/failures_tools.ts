// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import dayjs from 'dayjs';
import { nanoid } from 'nanoid';

/**
 * Creates a list of distinct variants found in the list of failures provided.
 *
 * @param {ClusterFailure[]} failures the failures list.
 * @return {VariantGroup[]} A list of distinct variants.
 */
export const countDistictVariantValues = (failures: ClusterFailure[]): VariantGroup[] => {
  if (!failures) {
    return [];
  }
  const variantGroups: VariantGroup[] = [];
  failures.forEach((failure) => {
    failure.variant.forEach((v) => {
      if (!v.key) {
        return;
      }
      const variant = variantGroups.filter((e) => e.key === v.key)?.[0];
      if (!variant) {
        variantGroups.push({ key: v.key, values: [v.value || ''], isSelected: false });
      } else {
        if (variant.values.indexOf(v.value || '') === -1) {
          variant.values.push(v.value || '');
        }
      }
    });
  });
  return variantGroups;
};

// group a number of failures into a tree of failure groups.
// grouper is a function that returns a list of keys, one corresponding to each level of the grouping tree.
// impactFilter controls how metric counts are aggregated from failures into parent groups (see treeCounts and rejected... functions).
export const groupFailures = (failures: ClusterFailure[], grouper: (f: ClusterFailure) => string[]): FailureGroup[] => {
  const topGroups: FailureGroup[] = [];
  failures.forEach((f) => {
    const keys = grouper(f);
    let groups = topGroups;
    const failureTime = dayjs(f.partitionTime);
    let level = 0;
    for (const key of keys) {
      const group = getOrCreateGroup(
          groups, key, failureTime.toISOString(),
      );
      group.level = level;
      level += 1;
      groups = group.children;
    }
    const failureGroup = newGroup('', failureTime.toISOString());
    failureGroup.failure = f;
    failureGroup.level = level;
    groups.push(failureGroup);
  });
  return topGroups;
};

// Create a new group.
export const newGroup = (name: string, failureTime: string): FailureGroup => {
  return {
    id: name || nanoid(),
    name: name,
    failures: 0,
    invocationFailures: 0,
    testRunFailures: 0,
    presubmitRejects: 0,
    children: [],
    isExpanded: false,
    latestFailureTime: failureTime,
    level: 0,
  };
};

// Find a group by name in the given list of groups, create a new one and insert it if it is not found.
// failureTime is only used when creating a new group.
export const getOrCreateGroup = (
    groups: FailureGroup[], name: string, failureTime: string,
): FailureGroup => {
  let group = groups.filter((g) => g.name == name)?.[0];
  if (group) {
    return group;
  }
  group = newGroup(name, failureTime);
  groups.push(group);
  return group;
};

// Returns the distinct values returned by featureExtractor for all children of the group.
// If featureExtractor returns undefined, the failure will be ignored.
// The distinct values for each group in the tree are also reported to `visitor` as the tree is traversed.
// A typical `visitor` function will store the count of distinct values in a property of the group.
export const treeDistinctValues = (
    group: FailureGroup,
    featureExtractor: FeatureExtractor,
    visitor: (group: FailureGroup, distinctValues: Set<string>) => void,
): Set<string> => {
  const values: Set<string> = new Set();
  if (group.failure) {
    for (const value of featureExtractor(group.failure)) {
      values.add(value);
    }
  } else {
    for (const child of group.children) {
      for (const value of treeDistinctValues(
          child, featureExtractor, visitor,
      )) {
        values.add(value);
      }
    }
  }
  visitor(group, values);
  return values;
};

// A FeatureExtractor returns a string representing some feature of a ClusterFailure.
// Returns undefined if there is no such feature for this failure.
export type FeatureExtractor = (failure: ClusterFailure) => Set<string>;

// failureIdExtractor returns an extractor that returns a unique failure id for each failure.
// As failures don't actually have ids, it just returns an incrementing integer.
export const failureIdsExtractor = (): FeatureExtractor => {
  let unique = 0;
  return (f) => {
    const values: Set<string> = new Set();
    for (let i = 0; i < f.count; i++) {
      unique += 1;
      values.add('' + unique);
    }
    return values;
  };
};

// Returns an extractor that returns the id of the test run that was rejected by this failure, if any.
// The impact filter is taken into account in determining if the run was rejected by this failure.
export const rejectedTestRunIdsExtractor = (impactFilter: ImpactFilter): FeatureExtractor => {
  return (f) => {
    const values: Set<string> = new Set();
    if (!impactFilter.ignoreTestRunBlocked && !f.isTestRunBlocked) {
      return values;
    }
    for (const testRunId of f.testRunIds) {
      if (testRunId) {
        values.add(testRunId);
      }
    }
    return values;
  };
};

// Returns whether the failure was exonerated because it occurs on other
// CLs or on mainline. These are exoneration types that appear only in CQ.
const isExoneratedByOccuranceElsewhere = (status: ExonerationStatus | null): boolean => {
  return status == 'WEETBIX' || // Deprecated. Can be removed from June 2022.
            status == 'OCCURS_ON_OTHER_CLS' ||
            status == 'OCCURS_ON_MAINLINE';
};

// Returns an extractor that returns the id of the ingested invocation that was rejected by this failure, if any.
// The impact filter is taken into account in determining if the invocation was rejected by this failure.
export const rejectedIngestedInvocationIdsExtractor = (impactFilter: ImpactFilter): FeatureExtractor => {
  return (failure) => {
    const values: Set<string> = new Set();
    if (isExoneratedByOccuranceElsewhere(failure.exonerationStatus) &&
                !(impactFilter.ignoreWeetbixExoneration || impactFilter.ignoreAllExoneration)) {
      return values;
    }
    if (failure.exonerationStatus != 'NOT_EXONERATED' &&
        !isExoneratedByOccuranceElsewhere(failure.exonerationStatus) &&
         !impactFilter.ignoreAllExoneration) {
      return values;
    }
    if (!failure.isIngestedInvocationBlocked && !impactFilter.ignoreIngestedInvocationBlocked) {
      return values;
    }
    if (!impactFilter.ignoreTestRunBlocked && !failure.isTestRunBlocked) {
      return values;
    }
    if (failure.ingestedInvocationId) {
      values.add(failure.ingestedInvocationId);
    }
    return values;
  };
};

// Returns an extractor that returns the identity of the CL that was rejected by this failure, if any.
// The impact filter is taken into account in determining if the CL was rejected by this failure.
export const rejectedPresubmitRunIdsExtractor = (impactFilter: ImpactFilter): FeatureExtractor => {
  return (failure) => {
    const values: Set<string> = new Set();
    if (isExoneratedByOccuranceElsewhere(failure.exonerationStatus) &&
                !(impactFilter.ignoreWeetbixExoneration || impactFilter.ignoreAllExoneration)) {
      return values;
    }
    if ((failure.exonerationStatus != 'NOT_EXONERATED' && !isExoneratedByOccuranceElsewhere(failure.exonerationStatus)) &&
                        !impactFilter.ignoreAllExoneration) {
      return values;
    }
    if (!failure.isIngestedInvocationBlocked && !impactFilter.ignoreIngestedInvocationBlocked) {
      return values;
    }
    if (!impactFilter.ignoreTestRunBlocked && !failure.isTestRunBlocked) {
      return values;
    }
    if (failure.presubmitRunCl && failure.presubmitRunOwner == 'user') {
      values.add(failure.presubmitRunCl.host + '/' + failure.presubmitRunCl.change.toFixed(0));
    }
    return values;
  };
};

// Sorts child failure groups at each node of the tree by the given metric.
export const sortFailureGroups = (
    groups: FailureGroup[],
    metric: MetricName,
    ascending: boolean,
): FailureGroup[] => {
  const cloneGroups = [...groups];
  const getMetric = (group: FailureGroup): number => {
    switch (metric) {
      case 'failures':
        return group.failures;
      case 'presubmitRejects':
        return group.presubmitRejects;
      case 'invocationFailures':
        return group.invocationFailures;
      case 'testRunFailures':
        return group.testRunFailures;
      case 'latestFailureTime':
        return dayjs(group.latestFailureTime).unix();
      default:
        throw new Error('unknown metric: ' + metric);
    }
  };
  cloneGroups.sort((a, b) => ascending ? (getMetric(a) - getMetric(b)) : (getMetric(b) - getMetric(a)));
  for (const group of cloneGroups) {
    if (group.children.length > 0) {
      group.children = sortFailureGroups(group.children, metric, ascending);
    }
  }
  return cloneGroups;
};

/**
 * Groups failures by the variant groups selected.
 *
 * @param {ClusterFailure} failures The list of  failures to group.
 * @param {VariantGroup} variantGroups The list of variant groups to use for grouping.
 * @param {FailureFilter} failureFilter The failure filter to filter out the failures.
 * @return {FailureGroup[]} The list of failures grouped by the variants.
 */
export const groupAndCountFailures = (
    failures: ClusterFailure[],
    variantGroups: VariantGroup[],
    failureFilter: FailureFilter,
): FailureGroup[] => {
  if (failures) {
    let currentFailures = failures;
    if (failureFilter == 'Presubmit Failures') {
      currentFailures = failures.filter((f) => f.presubmitRunId);
    } else if (failureFilter == 'Postsubmit Failures') {
      currentFailures = failures.filter((f) => !f.presubmitRunId);
    }
    const groups = groupFailures(currentFailures, (failure) => {
      const variantValues = variantGroups.filter((v) => v.isSelected)
          .map((v) => failure.variant.filter((fv) => fv.key === v.key)?.[0]?.value || '');
      return [...variantValues, failure.testId || ''];
    });
    return groups;
  }
  return [];
};

export const countAndSortFailures = (groups: FailureGroup[], impactFilter: ImpactFilter): FailureGroup[] => {
  const groupsClone = [...groups];
  groupsClone.forEach((group) => {
    treeDistinctValues(
        group, failureIdsExtractor(), (g, values) => g.failures = values.size,
    );
    treeDistinctValues(
        group, rejectedTestRunIdsExtractor(impactFilter), (g, values) => g.testRunFailures = values.size,
    );
    treeDistinctValues(
        group, rejectedIngestedInvocationIdsExtractor(impactFilter), (g, values) => g.invocationFailures = values.size,
    );
    treeDistinctValues(
        group, rejectedPresubmitRunIdsExtractor(impactFilter), (g, values) => g.presubmitRejects = values.size,
    );
  });
  return groupsClone;
};

// ImpactFilter represents what kind of impact should be counted or ignored in
// calculating impact for failures.
export interface ImpactFilter {
    name: string;
    ignoreWeetbixExoneration: boolean;
    ignoreAllExoneration: boolean;
    ignoreIngestedInvocationBlocked: boolean;
    ignoreTestRunBlocked: boolean;
}
export const ImpactFilters: ImpactFilter[] = [
  {
    name: 'Actual Impact',
    ignoreWeetbixExoneration: false,
    ignoreAllExoneration: false,
    ignoreIngestedInvocationBlocked: false,
    ignoreTestRunBlocked: false,
  }, {
    name: 'Without Weetbix Exoneration',
    ignoreWeetbixExoneration: true,
    ignoreAllExoneration: false,
    ignoreIngestedInvocationBlocked: false,
    ignoreTestRunBlocked: false,
  }, {
    name: 'Without All Exoneration',
    ignoreWeetbixExoneration: true,
    ignoreAllExoneration: true,
    ignoreIngestedInvocationBlocked: false,
    ignoreTestRunBlocked: false,
  }, {
    name: 'Without Retrying Test Runs',
    ignoreWeetbixExoneration: true,
    ignoreAllExoneration: true,
    ignoreIngestedInvocationBlocked: true,
    ignoreTestRunBlocked: false,
  }, {
    name: 'Without Any Retries',
    ignoreWeetbixExoneration: true,
    ignoreAllExoneration: true,
    ignoreIngestedInvocationBlocked: true,
    ignoreTestRunBlocked: true,
  },
];

export const defaultImpactFilter: ImpactFilter = ImpactFilters[1];

// Test result was no exonerated.
type ExonerationStatus =
      // The test was not exonerated.
    'NOT_EXONERATED'
    // Despite having an unexpected result, and no exoneration recorded
    // in ResultDB, the build did not end in the state "failed"
    // (e.g. the status was succeeded, cancelled or infra failure).
    // The test result is implicitly exonerated.
    | 'IMPLICIT'
    // DEPRECATED. The test was marked exonerated in ResultDB, for a reason
    // other than Weetbix or FindIt failure analysis.
    | 'EXPLICIT'
    // DEPRECATED. Test result was recorded as exonerated
    // based on Weetbix (or FindIt) data.
    | 'WEETBIX'
    // Similar unexpected results were observed in presubmit run(s) for other,
    // unrelated CL(s). (This is suggestive of the issue being present
    // on mainline but is not confirmed as there are possible confounding
    // factors, like how tests are run on CLs vs how tests are run on
    // mainline branches.)
    // Applies to unexpected results in presubmit/CQ runs only.
    | 'OCCURS_ON_OTHER_CLS'
    // Similar unexpected results were observed on a mainline branch
    // (i.e. against a build without unsubmitted changes applied).
    // (For avoidance of doubt, this includes both flakily and
    // deterministically occurring unexpected results.)
    // Applies to unexpected results in presubmit/CQ runs only.
    | 'OCCURS_ON_MAINLINE'
    // The tests are not critical to the test subject (e.g. CL) passing.
    // This could be because more data is being collected to determine if
    // the tests are stable enough to be made critical (as is often the
    // case for experimental test suites).
    | 'NOT_CRITICAL'
    // The test result was an unexpected pass.
    | 'UNEXPECTED_PASS'
    // The test was marked exonerated in ResultDB, but a machine-understandable
    // reason for the exoneration is not available.
    | 'OTHER_EXPLICIT';

// ClusterFailure is the data returned by the server for each failure.
export interface ClusterFailure {
    realm: string | null;
    testId: string | null;
    variant: Variant[];
    presubmitRunCl: Changelist | null;
    presubmitRunId: PresubmitRunId | null;
    presubmitRunOwner: string | null;
    partitionTime: string | null;
    exonerationStatus: ExonerationStatus | null;
    ingestedInvocationId: string | null;
    isIngestedInvocationBlocked: boolean | null;
    testRunIds: Array<string | null>;
    isTestRunBlocked: boolean | null;
    count: number;
}

// Key/Value Variant pairs for failures.
export interface Variant {
    key: string | null;
    value: string | null;
}

// Presubmit Run Ids of failures returned from the server.
export interface PresubmitRunId {
    system: string | null;
    id: string | null;
}

// Changelist represents a gerrit patchset.
export interface Changelist {
    host: string;
    change: number;
    patchset: number;
}

// Metrics that can be used for sorting FailureGroups.
// Each value is a property of FailureGroup.
export type MetricName = 'presubmitRejects' | 'invocationFailures' | 'testRunFailures' | 'failures' | 'latestFailureTime';

// FailureGroups are nodes in the failure tree hierarchy.
export interface FailureGroup {
    id: string;
    name: string;
    presubmitRejects: number;
    invocationFailures: number;
    testRunFailures: number;
    failures: number;
    latestFailureTime: string;
    level: number;
    children: FailureGroup[];
    isExpanded: boolean;
    failure?: ClusterFailure;
}

// VariantGroup represents variant key that appear on at least one failure.
export interface VariantGroup {
    key: string;
    values: string[];
    isSelected: boolean;
}

export const FailureFilters = ['All Failures', 'Presubmit Failures', 'Postsubmit Failures'] as const;
export type FailureFilter = typeof FailureFilters[number];
export const defaultFailureFilter: FailureFilter = FailureFilters[0];
