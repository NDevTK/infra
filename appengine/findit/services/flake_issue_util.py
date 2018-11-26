# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Utilities for assisting in bug management for flaky tests."""

from collections import defaultdict
import datetime
import logging

from google.appengine.ext import ndb

from googleapiclient.errors import HttpError
from libs import time_util
from model.flake.detection.flake_occurrence import FlakeOccurrence
from model.flake.flake import Flake
from model.flake.flake_issue import FlakeIssue
from model.flake.flake_type import FlakeType
from services import issue_constants
from services import monorail_util
from services.apis import AnalyzeDetectedFlakeOccurrence
from services.flake_failure import flake_constants
from services.issue_generator import FlakeDetectionIssueGenerator
from waterfall import waterfall_config


class FlakeGroup(object):

  def __init__(self, occurrences, flake_luci_project):
    self.flakes = []
    self.num_occurrences = len(occurrences)
    self.luci_project = flake_luci_project

  def AddFlake(self, flake, occurrences):
    """Tries to add a flake to group if it can belong.

    Returns:
      (bool), if the flake is added to the group.
    """
    raise NotImplementedError()


# Group of flakes that need the same new bug. Based on the heuristic rules,
# the flakes in the same group should have the same normalized_step_name and
# test_suite_name, and all flakes happen at the exactly same builds.
class FlakeGroupByOccurrences(FlakeGroup):

  def __init__(self, flake, occurrences):
    super(FlakeGroupByOccurrences, self).__init__(occurrences,
                                                  flake.luci_project)
    self.normalized_step_name = flake.normalized_step_name
    self.test_suite_name = flake.test_suite_name
    self.flakes.append(flake)
    self.builds = self._GetBuildsList(occurrences)

  def _GetBuildsList(self, occurrences):
    occurrence_builds = []
    for occurrence in occurrences:
      occurrence_builds.append('{}@{}@{}'.format(
          occurrence.build_configuration.luci_project,
          occurrence.build_configuration.luci_bucket, occurrence.build_id))
    return occurrence_builds

  def AddFlake(self, flake, occurrences):
    """Adds the flake if it happens at the same builds as other flakes in group.

    Returns:
      (bool), if the flake is added to the group.
    """
    if len(occurrences) != len(self.builds):
      return False

    occurrence_builds = self._GetBuildsList(occurrences)

    if sorted(occurrence_builds) == sorted(self.builds):
      self.flakes.append(flake)
      return True

    return False

  def ToDict(self):
    # For unittest purpose.
    return {
        'normalized_step_name': self.normalized_step_name,
        'test_suite_name': self.test_suite_name,
        'flakes': self.flakes,
        'builds': self.flakes,
        'num_occurrences': self.num_occurrences
    }


# Group of flakes that are already linked to the same issue directly or
# indirectly. The flakes will only be grouped by the issue in this case, and
# they may not meet the heuristic rules as FlakeGroupByOccurrences.
class FlakeGroupByFlakeIssue(FlakeGroup):

  def __init__(self, flake_issue, flake, occurrences, previous_flake_issue):
    super(FlakeGroupByFlakeIssue, self).__init__(occurrences,
                                                 flake.luci_project)
    self.flake_issue = flake_issue
    self.flakes_with_same_occurrences = True
    self.flakes.append(flake)

    if flake_issue.issue_id != previous_flake_issue.issue_id:
      self.previous_issue_id = previous_flake_issue.issue_id

  def AddFlake(self, flake, occurrences):
    """Adds the flake, also updates flakes_with_same_occurrences and
      num_occurrences if needed."""
    self.flakes.append(flake)
    if len(occurrences) < self.num_occurrences:
      # Only maintains a minimum num_occurrences to show in bug comments.
      self.num_occurrences = len(occurrences)
      self.flakes_with_same_occurrences = False
    return True

  def ToDict(self):
    # For unittest purpose.
    return {
        'flake_issue': self.flake_issue,
        'flakes': self.flakes,
        'flakes_with_same_occurrences': self.flakes_with_same_occurrences,
        'num_occurrences': self.num_occurrences
    }


