# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from collections import defaultdict

from google.appengine.ext import ndb

from findit_v2.model.atomic_failure import AtomicFailure
from findit_v2.model.base_failure_analysis import BaseFailureAnalysis
from findit_v2.model.failure_group import BaseFailureGroup
from findit_v2.model.gitiles_commit import GitilesCommit
from findit_v2.model.luci_build import LuciBuild
from gae_libs.model.versioned_model import VersionedModel


def GetTestFailures(test_failure_entities):
  """Gets test failures in dict format."""
  test_failures = {}
  for test_failure_entity in test_failure_entities:
    if not test_failures.get(test_failure_entity.step_ui_name):
      test_failures[test_failure_entity.step_ui_name] = {
          'tests': [],
      }
    step_failures = test_failures.get(test_failure_entity.step_ui_name)

    if test_failure_entity.test:
      step_failures['tests'].append({
          'name': test_failure_entity.test,
          'properties': test_failure_entity.properties
      })
    else:
      step_failures['properties'] = test_failure_entity.properties

  return test_failures


class TestFailure(AtomicFailure):
  """ Test failure that cannot be further divided.

  Usually each entity should be for a test. But it's possible that Findit cannot
  get test level information for any reason, if so the entity could be at
  suite or step level.
  """
  # Name of the failed test
  test = ndb.StringProperty()

  # Key to the failure that this failure merges into.
  # No analysis on current failure, instead use the results of merged_failure.
  merged_failure_key = ndb.KeyProperty(kind='TestFailure')

  # Arguments number differs from overridden method - pylint: disable=W0221
  @classmethod
  def Create(cls,
             failed_build_key,
             step_ui_name,
             test,
             first_failed_build_id=None,
             last_passed_build_id=None,
             failure_group_build_id=None,
             files=None,
             merged_failure_key=None,
             properties=None):
    instance = super(TestFailure, cls).Create(
        failed_build_key, step_ui_name, first_failed_build_id,
        last_passed_build_id, failure_group_build_id, files, properties)
    instance.test = test
    instance.merged_failure_key = merged_failure_key
    return instance

  def GetFailureIdentifier(self):
    """Gets the identifier to differentiate a test failure in a step."""
    return self.test

  def GetMergedFailure(self):
    """Gets the most up-to-date merged_failure for the current failure."""
    if self.merged_failure_key:
      return self.merged_failure_key.get()

    if (self.first_failed_build_id == self.build_id and
        self.failure_group_build_id == self.build_id):
      # First failure without being merged into any other failure group.
      return self

    # In a special case that a non-first failure was processed before the first
    # failure, it's possible that the merged_failure_key is not stored in the
    # non-first failure.
    merged_failure_key = self.GetMergedFailureKey(
        {}, self.first_failed_build_id, self.step_ui_name,
        self.GetFailureIdentifier())

    if merged_failure_key:
      self.merged_failure_key = merged_failure_key
      self.put()
      return merged_failure_key.get()

    return None


class TestFailureGroup(BaseFailureGroup):
  """Class for group of test failures."""

  # Keys to the failures in the first build of the group, remain
  # unchanged after the group is created.
  # These failures are used to compare to failures in other builds and decide
  # if those failures can be added to the group.
  # If they can, add the failures to this group by setting their
  # failure_group_build_id this group's id.
  test_failure_keys = ndb.KeyProperty(TestFailure, repeated=True)

  @property
  def test_failures(self):
    """Gets test failures that are included in the group."""
    failed_target_objects = ndb.get_multi(self.test_failure_keys)
    return GetTestFailures(failed_target_objects)

  # Arguments number differs from overridden method - pylint: disable=W0221
  @classmethod
  def Create(cls, luci_project, luci_bucket, build_id, gitiles_host,
             gitiles_project, gitiles_ref, last_passed_gitiles_id,
             last_passed_commit_position, first_failed_gitiles_id,
             first_failed_commit_position, test_failure_keys):
    assert test_failure_keys, (
        'no failed_targets when creating TestFailureGroup for {}'.format(
            build_id))

    instance = super(TestFailureGroup, cls).Create(
        luci_project, luci_bucket, build_id, gitiles_host, gitiles_project,
        gitiles_ref, last_passed_gitiles_id, last_passed_commit_position,
        first_failed_gitiles_id, first_failed_commit_position)

    instance.test_failure_keys = test_failure_keys
    return instance


