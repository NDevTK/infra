# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from datetime import datetime
import mock

from buildbucket_proto import common_pb2
from buildbucket_proto.build_pb2 import Build
from buildbucket_proto.build_pb2 import BuilderID

from findit_v2.model.compile_failure import CompileFailureGroup
from findit_v2.model.luci_build import LuciFailedBuild
from findit_v2.services.analysis.compile_failure.compile_analysis_api import (
    CompileAnalysisAPI)
from findit_v2.services.chromium_api import ChromiumProjectAPI
from findit_v2.services.context import Context
from findit_v2.services.failure_type import StepTypeEnum
from services import git
from waterfall.test import wf_testcase


class CompileAnalysisAPITest(wf_testcase.TestCase):

  def _MockBuild(self,
                 build_id,
                 build_number,
                 gitiles_commit_id,
                 builder_name='Linux Tests',
                 build_status=common_pb2.FAILURE):
    builder = BuilderID(project='chromium', bucket='ci', builder=builder_name)
    build = Build(
        id=build_id, builder=builder, number=build_number, status=build_status)
    build.input.gitiles_commit.host = 'gitiles.host.com'
    build.input.gitiles_commit.project = 'project/name'
    build.input.gitiles_commit.ref = 'ref/heads/master'
    build.input.gitiles_commit.id = gitiles_commit_id
    build.create_time.FromDatetime(datetime(2019, 4, 9))
    build.start_time.FromDatetime(datetime(2019, 4, 9, 0, 1))
    build.end_time.FromDatetime(datetime(2019, 4, 9, 1))
    return build

  def setUp(self):
    super(CompileAnalysisAPITest, self).setUp()
    self.build_id = 8000000000123
    self.build_number = 123
    self.builder = BuilderID(
        project='chromium', bucket='ci', builder='Linux Tests')
    self.build = self._MockBuild(self.build_id, self.build_number,
                                 'git_sha_123')
    self.rerun_builder = BuilderID(
        project='chromium', bucket='try', builder='findit_variables')

    self.context = Context(
        luci_project_name='chromium',
        gitiles_host='gitiles.host.com',
        gitiles_project='project/name',
        gitiles_ref='ref/heads/master',
        gitiles_id='git_sha_123')

    self.build_entity = LuciFailedBuild.Create(
        luci_project=self.build.builder.project,
        luci_bucket=self.build.builder.bucket,
        luci_builder=self.build.builder.builder,
        build_id=self.build.id,
        legacy_build_number=self.build.number,
        gitiles_host=self.context.gitiles_host,
        gitiles_project=self.context.gitiles_project,
        gitiles_ref=self.context.gitiles_ref,
        gitiles_id=self.context.gitiles_id,
        commit_position=123,
        status=self.build.status,
        create_time=self.build.create_time.ToDatetime(),
        start_time=self.build.start_time.ToDatetime(),
        end_time=self.build.end_time.ToDatetime(),
        build_failure_type=StepTypeEnum.COMPILE)
    self.build_entity.put()

    self.analysis_api = CompileAnalysisAPI()

    self.compile_failure = self.analysis_api.CreateFailure(
        self.build_entity.key, 'compile', self.build_id, 8000000000122, None,
        None, None)
    self.compile_failure.put()

  @mock.patch.object(git, 'GetCommitPositionFromRevision', return_value=67890)
  def testEntitiesCreation(self, _):
    self.analysis_api.CreateAndSaveFailureGroup(
        self.context, self.build, [self.compile_failure.key], '122', 122, 123)
    groups = CompileFailureGroup.query().fetch()
    self.assertEqual(1, len(groups))
    self.assertEqual(self.build_id, groups[0].key.id())

    analysis = self.analysis_api.CreateAndSaveFailureAnalysis(
        'chromium', self.context, self.build, 'git_sha_122', 122, 123,
        'preject/bucket/builder', [self.compile_failure.key])
    self.assertIsNotNone(analysis)
    self.assertEqual(self.build_id, analysis.build_id)

  def testAPIStepType(self):
    self.assertEqual(StepTypeEnum.COMPILE, self.analysis_api.step_type)

  def testGetFailureEntitiesForABuild(self):
    failure_entities = self.analysis_api.GetFailureEntitiesForABuild(self.build)
    self.assertEqual(1, len(failure_entities))
    self.assertEqual(self.compile_failure, failure_entities[0])

  @mock.patch.object(git, 'GetCommitPositionFromRevision', return_value=67890)
  def testGetMergedFailureKey(self, _):
    self.assertIsNone(
        self.analysis_api.GetMergedFailureKey([self.compile_failure], None,
                                              'compile', None))

  @mock.patch.object(ChromiumProjectAPI, 'GetCompileFailures')
  def testGetFailuresInBuild(self, mock_compile_failure):
    self.analysis_api.GetFailuresInBuild(ChromiumProjectAPI(), self.build,
                                         ['compile'])
    self.assertTrue(mock_compile_failure.called)

  @mock.patch.object(ChromiumProjectAPI,
                     'GetFailuresWithMatchingCompileFailureGroups')
  def testGetFailuresWithMatchingFailureGroups(self, mock_failures_in_group):
    self.analysis_api.GetFailuresWithMatchingFailureGroups(
        self.context, self.build, {})
    self.assertTrue(mock_failures_in_group.called)