def _GetOpenIssueIdForFlakyTestByCustomizedField(test_name,
                                                 monorail_project='chromium'):
  """Returns flaky tests related issue by searching customized field.

  Args:
    test_name: The name of the test to search for.
    monorail_project: The Monorail project to search for.

  Returns:
    Id of the issue if it exists, otherwise None.
  """
  query = issue_constants.FLAKY_TEST_CUSTOMIZED_FIELD_QUERY_TEMPLATE.format(
      test_name)
  open_issues = monorail_util.GetOpenIssues(query, monorail_project)
  return open_issues[0].id if open_issues else None


def _GetOpenIssueIdForFlakyTestBySummary(test_name,
                                         monorail_project='chromium'):
  """Returns flaky tests related issue by searching summary.

  Note that searching for |test_name| in the summary alone is not enough, for
  example: 'suite.test needs to be rewritten', so at least one of the following
  additional identifiers is also required:
  1. The issue has label: Test-Flaky.
  2. The issue has component: Tests>Flaky.
  3. The issue has one of the issue_constants.FLAKY_TEST_SUMMARY_KEYWORDS in
     the summary.

  Args:
    test_name: The name of the test to search for.
    monorail_project: The Monorail project to search for.

  Returns:
    Minimum id among the matched issues if exists, otherwise None.
  """

  def _is_issue_related_to_flake(issue):
    if issue_constants.FLAKY_TEST_LABEL in issue.labels:
      return True

    if issue_constants.FLAKY_TEST_COMPONENT in issue.components:
      return True

    return any(keyword in issue.summary.lower()
               for keyword in issue_constants.FLAKY_TEST_SUMMARY_KEYWORDS)

  def _is_not_test_findit_wrong_issue(issue):
    return issue_constants.TEST_FINDIT_WRONG_LABEL not in issue.labels

  query = issue_constants.FLAKY_TEST_SUMMARY_QUERY_TEMPLATE.format(test_name)
  open_issues = monorail_util.GetOpenIssues(query, monorail_project)
  flaky_test_open_issues = [
      issue for issue in open_issues if (_is_issue_related_to_flake(issue) and
                                         _is_not_test_findit_wrong_issue(issue))
  ]
  if not flaky_test_open_issues:
    return None

  return min([issue.id for issue in flaky_test_open_issues])


def OpenIssueAlreadyExistsForFlakyTest(test_name, monorail_project='chromium'):
  """Returns True if a related flaky test bug already exists on Monorail.

  Args:
    test_name: The test name to search for.
    monorail_project: The Monorail project to search for.

  Returns:
    True is there is already a bug about this test being flaky, False otherwise.
  """
  return SearchOpenIssueIdForFlakyTest(test_name, monorail_project) is not None


def SearchOpenIssueIdForFlakyTest(test_name, monorail_project='chromium'):
  """Searches for existing open issue for a flaky test on Monorail.

  Args:
    test_name: The test name to search for.
    monorail_project: The Monorail project to search for.

  Returns:
    Id of the issue if it exists, otherwise None.
  """
  # Prefer issues without customized field because it means that the bugs were
  # created manually by develoepers, so it is more likely to gain attentions.
  return (_GetOpenIssueIdForFlakyTestBySummary(test_name, monorail_project) or
          _GetOpenIssueIdForFlakyTestByCustomizedField(test_name,
                                                       monorail_project))


def GetFlakeIssue(flake):
  """Returns the associated flake issue to a flake.

  Tries to use the flake's issue's merge_destination, and falls back to the
    issue itself if it doesn't merge into any other ones.

  Args:
    flake: A Flake Model entity.

  Returns:
    A FlakeIssue entity if it exists, otherwise None. This entity should be the
      final merging destination issue.
  """
  if not flake or not flake.flake_issue_key:
    return None

  flake_issue = flake.flake_issue_key.get()
  if not flake_issue:
    # Data is inconsistent, reset the key to allow a new FlakeIssue to be
    # attached later.
    flake.flake_issue_key = None
    flake.put()
    return None

  return flake_issue.GetMostUpdatedIssue()


