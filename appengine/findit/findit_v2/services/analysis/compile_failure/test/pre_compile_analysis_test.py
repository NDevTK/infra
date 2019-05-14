# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from datetime import datetime
import mock

from buildbucket_proto import common_pb2
from buildbucket_proto.build_pb2 import Build
from buildbucket_proto.build_pb2 import BuilderID
from buildbucket_proto.rpc_pb2 import SearchBuildsResponse
from buildbucket_proto.step_pb2 import Step

from common.waterfall import buildbucket_client
from findit_v2.model.compile_failure import CompileFailure
from findit_v2.model.compile_failure import CompileFailureAnalysis
from findit_v2.model.luci_build import LuciFailedBuild
from findit_v2.services.analysis.compile_failure import pre_compile_analysis
from findit_v2.services.chromium_api import ChromiumProjectAPI
from findit_v2.services.context import Context
from services import git
from waterfall.test import wf_testcase


class PreCompileAnalysisTest(wf_testcase.TestCase):

  def setUp(self):
    super(PreCompileAnalysisTest, self).setUp()
    self.build_id = 8000000000123
    self.build_number = 123
    self.builder = BuilderID(
        project='chromium', bucket='try', builder='linux-rel')
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

    self.build_info = {
        'id': 8000000000123,
        'number': self.build_number,
        'commit_id': 'git_sha_123'
    }

  @mock.patch.object(git, 'GetCommitPositionFromRevision', return_value=67890)
  def testSaveCompileFailures(self, _):
    detailed_compile_failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': {
                        'id': 8000000000121,
                        'number': 121,
                        'commit_id': 'git_sha'
                    },
                    'last_passed_build': {
                        'id': 8000000000120,
                        'number': 120,
                        'commit_id': 'git_sha'
                    },
                },
            },
            'first_failed_build': {
                'id': 8000000000121,
                'number': 121,
                'commit_id': 'git_sha'
            },
            'last_passed_build': {
                'id': 8000000000120,
                'number': 120,
                'commit_id': 'git_sha'
            },
        },
    }

    pre_compile_analysis.SaveCompileFailures(self.context, self.build,
                                             detailed_compile_failures)

    build = LuciFailedBuild.get_by_id(self.build_id)
    self.assertIsNotNone(build)

    compile_failures = CompileFailure.query(ancestor=build.key).fetch()
    self.assertEqual(1, len(compile_failures))
    self.assertEqual(8000000000121, compile_failures[0].first_failed_build_id)

  @mock.patch.object(git, 'GetCommitPositionFromRevision', return_value=67890)
  def testSaveCompileFailuresOnlyStepLevelFailures(self, _):
    detailed_compile_failures = {
        'compile': {
            'failures': {},
            'first_failed_build': {
                'id': 8000000000121,
                'number': 121,
                'commit_id': 'git_sha'
            },
            'last_passed_build': {
                'id': 8000000000120,
                'number': 120,
                'commit_id': 'git_sha'
            },
        },
    }

    pre_compile_analysis.SaveCompileFailures(self.context, self.build,
                                             detailed_compile_failures)

    build_entity = LuciFailedBuild.get_by_id(self.build_id)
    self.assertIsNotNone(build_entity)

    compile_failures = CompileFailure.query(ancestor=build_entity.key).fetch()
    self.assertEqual(1, len(compile_failures))
    self.assertEqual(8000000000121, compile_failures[0].first_failed_build_id)
    self.assertEqual([], compile_failures[0].output_targets)

  @mock.patch.object(ChromiumProjectAPI, 'GetCompileFailures')
  @mock.patch.object(buildbucket_client, 'GetV2Build')
  @mock.patch.object(buildbucket_client, 'SearchV2BuildsOnBuilder')
  def testUpdateCompileFailuresWithFirstFailureInfo(
      self, mock_prev_builds, mock_get_build, mock_prev_failures):
    """Test for the most common case: found both first_failed_build_id and
      last_passed_build_id."""
    mock_step = Step()
    mock_step.name = 'compile'
    mock_step.status = common_pb2.FAILURE
    build_122_id = 8000000000122
    build_122 = Build(
        id=build_122_id,
        builder=self.builder,
        number=self.build_number - 1,
        status=common_pb2.FAILURE)
    build_122.steps.extend([mock_step])
    build_122.input.gitiles_commit.id = 'git_sha_122'
    build_122_info = {
        'id': build_122_id,
        'number': self.build_number - 1,
        'commit_id': 'git_sha_122'
    }

    build_121_id = 8000000000121
    build_121 = Build(
        id=build_121_id,
        builder=self.builder,
        number=self.build_number - 2,
        status=common_pb2.SUCCESS)
    build_121.input.gitiles_commit.id = 'git_sha_121'
    build_121_info = {
        'id': build_121_id,
        'number': self.build_number - 2,
        'commit_id': 'git_sha_121'
    }

    mock_prev_builds.return_value = SearchBuildsResponse(
        builds=[build_122, build_121])
    mock_get_build.return_value = build_122

    mock_prev_failures.return_value = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': build_122_info,
                    'last_passed_build': None,
                },
            },
            'first_failed_build': build_122_info,
            'last_passed_build': None,
        },
    }

    detailed_compile_failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': self.build_info,
                    'last_passed_build': None,
                },
            },
            'first_failed_build': self.build_info,
            'last_passed_build': None,
        },
    }

    pre_compile_analysis.UpdateCompileFailuresWithFirstFailureInfo(
        self.context, self.build, detailed_compile_failures)

    expected_failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': build_122_info,
                    'last_passed_build': build_121_info,
                },
            },
            'first_failed_build': build_122_info,
            'last_passed_build': build_121_info,
        },
    }

    self.assertEqual(expected_failures, detailed_compile_failures)

  @mock.patch.object(buildbucket_client, 'GetV2Build')
  @mock.patch.object(buildbucket_client, 'SearchV2BuildsOnBuilder')
  def testUpdateCompileFailuresWithFirstFailureInfoPrevBuildDifferentStep(
      self, mock_prev_builds, mock_get_build):
    """Test for previous build failed with different steps."""
    mock_step = Step()
    mock_step.name = 'test'
    mock_step.status = common_pb2.FAILURE
    mock_step1 = Step()
    mock_step1.name = 'compile'
    mock_step1.status = common_pb2.SUCCESS
    build_122_id = 8000000000122
    build_122 = Build(
        id=build_122_id,
        builder=self.builder,
        number=self.build_number - 1,
        status=common_pb2.FAILURE)
    build_122.steps.extend([mock_step, mock_step1])
    build_122.input.gitiles_commit.id = 'git_sha_122'
    build_122_info = {
        'id': build_122_id,
        'number': self.build_number - 1,
        'commit_id': 'git_sha_122'
    }

    build_121_id = 8000000000121
    build_121 = Build(
        id=build_121_id,
        builder=self.builder,
        number=self.build_number - 2,
        status=common_pb2.SUCCESS)

    mock_prev_builds.return_value = SearchBuildsResponse(
        builds=[build_122, build_121])
    mock_get_build.return_value = build_122

    detailed_compile_failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': self.build_info,
                    'last_passed_build': None,
                },
            },
            'first_failed_build': self.build_info,
            'last_passed_build': None,
        },
    }

    pre_compile_analysis.UpdateCompileFailuresWithFirstFailureInfo(
        self.context, self.build, detailed_compile_failures)

    expected_failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': self.build_info,
                    'last_passed_build': build_122_info,
                },
            },
            'first_failed_build': self.build_info,
            'last_passed_build': build_122_info,
        },
    }
    self.assertEqual(expected_failures, detailed_compile_failures)

  @mock.patch.object(buildbucket_client, 'GetV2Build')
  @mock.patch.object(buildbucket_client, 'SearchV2BuildsOnBuilder')
  def testUpdateCompileFailuresWithFirstFailureInfoPrevBuildNoCompile(
      self, mock_prev_builds, mock_get_build):
    """Test for previous build didn't run compile."""
    mock_step = Step()
    mock_step.name = 'test'
    mock_step.status = common_pb2.FAILURE
    build_122_id = 8000000000122
    build_122 = Build(
        id=build_122_id,
        builder=self.builder,
        number=self.build_number - 1,
        status=common_pb2.FAILURE)
    build_122.steps.extend([mock_step])

    build_121_id = 8000000000121
    build_121 = Build(
        id=build_121_id,
        builder=self.builder,
        number=self.build_number - 2,
        status=common_pb2.SUCCESS)
    build_121.input.gitiles_commit.id = 'git_sha_121'
    build_121_info = {
        'id': build_121_id,
        'number': self.build_number - 2,
        'commit_id': 'git_sha_121'
    }

    mock_prev_builds.return_value = SearchBuildsResponse(
        builds=[build_122, build_121])
    mock_get_build.return_value = build_122

    detailed_compile_failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': self.build_info,
                    'last_passed_build': None,
                },
            },
            'first_failed_build': self.build_info,
            'last_passed_build': None,
        },
    }

    pre_compile_analysis.UpdateCompileFailuresWithFirstFailureInfo(
        self.context, self.build, detailed_compile_failures)

    expected_failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': self.build_info,
                    'last_passed_build': build_121_info,
                },
            },
            'first_failed_build': self.build_info,
            'last_passed_build': build_121_info,
        },
    }
    self.assertEqual(expected_failures, detailed_compile_failures)

  @mock.patch.object(ChromiumProjectAPI, 'GetCompileFailures')
  @mock.patch.object(buildbucket_client, 'GetV2Build')
  @mock.patch.object(buildbucket_client, 'SearchV2BuildsOnBuilder')
  def testUpdateCompileFailuresWithFirstFailureInfoDifferentFirstFailure(
      self, mock_prev_builds, mock_get_build, mock_prev_failures):
    """Test for targets in current build failed from different builds."""
    mock_step = Step()
    mock_step.name = 'compile'
    mock_step.status = common_pb2.FAILURE
    build_122_id = 8000000000122
    build_122 = Build(
        id=build_122_id,
        builder=self.builder,
        number=self.build_number - 1,
        status=common_pb2.FAILURE)
    build_122.steps.extend([mock_step])
    build_122.input.gitiles_commit.id = 'git_sha_122'
    build_122_info = {
        'id': build_122_id,
        'number': self.build_number - 1,
        'commit_id': 'git_sha_122'
    }

    mock_step1 = Step()
    mock_step1.name = 'compile'
    mock_step1.status = common_pb2.FAILURE
    build_121_id = 8000000000121
    build_121 = Build(
        id=build_121_id,
        builder=self.builder,
        number=self.build_number - 2,
        status=common_pb2.FAILURE)
    build_121.steps.extend([mock_step1])
    build_121.input.gitiles_commit.id = 'git_sha_121'
    build_121_info = {
        'id': build_121_id,
        'number': self.build_number - 2,
        'commit_id': 'git_sha_121'
    }

    mock_step2 = Step()
    mock_step2.name = 'compile'
    mock_step2.status = common_pb2.FAILURE
    build_120_id = 8000000000121
    build_120 = Build(
        id=build_120_id,
        builder=self.builder,
        number=self.build_number - 3,
        status=common_pb2.FAILURE)
    build_120.steps.extend([mock_step2])
    build_120.input.gitiles_commit.id = 'git_sha_120'
    build_120_info = {
        'id': build_120_id,
        'number': self.build_number - 3,
        'commit_id': 'git_sha_120'
    }

    mock_prev_builds.return_value = SearchBuildsResponse(
        builds=[build_122, build_121, build_120])
    mock_get_build.side_effect = [build_122, build_121, build_120]

    # Failed compiling target3 but successfully compiled target1&2.
    failures_122 = {
        'compile': {
            'failures': {
                frozenset(['target3']): {
                    'rule': 'ACTION',
                    'first_failed_build': build_122_info,
                    'last_passed_build': None,
                },
            },
            'first_failed_build': build_122_info,
            'last_passed_build': None,
        },
    }
    # Has the same failed targets as current build.
    failures_121 = {
        'compile': {
            'failures': {
                frozenset(['target3']): {
                    'rule': 'ACTION',
                    'first_failed_build': build_121_info,
                    'last_passed_build': None,
                },
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': build_121_info,
                    'last_passed_build': None,
                },
            },
            'first_failed_build': build_121_info,
            'last_passed_build': None,
        },
    }
    # Failed compile step, but only different targets.
    failures_120 = {
        'compile': {
            'failures': {
                frozenset(['target4']): {
                    'rule': 'CC',
                    'first_failed_build': build_120_info,
                    'last_passed_build': None,
                },
            },
            'first_failed_build': build_120_info,
            'last_passed_build': None,
        },
    }
    mock_prev_failures.side_effect = [failures_122, failures_121, failures_120]

    detailed_compile_failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': self.build_info,
                    'last_passed_build': None,
                },
                frozenset(['target3']): {
                    'rule': 'ACTION',
                    'first_failed_build': self.build_info,
                    'last_passed_build': None,
                },
            },
            'first_failed_build': self.build_info,
            'last_passed_build': None,
        },
    }

    pre_compile_analysis.UpdateCompileFailuresWithFirstFailureInfo(
        self.context, self.build, detailed_compile_failures)

    expected_failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': self.build_info,
                    'last_passed_build': build_122_info,
                },
                frozenset(['target3']): {
                    'rule': 'ACTION',
                    'first_failed_build': build_121_info,
                    'last_passed_build': build_120_info,
                },
            },
            'first_failed_build': build_121_info,
            'last_passed_build': build_120_info,
        },
    }

    self.assertEqual(expected_failures, detailed_compile_failures)

  @mock.patch.object(buildbucket_client, 'GetV2Build')
  @mock.patch.object(buildbucket_client, 'SearchV2BuildsOnBuilder')
  def testUpdateCompileFailuresWithFirstFailureInfoPrevBuildInfraFailure(
      self, mock_prev_builds, mock_get_build):
    """Test for previous build failed with different steps."""
    mock_step1 = Step()
    mock_step1.name = 'compile'
    mock_step1.status = common_pb2.INFRA_FAILURE
    build_122_id = 8000000000122
    build_122 = Build(
        id=build_122_id,
        builder=self.builder,
        number=self.build_number - 1,
        status=common_pb2.FAILURE)
    build_122.steps.extend([mock_step1])

    build_121_id = 8000000000121
    build_121 = Build(
        id=build_121_id,
        builder=self.builder,
        number=self.build_number - 2,
        status=common_pb2.SUCCESS)
    build_121.input.gitiles_commit.id = 'git_sha_121'
    build_121_info = {
        'id': build_121_id,
        'number': self.build_number - 2,
        'commit_id': 'git_sha_121'
    }

    mock_prev_builds.return_value = SearchBuildsResponse(
        builds=[build_122, build_121])
    mock_get_build.return_value = build_122

    detailed_compile_failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': self.build_info,
                    'last_passed_build': None,
                },
            },
            'first_failed_build': self.build_info,
            'last_passed_build': None,
        },
    }

    pre_compile_analysis.UpdateCompileFailuresWithFirstFailureInfo(
        self.context, self.build, detailed_compile_failures)

    expected_failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': self.build_info,
                    'last_passed_build': build_121_info,
                },
            },
            'first_failed_build': self.build_info,
            'last_passed_build': build_121_info,
        },
    }
    self.assertEqual(expected_failures, detailed_compile_failures)

  def testGetFirstFailuresInCurrentBuild(self):
    build_122_info = {
        'id': 8000000000122,
        'number': self.build_number - 1,
        'commit_id': 'git_sha_122'
    }

    failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': self.build_info,
                    'last_passed_build': build_122_info,
                },
            },
            'first_failed_build': self.build_info,
            'last_passed_build': build_122_info,
        },
    }

    expected_res = {
        'failures': {
            'compile': {
                'output_targets': [{'target1', 'target2'}],
                'last_passed_build': build_122_info,
            },
        },
        'last_passed_build': build_122_info
    }

    self.assertEqual(
        expected_res,
        pre_compile_analysis.GetFirstFailuresInCurrentBuild(
            self.context, self.build, failures))

  def testGetFirstFailuresInCurrentBuildNoFirstFailures(self):
    build_122_info = {
        'id': 8000000000122,
        'number': self.build_number - 1,
        'commit_id': 'git_sha_122'
    }

    build_121_info = {
        'id': 8000000000121,
        'number': self.build_number - 2,
        'commit_id': 'git_sha_121'
    }

    failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': build_122_info,
                    'last_passed_build': build_121_info,
                },
            },
            'first_failed_build': build_122_info,
            'last_passed_build': build_121_info,
        },
    }

    expected_res = {'failures': {}, 'last_passed_build': None}

    self.assertEqual(
        expected_res,
        pre_compile_analysis.GetFirstFailuresInCurrentBuild(
            self.context, self.build, failures))

  def testGetFirstFailuresInCurrentBuildNoLastPass(self):

    failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': self.build_info,
                    'last_passed_build': None,
                },
            },
            'first_failed_build': self.build_info,
            'last_passed_build': None,
        },
    }

    expected_res = {'failures': {}, 'last_passed_build': None}

    self.assertEqual(
        expected_res,
        pre_compile_analysis.GetFirstFailuresInCurrentBuild(
            self.context, self.build, failures))

  def testGetFirstFailuresInCurrentBuildOnlyStep(self):
    build_122_info = {
        'id': 8000000000122,
        'number': self.build_number - 1,
        'commit_id': 'git_sha_122'
    }

    failures = {
        'compile': {
            'failures': {},
            'first_failed_build': self.build_info,
            'last_passed_build': build_122_info,
        },
    }

    expected_res = {
        'failures': {
            'compile': {
                'output_targets': [],
                'last_passed_build': build_122_info,
            },
        },
        'last_passed_build': build_122_info
    }

    self.assertEqual(
        expected_res,
        pre_compile_analysis.GetFirstFailuresInCurrentBuild(
            self.context, self.build, failures))

  def testGetFirstFailuresInCurrentBuildOnlyStepFailedBefore(self):
    build_122_info = {
        'id': 8000000000122,
        'number': self.build_number - 1,
        'commit_id': 'git_sha_122'
    }
    build_121_info = {
        'id': 8000000000121,
        'number': self.build_number - 2,
        'commit_id': 'git_sha_121'
    }

    failures = {
        'compile': {
            'failures': {},
            'first_failed_build': build_122_info,
            'last_passed_build': build_121_info,
        },
    }

    expected_res = {'failures': {}, 'last_passed_build': None}

    self.assertEqual(
        expected_res,
        pre_compile_analysis.GetFirstFailuresInCurrentBuild(
            self.context, self.build, failures))

  def testGetFirstFailuresInCurrentBuildFailureStartedInDifferentBuild(self):
    build_122_info = {
        'id': 8000000000122,
        'number': self.build_number - 1,
        'commit_id': 'git_sha_122'
    }
    build_121_info = {
        'id': 8000000000121,
        'number': self.build_number - 2,
        'commit_id': 'git_sha_121'
    }

    failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': self.build_info,
                    'last_passed_build': build_122_info,
                },
                frozenset(['target3']): {
                    'rule': 'ACTION',
                    'first_failed_build': build_122_info,
                    'last_passed_build': None,
                },
                frozenset(['target4']): {
                    'rule': 'ACTION',
                    'first_failed_build': self.build_info,
                    'last_passed_build': build_121_info,
                },
            },
            'first_failed_build': build_122_info,
            'last_passed_build': None,
        },
    }

    expected_res = {
        'failures': {
            'compile': {
                'output_targets': [{'target4'}, {'target1', 'target2'}],
                'last_passed_build': build_121_info,
            },
        },
        'last_passed_build': build_121_info
    }

    self.assertEqual(
        expected_res,
        pre_compile_analysis.GetFirstFailuresInCurrentBuild(
            self.context, self.build, failures))

  @mock.patch.object(
      ChromiumProjectAPI,
      'GetRerunBuilderId',
      return_value='chromium/findit/findit_variables')
  @mock.patch.object(
      git, 'GetCommitPositionFromRevision', side_effect=[66680, 66666, 66680])
  def testSaveCompileAnalysis(self, *_):
    build_121_info = {
        'id': 8000000000121,
        'number': self.build_number - 2,
        'commit_id': 'git_sha_121'
    }

    detailed_compile_failures = {
        'compile': {
            'failures': {
                frozenset(['target1', 'target2']): {
                    'rule': 'CXX',
                    'first_failed_build': {
                        'id': 8000000000121,
                        'number': 121,
                        'commit_id': 'git_sha'
                    },
                    'last_passed_build': {
                        'id': 8000000000120,
                        'number': 120,
                        'commit_id': 'git_sha'
                    },
                },
                frozenset(['target3']): {
                    'rule': 'ACTION',
                    'first_failed_build': build_121_info,
                    'last_passed_build': None,
                },
            },
            'first_failed_build': {
                'id': 8000000000121,
                'number': 121,
                'commit_id': 'git_sha'
            },
            'last_passed_build': {
                'id': 8000000000120,
                'number': 120,
                'commit_id': 'git_sha'
            },
        },
    }

    pre_compile_analysis.SaveCompileFailures(self.context, self.build,
                                             detailed_compile_failures)

    first_failures_in_current_build = {
        'failures': {
            'compile': {
                'output_targets': [{'target1', 'target2'}],
                'last_passed_build': build_121_info,
            },
        },
        'last_passed_build': build_121_info
    }
    pre_compile_analysis.SaveCompileAnalysis(self.context, self.build,
                                             first_failures_in_current_build)

    analysis = CompileFailureAnalysis.GetVersion(self.build_id)
    self.assertIsNotNone(analysis)
    self.assertEqual('git_sha_121', analysis.last_passed_commit.gitiles_id)
    self.assertEqual(66666, analysis.last_passed_commit.commit_position)
    self.assertEqual('chromium/findit/findit_variables',
                     analysis.rerun_builder_id)
    self.assertEqual(1, len(analysis.compile_failure_keys))
    self.assertItemsEqual(['target1', 'target2'],
                          analysis.compile_failure_keys[0].get().output_targets)
