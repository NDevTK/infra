// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {assert} from 'chai';
import {fieldTypes} from '../shared/field-types.js';
import * as issue from './issue.js';

suite('issue', () => {
  test('issue', () => {
    assert.isUndefined(issue.issue({}));
    assert.deepEqual(issue.issue({issue: {localId: 100}}),
      {localId: 100});
  });

  test('fieldValues', () => {
    assert.isUndefined(issue.fieldValues({}));
    assert.isUndefined(issue.fieldValues({issue: {}}));
    assert.deepEqual(issue.fieldValues({
      issue: {fieldValues: [{value: 'v'}]},
    }), [{value: 'v'}]);
  });

  test('type', () => {
    assert.isUndefined(issue.type({}));
    assert.isUndefined(issue.type({issue: {}}));
    assert.isUndefined(issue.type({
      issue: {fieldValues: [{value: 'v'}]},
    }));
    assert.deepEqual(issue.type({
      issue: {fieldValues: [
        {fieldRef: {fieldName: 'IgnoreMe'}, value: 'v'},
        {fieldRef: {fieldName: 'Type'}, value: 'Defect'},
      ]},
    }), 'Defect');
  });

  test('restrictions', () => {
    assert.deepEqual(issue.restrictions({}), {});
    assert.deepEqual(issue.restrictions({issue: {}}), {});
    assert.deepEqual(issue.restrictions({issue: {labelRefs: []}}), {});

    assert.deepEqual(issue.restrictions({issue: {labelRefs: [
      {label: 'IgnoreThis'},
      {label: 'IgnoreThis2'},
    ]}}), {});

    assert.deepEqual(issue.restrictions({issue: {labelRefs: [
      {label: 'IgnoreThis'},
      {label: 'IgnoreThis2'},
      {label: 'Restrict-View-Google'},
      {label: 'Restrict-EditIssue-hello'},
      {label: 'Restrict-EditIssue-test'},
      {label: 'Restrict-AddIssueComment-HELLO'},
    ]}}), {
      'view': ['Google'],
      'edit': ['hello', 'test'],
      'comment': ['HELLO'],
    });
  });

  test('isRestricted', () => {
    assert.isFalse(issue.isRestricted({}));
    assert.isFalse(issue.isRestricted({}));
    assert.isFalse(issue.isRestricted({issue: {}}));
    assert.isFalse(issue.isRestricted({issue: {labelRefs: []}}));

    assert.isTrue(issue.isRestricted({issue: {labelRefs: [
      {label: 'IgnoreThis'},
      {label: 'IgnoreThis2'},
      {label: 'Restrict-View-Google'},
    ]}}));

    assert.isFalse(issue.isRestricted({issue: {labelRefs: [
      {label: 'IgnoreThis'},
      {label: 'IgnoreThis2'},
      {label: 'Restrict-View'},
      {label: 'Restrict'},
      {label: 'RestrictView'},
      {label: 'Restt-View'},
    ]}}));

    assert.isTrue(issue.isRestricted({issue: {labelRefs: [
      {label: 'restrict-view-google'},
    ]}}));

    assert.isTrue(issue.isRestricted({issue: {labelRefs: [
      {label: 'restrict-EditIssue-world'},
    ]}}));

    assert.isTrue(issue.isRestricted({issue: {labelRefs: [
      {label: 'RESTRICT-ADDISSUECOMMENT-everyone'},
    ]}}));
  });

  test('fieldValueMap', () => {
    assert.deepEqual(issue.fieldValueMap({}), new Map());
    assert.deepEqual(issue.fieldValueMap({issue: {
      fieldValues: [],
    }}), new Map());
    assert.deepEqual(issue.fieldValueMap({
      issue: {fieldValues: [
        {fieldRef: {fieldName: 'hello'}, value: 'v1'},
        {fieldRef: {fieldName: 'hello'}, value: 'v2'},
        {fieldRef: {fieldName: 'world'}, value: 'v3'},
      ]},
    }), new Map([
      ['hello', ['v1', 'v2']],
      ['world', ['v3']],
    ]));
  });

  test('fieldDefs', () => {
    assert.deepEqual(issue.fieldDefs({project: {}}), []);

    // Remove approval-related fields, regardless of issue.
    assert.deepEqual(issue.fieldDefs({project: {config: {
      fieldDefs: [
        {fieldRef: {fieldName: 'test', type: fieldTypes.INT_TYPE}},
        {fieldRef: {fieldName: 'ignoreMe', type: fieldTypes.APPROVAL_TYPE}},
        {fieldRef: {fieldName: 'LookAway', approvalName: 'ThisIsAnApproval'}},
        {fieldRef: {fieldName: 'phaseField'}, isPhaseField: true},
      ],
    }}}), [
      {fieldRef: {fieldName: 'test', type: fieldTypes.INT_TYPE}},
    ]);

    // Filter defs by applicableType.
    assert.deepEqual(issue.fieldDefs({
      project: {config: {
        fieldDefs: [
          {fieldRef: {fieldName: 'intyInt', type: fieldTypes.INT_TYPE}},
          {fieldRef: {fieldName: 'enum', type: fieldTypes.ENUM_TYPE}},
          {fieldRef: {fieldName: 'nonApplicable', type: fieldTypes.STR_TYPE},
            applicableType: 'None'},
          {fieldRef: {fieldName: 'defectsOnly', type: fieldTypes.STR_TYPE},
            applicableType: 'Defect'},
        ],
      }},
      issue: {
        fieldValues: [
          {fieldRef: {fieldName: 'Type'}, value: 'Defect'},
        ],
      },
    }), [
      {fieldRef: {fieldName: 'intyInt', type: fieldTypes.INT_TYPE}},
      {fieldRef: {fieldName: 'enum', type: fieldTypes.ENUM_TYPE}},
      {fieldRef: {fieldName: 'defectsOnly', type: fieldTypes.STR_TYPE},
        applicableType: 'Defect'},
    ]);
  });
});