class TestFailureAnalysis(BaseFailureAnalysis, VersionedModel):
  """Class for a test analysis.

  This class stores information that is needed during the analysis, and also
  some metadata for the analysis.

  The objects are versioned, so when rerun, Findit will create an entity with
  newer version, instead of deleting the existing analysis.
  """
  # Key to the failed targets this analysis analyzes.
  test_failure_keys = ndb.KeyProperty(TestFailure, repeated=True)

  # Arguments number differs from overridden method - pylint: disable=W0221
  @classmethod
  def Create(cls, luci_project, luci_bucket, luci_builder, build_id,
             gitiles_host, gitiles_project, gitiles_ref, last_passed_gitiles_id,
             last_passed_commit_position, first_failed_gitiles_id,
             first_failed_commit_position, rerun_builder_id, test_failure_keys):
    instance = super(TestFailureAnalysis, cls).Create(build_id)

    last_passed_commit = GitilesCommit(
        gitiles_host=gitiles_host,
        gitiles_project=gitiles_project,
        gitiles_ref=gitiles_ref,
        gitiles_id=last_passed_gitiles_id,
        commit_position=last_passed_commit_position)

    first_failed_commit = GitilesCommit(
        gitiles_host=gitiles_host,
        gitiles_project=gitiles_project,
        gitiles_ref=gitiles_ref,
        gitiles_id=first_failed_gitiles_id,
        commit_position=first_failed_commit_position)

    instance.luci_project = luci_project
    instance.bucket_id = '{}/{}'.format(luci_project, luci_bucket)
    instance.builder_id = '{}/{}/{}'.format(luci_project, luci_bucket,
                                            luci_builder)
    instance.build_id = build_id
    instance.last_passed_commit = last_passed_commit
    instance.first_failed_commit = first_failed_commit
    instance.rerun_builder_id = rerun_builder_id
    instance.test_failure_keys = test_failure_keys

    return instance

  def Update(self, end_time=None, status=None, error=None):
    # pylint: disable=attribute-defined-outside-init
    self.end_time = self.end_time or end_time
    self.status = status if status is not None else self.status
    self.error = error if error else self.error
    self.put()


class TestFailureInRerunBuild(ndb.Model):
  """Atomic test failure in a rerun build.

  Since we only need to keep a simple record on what's failed in rerun build,
  it's no need to reuse TestFailure.
  """
  # Full step name.
  step_ui_name = ndb.StringProperty()

  # Failed test name.
  test = ndb.StringProperty()


class TestRerunBuild(LuciBuild):
  """Class for a rerun build for a compile failure analysis."""

  # Compile failures in the rerun build.
  failures = ndb.LocalStructuredProperty(TestFailureInRerunBuild, repeated=True)

  def GetFailuresInBuild(self):
    """Gets a list of failed tests of each step that failed in the rerun build.
    """
    test_failures = defaultdict(list)
    for failure in self.failures or []:
      test_failures[failure.step_ui_name].append(failure.test)

    return {
        step_ui_name: list(set(failed_targets_in_step) - {None})
        for step_ui_name, failed_targets_in_step in test_failures.iteritems()
    }

  def SaveRerunBuildResults(self, status, detailed_test_failures):
    """Saves the results of this rerun build.

    Args:
      status (int): status of the build. See common_pb2 for available values.
      detailed_test_failures (dict): Compile failures in the rerun build.
      Format is like:
      {
        'step_name': {
          'failures': {
            'test1': {
              'first_failed_build': {
                'id': 8765432109,
                'number': 123,
                'commit_id': 654321
              },
              'last_passed_build': None,
              'properties': {
                # Arbitrary information about the failure if exists.
              }
            },
          'first_failed_build': {
            'id': 8765432109,
            'number': 123,
            'commit_id': 654321
          },
          'last_passed_build': None,
          'properties': {
            # Arbitrary information about the failure if exists.
          }
        },
      }
    """
    self.status = status  # pylint: disable=attribute-defined-outside-init
    self.failures = []
    for step_ui_name, step_info in detailed_test_failures.iteritems():
      for test in step_info['failures']:
        failure_entity = TestFailureInRerunBuild(
            step_ui_name=step_ui_name, test=test)
        self.failures.append(failure_entity)
    self.put()

  @classmethod
  def SearchBuildOnCommit(cls, analysis_key, commit):
    return cls.query(ancestor=analysis_key).filter(
        cls.gitiles_commit.gitiles_id == commit.gitiles_id).fetch()
