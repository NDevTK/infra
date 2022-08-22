// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import dayjs from 'dayjs';
import { nanoid } from 'nanoid';
import { DistinctClusterFailure, Exoneration } from '../services/cluster';

/**
 * Creates a list of distinct variants found in the list of failures provided.
 *
 * @param {DistinctClusterFailure[]} failures the failures list.
 * @return {VariantGroup[]} A list of distinct variants.
 */
export const countDistictVariantValues = (failures: DistinctClusterFailure[]): VariantGroup[] => {
  if (!failures) {
    return [];
  }
  const variantGroups: VariantGroup[] = [];
  failures.forEach((failure) => {
    if (failure.variant === undefined) {
      return;
    }
    const def = failure.variant.def;
    for (const key in def) {
      if (!Object.prototype.hasOwnProperty.call(def, key)) {
        continue;
      }
      const value = def[key] || '';
      const variant = variantGroups.filter((e) => e.key === key)?.[0];
      if (!variant) {
        variantGroups.push({ key: key, values: [value], isSelected: false });
      } else {
        if (variant.values.indexOf(value) === -1) {
          variant.values.push(value);
        }
      }
    }
  });
  return variantGroups;
};

// group a number of failures into a tree of failure groups.
// grouper is a function that returns a list of keys, one corresponding to each level of the grouping tree.
// impactFilter controls how metric counts are aggregated from failures into parent groups (see treeCounts and rejected... functions).
export const groupFailures = (failures: DistinctClusterFailure[], grouper: (f: DistinctClusterFailure) => GroupKey[]): FailureGroup[] => {
  const topGroups: FailureGroup[] = [];
  const leafKey: GroupKey = { type: 'leaf', value: '' };
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
    const failureGroup = newGroup(leafKey, failureTime.toISOString());
    failureGroup.failure = f;
    failureGroup.level = level;
    groups.push(failureGroup);
  });
  return topGroups;
};

// Create a new group.
export const newGroup = (key: GroupKey, failureTime: string): FailureGroup => {
  return {
    id: key.value || nanoid(),
    key: key,
    criticalFailuresExonerated: 0,
    failures: 0,
    invocationFailures: 0,
    presubmitRejects: 0,
    children: [],
    isExpanded: false,
    latestFailureTime: failureTime,
    level: 0,
  };
};

