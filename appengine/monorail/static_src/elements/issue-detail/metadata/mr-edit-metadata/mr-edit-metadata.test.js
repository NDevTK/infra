// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {assert} from 'chai';
import sinon from 'sinon';
import {MrEditMetadata} from './mr-edit-metadata.js';
import {ISSUE_EDIT_PERMISSION, ISSUE_EDIT_SUMMARY_PERMISSION,
  ISSUE_EDIT_STATUS_PERMISSION, ISSUE_EDIT_OWNER_PERMISSION,
  ISSUE_EDIT_CC_PERMISSION,
} from 'elements/shared/permissions.js';
import {store} from 'elements/reducers/base.js';


let element;


describe('mr-edit-metadata', () => {
  beforeEach(() => {
    element = document.createElement('mr-edit-metadata');
    document.body.appendChild(element);

    sinon.stub(store, 'dispatch');

    element.issuePermissions = [ISSUE_EDIT_PERMISSION];
  });

  afterEach(() => {
    document.body.removeChild(element);
    store.dispatch.restore();
  });

  it('initializes', () => {
    assert.instanceOf(element, MrEditMetadata);
  });

  it('disconnecting element reports form is not dirty', () => {
    element.formName = 'test';

    assert.isFalse(store.dispatch.calledOnce);

    document.body.removeChild(element);

    assert.isTrue(store.dispatch.calledOnce);
    sinon.assert.calledWith(
      store.dispatch,
      {
        type: 'REPORT_DIRTY_FORM',
        name: 'test',
        isDirty: false,
      }
    );

    document.body.appendChild(element);
  });

  it('delta empty when no changes', () => {
    assert.deepEqual(element.delta, {});
  });

  it('toggling checkbox toggles sendEmail', async () => {
    element.sendEmail = false;

    await element.updateComplete;
    const checkbox = element.shadowRoot.querySelector('#sendEmail');

    await checkbox.updateComplete;

    checkbox.click();
    await element.updateComplete;

    assert.equal(checkbox.checked, true);
    assert.equal(element.sendEmail, true);

    checkbox.click();
    await element.updateComplete;

    assert.equal(checkbox.checked, false);
    assert.equal(element.sendEmail, false);

    checkbox.click();
    await element.updateComplete;

    assert.equal(checkbox.checked, true);
    assert.equal(element.sendEmail, true);
  });

  it('changing status produces delta change', async () => {
    element.statuses = [
      {'status': 'New'},
      {'status': 'Old'},
      {'status': 'Test'},
    ];
    element.status = 'New';

    await element.updateComplete;

    const statusComponent = element.shadowRoot.querySelector('#statusInput');
    statusComponent.status = 'Old';

    await element.updateComplete;

    assert.deepEqual(element.delta, {
      status: 'Old',
    });
  });

  it('invalid status throws', async () => {
    element.statuses = [
      {'status': 'New'},
      {'status': 'Old'},
      {'status': 'Duplicate'},
    ];
    element.status = 'Duplicate';

    await element.updateComplete;

    const statusComponent = element.shadowRoot.querySelector('#statusInput');
    statusComponent.shadowRoot.querySelector('#mergedIntoInput').setValue('xx');
    assert.deepEqual(element.delta, {});
    assert.equal(element.error, 'Invalid input for field: mergedInto');
  });

  it('not changing status produces no delta', async () => {
    element.statuses = [
      {'status': 'Duplicate'},
    ];
    element.status = 'Duplicate';

    element.mergedInto = {
      projectName: 'chromium',
      localId: 1234,
    };

    element.projectName = 'chromium';

    await element.updateComplete;
    await element.updateComplete; // Merged input updates its value.

    assert.deepEqual(element.delta, {});
  });

  it('changing status to duplicate produces delta change', async () => {
    element.statuses = [
      {'status': 'New'},
      {'status': 'Duplicate'},
    ];
    element.status = 'New';

    await element.updateComplete;

    const statusComponent = element.shadowRoot.querySelector(
      '#statusInput');
    const root = statusComponent.shadowRoot;
    const statusInput = root.querySelector('#statusInput');
    statusInput.value = 'Duplicate';
    statusInput.dispatchEvent(new Event('change'));

    await element.updateComplete;

    root.querySelector('#mergedIntoInput').setValue(
      'chromium:1234');
    assert.deepEqual(element.delta, {
      status: 'Duplicate',
      mergedIntoRef: {
        projectName: 'chromium',
        localId: 1234,
      },
    });
  });

  it('changing summary produces delta change', async () => {
    element.summary = 'Old summary';

    await element.updateComplete;

    element.shadowRoot.querySelector(
      '#summaryInput').value = 'newfangled fancy summary';
    assert.deepEqual(element.delta, {
      summary: 'newfangled fancy summary',
    });
  });

  it('changing custom fields produces delta', async () => {
    element.fieldValueMap = new Map([['fakeField', ['prev value']]]);
    element.fieldDefs = [
      {
        fieldRef: {
          fieldName: 'testField',
          fieldId: 1,
        },
      },
      {
        fieldRef: {
          fieldName: 'fakeField',
          fieldId: 2,
        },
      },
    ];

    await element.updateComplete;

    element.shadowRoot.querySelector('#testFieldInput').setValue('test value');
    element.shadowRoot.querySelector('#fakeFieldInput').setValue('');
    assert.deepEqual(element.delta, {
      fieldValsAdd: [
        {
          fieldRef: {
            fieldName: 'testField',
            fieldId: 1,
          },
          value: 'test value',
        },
      ],
      fieldValsRemove: [
        {
          fieldRef: {
            fieldName: 'fakeField',
            fieldId: 2,
          },
          value: 'prev value',
        },
      ],
    });
  });

  it('changing approvers produces delta', async () => {
    element.isApproval = true;
    element.hasApproverPrivileges = true;
    element.approvers = [
      {displayName: 'foo@example.com', userId: '1'},
      {displayName: 'bar@example.com', userId: '2'},
      {displayName: 'baz@example.com', userId: '3'},
    ];

    await element.updateComplete;

    element.shadowRoot.querySelector('#approversInput').setValue(
      ['chicken@example.com', 'foo@example.com', 'dog@example.com']);

    await element.updateComplete;

    assert.deepEqual(element.delta, {
      approverRefsAdd: [
        {displayName: 'chicken@example.com'},
        {displayName: 'dog@example.com'},
      ],
      approverRefsRemove: [
        {displayName: 'bar@example.com'},
        {displayName: 'baz@example.com'},
      ],
    });
  });

  it('changing blockedon produces delta change', async () => {
    element.blockedOn = [
      {projectName: 'chromium', localId: '1234'},
      {projectName: 'monorail', localId: '4567'},
    ];
    element.projectName = 'chromium';

    await element.updateComplete;

    const blockedOnInput = element.shadowRoot.querySelector('#blockedOnInput');
    blockedOnInput.setValue(['1234', 'v8:5678']);

    await element.updateComplete;

    assert.deepEqual(element.delta, {
      blockedOnRefsAdd: [{
        projectName: 'v8',
        localId: 5678,
      }],
      blockedOnRefsRemove: [{
        projectName: 'monorail',
        localId: 4567,
      }],
    });
  });

  it('_optionsForField computes options', () => {
    const optionsPerEnumField = new Map([
      ['enumfield', [{optionName: 'one'}, {optionName: 'two'}]],
    ]);
    assert.deepEqual(
      element._optionsForField(optionsPerEnumField, new Map(), 'enumField'), [
        {
          optionName: 'one',
        },
        {
          optionName: 'two',
        },
      ]);
  });

  it('changing enum fields produces delta', async () => {
    element.fieldDefs = [
      {
        fieldRef: {
          fieldName: 'enumField',
          fieldId: 1,
          type: 'ENUM_TYPE',
        },
        isMultivalued: true,
      },
    ];

    element.optionsPerEnumField = new Map([
      ['enumfield', [{optionName: 'one'}, {optionName: 'two'}]],
    ]);

    await element.updateComplete;
    await element.updateComplete;

    element.shadowRoot.querySelector(
      '#enumFieldInput').setValue(['one', 'two']);

    await element.updateComplete;

    assert.deepEqual(element.delta, {
      fieldValsAdd: [
        {
          fieldRef: {
            fieldName: 'enumField',
            fieldId: 1,
          },
          value: 'one',
        },
        {
          fieldRef: {
            fieldName: 'enumField',
            fieldId: 1,
          },
          value: 'two',
        },
      ],
    });
  });

  it('changing multiple single valued enum fields', async () => {
    element.fieldDefs = [
      {
        fieldRef: {
          fieldName: 'enumField',
          fieldId: 1,
          type: 'ENUM_TYPE',
        },
      },
      {
        fieldRef: {
          fieldName: 'enumField2',
          fieldId: 2,
          type: 'ENUM_TYPE',
        },
      },
    ];

    element.optionsPerEnumField = new Map([
      ['enumfield', [{optionName: 'one'}, {optionName: 'two'}]],
      ['enumfield2', [{optionName: 'three'}, {optionName: 'four'}]],
    ]);

    await element.updateComplete;

    element.shadowRoot.querySelector('#enumFieldInput').setValue(['two']);
    element.shadowRoot.querySelector('#enumField2Input').setValue(['three']);

    await element.updateComplete;

    assert.deepEqual(element.delta, {
      fieldValsAdd: [
        {
          fieldRef: {
            fieldName: 'enumField',
            fieldId: 1,
          },
          value: 'two',
        },
        {
          fieldRef: {
            fieldName: 'enumField2',
            fieldId: 2,
          },
          value: 'three',
        },
      ],
    });
  });

  it('adding components produces delta', async () => {
    element.isApproval = false;
    element.issuePermissions = [ISSUE_EDIT_PERMISSION];

    await element.updateComplete;

    const compInput = element.shadowRoot.querySelector('#componentsInput');

    compInput.setValue(['Hello>World']);

    await element.updateComplete;

    assert.deepEqual(element.delta, {
      compRefsAdd: [
        {path: 'Hello>World'},
      ],
    });

    compInput.setValue(['Hello>World', 'Test', 'Multi']);

    await element.updateComplete;

    assert.deepEqual(element.delta, {
      compRefsAdd: [
        {path: 'Hello>World'},
        {path: 'Test'},
        {path: 'Multi'},
      ],
    });

    compInput.setValue([]);
    await element.updateComplete;

    assert.deepEqual(element.delta, {});
  });

  it('approver input appears when user has privileges', async () => {
    assert.isNull(
      element.shadowRoot.querySelector('#approversInput'));
    element.isApproval = true;
    element.hasApproverPrivileges = true;

    await element.updateComplete;

    assert.isNotNull(
      element.shadowRoot.querySelector('#approversInput'));
  });

  it('reset empties form values', async () => {
    element.fieldDefs = [
      {
        fieldRef: {
          fieldName: 'testField',
          fieldId: 1,
        },
      },
      {
        fieldRef: {
          fieldName: 'fakeField',
          fieldId: 2,
        },
      },
    ];

    await element.updateComplete;

    const uploader = element.shadowRoot.querySelector('mr-upload');
    uploader.files = [
      {name: 'test.png'},
      {name: 'rutabaga.png'},
    ];

    element.shadowRoot.querySelector('#testFieldInput').setValue('testy test');
    element.shadowRoot.querySelector('#fakeFieldInput').setValue('hello world');

    await element.reset();

    assert.lengthOf(element.shadowRoot.querySelector(
      '#testFieldInput').value, 0);
    assert.lengthOf(element.shadowRoot.querySelector(
      '#fakeFieldInput').value, 0);
    assert.lengthOf(uploader.files, 0);
  });

  it('edit issue permissions', async () => {
    const allFields = ['summary', 'status', 'owner', 'cc'];
    const testCases = [
      {permissions: [], nonNull: []},
      {permissions: [ISSUE_EDIT_PERMISSION], nonNull: allFields},
      {permissions: [ISSUE_EDIT_SUMMARY_PERMISSION], nonNull: ['summary']},
      {permissions: [ISSUE_EDIT_STATUS_PERMISSION], nonNull: ['status']},
      {permissions: [ISSUE_EDIT_OWNER_PERMISSION], nonNull: ['owner']},
      {permissions: [ISSUE_EDIT_CC_PERMISSION], nonNull: ['cc']},
    ];
    element.statuses = [{'status': 'Foo'}];

    for (const testCase of testCases) {
      element.issuePermissions = testCase.permissions;
      await element.updateComplete;

      allFields.forEach((fieldName) => {
        const field = element.shadowRoot.querySelector(`#${fieldName}Input`);
        if (testCase.nonNull.includes(fieldName)) {
          assert.isNotNull(field);
        } else {
          assert.isNull(field);
        }
      });
    }
  });

  it('duplicate issue is rendered correctly', async () => {
    element.statuses = [
      {'status': 'Duplicate'},
    ];
    element.status = 'Duplicate';
    element.projectName = 'chromium';
    element.mergedInto = {
      projectName: 'chromium',
      localId: 1234,
    };

    await element.updateComplete;
    await element.updateComplete;

    const statusComponent = element.shadowRoot.querySelector(
      '#statusInput');
    const root = statusComponent.shadowRoot;
    assert.equal(
      root.querySelector('#mergedIntoInput').value, '1234');
  });

  it('duplicate issue on different project is rendered correctly', async () => {
    element.statuses = [
      {'status': 'Duplicate'},
    ];
    element.status = 'Duplicate';
    element.projectName = 'chromium';
    element.mergedInto = {
      projectName: 'monorail',
      localId: 1234,
    };

    await element.updateComplete;
    await element.updateComplete;

    const statusComponent = element.shadowRoot.querySelector(
      '#statusInput');
    const root = statusComponent.shadowRoot;
    assert.equal(
      root.querySelector('#mergedIntoInput').value, 'monorail:1234');
  });

  it('blocking issues are rendered correctly', async () => {
    element.blocking = [
      {projectName: 'chromium', localId: '1234'},
      {projectName: 'monorail', localId: '4567'},
    ];
    element.projectName = 'chromium';

    await element.updateComplete;

    const blockingInput = element.shadowRoot.querySelector('#blockingInput');

    assert.deepEqual(['1234', 'monorail:4567'], blockingInput.values);
  });

  it('filter out deleted users', async () => {
    element.cc = [
      {displayName: 'test@example.com', userId: '1234'},
      {displayName: 'a deleted user'},
      {displayName: 'someone@example.com', userId: '5678'},
    ];

    await element.updateComplete;

    const actualValues =
      element.shadowRoot.querySelector('#ccInput').values;
    assert.deepEqual(actualValues, [
      'test@example.com',
      'someone@example.com',
    ]);
  });
});
