# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is govered by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""Unittests for the flt launch issues conversion task."""

import logging
import unittest
import settings
import mock

from framework import exceptions
from framework import permissions
from services import service_manager
from tracker import fltconversion
from tracker import tracker_bizobj
from testing import fake
from testing import testing_helpers
from proto import tracker_pb2

class FLTConvertTask(unittest.TestCase):

  def setUp(self):
    self.services = service_manager.Services(
        issue=fake.IssueService()
    )
    self.mr = testing_helpers.MakeMonorailRequest()
    self.task = fltconversion.FLTConvertTask(
        'req', 'res', services=self.services)
    self.task.mr = self.mr
    self.issue = fake.MakeTestIssue(
        789, 1, 'summary', 'New', 111L, issue_id=78901)
    self.config = tracker_bizobj.MakeDefaultProjectIssueConfig(789)

  def testAssertBasePermission(self):
    self.mr.auth.user_pb.is_site_admin = True
    settings.app_id = 'monorail-staging'
    self.task.AssertBasePermission(self.mr)

    self.mr.auth.user_pb.is_site_admin = False
    self.assertRaises(permissions.PermissionException,
                      self.task.AssertBasePermission, self.mr)

    self.mr.auth.user_pb.is_site_admin = True
    settings.app_id = 'monorail-prod'
    self.assertRaises(exceptions.ActionNotSupported,
                      self.task.AssertBasePermission, self.mr)

  def testExecuteIssueChanges(self):
    self.task.services.issue._UpdateIssuesApprovals = mock.Mock()
    self.task.services.issue.DeltaUpdateIssue = mock.Mock(
        return_value=([], None))
    self.task.services.issue.InsertComment = mock.Mock()
    self.config.approval_defs = [
        tracker_pb2.ApprovalDef(approval_id=1, survey=''), # test empty survey
        tracker_pb2.ApprovalDef(approval_id=2), # test missing survey
        tracker_pb2.ApprovalDef(survey='Missing approval_id should not error.'),
        tracker_pb2.ApprovalDef(approval_id=3, survey='Q1\nQ2\n\nQ3'),
        tracker_pb2.ApprovalDef(approval_id=4, survey='Q1\nQ2\n\nQ3 two'),
        tracker_pb2.ApprovalDef()]

    new_avs = [tracker_pb2.ApprovalValue(
        approval_id=1, status=tracker_pb2.ApprovalStatus.APPROVED,
        approver_ids=[111L, 222L]),
               tracker_pb2.ApprovalValue(approval_id=4),
               tracker_pb2.ApprovalValue(approval_id=2),
               tracker_pb2.ApprovalValue(approval_id=3)]

    phases = [tracker_pb2.Phase(phase_id=1, name='Phase1', rank=1)]
    new_fvs = [tracker_bizobj.MakeFieldValue(
        11, 70, None, None, None, None, False, phase_id=1),
               tracker_bizobj.MakeFieldValue(
                   12, None, 'strfield', None, None, None, False)]
    _amendments = self.task.ExecuteIssueChanges(
        self.config, self.issue, new_avs, phases, new_fvs)

    self.issue.approval_values = new_avs
    self.issue.phases = phases
    delta = tracker_pb2.IssueDelta(
        labels_add=['Type-FLT-Launch', 'FLT-Conversion'],
        labels_remove=['Type-Launch'], field_vals_add=new_fvs)
    cmt_1 = tracker_pb2.IssueComment(
        issue_id=78901, project_id=789, user_id=self.mr.auth.user_id,
        content='', is_description=True, approval_id=1)
    cmt_2 = tracker_pb2.IssueComment(
        issue_id=78901, project_id=789, user_id=self.mr.auth.user_id,
        content='', is_description=True, approval_id=2)
    cmt_3 = tracker_pb2.IssueComment(
        issue_id=78901, project_id=789, user_id=self.mr.auth.user_id,
        content='<b>Q1</b>\n<b>Q2</b>\n<b></b>\n<b>Q3</b>',
        is_description=True, approval_id=3)
    cmt_4 = tracker_pb2.IssueComment(
        issue_id=78901, project_id=789, user_id=self.mr.auth.user_id,
        content='<b>Q1</b>\n<b>Q2</b>\n<b></b>\n<b>Q3 two</b>',
        is_description=True, approval_id=4)


    comment_calls = [mock.call(self.mr.cnxn, cmt_1),
                     mock.call(self.mr.cnxn, cmt_4),
                     mock.call(self.mr.cnxn, cmt_2),
                     mock.call(self.mr.cnxn, cmt_3)]
    self.task.services.issue.InsertComment.assert_has_calls(comment_calls)

    self.task.services.issue._UpdateIssuesApprovals.assert_called_once_with(
        self.mr.cnxn, self.issue)
    self.task.services.issue.DeltaUpdateIssue.assert_called_once_with(
        self.mr.cnxn, self.task.services, self.mr.auth.user_id, 789,
        self.config, self.issue, delta,
        comment=fltconversion.CONVERSION_COMMENT)