// Find a group by key in the given list of groups, create a new one and insert it if it is not found.
// failureTime is only used when creating a new group.
export const getOrCreateGroup = (
    groups: FailureGroup[], key: GroupKey, failureTime: string,
): FailureGroup => {
  let group = groups.filter((g) => keyEqual(g.key, key))?.[0];
  if (group) {
    return group;
  }
  group = newGroup(key, failureTime);
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
export type FeatureExtractor = (failure: DistinctClusterFailure) => Set<string>;

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

// criticalFailuresExoneratedIdsExtractor returns an extractor that returns
// a unique failure id for each failure of a critical test that is exonerated.
// As failures don't actually have ids, it just returns an incrementing integer.
export const criticalFailuresExoneratedIdsExtractor = (): FeatureExtractor => {
  let unique = 0;
  return (f) => {
    const values: Set<string> = new Set();
    if (!f.isBuildCritical) {
      return values;
    }
    let exoneratedByCQ = false;
    if (f.exonerations != null) {
      for (let i = 0; i < f.exonerations.length; i++) {
        // Do not count the exoneration reason NOT_CRITICAL
        // (as it implies the test is not critical), or the
        // exoneration reason UNEXPECTED_PASS as the test is considered
        // passing.
        if (f.exonerations[i].reason == 'OCCURS_ON_MAINLINE' ||
              f.exonerations[i].reason == 'OCCURS_ON_OTHER_CLS') {
          exoneratedByCQ = true;
        }
      }
    }
    if (!exoneratedByCQ) {
      return values;
    }

    for (let i = 0; i < f.count; i++) {
      unique += 1;
      values.add('' + unique);
    }
    return values;
  };
};

// Returns whether the failure was exonerated for a reason other than it occurred
// on other CLs or on mainline.
const isExoneratedByNonWeetbix = (exonerations: Exoneration[] | undefined): boolean => {
  if (exonerations === undefined) {
    return false;
  }
  let hasOtherExoneration = false;
  for (let i = 0; i < exonerations.length; i++) {
    if (exonerations[i].reason != 'OCCURS_ON_MAINLINE' &&
          exonerations[i].reason != 'OCCURS_ON_OTHER_CLS') {
      hasOtherExoneration = true;
    }
  }
  return hasOtherExoneration;
};

// Returns an extractor that returns the id of the ingested invocation that was rejected by this failure, if any.
// The impact filter is taken into account in determining if the invocation was rejected by this failure.
export const rejectedIngestedInvocationIdsExtractor = (impactFilter: ImpactFilter): FeatureExtractor => {
  return (failure) => {
    const values: Set<string> = new Set();
    // If neither Weetbix nor all exoneration is ignored, we want actual impact.
    // This requires exclusion of all exonerated test results, as well as
    // test results from builds which passed (which implies the test results
    // could not have caused the presubmit run to fail).
    if (((failure.exonerations !== undefined && failure.exonerations.length > 0) || failure.buildStatus != 'BUILD_STATUS_FAILURE') &&
                !(impactFilter.ignoreWeetbixExoneration || impactFilter.ignoreAllExoneration)) {
      return values;
    }
    // If not all exoneration is ignored, it means we want actual or without weetbix impact.
    // All exonerations not made by weetbix should be applied, those made by Weetbix should not
    // be applied (or will have already been applied).
    if (isExoneratedByNonWeetbix(failure.exonerations) &&
        !impactFilter.ignoreAllExoneration) {
      return values;
    }
    if (!failure.isIngestedInvocationBlocked && !impactFilter.ignoreIngestedInvocationBlocked) {
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
    // If neither Weetbix nor all exoneration is ignored, we want actual impact.
    // This requires exclusion of all exonerated test results, as well as
    // test results from builds which passed (which implies the test results
    // could not have caused the presubmit run to fail).
    if (((failure.exonerations !== undefined && failure.exonerations.length > 0) || failure.buildStatus != 'BUILD_STATUS_FAILURE') &&
                !(impactFilter.ignoreWeetbixExoneration || impactFilter.ignoreAllExoneration)) {
      return values;
    }
    // If not all exoneration is ignored, it means we want actual or without weetbix impact.
    // All test results exonerated, but not exonerated by weetbix should be ignored.
    if (isExoneratedByNonWeetbix(failure.exonerations) &&
        !impactFilter.ignoreAllExoneration) {
      return values;
    }
    if (!failure.isIngestedInvocationBlocked && !impactFilter.ignoreIngestedInvocationBlocked) {
      return values;
    }
    if (failure.changelists !== undefined && failure.changelists.length > 0 &&
        failure.presubmitRun !== undefined && failure.presubmitRun.owner == 'user' &&
        failure.isBuildCritical && failure.presubmitRun.mode == 'FULL_RUN') {
      values.add(failure.changelists[0].host + '/' + failure.changelists[0].change);
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
      case 'criticalFailuresExonerated':
        return group.criticalFailuresExonerated;
      case 'failures':
        return group.failures;
      case 'invocationFailures':
        return group.invocationFailures;
      case 'presubmitRejects':
        return group.presubmitRejects;
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
 * @param {DistinctClusterFailure} failures The list of failures to group.
 * @param {VariantGroup} variantGroups The list of variant groups to use for grouping.
 * @param {FailureFilter} failureFilter The failure filter to filter out the failures.
 * @return {FailureGroup[]} The list of failures grouped by the variants.
 */
export const groupAndCountFailures = (
    failures: DistinctClusterFailure[],
    variantGroups: VariantGroup[],
    failureFilter: FailureFilter,
): FailureGroup[] => {
  if (failures) {
    let currentFailures = failures;
    if (failureFilter == 'Presubmit Failures') {
      currentFailures = failures.filter((f) => f.presubmitRun);
    } else if (failureFilter == 'Postsubmit Failures') {
      currentFailures = failures.filter((f) => !f.presubmitRun);
    }
    const groups = groupFailures(currentFailures, (failure) => {
      const variantValues = variantGroups.filter((v) => v.isSelected)
          .map((v) => {
            const key: GroupKey = { type: 'variant', key: v.key, value: failure.variant?.def[v.key] || '' };
            return key;
          });
      return [...variantValues, { type: 'test', value: failure.testId || '' }];
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
        group, criticalFailuresExoneratedIdsExtractor(), (g, values) => g.criticalFailuresExonerated = values.size,
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
}
export const ImpactFilters: ImpactFilter[] = [
  {
    name: 'Actual Impact',
    ignoreWeetbixExoneration: false,
    ignoreAllExoneration: false,
    ignoreIngestedInvocationBlocked: false,
  }, {
    name: 'Without Weetbix Exoneration',
    ignoreWeetbixExoneration: true,
    ignoreAllExoneration: false,
    ignoreIngestedInvocationBlocked: false,
  }, {
    name: 'Without All Exoneration',
    ignoreWeetbixExoneration: true,
    ignoreAllExoneration: true,
    ignoreIngestedInvocationBlocked: false,
  }, {
    name: 'Without Any Retries',
    ignoreWeetbixExoneration: true,
    ignoreAllExoneration: true,
    ignoreIngestedInvocationBlocked: true,
  },
];

export const defaultImpactFilter: ImpactFilter = ImpactFilters[0];

// Metrics that can be used for sorting FailureGroups.
// Each value is a property of FailureGroup.
export type MetricName = 'presubmitRejects' | 'invocationFailures' | 'criticalFailuresExonerated' | 'failures' | 'latestFailureTime';

export type GroupType = 'test' | 'variant' | 'leaf';

export interface GroupKey {
  // The type of group.
  // This could be a group for a test, for a variant value,
  // or a leaf (for individual failures).
  type: GroupType;

  // For variant-based grouping keys, the name of the variant.
  // Unspecified otherwise.
  key?: string;

  // The name of the group. E.g. the name of the test or the variant value.
  // May be empty for leaf nodes.
  value: string;
}

export const keyEqual = (a: GroupKey, b: GroupKey) => {
  return a.value === b.value && a.key === b.key && a.type === b.type;
};

// FailureGroups are nodes in the failure tree hierarchy.
export interface FailureGroup {
    id: string;
    key: GroupKey;
    criticalFailuresExonerated: number;
    failures: number;
    invocationFailures: number;
    presubmitRejects: number;
    latestFailureTime: string;
    level: number;
    children: FailureGroup[];
    isExpanded: boolean;
    failure?: DistinctClusterFailure;
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
