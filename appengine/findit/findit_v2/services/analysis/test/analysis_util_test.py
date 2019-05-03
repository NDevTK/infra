# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from datetime import datetime

from findit_v2.model.compile_failure import CompileFailure
from findit_v2.model.gitiles_commit import GitilesCommit
from findit_v2.model.luci_build import LuciFailedBuild
from findit_v2.services.analysis import analysis_util
from findit_v2.services.failure_type import StepTypeEnum
from waterfall.test import wf_testcase


class AnalysisUtilTest(wf_testcase.TestCase):

  def setUp(self):
    super(AnalysisUtilTest, self).setUp()
    self.gitiles_host = 'gitiles.host.com'
    self.gitiles_project = 'project/name'
    self.gitiles_ref = 'ref/heads/master'

    self.build = LuciFailedBuild.Create(
        luci_project='chromium',
        luci_bucket='ci',
        luci_builder='Linux Builder',
        build_id=9876543210,
        legacy_build_number=12345,
        gitiles_host='chromium.googlesource.com',
        gitiles_project='chromium/src',
        gitiles_ref='refs/heads/master',
        gitiles_id='git_hash',
        commit_position=65450,
        status=20,
        create_time=datetime(2019, 3, 28),
        start_time=datetime(2019, 3, 28, 0, 1),
        end_time=datetime(2019, 3, 28, 1),
        build_failure_type=StepTypeEnum.COMPILE)
    self.build.put()

    self.compile_failure_1 = CompileFailure.Create(self.build.key, 'compile',
                                                   ['a.o'], 'CC')
    self.compile_failure_1.put()
    self.compile_failure_2 = CompileFailure.Create(self.build.key, 'compile',
                                                   ['b.o'], 'CC')
    self.compile_failure_2.put()

    self.commits = []
    for i in xrange(0, 11):
      self.commits.append(self._CreateGitilesCommit('r%d' % i, 100 + i))

  def _CreateGitilesCommit(self, gitiles_id, commit_position):
    return GitilesCommit(
        gitiles_host=self.gitiles_host,
        gitiles_project=self.gitiles_project,
        gitiles_ref=self.gitiles_ref,
        gitiles_id=gitiles_id,
        commit_position=commit_position)

  def testUpdateFailureRegressionRanges(self):
    rerun_builds_info = [(self.commits[5], {}),
                         (self.commits[7], {
                             'compile': ['a.o']
                         }), (self.commits[6], {
                             'compile': ['a.o']
                         }), (self.commits[8], {
                             'compile': ['b.o']
                         })]
    failures_with_range = [{
        'failure': self.compile_failure_1,
        'last_passed_commit': self.commits[0],
        'first_failed_commit': self.commits[10],
    },
                           {
                               'failure': self.compile_failure_2,
                               'last_passed_commit': self.commits[0],
                               'first_failed_commit': self.commits[10],
                           }]

    expected_results = [{
        'failure': self.compile_failure_1,
        'last_passed_commit': self.commits[5],
        'first_failed_commit': self.commits[6],
    },
                        {
                            'failure': self.compile_failure_2,
                            'last_passed_commit': self.commits[7],
                            'first_failed_commit': self.commits[8],
                        }]

    analysis_util.UpdateFailureRegressionRanges(rerun_builds_info,
                                                failures_with_range)

    for real_failure in failures_with_range:
      for expected_result in expected_results:
        if real_failure['failure'].output_targets == expected_result[
            'failure'].output_targets:
          self.assertEqual(expected_result['last_passed_commit'].gitiles_id,
                           real_failure['last_passed_commit'].gitiles_id)
          self.assertEqual(expected_result['first_failed_commit'].gitiles_id,
                           real_failure['first_failed_commit'].gitiles_id)

  def testGroupFailuresByRegerssionRange(self):
    compile_failure_3 = CompileFailure.Create(self.build.key, 'compile',
                                              ['c.o'], 'CC')
    compile_failure_3.put()

    failures_with_range = [{
        'failure': self.compile_failure_1,
        'last_passed_commit': self.commits[5],
        'first_failed_commit': self.commits[6],
    },
                           {
                               'failure': self.compile_failure_2,
                               'last_passed_commit': self.commits[7],
                               'first_failed_commit': self.commits[8],
                           },
                           {
                               'failure': compile_failure_3,
                               'last_passed_commit': self.commits[5],
                               'first_failed_commit': self.commits[6],
                           }]

    expected_result = [
        {
            'failures': [self.compile_failure_1, compile_failure_3],
            'last_passed_commit': self.commits[5],
            'first_failed_commit': self.commits[6],
        },
        {
            'failures': [self.compile_failure_2],
            'last_passed_commit': self.commits[7],
            'first_failed_commit': self.commits[8],
        },
    ]

    result = analysis_util.GroupFailuresByRegerssionRange(failures_with_range)
    self.assertItemsEqual(expected_result, result)
