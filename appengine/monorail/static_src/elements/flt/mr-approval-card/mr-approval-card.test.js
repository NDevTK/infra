// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {assert} from 'chai';
import {MrApprovalCard} from './mr-approval-card.js';
import {flush} from '@polymer/polymer/lib/utils/flush.js';
import {resetState} from '../../redux/redux-mixin.js';

let element;

suite('mr-approval-card', () => {
  setup(() => {
    element = document.createElement('mr-approval-card');
    document.body.appendChild(element);
  });

  teardown(() => {
    document.body.removeChild(element);
    element.dispatchAction(resetState());
  });

  test('initializes', () => {
    assert.instanceOf(element, MrApprovalCard);
  });

  test('_isApprover true when user is an approver', () => {
    const userNotInList = element._computeIsApprover([
      {displayName: 'tester@user.com'},
      {displayName: 'test@notuser.com'},
      {displayName: 'hello@world.com'},
    ], 'test@user.com', []);
    assert.isFalse(userNotInList);

    const userInList = element._computeIsApprover([
      {displayName: 'tester@user.com'},
      {displayName: 'test@notuser.com'},
      {displayName: 'hello@world.com'},
      {displayName: 'test@user.com'},
    ], 'test@user.com', []);
    assert.isTrue(userInList);

    const userGroupNotInList = element._computeIsApprover([
      {displayName: 'tester@user.com'},
      {displayName: 'nongroup@group.com'},
      {displayName: 'group@nongroup.com'},
      {displayName: 'ignore@test.com'},
    ], 'test@user.com', [
      {displayName: 'group@group.com'},
      {displayName: 'test@group.com'},
      {displayName: 'group@user.com'},
    ]);
    assert.isFalse(userGroupNotInList);

    const userGroupInList = element._computeIsApprover([
      {displayName: 'tester@user.com'},
      {displayName: 'group@group.com'},
      {displayName: 'test@notuser.com'},
    ], 'test@user.com', [
      {displayName: 'group@group.com'},
    ]);
    assert.isTrue(userGroupInList);
  });

  test('site admins have approver privileges', () => {
    const notice = element.shadowRoot.querySelector('.approver-notice');
    assert.equal(notice.textContent.trim(), '');

    element.user = {isSiteAdmin: true};
    assert.isTrue(element._hasApproverPrivileges);

    flush(() => {
      assert.equal(notice.textContent.trim(),
        'Your site admin privileges give you full access to edit this approval.'
      );
    });
  });

  test('site admins see all approval statuses except NotSet', () => {
    element.user = {isSiteAdmin: true};

    assert.isFalse(element._isApprover);

    element.statusEnum = 'NEEDS_REVIEW';

    assert.equal(element._availableStatuses.length, 7);
    assert.equal(element._availableStatuses[0].status, 'NeedsReview');
    assert.equal(element._availableStatuses[1].status, 'NA');
    assert.equal(element._availableStatuses[2].status, 'ReviewRequested');
    assert.equal(element._availableStatuses[3].status, 'ReviewStarted');
    assert.equal(element._availableStatuses[4].status, 'NeedInfo');
    assert.equal(element._availableStatuses[5].status, 'Approved');
    assert.equal(element._availableStatuses[6].status, 'NotApproved');
  });

  test('approvers see all approval statuses except NotSet', () => {
    element.user = {isSiteAdmin: false, email: 'test@email.com'};
    element.approvers = [{displayName: 'test@email.com'}];

    assert.isTrue(element._isApprover);

    element.statusEnum = 'NEEDS_REVIEW';

    assert.equal(element._availableStatuses.length, 7);
    assert.equal(element._availableStatuses[0].status, 'NeedsReview');
    assert.equal(element._availableStatuses[1].status, 'NA');
    assert.equal(element._availableStatuses[2].status, 'ReviewRequested');
    assert.equal(element._availableStatuses[3].status, 'ReviewStarted');
    assert.equal(element._availableStatuses[4].status, 'NeedInfo');
    assert.equal(element._availableStatuses[5].status, 'Approved');
    assert.equal(element._availableStatuses[6].status, 'NotApproved');
  });

  test('non-approvers see non-restricted approval statuses', () => {
    element.user = {isSiteAdmin: false, displayName: 'test@email.com'};
    element.approvers = [{displayName: 'test@otheremail.com'}];

    assert.isFalse(element._isApprover);

    element.statusEnum = 'NEEDS_REVIEW';

    assert.equal(element._availableStatuses.length, 4);
    assert.equal(element._availableStatuses[0].status, 'NeedsReview');
    assert.equal(element._availableStatuses[1].status, 'ReviewRequested');
    assert.equal(element._availableStatuses[2].status, 'ReviewStarted');
    assert.equal(element._availableStatuses[3].status, 'NeedInfo');
  });

  test('non-approvers see restricted approval status when set', () => {
    element.user = {isSiteAdmin: false, displayName: 'test@email.com'};
    element.approvers = [{displayName: 'test@otheremail.com'}];

    assert.isFalse(element._isApprover);

    element.statusEnum = 'APPROVED';

    assert.equal(element._availableStatuses.length, 5);
    assert.equal(element._availableStatuses[0].status, 'NeedsReview');
    assert.equal(element._availableStatuses[1].status, 'ReviewRequested');
    assert.equal(element._availableStatuses[2].status, 'ReviewStarted');
    assert.equal(element._availableStatuses[3].status, 'NeedInfo');
    assert.equal(element._availableStatuses[4].status, 'Approved');
  });

  test('expands to show focused comment', () => {
    element.focusId = 'c4';
    element.fieldName = 'field';
    element.comments = [
      {
        sequenceNum: 1,
        approvalRef: {fieldName: 'other-field'},
      },
      {
        sequenceNum: 2,
        approvalRef: {fieldName: 'field'},
      },
      {
        sequenceNum: 3,
      },
      {
        sequenceNum: 4,
        approvalRef: {fieldName: 'field'},
      },
    ];

    flush();

    assert.isTrue(element.opened);
  });

  test('does not expands to show focused comment on other elements', () => {
    element.focusId = 'c3';
    element.fieldName = 'field';
    element.comments = [
      {
        sequenceNum: 1,
        approvalRef: {fieldName: 'other-field'},
      },
      {
        sequenceNum: 2,
        approvalRef: {fieldName: 'field'},
      },
      {
        sequenceNum: 3,
      },
      {
        sequenceNum: 4,
        approvalRef: {fieldName: 'field'},
      },
    ];

    flush();

    assert.isFalse(element.opened);
  });
});
