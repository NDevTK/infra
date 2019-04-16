// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {assert} from 'chai';
import * as project from './project.js';
import {fieldTypes} from '../shared/field-types.js';

suite('project', () => {
  test('fieldDefs', () => {
    assert.deepEqual(project.fieldDefs({project: {}}), []);
    assert.deepEqual(project.fieldDefs({project: {config: {}}}), []);
    assert.deepEqual(project.fieldDefs({
      project: {config: {fieldDefs: [{fieldRef: {fieldName: 'test'}}]}},
    }), [{fieldRef: {fieldName: 'test'}}]);
  });

  test('enumFieldDefs', () => {
    assert.deepEqual(project.enumFieldDefs({project: {}}), []);
    assert.deepEqual(project.enumFieldDefs({project: {config: {}}}), []);
    assert.deepEqual(project.enumFieldDefs({
      project: {config: {fieldDefs: [
        {fieldRef: {fieldName: 'test'}},
        {fieldRef: {fieldName: 'enum', type: fieldTypes.ENUM_TYPE}},
        {fieldRef: {fieldName: 'ignore', type: fieldTypes.DATE_TYPE}},
      ]}},
    }), [{fieldRef: {fieldName: 'enum', type: fieldTypes.ENUM_TYPE}}]);
  });

  test('optionsPerEnumField', () => {
    assert.deepEqual(project.optionsPerEnumField({project: {}}), new Map());
    assert.deepEqual(project.optionsPerEnumField({
      project: {config: {
        fieldDefs: [
          {fieldRef: {fieldName: 'ignore', type: fieldTypes.DATE_TYPE}},
          {fieldRef: {fieldName: 'eNum', type: fieldTypes.ENUM_TYPE}},
        ],
        labelDefs: [
          {label: 'enum-one'},
          {label: 'ENUM-tWo'},
          {label: 'not-enum-three'},
        ],
      }},
    }), new Map([
      ['enum', [
        {label: 'enum-one', optionName: 'one'},
        {label: 'ENUM-tWo', optionName: 'tWo'},
      ]],
    ]));
  });

  test('fieldDefsByApprovalName', () => {
    assert.deepEqual(project.fieldDefsByApprovalName({project: {}}), new Map());

    assert.deepEqual(project.fieldDefsByApprovalName({project: {config: {
      fieldDefs: [
        {fieldRef: {fieldName: 'test', type: fieldTypes.INT_TYPE}},
        {fieldRef: {fieldName: 'ignoreMe', type: fieldTypes.APPROVAL_TYPE}},
        {fieldRef: {fieldName: 'yay', approvalName: 'ThisIsAnApproval'}},
        {fieldRef: {fieldName: 'ImAField', approvalName: 'ThisIsAnApproval'}},
        {fieldRef: {fieldName: 'TalkToALawyer', approvalName: 'Legal'}},
      ],
    }}}), new Map([
      ['ThisIsAnApproval', [
        {fieldRef: {fieldName: 'yay', approvalName: 'ThisIsAnApproval'}},
        {fieldRef: {fieldName: 'ImAField', approvalName: 'ThisIsAnApproval'}},
      ]],
      ['Legal', [
        {fieldRef: {fieldName: 'TalkToALawyer', approvalName: 'Legal'}},
      ]],
    ]));
  });
});
