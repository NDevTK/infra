# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from datetime import datetime
import mock

from buildbucket_proto import common_pb2
from buildbucket_proto.build_pb2 import Build
from buildbucket_proto.build_pb2 import BuilderID

from common.waterfall import buildbucket_client
from findit_v2.model.compile_failure import CompileFailure
from findit_v2.model.compile_failure import CompileFailureAnalysis
from findit_v2.model.compile_failure import CompileRerunBuild
from findit_v2.model.gitiles_commit import GitilesCommit
from findit_v2.model.luci_build import LuciFailedBuild
from findit_v2.services.analysis.compile_failure import (
    compile_failure_rerun_analysis)
from findit_v2.services.chromium_api import ChromiumProjectAPI
from findit_v2.services.context import Context
from findit_v2.services.failure_type import StepTypeEnum
from waterfall.test import wf_testcase


class CompileFailureRerunAnalysisTest(wf_testcase.TestCase):

  def setUp(self):
    super(CompileFailureRerunAnalysisTest, self).setUp()
    self.build_id = 8000000000123
    self.build_number = 123
    self.builder = BuilderID(
        project='chromium', bucket='ci', builder='linux-rel')
    self.build = Build(
        id=self.build_id,
        builder=self.builder,
        number=self.build_number,
        status=common_pb2.FAILURE)
    self.build.input.gitiles_commit.host = 'gitiles.host.com'
    self.build.input.gitiles_commit.project = 'project/name'
    self.build.input.gitiles_commit.ref = 'ref/heads/master'
    self.build.input.gitiles_commit.id = 'git_sha_123'
    self.build.create_time.FromDatetime(datetime(2019, 4, 9))
    self.build.start_time.FromDatetime(datetime(2019, 4, 9, 0, 1))
    self.build.end_time.FromDatetime(datetime(2019, 4, 9, 1))

    self.context = Context(
        luci_project_name='chromium',
        gitiles_host='gitiles.host.com',
        gitiles_project='project/name',
        gitiles_ref='ref/heads/master',
        gitiles_id='git_sha')

    build_entity = LuciFailedBuild.Create(
        luci_project=self.context.luci_project_name,
        luci_bucket=self.build.builder.bucket,
        luci_builder=self.build.builder.builder,
        build_id=self.build_id,
        legacy_build_number=self.build_number,
        gitiles_host=self.context.gitiles_host,
        gitiles_project=self.context.gitiles_project,
        gitiles_ref=self.context.gitiles_ref,
        gitiles_id=self.context.gitiles_id,
        commit_position=6000005,
        status=20,
        create_time=datetime(2019, 3, 28),
        start_time=datetime(2019, 3, 28, 0, 1),
        end_time=datetime(2019, 3, 28, 1),
        build_failure_type=StepTypeEnum.COMPILE)
    build_entity.put()

    self.compile_failure = CompileFailure.Create(
        failed_build_key=build_entity.key,
        step_ui_name='compile',
        output_targets=['a.o'],
        first_failed_build_id=self.build_id,
        failure_group_build_id=None)
    self.compile_failure.put()

    self.analysis = CompileFailureAnalysis.Create(
        luci_project=self.context.luci_project_name,
        luci_bucket=self.build.builder.bucket,
        luci_builder=self.build.builder.builder,
        build_id=self.build_id,
        gitiles_host=self.context.gitiles_host,
        gitiles_project=self.context.gitiles_project,
        gitiles_ref=self.context.gitiles_ref,
        last_passed_gitiles_id='left_sha',
        last_passed_cp=6000000,
        first_failed_gitiles_id=self.context.gitiles_id,
        first_failed_cp=6000005,
        rerun_builder_id='chromium/findit/findit-variables',
        compile_failure_keys=[self.compile_failure.key])
    self.analysis.Save()

  def _CreateCompileRerunBuild(self, commit_position=6000002):
    rerun_commit = GitilesCommit(
        gitiles_host=self.context.gitiles_host,
        gitiles_project=self.context.gitiles_project,
        gitiles_ref=self.context.gitiles_ref,
        gitiles_id=str(commit_position),
        commit_position=commit_position)

    rerun_builder = BuilderID(
        project='chromium', bucket='findit', builder='findit-variables')

    rerun_build = CompileRerunBuild.Create(
        luci_project=rerun_builder.project,
        luci_bucket=rerun_builder.bucket,
        luci_builder=rerun_builder.builder,
        build_id=8000000000789,
        legacy_build_number=60789,
        gitiles_host=rerun_commit.gitiles_host,
        gitiles_project=rerun_commit.gitiles_project,
        gitiles_ref=rerun_commit.gitiles_ref,
        gitiles_id=rerun_commit.gitiles_id,
        commit_position=rerun_commit.commit_position,
        status=1,
        create_time=datetime(2019, 3, 28),
        parent_key=self.analysis.key)
    rerun_build.put()
    return rerun_build

  @mock.patch.object(
      ChromiumProjectAPI,
      'GetCompileRerunBuildInputProperties',
      return_value={'recipe': 'compile'})
  @mock.patch.object(buildbucket_client, 'TriggerV2Build')
  def testTriggerRerunBuild(self, mock_trigger_build, _):
    new_build_id = 800000024324
    new_build = Build(id=new_build_id, number=300)
    new_build.status = common_pb2.SCHEDULED
    new_build.create_time.FromDatetime(datetime(2019, 4, 20))
    rerun_builder = BuilderID(
        project='chromium', bucket='findit', builder='findit-variables')
    rerun_commit = GitilesCommit(
        gitiles_host=self.context.gitiles_host,
        gitiles_project=self.context.gitiles_project,
        gitiles_ref=self.context.gitiles_ref,
        gitiles_id='6000002',
        commit_position=6000002)
    output_targets = {'compile': ['a.o']}

    mock_trigger_build.return_value = new_build

    compile_failure_rerun_analysis.TriggerRerunBuild(
        self.context, self.build_id, self.build, self.analysis.key,
        rerun_builder, rerun_commit, output_targets)

    rerun_build = CompileRerunBuild.get_by_id(
        new_build_id, parent=self.analysis.key)
    self.assertIsNotNone(rerun_build)

  @mock.patch.object(
      ChromiumProjectAPI,
      'GetCompileRerunBuildInputProperties',
      return_value={'recipe': 'compile'})
  @mock.patch.object(buildbucket_client, 'TriggerV2Build')
  def testTriggerRerunBuildFoundRunningBuild(self, mock_trigger_build, _):
    """This test is for the case where there's already an existing rerun build,
      so no new rerun-build should be scheduled."""
    rerun_commit = GitilesCommit(
        gitiles_host=self.context.gitiles_host,
        gitiles_project=self.context.gitiles_project,
        gitiles_ref=self.context.gitiles_ref,
        gitiles_id='6000002',
        commit_position=6000002)

    rerun_builder = BuilderID(
        project='chromium', bucket='findit', builder='findit-variables')
    output_targets = {'compile': ['a.o']}

    self._CreateCompileRerunBuild()

    compile_failure_rerun_analysis.TriggerRerunBuild(
        self.context, self.build_id, self.build, self.analysis.key,
        rerun_builder, rerun_commit, output_targets)

    self.assertFalse(mock_trigger_build.called)

  @mock.patch.object(
      ChromiumProjectAPI,
      'GetCompileRerunBuildInputProperties',
      return_value=None)
  @mock.patch.object(buildbucket_client, 'TriggerV2Build')
  def testTriggerRerunBuildFailedToGetProperty(self, mock_trigger_build, _):
    """This test is for the case where there's already an existing rerun build,
      so no new rerun-build should be scheduled."""
    rerun_commit = GitilesCommit(
        gitiles_host=self.context.gitiles_host,
        gitiles_project=self.context.gitiles_project,
        gitiles_ref=self.context.gitiles_ref,
        gitiles_id='6000002',
        commit_position=6000002)

    rerun_builder = BuilderID(
        project='chromium', bucket='findit', builder='findit-variables')
    output_targets = {'compile': ['a.o']}

    compile_failure_rerun_analysis.TriggerRerunBuild(
        self.context, self.build_id, self.build, self.analysis.key,
        rerun_builder, rerun_commit, output_targets)

    self.assertFalse(mock_trigger_build.called)

  @mock.patch.object(
      ChromiumProjectAPI,
      'GetCompileRerunBuildInputProperties',
      return_value={'recipe': 'compile'})
  @mock.patch.object(buildbucket_client, 'TriggerV2Build', return_value=None)
  def testTriggerRerunBuildFailedToTriggerBuild(self, mock_trigger_build, _):
    """This test is for the case where there's already an existing rerun build,
      so no new rerun-build should be scheduled."""
    rerun_commit = GitilesCommit(
        gitiles_host=self.context.gitiles_host,
        gitiles_project=self.context.gitiles_project,
        gitiles_ref=self.context.gitiles_ref,
        gitiles_id='6000002',
        commit_position=6000002)

    rerun_builder = BuilderID(
        project='chromium', bucket='findit', builder='findit-variables')
    output_targets = {'compile': ['a.o']}

    compile_failure_rerun_analysis.TriggerRerunBuild(
        self.context, self.build_id, self.build, self.analysis.key,
        rerun_builder, rerun_commit, output_targets)

    self.assertTrue(mock_trigger_build.called)
    rerun_builds = CompileRerunBuild.query(ancestor=self.analysis.key).fetch()
    self.assertEqual([], rerun_builds)

  def testGetRegressionRangesForCompileFailuresNoRerunBuilds(self):
    result = (
        compile_failure_rerun_analysis._GetRegressionRangesForCompileFailures(
            self.analysis))
    self.assertItemsEqual([self.compile_failure], result[0]['failures'])
    self.assertEqual(6000000, result[0]['last_passed_commit'].commit_position)
    self.assertEqual(6000005, result[0]['first_failed_commit'].commit_position)

  def testGetRegressionRangesForCompileFailures(self):
    rerun_build_failures = {
        'compile': {
            'failures': {
                'target_str': {
                    'output_targets': ['a.o'],
                    'rule': 'CXX'
                }
            }
        }
    }

    rerun_build = self._CreateCompileRerunBuild()
    rerun_build.SaveRerunBuildResults(20, rerun_build_failures)

    results = (
        compile_failure_rerun_analysis._GetRegressionRangesForCompileFailures(
            self.analysis))
    self.assertEqual([self.compile_failure], results[0]['failures'])
    self.assertEqual(6000002, results[0]['first_failed_commit'].commit_position)
    self.assertEqual(6000000, results[0]['last_passed_commit'].commit_position)
