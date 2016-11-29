# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from collections import defaultdict
import copy

from common.dependency import Dependency
from common.dependency import DependencyRoll
from common import chrome_dependency_fetcher
from crash import changelist_classifier
from crash.crash_report import CrashReport
from crash.results import AnalysisInfo
from crash.results import MatchResult
from crash.stacktrace import CallStack
from crash.stacktrace import StackFrame
from crash.stacktrace import Stacktrace
from crash.test.crash_test_suite import CrashTestSuite
from lib.gitiles.blame import Blame
from lib.gitiles.blame import Region
from lib.gitiles.change_log import ChangeLog
from lib.gitiles.gitiles_repository import GitilesRepository

DUMMY_CHANGELOG1 = ChangeLog.FromDict({
    'author_name': 'r@chromium.org',
    'message': 'dummy',
    'committer_email': 'r@chromium.org',
    'commit_position': 175900,
    'author_email': 'r@chromium.org',
    'touched_files': [
        {
            'change_type': 'add',
            'new_path': 'a.cc',
            'old_path': None,
        },
    ],
    'author_time': 'Thu Mar 31 21:24:43 2016',
    'committer_time': 'Thu Mar 31 21:28:39 2016',
    'commit_url':
        'https://repo.test/+/1',
    'code_review_url': 'https://codereview.chromium.org/3281',
    'committer_name': 'example@chromium.org',
    'revision': '1',
    'reverted_revision': None
})

DUMMY_CHANGELOG2 = ChangeLog.FromDict({
    'author_name': 'example@chromium.org',
    'message': 'dummy',
    'committer_email': 'example@chromium.org',
    'commit_position': 175976,
    'author_email': 'example@chromium.org',
    'touched_files': [
        {
            'change_type': 'add',
            'new_path': 'f0.cc',
            'old_path': 'b/f0.cc'
        },
    ],
    'author_time': 'Thu Mar 31 21:24:43 2016',
    'committer_time': 'Thu Mar 31 21:28:39 2016',
    'commit_url':
        'https://repo.test/+/2',
    'code_review_url': 'https://codereview.chromium.org/3281',
    'committer_name': 'example@chromium.org',
    'revision': '2',
    'reverted_revision': '1'
})

DUMMY_CHANGELOG3 = ChangeLog.FromDict({
    'author_name': 'e@chromium.org',
    'message': 'dummy',
    'committer_email': 'e@chromium.org',
    'commit_position': 176000,
    'author_email': 'e@chromium.org',
    'touched_files': [
        {
            'change_type': 'modify',
            'new_path': 'f.cc',
            'old_path': 'f.cc'
        },
        {
            'change_type': 'delete',
            'new_path': None,
            'old_path': 'f1.cc'
        },
    ],
    'author_time': 'Thu Apr 1 21:24:43 2016',
    'committer_time': 'Thu Apr 1 21:28:39 2016',
    'commit_url':
        'https://repo.test/+/3',
    'code_review_url': 'https://codereview.chromium.org/3281',
    'committer_name': 'example@chromium.org',
    'revision': '3',
    'reverted_revision': None
})

# TODO(wrengr): re crrev.com/2414523002: we need to have a specified
# revision_range (even if the versions therein are None), because
# ChangelistClassifier.__call__ will take it apart in order to call
# GetDEPSRollsDict; if it can't then it will immediately return the
# empty list of results, breaking many of the tests here. Of course,
# taking revision_range apart isn't actually required for the tests,
# since we mock GetDEPSRollsDict. So, really what we ought to do in the
# long run is redesign things so that GetDEPSRollsDict takes the
# CrashReport directly and pulls out the revision_range and platform
# itself; that way ChangelistClassifier.__call__ needn't worry about it,
# allowing us to clean up the tests here.
DUMMY_REPORT = CrashReport(None, None, None, Stacktrace(), (None, None))