def _FlakeHasEnoughOccurrences(unreported_occurrences):
  """Returns True if there are enough occurrences to worth reporting the flake.

  A flake has enough occurrences if the occurrences cover at least 3 different
  CLs.

  Args:
    unreported_occurrences: A list of occurrence that share the same parent
                            flake and haven't been reported yet.
                            The caller is responsible for making sure of it.
  """
  flake_detection_settings = waterfall_config.GetFlakeDetectionSettings()
  required_falsely_rejected_cls = flake_detection_settings.get(
      'min_required_impacted_cls_per_day',
      flake_constants.DEFAULT_MINIMUM_REQUIRED_IMPACTED_CLS_PER_DAY)
  cl_ids = [occurrence.gerrit_cl_id for occurrence in unreported_occurrences]
  unique_cl_ids = set(cl_ids)
  return len(unique_cl_ids) >= required_falsely_rejected_cls


def GetFlakesWithEnoughOccurrences():
  """Queries Datastore and returns flakes that has enough occurrences.

  The most intuitive algorithm is to fetch all flakes first, and then for each
  flake, fetch its recent and unreported flake occurrences, but it has
  performance implications when there are a lot of flakes (too many RPC calls)
  because a large number of calls are wasted on flakes that don't even have any
  recent and unreported flake occurrence.

  So, instead, this algorithm first fetches all recent and unreported flake
  occurrences, and then by looking up their parents to figure out the subset of
  flakes that need to be fetched.

  Returns:
    A list of tuples whose first element is a flake entity, second element is
    number of corresponding recent and unreported occurrences, third element is
    the flake issue the flake links to if exist.
  """
  utc_one_day_ago = time_util.GetUTCNow() - datetime.timedelta(days=1)
  occurrences = FlakeOccurrence.query(
      ndb.AND(FlakeOccurrence.flake_type == FlakeType.CQ_FALSE_REJECTION,
              FlakeOccurrence.time_happened > utc_one_day_ago)).fetch()

  logging.info(
      'There are %d cq false rejection occurrences within the past 24h.' %
      len(occurrences))

  flake_key_to_occurrences = defaultdict(list)
  for occurrence in occurrences:
    flake_key_to_occurrences[occurrence.key.parent()].append(occurrence)

  unique_flake_keys = flake_key_to_occurrences.keys()
  flakes = ndb.get_multi(unique_flake_keys)

  key_to_flake = dict(zip(unique_flake_keys, flakes))

  # Filter out occurrences that have already been reported according to the
  # last update time of the associated flake issue.
  flake_key_to_enough_unreported_occurrences = {}
  for flake_key, occurrences in flake_key_to_occurrences.iteritems():
    if not key_to_flake[flake_key]:
      logging.error('Flake not found for key %s', flake_key.urlsafe())
      continue

    if not _FlakeHasEnoughOccurrences(occurrences):
      continue

    flake_issue = GetFlakeIssue(flake_key.get())
    last_updated_time_by_flake_detection = (
        flake_issue.last_updated_time_by_flake_detection
        if flake_issue else None)
    if (last_updated_time_by_flake_detection and
        last_updated_time_by_flake_detection > utc_one_day_ago):
      # An issue can be updated at most once in any 24h window avoid noises.
      continue

    flake_key_to_enough_unreported_occurrences[flake_key] = {
        'occurrences': occurrences,
        'flake_issue': flake_issue
    }

  return [(key_to_flake[flake_key], info['occurrences'], info['flake_issue'])
          for flake_key, info in flake_key_to_enough_unreported_occurrences
          .iteritems()]