class ConvertMLabels(unittest.TestCase):

  def setUp(self):
    self.target_id = 24
    self.approved_id = 27
    self.beta_phase = tracker_pb2.Phase(phase_id=1, name='bEtA')
    self.stable_phase = tracker_pb2.Phase(phase_id=2, name='StAbLe')
    self.stable_full_phase = tracker_pb2.Phase(phase_id=3, name='stable-FULL')
    self.stable_exp_phase = tracker_pb2.Phase(phase_id=4, name='STABLE-exp')

  def testConvertMLabels_NormalFinch(self):

    phases = [self.stable_exp_phase, self.beta_phase, self.stable_full_phase]
    labels = [
        'launch-m-approved-81-beta',  # beta:M-Approved=81
        'launch-m-target-80-stable-car',  # ignore
        'a-Launch-M-Target-80-Stable-car',  # ignore
        'launch-m-target-70-Stable',  # stable-full:M-Target=70
        'LAUNCH-M-TARGET-71-STABLE',  # stable-full:M-Target=71
        'launch-m-target-70-stable-exp',  # stable-exp:M-Target=70
        'launch-m-target-69-stable-exp',  # stable-exp:M-Target=69
        'launch-M-APPROVED-70-Stable-Exp',  # stable-exp:M-Approved-70
        'launch-m-approved-73-stable',  # stable-full:M-Approved-73
        'launch-m-error-73-stable',  # ignore
        'launch-m-approved-8-stable',  #ignore
        'irrelevant label-weird',  # ignore
    ]
    actual_fvs = fltconversion.ConvertMLabels(
        labels, phases, self.target_id, self.approved_id)

    expected_fvs = [
      tracker_pb2.FieldValue(
          field_id=self.approved_id, int_value=81,
          phase_id=self.beta_phase.phase_id, derived=False,),
      tracker_pb2.FieldValue(
          field_id=self.target_id, int_value=70,
          phase_id=self.stable_full_phase.phase_id, derived=False),
      tracker_pb2.FieldValue(
          field_id=self.target_id, int_value=71,
          phase_id=self.stable_full_phase.phase_id, derived=False),
      tracker_pb2.FieldValue(
          field_id=self.target_id, int_value=70,
          phase_id=self.stable_exp_phase.phase_id, derived=False),
      tracker_pb2.FieldValue(
          field_id=self.target_id, int_value=69,
          phase_id=self.stable_exp_phase.phase_id, derived=False),
      tracker_pb2.FieldValue(
          field_id=self.approved_id, int_value=70,
          phase_id=self.stable_exp_phase.phase_id, derived=False),
      tracker_pb2.FieldValue(
          field_id=self.approved_id, int_value=73,
          phase_id=self.stable_full_phase.phase_id, derived=False)
    ]

    self.assertEqual(actual_fvs, expected_fvs)


class ConvertLaunchLabels(unittest.TestCase):

  def setUp(self):
    self.project_fds = [
        tracker_pb2.FieldDef(
            field_id=1, project_id=789, field_name='String',
            field_type=tracker_pb2.FieldTypes.STR_TYPE),
        tracker_pb2.FieldDef(
            field_id=2, project_id=789, field_name='Chrome-UX',
            field_type=tracker_pb2.FieldTypes.APPROVAL_TYPE),
        tracker_pb2.FieldDef(
            field_id=3, project_id=789, field_name='Chrome-Privacy',
            field_type=tracker_pb2.FieldTypes.APPROVAL_TYPE)
        ]
    approvalUX = tracker_pb2.ApprovalValue(
        approval_id=2, status=tracker_pb2.ApprovalStatus.NEEDS_REVIEW)
    approvalPrivacy = tracker_pb2.ApprovalValue(approval_id=3)
    self.approvals = [approvalUX, approvalPrivacy]
    self.issue = fake.MakeTestIssue(001, 1, 'summary', 'New', 111L)

  def testConvertLaunchLabels_Normal(self):
    self.issue.labels = [
        'Launch-UX-NotReviewed', 'Launch-Privacy-Yes', 'Launch-NotRelevant']
    actual = fltconversion.ConvertLaunchLabels(
        self.issue, self.approvals, self.project_fds)
    expected = [
      tracker_pb2.ApprovalValue(
          approval_id=2, status=tracker_pb2.ApprovalStatus.NEEDS_REVIEW),
      tracker_pb2.ApprovalValue(
          approval_id=3, status=tracker_pb2.ApprovalStatus.APPROVED)
    ]
    self.assertEqual(actual, expected)

  def testConvertLaunchLabels_ExtraAndMissingLabels(self):
    self.issue.labels = [
        'Blah-Launch-Privacy-Yes',  # Missing, this is not a valid Label
        'Launch-Security-Yes',  # Extra, no matching approval in given approvals
        'Launch-UI-Yes']  # Missing Launch-Privacy
    actual = fltconversion.ConvertLaunchLabels(
        self.issue, self.approvals, self.project_fds)
    expected = [
        tracker_pb2.ApprovalValue(
            approval_id=2, status=tracker_pb2.ApprovalStatus.APPROVED),
      tracker_pb2.ApprovalValue(
          approval_id=3, status=tracker_pb2.ApprovalStatus.NOT_SET)
        ]
    self.assertEqual(actual, expected)
