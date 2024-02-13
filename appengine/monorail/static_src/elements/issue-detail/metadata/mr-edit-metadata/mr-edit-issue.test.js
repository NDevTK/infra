// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import sinon from 'sinon';
import {assert} from 'chai';
import {prpcClient} from 'prpc-client-instance.js';
import {MrEditIssue, allowRemovedRestrictions} from './mr-edit-issue.js';
import {clientLoggerFake} from 'shared/test/fakes.js';
import {migratedTypes} from 'shared/issue-fields.js';

let element;
let clock;

describe('mr-edit-issue', () => {
  beforeEach(() => {
    element = document.createElement('mr-edit-issue');
    document.body.appendChild(element);
    sinon.stub(prpcClient, 'call');

    element.clientLogger = clientLoggerFake();
    clock = sinon.useFakeTimers();
  });

  afterEach(() => {
    document.body.removeChild(element);
    prpcClient.call.restore();

    clock.restore();
  });

  it('initializes', () => {
    assert.instanceOf(element, MrEditIssue);
  });

  it('scrolls into view on #makechanges hash', async () => {
    await element.updateComplete;

    const header = element.querySelector('#makechanges');
    sinon.stub(header, 'scrollIntoView');

    element.focusId = 'makechanges';
    await element.updateComplete;

    assert.isTrue(header.scrollIntoView.calledOnce);

    header.scrollIntoView.restore();
  });

  it('shows snackbar and resets form when editing finishes', async () => {
    sinon.stub(element, 'reset');
    sinon.stub(element, '_showCommentAddedSnackbar');

    element.updatingIssue = true;
    await element.updateComplete;

    sinon.assert.notCalled(element._showCommentAddedSnackbar);
    sinon.assert.notCalled(element.reset);

    element.updatingIssue = false;
    await element.updateComplete;

    sinon.assert.calledOnce(element._showCommentAddedSnackbar);
    sinon.assert.calledOnce(element.reset);
  });

  it('does not show snackbar or reset form on edit error', async () => {
    sinon.stub(element, 'reset');
    sinon.stub(element, '_showCommentAddedSnackbar');

    element.updatingIssue = true;
    await element.updateComplete;

    element.updateError = 'The save failed';
    element.updatingIssue = false;
    await element.updateComplete;

    sinon.assert.notCalled(element._showCommentAddedSnackbar);
    sinon.assert.notCalled(element.reset);
  });

  it('shows current status even if not defined for project', async () => {
    await element.updateComplete;

    const editMetadata = element.querySelector('mr-edit-metadata');
    assert.deepEqual(editMetadata.statuses, []);

    element.projectConfig = {statusDefs: [
      {status: 'hello'},
      {status: 'world'},
    ]};

    await editMetadata.updateComplete;

    assert.deepEqual(editMetadata.statuses, [
      {status: 'hello'},
      {status: 'world'},
    ]);

    element.issue = {
      statusRef: {status: 'hello'},
    };

    await editMetadata.updateComplete;

    assert.deepEqual(editMetadata.statuses, [
      {status: 'hello'},
      {status: 'world'},
    ]);

    element.issue = {
      statusRef: {status: 'weirdStatus'},
    };

    await editMetadata.updateComplete;

    assert.deepEqual(editMetadata.statuses, [
      {status: 'weirdStatus'},
      {status: 'hello'},
      {status: 'world'},
    ]);
  });

  it('ignores deprecated statuses, unless used on current issue', async () => {
    await element.updateComplete;

    const editMetadata = element.querySelector('mr-edit-metadata');
    assert.deepEqual(editMetadata.statuses, []);

    element.projectConfig = {statusDefs: [
      {status: 'new'},
      {status: 'accepted', deprecated: false},
      {status: 'compiling', deprecated: true},
    ]};

    await editMetadata.updateComplete;

    assert.deepEqual(editMetadata.statuses, [
      {status: 'new'},
      {status: 'accepted', deprecated: false},
    ]);


    element.issue = {
      statusRef: {status: 'compiling'},
    };

    await editMetadata.updateComplete;

    assert.deepEqual(editMetadata.statuses, [
      {status: 'compiling'},
      {status: 'new'},
      {status: 'accepted', deprecated: false},
    ]);
  });

  it('filter out empty or deleted user owners', () => {
    assert.equal(
        element._ownerDisplayName({displayName: 'a_deleted_user'}),
        '');
    assert.equal(
        element._ownerDisplayName({
          displayName: 'test@example.com',
          userId: '1234',
        }),
        'test@example.com');
  });

  it('logs issue-update metrics', async () => {
    await element.updateComplete;

    const editMetadata = element.querySelector('mr-edit-metadata');

    sinon.stub(editMetadata, 'delta').get(() => ({summary: 'test'}));

    await element.save();

    sinon.assert.calledOnce(element.clientLogger.logStart);
    sinon.assert.calledWith(element.clientLogger.logStart,
        'issue-update', 'computer-time');

    // Simulate a response updating the UI.
    element.issue = {summary: 'test'};

    await element.updateComplete;
    await element.updateComplete;

    sinon.assert.calledOnce(element.clientLogger.logEnd);
    sinon.assert.calledWith(element.clientLogger.logEnd,
        'issue-update', 'computer-time', 120 * 1000);
  });

  it('presubmits issue on metadata change', async () => {
    element.issueRef = {};

    await element.updateComplete;
    const editMetadata = element.querySelector('mr-edit-metadata');
    editMetadata.dispatchEvent(new CustomEvent('change', {
      detail: {
        delta: {
          summary: 'Summary',
        },
      },
    }));

    // Wait for debouncer.
    clock.tick(element.presubmitDebounceTimeOut + 1);

    sinon.assert.calledWith(prpcClient.call, 'monorail.Issues',
        'PresubmitIssue',
        {issueDelta: {summary: 'Summary'}, issueRef: {}});
  });

  it('presubmits issue on comment change', async () => {
    element.issueRef = {};

    await element.updateComplete;
    const editMetadata = element.querySelector('mr-edit-metadata');
    editMetadata.dispatchEvent(new CustomEvent('change', {
      detail: {
        delta: {},
        commentContent: 'test',
      },
    }));

    // Wait for debouncer.
    clock.tick(element.presubmitDebounceTimeOut + 1);

    sinon.assert.calledWith(prpcClient.call, 'monorail.Issues',
        'PresubmitIssue',
        {issueDelta: {}, issueRef: {}});
  });


  it('does not presubmit issue when no changes', () => {
    element._presubmitIssue({});

    sinon.assert.notCalled(prpcClient.call);
  });

  it('editing form runs _presubmitIssue debounced', async () => {
    sinon.stub(element, '_presubmitIssue');

    await element.updateComplete;

    // User makes some changes.
    const comment = element.querySelector('#commentText');
    comment.value = 'Value';
    comment.dispatchEvent(new Event('keyup'));

    clock.tick(5);

    // User makes more changes before debouncer timeout is done.
    comment.value = 'more changes';
    comment.dispatchEvent(new Event('keyup'));

    clock.tick(10);

    sinon.assert.notCalled(element._presubmitIssue);

    // Wait for debouncer.
    clock.tick(element.presubmitDebounceTimeOut + 1);

    sinon.assert.calledOnce(element._presubmitIssue);
  });
});