def _GetAndUpdateMergedIssue(flake_issue):
  """Gets the most up-to-date merged issue and update data in data store."""
  monorail_project = flake_issue.monorail_project
  merged_issue = monorail_util.GetMergedDestinationIssueForId(
      flake_issue.issue_id, monorail_project)
  if flake_issue.issue_id != merged_issue.id:
    logging.info(
        'Flake issue %s was merged to %s, updates this issue and'
        ' all issues were merged into it.',
        FlakeIssue.GetLinkForIssue(monorail_project, flake_issue.issue_id),
        FlakeIssue.GetLinkForIssue(monorail_project, merged_issue.id))
    _UpdateMergeDestination(flake_issue, merged_issue.id)

  return merged_issue


def GetFlakeGroupsForActionsOnBugs(flake_tuples_to_report):
  """Groups the flakes either by heuristic rules (if they linked to no issue)
    or by issue (if they have linked to an issue).

  Returns:
    (list, list): A list of FlakeGroupByOccurrences entities and a list of
      FlakeGroupByFlakeIssue entities.
  """
  # Groups of flakes that need the same new bug.
  # Keyed by normalized_step_name and test_suite_name.
  # Since we will group flakes only when they happen at the same builds,
  # it's possible to have different FlakeGroupByOccurrences with the same key,
  # so use a list to save them.
  flake_groups_by_heuristic = defaultdict(list)
  # Groups of flakes that are already linked to issues.
  flake_groups_by_issue = {}

  for flake_tuple in flake_tuples_to_report:
    flake = flake_tuple[0]
    occurrences = flake_tuple[1]
    flake_issue = flake_tuple[2]

    if flake_issue:
      flake_group = flake_groups_by_issue.get(flake_issue.issue_id)
      if not flake_group:
        merged_issue = _GetAndUpdateMergedIssue(flake_issue)
        if merged_issue.open:
          # Only creates new group of flakes by FlakeIssue if the issue is still
          # open.
          merged_flake_issue = flake_issue.GetMostUpdatedIssue()

          flake_group = FlakeGroupByFlakeIssue(merged_flake_issue, flake,
                                               occurrences, flake_issue)
          flake_groups_by_issue[merged_flake_issue.issue_id] = flake_group
          continue
      else:
        flake_group.AddFlake(flake, occurrences)
        continue

    # Group by heuristic for flakes not linked to any issue or the issue has
    # been closed.
    basic_group_key = '{}@{}@{}'.format(
        flake.luci_project, flake.normalized_step_name, flake.test_suite_name)

    grouped = False
    for flake_group in flake_groups_by_heuristic[basic_group_key]:
      grouped = flake_group.AddFlake(flake, occurrences)
      if grouped:
        break

    if not grouped:
      flake_group = FlakeGroupByOccurrences(flake, occurrences)
      flake_groups_by_heuristic[basic_group_key].append(flake_group)

  flake_groups_by_heuristic_list = []
  for heuristic_groups in flake_groups_by_heuristic.values():
    flake_groups_by_heuristic_list.extend(heuristic_groups)

  return flake_groups_by_heuristic_list, flake_groups_by_issue.values()


