# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from datetime import datetime
import mock
import textwrap
import urllib

from common import constants
from common import rotations
from infra_api_clients.codereview import codereview_util
from infra_api_clients.codereview.cl_info import ClInfo
from infra_api_clients.codereview.cl_info import Commit
from infra_api_clients.codereview.cl_info import Revert
from infra_api_clients.codereview.gerrit import Gerrit
from libs import analysis_status as status
from libs import time_util
from model.base_suspected_cl import RevertCL
from model.wf_suspected_cl import WfSuspectedCL
from pipelines.pipeline_inputs_and_outputs import CLKey
from pipelines.pipeline_inputs_and_outputs import CreateRevertCLPipelineInput
from services import revert
from waterfall import buildbot
from waterfall import suspected_cl_util
from waterfall import waterfall_config
from waterfall.test import wf_testcase

_CODEREVIEW = Gerrit('chromium-review.googlesource.com')


class RevertUtilTest(wf_testcase.WaterfallTestCase):

  def setUp(self):
    super(RevertUtilTest, self).setUp()
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
  def testRevertCLSucceed(self, mock_fn, mock_revert, mock_add, *_):
    repo_name = 'chromium'
    revision = 'rev1'
    commit_position = 123

    cl_info = ClInfo(self.review_server_host, self.review_change_id)
    cl_info.commits.append(
        Commit('20001', 'rev1', datetime(2017, 2, 1, 0, 0, 0)))
    cl_info.owner_email = 'abc@chromium.org'
    mock_fn.return_value = cl_info
    mock_revert.return_value = '54321'

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

    revert_status = revert.RevertCulprit(
        CreateRevertCLPipelineInput(
            cl_key=CLKey(
                repo_name=repo_name.decode('utf-8'),
                revision=revision.decode('utf-8')),
            build_id='m/b/1'.decode('utf-8')))

    self.assertEquals(revert_status, revert.CREATED_BY_FINDIT)

    culprit = WfSuspectedCL.Get(repo_name, revision)
    self.assertEqual(culprit.revert_status, status.COMPLETED)
    self.assertIsNotNone(culprit.revert_cl)

    reason = textwrap.dedent("""
        Findit (https://goo.gl/kROfz5) identified CL at revision %s as the
        culprit for failures in the build cycles as shown on:
        https://findit-for-me.appspot.com/waterfall/culprit?key=%s\n
        Sample Failed Build: %s""") % (commit_position, culprit.key.urlsafe(),
                                       buildbot.CreateBuildUrl('m', 'b', '1'))
    mock_revert.assert_called_once_with(reason, self.review_change_id, '20001')

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
        Please approve and submit this revert if it is correct.
        If it is a false positive, please abandon and report it
        at %s.
        If failed to submit the revert, please abandon it and report the failure
        at %s.""") % (false_positive_bug_link, auto_revert_bug_link)
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

    revert_status = revert.RevertCulprit(
        CreateRevertCLPipelineInput(
            cl_key=CLKey(
                repo_name=repo_name.decode('utf-8'),
                revision=revision.decode('utf-8')),
            build_id='m/b/1'.decode('utf-8')))

    self.assertEquals(revert_status, revert.CREATED_BY_SHERIFF)

    committed = revert.CommitRevert(repo_name, revision, revert_status, None)
    self.assertFalse(committed)

    culprit = WfSuspectedCL.Get(repo_name, revision)
    self.assertEquals(culprit.revert_status, status.SKIPPED)
    self.assertIsNone(culprit.revert_cl)
    self.assertIsNone(culprit.revert_pipeline_id)

  @mock.patch.object(waterfall_config, 'GetActionSettings', return_value={})
  @mock.patch.object(_CODEREVIEW, 'PostMessage', return_value=True)
  @mock.patch.object(_CODEREVIEW, 'AddReviewers', return_value=False)
  @mock.patch.object(rotations, 'current_sheriffs', return_value=['a@b.com'])
  @mock.patch.object(
      codereview_util, 'GetCodeReviewForReview', return_value=_CODEREVIEW)
  @mock.patch.object(_CODEREVIEW, 'GetClDetails')
  def testRevertAddReviewerFailed(self, mock_fn, *_):
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

    WfSuspectedCL.Create(repo_name, revision, 123).put()

    revert_status = revert.RevertCulprit(
        CreateRevertCLPipelineInput(
            cl_key=CLKey(
                repo_name=repo_name.decode('utf-8'),
                revision=revision.decode('utf-8')),
            build_id='m/b/1'.decode('utf-8')))

    self.assertEquals(revert_status, revert.ERROR)

    committed = revert.CommitRevert(repo_name, revision, revert_status, None)
    self.assertFalse(committed)

    culprit = WfSuspectedCL.Get(repo_name, revision)
    self.assertEqual(culprit.revert_status, status.ERROR)
    self.assertIsNotNone(culprit.revert_cl)

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

    revert_status = revert.RevertCulprit(
        CreateRevertCLPipelineInput(
            cl_key=CLKey(
                repo_name=repo_name.decode('utf-8'),
                revision=revision.decode('utf-8')),
            build_id='m/b/1'.decode('utf-8')))

    self.assertEquals(revert_status, revert.CREATED_BY_FINDIT)

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

    revert_status = revert.RevertCulprit(
        CreateRevertCLPipelineInput(
            cl_key=CLKey(
                repo_name=repo_name.decode('utf-8'),
                revision=revision.decode('utf-8')),
            build_id='m/b/1'.decode('utf-8')))

    self.assertEquals(revert_status, revert.SKIPPED)

    culprit = WfSuspectedCL.Get(repo_name, revision)
    self.assertEqual(culprit.revert_status, status.SKIPPED)
    self.assertIsNone(culprit.revert_cl)
    self.assertEqual(culprit.skip_revert_reason, revert.CULPRIT_IS_A_REVERT)

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

    revert_status = revert.RevertCulprit(
        CreateRevertCLPipelineInput(
            cl_key=CLKey(
                repo_name=repo_name.decode('utf-8'),
                revision=revision.decode('utf-8')),
            build_id='m/b/1'.decode('utf-8')))

    self.assertEquals(revert_status, revert.SKIPPED)

    culprit = WfSuspectedCL.Get(repo_name, revision)
    self.assertEqual(culprit.revert_status, status.SKIPPED)
    self.assertIsNone(culprit.revert_cl)
    self.assertEqual(culprit.skip_revert_reason, revert.AUTO_REVERT_OFF)

  def testRevertHasCompleted(self):
    repo_name = 'chromium'
    revision = 'rev1'
    pipeline_id = 'pipeline_id'

    culprit = WfSuspectedCL.Create(repo_name, revision, 123)
    culprit.revert_cl = RevertCL()
    culprit.revert_status = status.COMPLETED
    culprit.put()

    self.assertFalse(revert._CanRevert(repo_name, revision, pipeline_id))

  def testRevertACulpritIsBeingReverted(self):
    repo_name = 'chromium'
    revision = 'rev1'
    pipeline_id = 'another_pipeline'

    culprit = WfSuspectedCL.Create(repo_name, revision, 123)
    culprit.revert_status = status.RUNNING
    culprit.revert_pipeline_id = 'pipeline_id'
    culprit.put()

    self.assertFalse(revert._CanRevert(repo_name, revision, pipeline_id))

  @mock.patch.object(waterfall_config, 'GetActionSettings', return_value={})
  def testRevertTurnedOff(self, _):
    repo_name = 'chromium'
    revision = 'rev1'

    pipeline_input = CreateRevertCLPipelineInput(
        cl_key=CLKey(
            repo_name=repo_name.decode('utf-8'),
            revision=revision.decode('utf-8')),
        build_id='m/b/1'.decode('utf-8'))
    self.assertFalse(revert.ShouldRevert(pipeline_input, None))

  def testShouldNotRevertIfRevertIsSkipped(self):
    repo_name = 'chromium'
    revision = 'rev1'
    culprit = WfSuspectedCL.Create(repo_name, revision, 123)
    culprit.revert_status = status.SKIPPED
    culprit.put()
    pipeline_input = CreateRevertCLPipelineInput(
        cl_key=CLKey(
            repo_name=repo_name.decode('utf-8'),
            revision=revision.decode('utf-8')),
        build_id='m/b/1'.decode('utf-8'))
    self.assertFalse(revert.ShouldRevert(pipeline_input, None))

  @mock.patch.object(
      revert, '_GetDailyNumberOfRevertedCulprits', return_value=10)
  def testAutoRevertExceedsLimit(self, _):
    repo_name = 'chromium'
    revision = 'rev1'

    culprit = WfSuspectedCL.Create(repo_name, revision, 123)
    culprit.put()

    pipeline_input = CreateRevertCLPipelineInput(
        cl_key=CLKey(
            repo_name=repo_name.decode('utf-8'),
            revision=revision.decode('utf-8')),
        build_id='m/b/1'.decode('utf-8'))
    self.assertFalse(revert.ShouldRevert(pipeline_input, None))

  @mock.patch.object(revert, '_CanRevert', return_value=True)
  def testShouldRevert(self, _):
    repo_name = 'chromium'
    revision = 'rev1'
    pipeline_id = 'pipeline_id'
    pipeline_input = CreateRevertCLPipelineInput(
        cl_key=CLKey(
            repo_name=repo_name.decode('utf-8'),
            revision=revision.decode('utf-8')),
        build_id='m/b/1'.decode('utf-8'))
    self.assertTrue(revert.ShouldRevert(pipeline_input, pipeline_id))

  @mock.patch.object(
      time_util, 'GetUTCNow', return_value=datetime(2017, 02, 01, 0, 0, 0))
  def testGetDailyNumberOfRevertedCulprits(self, _):
    self.assertEqual(0, revert._GetDailyNumberOfRevertedCulprits(10))

  def testRevertHasCommitted(self):
    repo_name = 'chromium'
    revision = 'rev1'

    culprit = WfSuspectedCL.Create(repo_name, revision, 123)
    culprit.revert_cl = RevertCL()
    culprit.revert_submission_status = status.COMPLETED
    culprit.put()

    revert_status = revert.CREATED_BY_FINDIT
    committed = revert.CommitRevert(repo_name, revision, revert_status,
                                    'pipeline_id')

    self.assertFalse(committed)

  @mock.patch.object(revert, '_GetDailyNumberOfCommits', return_value=4)
  def testCommitExceedsLimit(self, _):
    repo_name = 'chromium'
    revision = 'rev1'

    culprit = WfSuspectedCL.Create(repo_name, revision, 123)
    culprit.revert_cl = RevertCL()
    culprit.revert_status = status.COMPLETED
    culprit.put()

    revert_status = revert.CREATED_BY_FINDIT
    committed = revert.CommitRevert(repo_name, revision, revert_status,
                                    'pipeline_id')

    self.assertFalse(committed)

  @mock.patch.object(waterfall_config, 'GetActionSettings', return_value={})
  def testRevertTurnedOffNoCommit(self, _):
    repo_name = 'chromium'
    revision = 'rev1'

    revert_status = revert.CREATED_BY_FINDIT
    committed = revert.CommitRevert(repo_name, revision, revert_status,
                                    'pipeline_id')

    self.assertFalse(committed)

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

    revert_status = revert.CREATED_BY_FINDIT
    committed = revert.CommitRevert(repo_name, revision, revert_status,
                                    'pipeline_id')
    self.assertFalse(committed)

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

    revert_status = revert.CREATED_BY_FINDIT
    committed = revert.CommitRevert(repo_name, revision, revert_status,
                                    'pipeline_id')

    self.assertFalse(committed)
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
        Please approve and submit this revert if it is correct.
        If it is a false positive, please abandon and report it
        at %s.
        If failed to submit the revert, please abandon it and report the failure
        at %s.""") % (false_positive_bug_link, auto_revert_bug_link)
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

    revert_status = revert.CREATED_BY_FINDIT
    committed = revert.CommitRevert(repo_name, revision, revert_status,
                                    'pipeline_id')

    self.assertFalse(committed)

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
    revert_status = revert.CREATED_BY_FINDIT
    committed = revert.CommitRevert(repo_name, revision, revert_status,
                                    'pipeline_id')

    self.assertTrue(committed)

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
        at %s.""") % bug_link
    mock_add.assert_called_once_with(revert_change_id, ['a@b.com'], message)

  def testUpdateCulprit(self):
    repo_name = 'chromium'
    revision = 'rev1'
    culprit = WfSuspectedCL.Create(repo_name, revision, 123)
    culprit.revert_submission_status = status.RUNNING
    culprit.submit_revert_pipeline_id = 'some_id'
    culprit.put()

    culprit = revert._UpdateCulprit(repo_name, revision)
    self.assertEqual(culprit.submit_revert_pipeline_id, 'some_id')

  @mock.patch.object(revert, '_CanCommitRevert', return_value=True)
  @mock.patch.object(suspected_cl_util, 'GetCulpritInfo')
  def testShouldNotCommitRevertForAutoRoll(self, mock_info, _):
    repo_name = 'chromium'
    revision = 'rev1'
    revert_status = revert.CREATED_BY_FINDIT
    pipeline_id = 'pipeline_id'
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

    self.assertFalse(
        revert._ShouldCommitRevert(repo_name, revision, revert_status,
                                   pipeline_id))