class ChangelistClassifierTest(CrashTestSuite):

  def setUp(self):
    super(ChangelistClassifierTest, self).setUp()
    self.changelist_classifier = changelist_classifier.ChangelistClassifier(
        GitilesRepository(self.GetMockHttpClient()), 7)

  def testSkipAddedAndDeletedRegressionRolls(self):
    self.mock(chrome_dependency_fetcher.ChromeDependencyFetcher,
              'GetDependency', lambda *_: {})
    dep_rolls = {
        'src/dep': DependencyRoll('src/dep1', 'https://url_dep1', None, '9'),
        'src/': DependencyRoll('src/', ('https://chromium.googlesource.com/'
                                        'chromium/src.git'), '4', '5')
    }
    self.mock(chrome_dependency_fetcher.ChromeDependencyFetcher,
              'GetDependencyRollsDict', lambda *_: dep_rolls)

    passed_in_regression_deps_rolls = []
    def _MockGetChangeLogsForFilesGroupedByDeps(regression_deps_rolls, *_):
      passed_in_regression_deps_rolls.append(regression_deps_rolls)
      return {}, None

    self.mock(changelist_classifier, 'GetChangeLogsForFilesGroupedByDeps',
              _MockGetChangeLogsForFilesGroupedByDeps)
    self.mock(changelist_classifier, 'GetStackInfosForFilesGroupedByDeps',
              lambda *_: {})
    self.mock(changelist_classifier, 'FindMatchResults', lambda *_: None)

    self.changelist_classifier(CrashReport(crashed_version = '5',
                               signature = 'sig',
                               platform = 'canary',
                               stacktrace = Stacktrace([CallStack(0)]),
                               regression_range = ['4', '5']))
    expected_regression_deps_rolls = copy.deepcopy(dep_rolls)

    # Regression of a dep added/deleted (old_revision/new_revision is None) can
    # not be known for sure and this case rarely happens, so just filter them
    # out.
    del expected_regression_deps_rolls['src/dep']
    self.assertEqual(passed_in_regression_deps_rolls[0],
                     expected_regression_deps_rolls)

  def testGetDepsInCrashStack(self):
    crash_stack = CallStack(0)
    crash_stack.extend([
        StackFrame(0, 'src/', 'func0', 'f0.cc', 'src/f0.cc', [1]),
        StackFrame(1, 'src/', 'func1', 'f1.cc', 'src/f1.cc', [2, 3]),
        StackFrame(1, '', 'func2', 'f2.cc', 'src/f2.cc', [2, 3]),
    ])
    crash_deps = {'src/': Dependency('src/', 'https://chromium_repo', '1'),
                  'src/v8/': Dependency('src/v8/', 'https://v8_repo', '2')}

    expected_stack_deps = {'src/': crash_deps['src/']}

    self.assertDictEqual(
        changelist_classifier.GetDepsInCrashStack(crash_stack, crash_deps),
        expected_stack_deps)

  def testGetChangeLogsForFilesGroupedByDeps(self):
    regression_deps_rolls = {
        'src/dep': DependencyRoll('src/dep1', 'https://url_dep1', '7', '9'),
        'src/': DependencyRoll('src/', ('https://chromium.googlesource.com/'
                                        'chromium/src.git'), '4', '5')
    }

    stack_deps = {
        'src/': Dependency('src/', 'https://url_src', 'rev1', 'DEPS'),
        'src/new': Dependency('src/new', 'https://new', 'rev2', 'DEPS'),
        'src/dep': Dependency('src/dep', 'https://url_dep', 'rev', 'DEPS'),
    }

    def _MockGetChangeLogs(_, start_rev, end_rev):
      if start_rev == '4' and end_rev == '5':
        return [DUMMY_CHANGELOG1, DUMMY_CHANGELOG2, DUMMY_CHANGELOG3]

      return []

    self.mock(GitilesRepository, 'GetChangeLogs', _MockGetChangeLogs)

    dep_file_to_changelogs, ignore_cls = (
        changelist_classifier.GetChangeLogsForFilesGroupedByDeps(
            regression_deps_rolls, stack_deps,
            GitilesRepository(self.GetMockHttpClient())))
    dep_file_to_changelogs_json = defaultdict(lambda: defaultdict(list))
    for dep, file_to_changelogs in dep_file_to_changelogs.iteritems():
      for file_path, changelogs in file_to_changelogs.iteritems():
        for changelog in changelogs:
          dep_file_to_changelogs_json[dep][file_path].append(changelog.ToDict())

    expected_dep_file_to_changelogs_json = {
        'src/': {
            'a.cc': [DUMMY_CHANGELOG1.ToDict()],
            'f.cc': [DUMMY_CHANGELOG3.ToDict()]
        }
    }
    self.assertDictEqual(dep_file_to_changelogs_json,
                     expected_dep_file_to_changelogs_json)
    self.assertSetEqual(ignore_cls, set(['1']))

  def testGetStackInfosForFilesGroupedByDeps(self):
    main_stack = CallStack(0)
    main_stack.extend(
        [StackFrame(0, 'src/', 'c(p* &d)', 'a.cc', 'src/a.cc', [177]),
         StackFrame(1, 'src/', 'd(a* c)', 'a.cc', 'src/a.cc', [227, 228, 229]),
         StackFrame(2, 'src/v8/', 'e(int)', 'b.cc', 'src/v8/b.cc', [89, 90])])

    low_priority_stack = CallStack(1)
    low_priority_stack.append(
        StackFrame(0, 'src/dummy/', 'c(p* &d)', 'd.cc', 'src/dummy/d.cc', [17]))

    stacktrace = Stacktrace()
    stacktrace.extend([main_stack, low_priority_stack])

    crashed_deps = {'src/': Dependency('src/', 'https//repo', '2'),
                    'src/v8/': Dependency('src/v8', 'https//repo', '1')}

    expected_dep_file_to_stack_infos = {
        'src/': {
            'a.cc': [
                (main_stack[0], 0),
                (main_stack[1], 0),
            ],
        },
        'src/v8/': {
            'b.cc': [
                (main_stack[2], 0),
            ]
        }
    }

    dep_file_to_stack_infos = (
        changelist_classifier.GetStackInfosForFilesGroupedByDeps(
            stacktrace, crashed_deps))

    self.assertEqual(len(dep_file_to_stack_infos),
                     len(expected_dep_file_to_stack_infos))

    for dep, file_to_stack_infos in dep_file_to_stack_infos.iteritems():
      self.assertTrue(dep in expected_dep_file_to_stack_infos)
      expected_file_to_stack_infos = expected_dep_file_to_stack_infos[dep]

      for file_path, stack_infos in file_to_stack_infos.iteritems():
        self.assertTrue(file_path in expected_file_to_stack_infos)
        expected_stack_infos = expected_file_to_stack_infos[file_path]

        self._VerifyTwoStackInfosEqual(stack_infos, expected_stack_infos)

  def testFindMatchResults(self):
    dep_file_to_changelogs = {
        'src/': {
            'a.cc': [
                DUMMY_CHANGELOG1,
            ]
        }
    }

    dep_file_to_stack_infos = {
        'src/': {
            'a.cc': [
                (StackFrame(0, 'src/', 'func', 'a.cc', 'src/a.cc', [1]), 0),
                (StackFrame(1, 'src/', 'func', 'a.cc', 'src/a.cc', [7]), 0),
            ],
            'b.cc': [
                (StackFrame(2, 'src/', 'func', 'b.cc', 'src/b.cc', [36]), 0),
            ]
        }
    }

    dummy_blame = Blame('9', 'a.cc')
    dummy_blame.AddRegion(
        Region(1, 5, '6', 'a', 'a@chromium.org', 'Thu Mar 31 21:24:43 2016'))
    dummy_blame.AddRegion(
        Region(6, 10, '1', 'b', 'b@chromium.org', 'Thu Jun 19 12:11:40 2015'))

    self.mock(GitilesRepository, 'GetBlame', lambda *_: dummy_blame)

    stack_deps = {
        'src/': Dependency('src/', 'https://url_src', 'rev1', 'DEPS'),
    }

    expected_match_results = [{
        'url': 'https://repo.test/+/1',
        'review_url': 'https://codereview.chromium.org/3281',
        'revision': '1',
        'project_path': 'src/',
        'author': 'r@chromium.org',
        'time': 'Thu Mar 31 21:24:43 2016',
        'reasons': None,
        'confidence': None,
        'changed_files': None
    }]

    match_results = changelist_classifier.FindMatchResults(
        dep_file_to_changelogs, dep_file_to_stack_infos, stack_deps,
        GitilesRepository(self.GetMockHttpClient()))
    self.assertListEqual([result.ToDict() for result in match_results],
                         expected_match_results)

  # TODO(http://crbug.com/659346): why do these mocks give coverage
  # failures? That's almost surely hiding a bug in the tests themselves.
  def testFindItForCrashNoRegressionRange(self): # pragma: no cover
    self.mock(chrome_dependency_fetcher.ChromeDependencyFetcher,
        'GetDependencyRollsDict', lambda *_: {})
    self.mock(chrome_dependency_fetcher.ChromeDependencyFetcher,
        'GetDependency', lambda *_: {})
    # N.B., for this one test we really do want regression_range=None.
    report = DUMMY_REPORT._replace(regression_range=None)
    self.assertListEqual(self.changelist_classifier(report), [])

  def testFindItForCrashNoMatchFound(self):
    self.mock(changelist_classifier, 'FindMatchResults', lambda *_: [])
    self.mock(chrome_dependency_fetcher.ChromeDependencyFetcher,
        'GetDependencyRollsDict',
        lambda *_: {'src/': DependencyRoll('src/', 'https://repo', '1', '2')})
    self.mock(chrome_dependency_fetcher.ChromeDependencyFetcher,
        'GetDependency', lambda *_: {})
    self.assertListEqual(self.changelist_classifier(DUMMY_REPORT), [])

  def testFindItForCrash(self):

    def _MockFindMatchResults(*_):
      match_result1 = MatchResult(DUMMY_CHANGELOG1, 'src/', '')
      frame1 = StackFrame(0, 'src/', 'func', 'a.cc', 'src/a.cc', [1])
      frame2 = StackFrame(1, 'src/', 'func', 'a.cc', 'src/a.cc', [7])
      match_result1.file_to_stack_infos = {
          'a.cc': [(frame1, 0), (frame2, 0)]
      }
      match_result1.file_to_analysis_info = {
          'a.cc': AnalysisInfo(min_distance=0, min_distance_frame=frame1)
      }

      match_result2 = MatchResult(DUMMY_CHANGELOG3, 'src/', '')
      frame3 = StackFrame(5, 'src/', 'func', 'f.cc', 'src/f.cc', [1])
      match_result2.file_to_stack_infos = {
          'f.cc': [(frame3, 0)]
      }
      match_result2.file_to_analysis_info = {
          'a.cc': AnalysisInfo(min_distance=20, min_distance_frame=frame3)
      }

      return [match_result1, match_result2]

    self.mock(changelist_classifier, 'FindMatchResults', _MockFindMatchResults)
    self.mock(chrome_dependency_fetcher.ChromeDependencyFetcher,
        'GetDependencyRollsDict',
        lambda *_: {'src/': DependencyRoll('src/', 'https://repo', '1', '2')})
    self.mock(chrome_dependency_fetcher.ChromeDependencyFetcher,
        'GetDependency', lambda *_: {})
    results = self.changelist_classifier(DUMMY_REPORT)
    expected_match_results = [
        {
            'reasons': [('TopFrameIndex', 1.0, 'Top frame is #0'),
                        ('MinDistance', 1, 'Minimum distance is 0')],
            'changed_files': [{'info': 'Minimum distance (LOC) 0, frame #0',
                               'blame_url': None, 'file': 'a.cc'}],
            'time': 'Thu Mar 31 21:24:43 2016',
            'author': 'r@chromium.org',
            'url': 'https://repo.test/+/1',
            'project_path': 'src/',
            'review_url': 'https://codereview.chromium.org/3281',
            'confidence': 1.0, 'revision': '1'
        },
    ]
    self.assertListEqual([result.ToDict() for result in results],
                         expected_match_results)

  def testFinditForCrashFilterZeroConfidentResults(self):
    def _MockFindMatchResults(*_):
      match_result1 = MatchResult(DUMMY_CHANGELOG1, 'src/', '')
      frame1 = StackFrame(0, 'src/', 'func', 'a.cc', 'src/a.cc', [1])
      frame2 = StackFrame(1, 'src/', 'func', 'a.cc', 'src/a.cc', [7])
      match_result1.file_to_stack_infos = {
          'a.cc': [(frame1, 0), (frame2, 0)]
      }
      match_result1.file_to_analysis_info = {
          'a.cc': AnalysisInfo(min_distance=1, min_distance_frame=frame1)
      }

      match_result2 = MatchResult(DUMMY_CHANGELOG3, 'src/', '')
      frame3 = StackFrame(15, 'src/', 'func', 'f.cc', 'src/f.cc', [1])
      match_result2.file_to_stack_infos = {
          'f.cc': [(frame3, 0)]
      }
      match_result2.file_to_analysis_info = {
          'f.cc': AnalysisInfo(min_distance=20, min_distance_frame=frame3)
      }

      match_result3 = MatchResult(DUMMY_CHANGELOG3, 'src/', '')
      frame4 = StackFrame(3, 'src/', 'func', 'ff.cc', 'src/ff.cc', [1])
      match_result3.file_to_stack_infos = {
          'f.cc': [(frame4, 0)]
      }
      match_result3.file_to_analysis_info = {
          'f.cc': AnalysisInfo(min_distance=60, min_distance_frame=frame4)
      }

      return [match_result1, match_result2, match_result3]

    self.mock(changelist_classifier, 'FindMatchResults', _MockFindMatchResults)
    self.mock(chrome_dependency_fetcher.ChromeDependencyFetcher,
        'GetDependencyRollsDict',
        lambda *_: {'src/': DependencyRoll('src/', 'https://repo', '1', '2')})
    self.mock(chrome_dependency_fetcher.ChromeDependencyFetcher,
        'GetDependency', lambda *_: {})

    results = self.changelist_classifier(DUMMY_REPORT)
    expected_match_results = [
        {
            'author': 'r@chromium.org',
            'changed_files': [
                {
                    'blame_url': None,
                    'file': 'a.cc',
                    'info': 'Minimum distance (LOC) 1, frame #0'
                }
            ],
            'confidence': 0.8,
            'project_path': 'src/',
            'reasons': [
                ('TopFrameIndex', 1.0, 'Top frame is #0'),
                ('MinDistance', 0.8, 'Minimum distance is 1')
            ],
            'review_url': 'https://codereview.chromium.org/3281',
            'revision': '1',
            'time': 'Thu Mar 31 21:24:43 2016',
            'url': 'https://repo.test/+/1'
        }
    ]
    self.assertListEqual([result.ToDict() for result in results],
                         expected_match_results)

  def testFinditForCrashAllMatchResultsWithZeroConfidences(self):
    def _MockFindMatchResults(*_):
      match_result1 = MatchResult(DUMMY_CHANGELOG1, 'src/', '')
      frame1 = StackFrame(20, 'src/', '', 'func', 'a.cc', [1])
      frame2 = StackFrame(21, 'src/', '', 'func', 'a.cc', [7])
      match_result1.file_to_stack_infos = {
          'a.cc': [(frame1, 0), (frame2, 0)]
      }
      match_result1.file_to_analysis_info = {
          'a.cc': AnalysisInfo(min_distance=1, min_distance_frame=frame1)
      }

      match_result2 = MatchResult(DUMMY_CHANGELOG3, 'src/', '')
      frame3 = StackFrame(15, 'src/', '', 'func', 'f.cc', [1])
      match_result2.file_to_stack_infos = {
          'f.cc': [(frame3, 0)]
      }
      match_result2.min_distance = 20
      match_result2.file_to_analysis_info = {
          'f.cc': AnalysisInfo(min_distance=20, min_distance_frame=frame3)
      }

      return [match_result1, match_result2]

    self.mock(changelist_classifier, 'FindMatchResults', _MockFindMatchResults)
    self.mock(chrome_dependency_fetcher.ChromeDependencyFetcher,
        'GetDependencyRollsDict',
        lambda *_: {'src/': DependencyRoll('src/', 'https://repo', '1', '2')})
    self.mock(chrome_dependency_fetcher.ChromeDependencyFetcher,
        'GetDependency', lambda *_: {})

    self.assertListEqual(self.changelist_classifier(DUMMY_REPORT), [])