# TODO(crbug.com/903459): Move ReportFlakesToMonorail and
#  ReportFlakesToFlakeAnalyzer to auto action layer.
def ReportFlakesToMonorail(flake_tuples_to_report):
  """Reports newly detected flakes and occurrences to Monorail.

  ONLY create or update a bug if:
    rule 1. Has NOT reached the maximum configured bug update limit within 24h.
    rule 2. The bug wasn't created or updated within the past 24h.

  Args:
    flake_tuples_to_report: A list of tuples whose first element is a Flake
                            entity and second element is a list of corresponding
                            occurrences to report.
  """
  action_settings = waterfall_config.GetActionSettings()
  limit = action_settings.get('max_flake_bug_updates_per_day',
                              flake_constants.DEFAULT_MAX_BUG_UPDATES_PER_DAY)

  utc_one_day_ago = time_util.GetUTCNow() - datetime.timedelta(days=1)
  num_updated_issues_24h = (
      FlakeIssue.query(
          FlakeIssue.last_updated_time_by_flake_detection > utc_one_day_ago)
      .count())

  if num_updated_issues_24h >= limit:
    logging.info('Issues created or updated during the past 24 hours has '
                 'reached the limit.')
    return

  num_of_flakes_to_report = min(
      len(flake_tuples_to_report), limit - num_updated_issues_24h)
  flake_tuples_to_report = flake_tuples_to_report[:num_of_flakes_to_report]
  logging.info('There are %d flakes whose issues will be created or updated.' %
               num_of_flakes_to_report)

  for flake, occurrences in flake_tuples_to_report:
    issue_generator = FlakeDetectionIssueGenerator(flake, len(occurrences))
    try:
      CreateOrUpdateIssue(issue_generator, flake.luci_project)
      # Update FlakeIssue's last_updated_time_by_flake_detection property. This
      # property is only applicable to Flake Detection because Flake Detection
      # can update an issue at most once every 24 hours.
      flake_issue = GetFlakeIssue(flake)
      flake_issue.last_updated_time_by_flake_detection = time_util.GetUTCNow()
      flake_issue.put()
    except HttpError as error:
      # benign exceptions (HttpError 403) may happen when FindIt tries to
      # update an issue that it doesn't have permission to. Do not raise
      # exception so that the for loop can move on to create or update next
      # issues.
      logging.warning('Failed to create or update issue due to error: %s',
                      error)


def _IsReportFlakesToFlakeAnalyzerEnabled():
  """Returns True if the feature to report flakes to Flake Analyzer is enabled.

  Returns:
    Returns True if it is enabled, otherwise False.
  """
  # Unless the flag is explicitly set, assumes disabled by default.
  return waterfall_config.GetFlakeDetectionSettings().get(
      'report_flakes_to_flake_analyzer', False)


def ReportFlakesToFlakeAnalyzer(flake_tuples_to_report):
  """Reports newly detected flakes and occurrences to Flake Analyzer.

  Args:
    flake_tuples_to_report: A list of tuples whose first element is a Flake
                            entity and second element is a list of corresponding
                            occurrences to report.
  """
  if not _IsReportFlakesToFlakeAnalyzerEnabled():
    logging.info('Skip reporting flakes to Flake Analyzer because the feature '
                 'is disabled.')
    return

  for flake, occurrences in flake_tuples_to_report:
    flake_issue = GetFlakeIssue(flake)
    issue_id = flake_issue.issue_id if flake_issue else None
    for occurrence in occurrences:
      AnalyzeDetectedFlakeOccurrence(flake, occurrence, issue_id)


def _GetOrCreateFlakeIssue(issue_id, monorail_project):
  """Gets or creates a FlakeIssue entity for the monorail issue.

  Args:
    issue_id (int): Id of the issue.
    monorail_project (str): Monorail project of the issue.
  Returns:
    (FlakeIssue): a FlakeIssue entity of the issue.
  """
  flake_issue = FlakeIssue.Get(monorail_project, issue_id)
  if flake_issue:
    return flake_issue

  flake_issue = FlakeIssue.Create(monorail_project, issue_id)
  flake_issue.put()
  return flake_issue


def _UpdateMergeDestination(flake_issue, merged_issue_id):
  """Updates flake_issue and all other issues that are merged into it to
    store the new merging_destination.

  Args:
    flake_issue (FlakeIssue): A FlakeIssue to update.
    merged_issue_id (int): Id of the merged_issue.
  """
  merged_flake_issue = _GetOrCreateFlakeIssue(merged_issue_id,
                                              flake_issue.monorail_project)
  assert merged_flake_issue, (
      'Failed to get or create FlakeIssue for merged_issue %s' %
      FlakeIssue.GetLinkForIssue(flake_issue.monorail_project, merged_issue_id))

  merged_flake_issue_key = merged_flake_issue.key
  flake_issue.merge_destination_key = merged_flake_issue_key
  flake_issue.put()

  # Updates all issues that were merged into this flake_issue before.
  issue_leaves = FlakeIssue.query(
      FlakeIssue.merge_destination_key == flake_issue.key).fetch()
  if not issue_leaves:
    return

  for issue_leaf in issue_leaves:
    issue_leaf.merge_destination_key = merged_flake_issue_key
  ndb.put_multi(issue_leaves)