describe('allowRemovedRestrictions', () => {
  beforeEach(() => {
    sinon.stub(window, 'confirm');
  });

  afterEach(() => {
    window.confirm.restore();
  });

  it('returns true if no restrictions removed', () => {
    assert.isTrue(allowRemovedRestrictions([
      {label: 'not-restricted'},
      {label: 'fine'},
    ]));
  });

  it('returns false if restrictions removed and confirmation denied', () => {
    window.confirm.returns(false);
    assert.isFalse(allowRemovedRestrictions([
      {label: 'not-restricted'},
      {label: 'restrict-view-people'},
    ]));
  });

  it('returns true if restrictions removed and confirmation accepted', () => {
    window.confirm.returns(true);
    assert.isTrue(allowRemovedRestrictions([
      {label: 'not-restricted'},
      {label: 'restrict-view-people'},
    ]));
  });

  describe('migrated issue', () => {
    it('does not show notice if issue not migrated', async () => {
      element.migratedId = '';

      await element.updateComplete;

      assert.isNull(element.querySelector('.migrated-banner'));
      assert.isNull(element.querySelector('.legacy-edit'));
    });

    it('shows notice if issue migrated', async () => {
      element.migratedId = '1234';
      element.migratedType = migratedTypes.LAUNCH_TYPE
      await element.updateComplete;

      assert.isNotNull(element.querySelector('.migrated-banner'));
      assert.isNotNull(element.querySelector('.legacy-edit'));
    });

    it('shows issue link when migrated off monorail', async () => {
      element.migratedId = '1234';
      element.migratedType = migratedTypes.BUGANIZER_TYPE
      await element.updateComplete;

      const link = element.querySelector('.migrated-banner a');
      assert.include(link.textContent, 'crbug.com/1234');
    });

    it('shows launch banner when migrated to launch', async () => {
      element.migratedId = '1234';
      element.migratedType = migratedTypes.LAUNCH_TYPE
      await element.updateComplete;

      const link = element.querySelector('.migrated-banner');
      assert.include(link.textContent, 'This issue has been migrated to Launch, see link in final comment below');
    });

    it('hides edit form if issue migrated', async () => {
      element.migratedId = '1234';
      element.migratedType = migratedTypes.LAUNCH_TYPE
      await element.updateComplete;

      const editForm = element.querySelector('mr-edit-metadata');
      assert.isTrue(editForm.hasAttribute('hidden'));
    });

    it('unhides edit form on button click', async () => {
      element.migratedId = '1234';
      element.migratedType = migratedTypes.LAUNCH_TYPE
      await element.updateComplete;

      const button = element.querySelector('.legacy-edit');
      button.click();

      await element.updateComplete;

      const editForm = element.querySelector('mr-edit-metadata');
      assert.isFalse(editForm.hasAttribute('hidden'));
    });
  });
});
