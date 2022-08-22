// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import dayjs from 'dayjs';
import { DistinctClusterFailure } from '../../services/cluster';

import {
  FailureGroup,
  GroupKey,
  VariantGroup,
  ImpactFilter,
  ImpactFilters,
} from '../../tools/failures_tools';

class ClusterFailureBuilder {
  failure: DistinctClusterFailure;
  constructor() {
    this.failure = {
      testId: 'ninja://dir/test.param',
      variant: undefined,
      presubmitRun: {
        presubmitRunId: { system: 'cv', id: 'presubmitRunId' },
        owner: 'user',
        mode: 'FULL_RUN',
      },
      changelists: [{
        host: 'clproject-review.googlesource.com',
        change: '123456',
        patchset: 7,
      }],
      partitionTime: '2021-05-12T19:05:34',
      exonerations: undefined,
      buildStatus: 'BUILD_STATUS_SUCCESS',
      isBuildCritical: true,
      ingestedInvocationId: 'ingestedInvocationId',
      isIngestedInvocationBlocked: false,
      count: 1,
    };
  }
  build(): DistinctClusterFailure {
    return this.failure;
  }
  ingestedInvocationBlocked() {
    this.failure.isIngestedInvocationBlocked = true;
    return this;
  }
  notPresubmitCritical() {
    this.failure.isBuildCritical = false;
    return this;
  }
  buildFailed() {
    this.failure.buildStatus = 'BUILD_STATUS_FAILURE';
    return this;
  }
  dryRun() {
    this.failure.presubmitRun = {
      presubmitRunId: { system: 'cv', id: 'presubmitRunId' },
      owner: 'user',
      mode: 'DRY_RUN',
    };
    return this;
  }
  exonerateOccursOnOtherCLs() {
    this.failure.exonerations = [];
    this.failure.exonerations.push({ reason: 'OCCURS_ON_OTHER_CLS' });
    return this;
  }
  exonerateNotCritical() {
    this.failure.exonerations = [];
    this.failure.exonerations.push({ reason: 'NOT_CRITICAL' });
    return this;
  }
  withVariantGroups(key: string, value: string) {
    if (this.failure.variant === undefined) {
      this.failure.variant = { def: {} };
    }
    this.failure.variant.def[key] = value;
    return this;
  }
  withTestId(id: string) {
    this.failure.testId = id;
    return this;
  }
  withoutPresubmit() {
    this.failure.changelists = undefined;
    this.failure.presubmitRun = undefined;
    return this;
  }
}

export const newMockGroup = (key: GroupKey): FailureGroupBuilder => {
  return new FailureGroupBuilder(key);
};

class FailureGroupBuilder {
  failureGroup: FailureGroup;
  constructor(key: GroupKey) {
    this.failureGroup = {
      id: key.value,
      key,
      children: [],
      criticalFailuresExonerated: 0,
      failures: 0,
      invocationFailures: 0,
      presubmitRejects: 0,
      latestFailureTime: dayjs().toISOString(),
      isExpanded: false,
      level: 0,
      failure: undefined,
    };
  }

  build(): FailureGroup {
    return this.failureGroup;
  }

  withFailures(failures: number) {
    this.failureGroup.failures = failures;
    return this;
  }

  withPresubmitRejects(presubmitRejects: number) {
    this.failureGroup.presubmitRejects = presubmitRejects;
    return this;
  }

  withCriticalFailuresExonerated(criticalFailuresExonerated: number) {
    this.failureGroup.criticalFailuresExonerated = criticalFailuresExonerated;
    return this;
  }

  withInvocationFailures(invocationFailures: number) {
    this.failureGroup.invocationFailures =invocationFailures;
    return this;
  }

  withFailure(failure: DistinctClusterFailure) {
    this.failureGroup.failure = failure;
    return this;
  }

  withChildren(children: FailureGroup[]) {
    this.failureGroup.children = children;
    return this;
  }
}

// Helper functions.
export const impactFilterNamed = (name: string) => {
  return ImpactFilters.filter((f: ImpactFilter) => name == f.name)?.[0];
};

export const newMockFailure = (): ClusterFailureBuilder => {
  return new ClusterFailureBuilder();
};

export const createDefaultMockFailure = (): DistinctClusterFailure => {
  return newMockFailure().build();
};

export const createDefaultMockFailures = (num = 5): Array<DistinctClusterFailure> => {
  return Array.from(Array(num).keys())
      .map(() => createDefaultMockFailure());
};


export const createMockVariantGroups = (): VariantGroup[] => {
  return Array.from(Array(4).keys())
      .map((k) =>(
        {
          key: `v${k}`,
          values: [
            `value${k}`,
          ],
          isSelected: false,
        }
      ));
};

export const createDefaultMockFailureGroup = (key: GroupKey | null = null): FailureGroup => {
  if (!key) {
    key = { type: 'test', value: 'testgroup' };
  }
  return newMockGroup(key).withFailures(1).build();
};

export const createDefaultMockFailureGroupWithChildren = (): FailureGroup => {
  return newMockGroup({ type: 'test', value: 'testgroup' })
      .withChildren([
        newMockGroup({ type: 'leaf', value: 'a3' }).withFailures(3).build(),
        newMockGroup({ type: 'leaf', value: 'a2' }).withFailures(2).build(),
        newMockGroup({ type: 'leaf', value: 'a1' }).withFailures(1).build(),
      ]).build();
};
