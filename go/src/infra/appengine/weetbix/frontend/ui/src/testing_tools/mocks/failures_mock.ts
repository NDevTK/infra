// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import dayjs from 'dayjs';

import {
  ClusterFailure,
  FailureGroup,
  VariantGroup,
  ImpactFilter,
  ImpactFilters,
} from '../../tools/failures_tools';

class ClusterFailureBuilder {
  failure: ClusterFailure;
  constructor() {
    this.failure = {
      realm: 'testproject/testrealm',
      testId: 'ninja://dir/test.param',
      variant: [],
      presubmitRunCl: { host: 'clproject-review.googlesource.com', change: 123456, patchset: 7 },
      presubmitRunId: { system: 'cv', id: 'presubmitRunId' },
      presubmitRunOwner: 'user',
      partitionTime: '2021-05-12T19:05:34',
      exonerationStatus: 'NOT_EXONERATED',
      ingestedInvocationId: 'ingestedInvocationId',
      isIngestedInvocationBlocked: false,
      testRunIds: ['testRunId'],
      isTestRunBlocked: false,
      count: 1,
    };
  }
  build(): ClusterFailure {
    return this.failure;
  }
  testRunBlocked() {
    this.failure.isTestRunBlocked = true;
    return this;
  }
  ingestedInvocationBlocked() {
    this.failure.isIngestedInvocationBlocked = true;
    this.failure.isTestRunBlocked = true;
    return this;
  }
  exonerateOccursOnOtherCLs() {
    this.failure.exonerationStatus = 'OCCURS_ON_OTHER_CLS';
    return this;
  }
  exonerateNotCritical() {
    this.failure.exonerationStatus = 'NOT_CRITICAL';
    return this;
  }
  withVariantGroups(key: string, value: string) {
    this.failure.variant.push({ key, value });
    return this;
  }
  withTestId(id: string) {
    this.failure.testId = id;
    return this;
  }
  withoutPresubmit() {
    this.failure.presubmitRunCl = null;
    this.failure.presubmitRunId = null;
    this.failure.presubmitRunOwner = null;
    return this;
  }
}

export const newMockGroup = (name: string): FailureGroupBuilder => {
  return new FailureGroupBuilder(name);
};

class FailureGroupBuilder {
  failureGroup: FailureGroup;
  constructor(name: string) {
    this.failureGroup = {
      id: name,
      name,
      children: [],
      failures: 0,
      testRunFailures: 0,
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

  withTestRunFailures(testRunFailures: number) {
    this.failureGroup.testRunFailures = testRunFailures;
    return this;
  }

  withInvocationFailures(invocationFailures: number) {
    this.failureGroup.invocationFailures =invocationFailures;
    return this;
  }

  withFailure(failure: ClusterFailure) {
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

export const createDefaultMockFailure = (): ClusterFailure => {
  return newMockFailure().build();
};

export const createDefaultMockFailures = (num = 5): Array<ClusterFailure> => {
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

export const createDefaultMockFailureGroup = (name = 'testgroup'): FailureGroup => {
  return newMockGroup(name).withFailures(1).build();
};

export const createDefaultMockFailureGroupWithChildren = (): FailureGroup => {
  return newMockGroup('testgroup')
      .withChildren([
        newMockGroup('a3').withFailures(3).build(),
        newMockGroup('a2').withFailures(2).build(),
        newMockGroup('a1').withFailures(1).build(),
      ]).build();
};
