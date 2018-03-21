# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from datetime import datetime
import mock
import textwrap
import urllib

from common import constants
from common import rotations
from common.waterfall import failure_type
from infra_api_clients.codereview import codereview_util
from infra_api_clients.codereview.cl_info import ClInfo
from infra_api_clients.codereview.cl_info import Commit
from infra_api_clients.codereview.cl_info import Revert
from infra_api_clients.codereview.gerrit import Gerrit
from libs import analysis_status as status
from libs import time_util
from model.base_suspected_cl import RevertCL
from model.wf_suspected_cl import WfSuspectedCL
from services import gerrit
from services.parameters import CLKey
from services.parameters import CreateRevertCLParameters
from services.parameters import SubmitRevertCLParameters
from waterfall import buildbot
from waterfall import suspected_cl_util
from waterfall import waterfall_config
from waterfall.test import wf_testcase

_CODEREVIEW = Gerrit('chromium-review.googlesource.com')


class GerritTest(wf_testcase.WaterfallTestCase):

  def setUp(self):
    super(GerritTest, self).setUp()
    self.culprit_commit_position = 123
    self.culprit_code_review_url = (
        'https://chromium-review.googlesource.com/12345')
    self.review_server_host = 'chromium-review.googlesource.com'
    self.review_change_id = '12345'

    def MockGetCulpritInfo(*_):
      culprit_info = {
          'commit_position': self.culprit_commit_position,
          'code_review_url': self.culprit_code_review_url,
          'review_server_host': self.review_server_host,
          'review_change_id': self.review_change_id,
          'author': {
              'email': 'author@chromium.org'
          }
      }
      return culprit_info

    self.mock(suspected_cl_util, 'GetCulpritInfo', MockGetCulpritInfo)

  @mock.patch.object(waterfall_config, 'GetActionSettings', return_value={})
  @mock.patch.object(
      time_util, 'GetUTCNow', return_value=datetime(2017, 2, 1, 16, 0, 0))
  @mock.patch.object(_CODEREVIEW, 'PostMessage', return_value=True)
  @mock.patch.object(rotations, 'current_sheriffs', return_value=['a@b.com'])
  @mock.patch.object(
      codereview_util, 'GetCodeReviewForReview', return_value=_CODEREVIEW)
  @mock.patch.object(_CODEREVIEW, 'AddReviewers', return_value=True)
  @mock.patch.object(_CODEREVIEW, 'CreateRevert')
  @mock.patch.object(_CODEREVIEW, 'GetClDetails')
  def testRevertCLSucceed(self, mock_fn, mock_gerrit, mock_add, *_):
    repo_name = 'chromium'
    revision = 'rev1'
    commit_position = 123
    sample_failed_step = 'compile'

    cl_info = ClInfo(self.review_server_host, self.review_change_id)
    cl_info.commits.append(
        Commit('20001', 'rev1', datetime(2017, 2, 1, 0, 0, 0)))
    cl_info.owner_email = 'abc@chromium.org'
    mock_fn.return_value = cl_info
    mock_gerrit.return_value = '54321'

    culprit = WfSuspectedCL.Create(repo_name, revision, commit_position)
    culprit.builds = {
        'm/b/2': {
            'status': None
        },
        'm/b/1': {
            'status': None
        },
        'm/b/3': {
            'status': None
        },
        'm1/b/0': {
            'status': None
        },
    }
    culprit.put()

    revert_status = gerrit.RevertCulprit(
        repo_name, revision, 'm/b/1', failure_type.COMPILE, sample_failed_step)

    self.assertEquals(revert_status, gerrit.CREATED_BY_FINDIT)

    culprit = WfSuspectedCL.Get(repo_name, revision)
    self.assertEqual(culprit.revert_status, status.COMPLETED)
    self.assertIsNotNone(culprit.revert_cl)

    reason = textwrap.dedent("""
        Findit (https://goo.gl/kROfz5) identified CL at revision %s as the
        culprit for failures in the build cycles as shown on:
        https://findit-for-me.appspot.com/waterfall/culprit?key=%s\n
        Sample Failed Build: %s\n
        Sample Failed Step: %s""") % (commit_position, culprit.key.urlsafe(),
                                      buildbot.CreateBuildUrl('m', 'b', '1'),
                                      sample_failed_step)
    mock_gerrit.assert_called_once_with(reason, self.review_change_id, '20001')

    culprit_link = (
        'https://findit-for-me.appspot.com/waterfall/culprit?key=%s' %
        (culprit.key.urlsafe()))
    false_positive_bug_query = urllib.urlencode({
        'status': 'Available',
        'labels': 'Test-Findit-Wrong',
        'components': 'Tools>Test>FindIt',
        'summary': 'Wrongly blame %s' % revision,
        'comment': 'Detail is %s' % culprit_link
    })
    false_positive_bug_link = (
        'https://bugs.chromium.org/p/chromium/issues/entry?%s') % (
            false_positive_bug_query)

    auto_revert_bug_query = urllib.urlencode({
        'status': 'Available',
        'components': 'Tools>Test>FindIt>Autorevert',
        'summary': 'Auto Revert failed on %s' % revision,
        'comment': 'Detail is %s' % culprit_link
    })
    auto_revert_bug_link = (
        'https://bugs.chromium.org/p/chromium/issues/entry?%s') % (
            auto_revert_bug_query)
    message = textwrap.dedent("""
        Sheriffs, CL owner or CL reviewers:
        Please submit this revert if it is correct.
        If it is a false positive, please abandon and report it
        at %s.
        If failed to submit the revert, please abandon it and report the failure
        at %s.

        For more information about Findit auto-revert: %s.

        Sheriffs, it'll be much appreciated if you could take several minutes
        to fill out this survey: %s.""") % (false_positive_bug_link,
                                            auto_revert_bug_link,
                                            gerrit._MANUAL_LINK,
                                            gerrit._SURVEY_LINK)
    mock_add.assert_called_once_with('54321', ['a@b.com'], message)

  @mock.patch.object(
      codereview_util, 'GetCodeReviewForReview', return_value=_CODEREVIEW)
  @mock.patch.object(_CODEREVIEW, 'GetClDetails')
  def testSheriffRevertedIt(self, mock_fn, *_):
    repo_name = 'chromium'
    revision = 'rev1'

    cl_info = ClInfo(self.review_server_host, self.review_change_id)
    cl_info.commits.append(
        Commit('20001', 'rev1', datetime(2017, 2, 1, 0, 0, 0)))
    cl_info.owner_email = 'abc@chromium.org'
    revert_cl = ClInfo('revert_review_host', '123V3137')
    cl_info.reverts.append(
        Revert('20001', revert_cl, 'a@b', datetime(2017, 2, 1, 1, 0, 0)))
    mock_fn.return_value = cl_info

    WfSuspectedCL.Create(repo_name, revision, 123).put()

    revert_status = gerrit.RevertCulprit(repo_name, revision, 'm/b/1',
                                         failure_type.COMPILE, 'compile')

    self.assertEquals(revert_status, gerrit.CREATED_BY_SHERIFF)

    commit_status = gerrit.CommitRevert(
        SubmitRevertCLParameters(
            cl_key=CLKey(repo_name=repo_name, revision=revision),
            revert_status=revert_status))
    self.assertEqual(gerrit.SKIPPED, commit_status)

    culprit = WfSuspectedCL.Get(repo_name, revision)
    self.assertEquals(culprit.revert_status, status.SKIPPED)
    self.assertIsNone(culprit.revert_cl)
    self.assertIsNone(culprit.revert_pipeline_id)

  @mock.patch.object(_CODEREVIEW, 'AddReviewers', return_value=True)
  @mock.patch.object(rotations, 'current_sheriffs', return_value=['a@b.com'])
  @mock.patch.object(
      codereview_util, 'GetCodeReviewForReview', return_value=_CODEREVIEW)
  @mock.patch.object(_CODEREVIEW, 'PostMessage')
  @mock.patch.object(_CODEREVIEW, 'GetClDetails')
  def testAddedReviewerFailedBefore(self, mock_fn, mock_send, *_):
    repo_name = 'chromium'
    revision = 'rev1'

    cl_info = ClInfo(self.review_server_host, self.review_change_id)
    cl_info.commits.append(
        Commit('20001', 'rev1', datetime(2017, 2, 1, 0, 0, 0)))
    cl_info.owner_email = 'abc@chromium.org'
    revert_cl = ClInfo('revert_review_host', '123V3127')
    revert_cl.url = 'https://chromium-review.googlesource.com/54321'
    cl_info.reverts.append(
        Revert('20001', revert_cl, constants.DEFAULT_SERVICE_ACCOUNT,
               datetime(2017, 2, 1, 1, 0, 0)))
    mock_fn.return_value = cl_info

    culprit = WfSuspectedCL.Create(repo_name, revision, 123)
    culprit.revert_cl = RevertCL()
    culprit.revert_status = status.RUNNING
    culprit.cr_notification_status = status.COMPLETED
    culprit.put()

    revert_status = gerrit.RevertCulprit(repo_name, revision, 'm/b/1',
                                         failure_type.COMPILE, 'compile')

    self.assertEquals(revert_status, gerrit.CREATED_BY_FINDIT)

    culprit = WfSuspectedCL.Get(repo_name, revision)
    self.assertEqual(culprit.revert_status, status.COMPLETED)
    self.assertIsNotNone(culprit.revert_cl)
    mock_send.assert_has_not_called()

  @mock.patch.object(
      codereview_util, 'GetCodeReviewForReview', return_value=_CODEREVIEW)
  @mock.patch.object(_CODEREVIEW, 'GetClDetails')
  def testCulpritIsARevert(self, mock_fn, _):
    repo_name = 'chromium'
    revision = 'rev1'

    cl_info = ClInfo(self.review_server_host, self.review_change_id)
    cl_info.revert_of = 123456
    cl_info.commits.append(
        Commit('20001', 'rev1', datetime(2017, 2, 1, 0, 0, 0)))
    mock_fn.return_value = cl_info

    culprit = WfSuspectedCL.Create(repo_name, revision, 123)
    culprit.put()

    revert_status = gerrit.RevertCulprit(repo_name, revision, 'm/b/1',
                                         failure_type.COMPILE, 'compile')

    self.assertEquals(revert_status, gerrit.SKIPPED)

    culprit = WfSuspectedCL.Get(repo_name, revision)
    self.assertEqual(culprit.revert_status, status.SKIPPED)
    self.assertIsNone(culprit.revert_cl)
    self.assertEqual(culprit.skip_revert_reason, gerrit.CULPRIT_IS_A_REVERT)

  @mock.patch.object(
      codereview_util, 'GetCodeReviewForReview', return_value=_CODEREVIEW)
  @mock.patch.object(_CODEREVIEW, 'GetClDetails')
  def testAutoRevertOff(self, mock_fn, _):
    repo_name = 'chromium'
    revision = 'rev1'

    cl_info = ClInfo(self.review_server_host, self.review_change_id)
    cl_info.auto_revert_off = True
    cl_info.commits.append(
        Commit('20001', 'rev1', datetime(2017, 2, 1, 0, 0, 0)))
    mock_fn.return_value = cl_info

    culprit = WfSuspectedCL.Create(repo_name, revision, 123)
    culprit.put()

    revert_status = gerrit.RevertCulprit(repo_name, revision, 'm/b/1',
                                         failure_type.COMPILE, 'compile')

    self.assertEquals(revert_status, gerrit.SKIPPED)

    culprit = WfSuspectedCL.Get(repo_name, revision)
    self.assertEqual(culprit.revert_status, status.SKIPPED)
    self.assertIsNone(culprit.revert_cl)
    self.assertEqual(culprit.skip_revert_reason, gerrit.AUTO_REVERT_OFF)

  @mock.patch.object(codereview_util, 'IsCodeReviewGerrit', return_value=False)
  def testSubmitRevertForRietveld(self, _):
    repo_name = 'chromium'
    revision = 'rev1'
    commit_position = 123

    cl_info = ClInfo(self.review_server_host, self.review_change_id)
    cl_info.commits.append(
        Commit('20001', 'rev1', datetime(2017, 2, 1, 0, 0, 0)))

    culprit = WfSuspectedCL.Create(repo_name, revision, commit_position)
    revert_for_culprit = RevertCL()
    revert_change_id = '54321'
    revert_for_culprit.revert_cl_url = 'https://%s/q/%s' % (
        self.review_server_host, revert_change_id)
    culprit.revert_cl = revert_for_culprit
    culprit.revert_status = status.COMPLETED
    culprit.put()

    revert_status = gerrit.CREATED_BY_FINDIT
    commit_status = gerrit.CommitRevert(
        SubmitRevertCLParameters(
            cl_key=CLKey(repo_name=repo_name, revision=revision),
            revert_status=revert_status))
    self.assertEqual(gerrit.SKIPPED, commit_status)

    culprit = WfSuspectedCL.Get(repo_name, revision)
    self.assertEqual(culprit.revert_submission_status, status.SKIPPED)

  @mock.patch.object(
      time_util, 'GetUTCNow', return_value=datetime(2017, 2, 1, 5, 0, 0))
  @mock.patch.object(
      codereview_util, 'GetCodeReviewForReview', return_value=_CODEREVIEW)
  @mock.patch.object(rotations, 'current_sheriffs', return_value=['a@b.com'])
  @mock.patch.object(_CODEREVIEW, 'AddReviewers', return_value=True)
  @mock.patch.object(_CODEREVIEW, 'SubmitRevert')
  @mock.patch.object(_CODEREVIEW, 'GetClDetails')
  def testSubmitRevertFailed(self, mock_fn, mock_commit, mock_add, *_):
    repo_name = 'chromium'
    revision = 'rev1'
    commit_position = 123

    cl_info = ClInfo(self.review_server_host, self.review_change_id)
    cl_info.commits.append(
        Commit('20001', 'rev1', datetime(2017, 2, 1, 0, 0, 0)))
    mock_fn.return_value = cl_info
    mock_commit.return_value = False

    culprit = WfSuspectedCL.Create(repo_name, revision, commit_position)
    revert_for_culprit = RevertCL()
    revert_change_id = '54321'
    revert_for_culprit.revert_cl_url = 'https://%s/q/%s' % (
        self.review_server_host, revert_change_id)
    culprit.revert_cl = revert_for_culprit
    culprit.revert_status = status.COMPLETED
    culprit.put()

    revert_status = gerrit.CREATED_BY_FINDIT
    commit_status = gerrit.CommitRevert(
        SubmitRevertCLParameters(
            cl_key=CLKey(repo_name=repo_name, revision=revision),
            revert_status=revert_status))

    self.assertEqual(gerrit.ERROR, commit_status)
    mock_commit.assert_called_once_with(revert_change_id)

    culprit = WfSuspectedCL.Get(repo_name, revision)
    self.assertEqual(culprit.revert_submission_status, status.ERROR)
    culprit_link = ('https://findit-for-me.appspot.com/waterfall/culprit?key=%s'
                    % culprit.key.urlsafe())
    false_positive_bug_query = urllib.urlencode({
        'status': 'Available',
        'labels': 'Test-Findit-Wrong',
        'components': 'Tools>Test>FindIt',
        'summary': 'Wrongly blame %s' % revision,
        'comment': 'Detail is %s' % culprit_link
    })
    false_positive_bug_link = (
        'https://bugs.chromium.org/p/chromium/issues/entry?%s') % (
            false_positive_bug_query)

    auto_revert_bug_query = urllib.urlencode({
        'status': 'Available',
        'components': 'Tools>Test>FindIt>Autorevert',
        'summary': 'Auto Revert failed on %s' % revision,
        'comment': 'Detail is %s' % culprit_link
    })
    auto_revert_bug_link = (
        'https://bugs.chromium.org/p/chromium/issues/entry?%s') % (
            auto_revert_bug_query)
    message = textwrap.dedent("""
        Sheriffs, CL owner or CL reviewers:
        Please submit this revert if it is correct.
        If it is a false positive, please abandon and report it
        at %s.
        If failed to submit the revert, please abandon it and report the failure
        at %s.

        For more information about Findit auto-revert: %s.

        Sheriffs, it'll be much appreciated if you could take several minutes
        to fill out this survey: %s.""") % (false_positive_bug_link,
                                            auto_revert_bug_link,
                                            gerrit._MANUAL_LINK,
                                            gerrit._SURVEY_LINK)
    mock_add.assert_called_once_with('54321', ['a@b.com'], message)

  @mock.patch.object(
      time_util, 'GetUTCNow', return_value=datetime(2017, 2, 4, 5, 0, 0))
  @mock.patch.object(
      codereview_util, 'GetCodeReviewForReview', return_value=_CODEREVIEW)
  @mock.patch.object(_CODEREVIEW, 'SubmitRevert')
  @mock.patch.object(_CODEREVIEW, 'GetClDetails')
  def testSubmitRevertCulpritTooOld(self, mock_fn, mock_commit, *_):
    repo_name = 'chromium'
    revision = 'rev1'
    commit_position = 123

    cl_info = ClInfo(self.review_server_host, self.review_change_id)
    cl_info.commits.append(
        Commit('20001', 'rev1', datetime(2017, 2, 1, 0, 0, 0)))
    mock_fn.return_value = cl_info
    mock_commit.return_value = True

    culprit = WfSuspectedCL.Create(repo_name, revision, commit_position)
    revert_for_culprit = RevertCL()
    revert_change_id = '54321'
    revert_for_culprit.revert_cl_url = 'https://%s/q/%s' % (
        self.review_server_host, revert_change_id)
    culprit.revert_cl = revert_for_culprit
    culprit.revert_status = status.COMPLETED
    culprit.put()

    revert_status = gerrit.CREATED_BY_FINDIT
    commit_status = gerrit.CommitRevert(
        SubmitRevertCLParameters(
            cl_key=CLKey(repo_name=repo_name, revision=revision),
            revert_status=revert_status))

    self.assertEqual(gerrit.SKIPPED, commit_status)

    culprit = WfSuspectedCL.Get(repo_name, revision)
    self.assertEqual(culprit.revert_submission_status, status.SKIPPED)

  @mock.patch.object(
      time_util, 'GetUTCNow', return_value=datetime(2017, 2, 1, 5, 0, 0))
  @mock.patch.object(
      codereview_util, 'GetCodeReviewForReview', return_value=_CODEREVIEW)
  @mock.patch.object(rotations, 'current_sheriffs', return_value=['a@b.com'])
  @mock.patch.object(_CODEREVIEW, 'AddReviewers', return_value=True)
  @mock.patch.object(_CODEREVIEW, 'SubmitRevert')
  @mock.patch.object(_CODEREVIEW, 'GetClDetails')
  def testSubmitRevertSucceed(self, mock_fn, mock_commit, mock_add, *_):
    repo_name = 'chromium'
    revision = 'rev1'
    commit_position = 123

    cl_info = ClInfo(self.review_server_host, self.review_change_id)
    cl_info.commits.append(
        Commit('20001', 'rev1', datetime(2017, 2, 1, 0, 0, 0)))
    mock_fn.return_value = cl_info
    mock_commit.return_value = True

    culprit = WfSuspectedCL.Create(repo_name, revision, commit_position)
    revert_for_culprit = RevertCL()
    revert_change_id = '54321'
    revert_for_culprit.revert_cl_url = 'https://%s/q/%s' % (
        self.review_server_host, revert_change_id)
    culprit.revert_cl = revert_for_culprit
    culprit.revert_status = status.COMPLETED
    culprit.put()
    revert_status = gerrit.CREATED_BY_FINDIT
    commit_status = gerrit.CommitRevert(
        SubmitRevertCLParameters(
            cl_key=CLKey(repo_name=repo_name, revision=revision),
            revert_status=revert_status))

    self.assertEqual(gerrit.COMMITTED, commit_status)

    culprit = WfSuspectedCL.Get(repo_name, revision)
    self.assertEqual(culprit.revert_submission_status, status.COMPLETED)

    mock_commit.assert_called_once_with(revert_change_id)

    culprit_link = (
        'https://findit-for-me.appspot.com/waterfall/culprit?key=%s' %
        (culprit.key.urlsafe()))
    false_positive_bug_query = urllib.urlencode({
        'status': 'Available',
        'labels': 'Test-Findit-Wrong',
        'components': 'Tools>Test>FindIt',
        'summary': 'Wrongly blame %s' % revision,
        'comment': 'Detail is %s' % culprit_link
    })
    bug_link = ('https://bugs.chromium.org/p/chromium/issues/entry?%s') % (
        false_positive_bug_query)
    message = textwrap.dedent("""
        Sheriffs, CL owner or CL reviewers:
        Please confirm this revert if it is correct.
        If it is a false positive, please revert and report it
        at %s.

        For more information about Findit auto-revert: %s.

        Sheriffs, it'll be much appreciated if you could take several minutes
        to fill out this survey: %s.""") % (bug_link, gerrit._MANUAL_LINK,
                                            gerrit._SURVEY_LINK)
    mock_add.assert_called_once_with(revert_change_id, ['a@b.com'], message)

  def testUpdateCulprit(self):
    repo_name = 'chromium'
    revision = 'rev1'
    culprit = WfSuspectedCL.Create(repo_name, revision, 123)
    culprit.revert_submission_status = status.RUNNING
    culprit.submit_revert_pipeline_id = 'some_id'
    culprit.put()

    culprit = gerrit._UpdateCulprit(repo_name, revision)
    self.assertEqual(culprit.submit_revert_pipeline_id, 'some_id')

  @mock.patch.object(suspected_cl_util, 'GetCulpritInfo')
  def testShouldNotCommitRevertForAutoRoll(self, mock_info):
    repo_name = 'chromium'
    revision = 'rev1'
    mock_info.return_value = {
        'commit_position': self.culprit_commit_position,
        'code_review_url': self.culprit_code_review_url,
        'review_server_host': self.review_server_host,
        'review_change_id': self.review_change_id,
        'author': {
            'email': 'v8-autoroll@chromium.org'
        }
    }
    culprit = WfSuspectedCL.Create(repo_name, revision, 123)
    culprit.revert_cl = RevertCL()
    culprit.revert_status = status.COMPLETED
    culprit.put()

    self.assertFalse(gerrit._CanAutoCommitRevertByGerrit(repo_name, revision))

  @mock.patch.object(
      codereview_util, 'GetCodeReviewForReview', return_value=None)
  def testSendNotificationForCulpritNoCodeReview(self, _):
    repo_name = 'chromium'
    revision = 'rev1'
    revert_status = gerrit.CREATED_BY_SHERIFF

    culprit = WfSuspectedCL.Create(repo_name, revision, 123)
    culprit.put()

    self.assertFalse(
        gerrit.SendNotificationForCulprit(repo_name, revision, revert_status))

  @mock.patch.object(
      codereview_util, 'GetCodeReviewForReview', return_value=_CODEREVIEW)
  @mock.patch.object(_CODEREVIEW, 'PostMessage', return_value=True)
  def testSendNotificationForCulpritConfirm(self, mock_post, _):
    repo_name = 'chromium'
    revision = 'rev1'
    revert_status = gerrit.CREATED_BY_SHERIFF

    culprit = WfSuspectedCL.Create(repo_name, revision, 123)
    culprit.put()

    self.assertTrue(
        gerrit.SendNotificationForCulprit(repo_name, revision, revert_status))
    message = textwrap.dedent("""
    Findit (https://goo.gl/kROfz5) %s this CL at revision %s as the culprit for
    failures in the build cycles as shown on:
    https://findit-for-me.appspot.com/waterfall/culprit?key=%s""") % (
        'confirmed', self.culprit_commit_position, culprit.key.urlsafe())
    mock_post.assert_called_once_with(self.review_change_id, message, False)

  @mock.patch.object(
      codereview_util, 'GetCodeReviewForReview', return_value=_CODEREVIEW)
  @mock.patch.object(_CODEREVIEW, 'PostMessage', return_value=True)
  def testSendNotificationForCulprit(self, mock_post, _):
    repo_name = 'chromium'
    revision = 'rev1'
    revert_status = None

    culprit = WfSuspectedCL.Create(repo_name, revision, 123)
    culprit.put()

    self.assertTrue(
        gerrit.SendNotificationForCulprit(repo_name, revision, revert_status))
    message = textwrap.dedent("""
    Findit (https://goo.gl/kROfz5) %s this CL at revision %s as the culprit for
    failures in the build cycles as shown on:
    https://findit-for-me.appspot.com/waterfall/culprit?key=%s""") % (
        'identified', self.culprit_commit_position, culprit.key.urlsafe())
    mock_post.assert_called_once_with(self.review_change_id, message, True)

  def testCulpritWasAutoCommitted(self):
    culprit_info = {
        'author': {
            'email': 'skia-chromium-autoroll@skia-buildbots.google.com.'
                     'iam.gserviceaccount.com'
        }
    }
    self.assertTrue(gerrit._CulpritWasAutoCommitted(culprit_info))