def CreateOrUpdateIssue(issue_generator, luci_project='chromium'):
  """Updates an issue if it exists, otherwise creates a new one.

  This method uses the best-effort to search the existing FlakeIssue entities
  and open issues on Monorail that are related to the flaky tests and reuse them
  if found, otherwise, creates a new issue and attach it to a Flake Model entity
  so that the newly created issue can be reused in the future.

  Args:
    issue_generator: A FlakyTestIssueGenerator object.
    luci_project: Name of the LUCI project that the flaky test is in, it is
                  used for searching existing Flake and FlakeIssue entities.

  Returns:
    Id of the issue that was eventually created or updated.
  """
  step_name = issue_generator.GetStepName()
  test_name = issue_generator.GetTestName()
  test_label_name = issue_generator.GetTestLabelName()

  target_flake = Flake.Get(luci_project, step_name, test_name)
  if not target_flake:
    target_flake = Flake.Create(luci_project, step_name, test_name,
                                test_label_name)
    target_flake.put()

  monorail_project = issue_generator.GetMonorailProject()
  flake_issue = GetFlakeIssue(target_flake)
  previous_tracking_bug_id = None

  if flake_issue:
    merged_issue = _GetAndUpdateMergedIssue(flake_issue)
    if flake_issue.issue_id != merged_issue.id:
      previous_tracking_bug_id = flake_issue.issue_id

    if merged_issue.open:
      logging.info(
          'Currently attached issue %s is open, update flake: %s with new '
          'occurrences.',
          FlakeIssue.GetLinkForIssue(monorail_project, merged_issue.id),
          target_flake.key)
      issue_generator.SetPreviousTrackingBugId(previous_tracking_bug_id)
      monorail_util.UpdateIssueWithIssueGenerator(
          issue_id=merged_issue.id, issue_generator=issue_generator)
      return merged_issue.id

    previous_tracking_bug_id = merged_issue.id

  logging.info(
      'flake %s has no issue attached or the attached issue was closed.' %
      target_flake.key)

  # Re-use an existing open bug if possible.
  issue_id = SearchOpenIssueIdForFlakyTest(target_flake.normalized_test_name,
                                           monorail_project)
  if issue_id:
    logging.info(
        'An existing issue %s was found, attach it to flake: %s and update it '
        'with new occurrences.',
        FlakeIssue.GetLinkForIssue(monorail_project, issue_id),
        target_flake.key)
    _AssignIssueToFlake(issue_id, target_flake)
    issue_generator.SetPreviousTrackingBugId(previous_tracking_bug_id)
    monorail_util.UpdateIssueWithIssueGenerator(
        issue_id=issue_id, issue_generator=issue_generator)
    return issue_id

  logging.info('No existing open issue was found, create a new one.')
  issue_generator.SetPreviousTrackingBugId(previous_tracking_bug_id)
  issue_id = monorail_util.CreateIssueWithIssueGenerator(
      issue_generator=issue_generator)
  logging.info('%s was created for flake: %s.',
               FlakeIssue.GetLinkForIssue(monorail_project, issue_id),
               target_flake.key)
  _AssignIssueToFlake(issue_id, target_flake)
  return issue_id


def _AssignIssueToFlake(issue_id, flake):
  """Assigns an issue id to a flake, and created a FlakeIssue if necessary.

  Args:
    issue_id: Id of a Monorail issue.
    flake: A Flake Model entity.
  """
  assert flake, 'The flake entity cannot be None.'

  monorail_project = FlakeIssue.GetMonorailProjectFromLuciProject(
      flake.luci_project)
  flake_issue = _GetOrCreateFlakeIssue(issue_id, monorail_project)
  flake.flake_issue_key = flake_issue.key
  flake.put()
