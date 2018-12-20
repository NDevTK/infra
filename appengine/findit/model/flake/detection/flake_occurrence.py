# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from google.appengine.ext import ndb
from google.appengine.ext.ndb import msgprop

from model.flake.flake_type import FlakeType


class BuildConfiguration(ndb.Model):
  """Tracks the build configuration of a flake occurrence."""

  # Used to identify a build configuration.
  luci_project = ndb.StringProperty(required=True)
  luci_bucket = ndb.StringProperty(required=True)
  luci_builder = ndb.StringProperty(required=True)

  # Required for legacy reasons, such as:
  # 1. Flake Analyzer to trigger flake analysis.
  # 2. To obtain isolate_target_name/step_metadata of a step in a build.
  # This should be removed once all builders are migrated to LUCI.
  legacy_master_name = ndb.StringProperty(indexed=False, required=True)
  legacy_build_number = ndb.IntegerProperty(indexed=False, required=True)


class FlakeOccurrence(ndb.Model):
  """Tracks a flake occurrence that caused build or step level retries."""

  # The type of the flake.
  flake_type = msgprop.EnumProperty(FlakeType, required=True)

  # Used to identify the flaky build.
  build_id = ndb.IntegerProperty(indexed=False, required=True)

  # Used to identify the original name of a step displayed in a given build.
  # step_ui_name may include hardware information, 'with patch' and
  # 'without patch' postfix. For example: 'unit_tests (with patch) on Android'.
  step_ui_name = ndb.StringProperty(required=True)

  # Used to identify the original name of a test in a given test binary.
  # test_name may include 'PRE_' prefixes and parameters if it's a gtest.
  # For example: 'A/ColorSpaceTest.PRE_PRE_testNullTransform/137'.
  test_name = ndb.StringProperty(required=True)

  # Used to identify the build configuration.
  build_configuration = ndb.StructuredProperty(
      BuildConfiguration, required=True)

  # The time the flake occurrence happened (test start time).
  time_happened = ndb.DateTimeProperty(required=True)

  # The time the flake occurrence was detected, used to track the delay of flake
  # detection for this occurrence.
  time_detected = ndb.DateTimeProperty(required=True, auto_now_add=True)

  # The id of the gerrit cl this occurrence is associated with.
  gerrit_cl_id = ndb.IntegerProperty(required=True)

  # Tags that specify the category of the flake occurrence, e.g. builder name,
  # master name, step name, etc.
  tags = ndb.StringProperty(repeated=True)

  @staticmethod
  def GetId(flake_type, build_id, step_ui_name, test_name):
    return '%s@%s@%s@%s' % (flake_type, build_id, step_ui_name, test_name)

  @classmethod
  def Create(cls,
             flake_type,
             build_id,
             step_ui_name,
             test_name,
             luci_project,
             luci_bucket,
             luci_builder,
             legacy_master_name,
             legacy_build_number,
             time_happened,
             gerrit_cl_id,
             parent_flake_key,
             tags=None):
    """Creates a cq false rejection flake occurrence.

    Args:
      step_ui_name: The original displayed name of a step in a given build.
      test_name: The original name of a test in a give test binary.
      parent_flake_key: parent Flake model this occurrence is grouped under.
                        This method assumes that the parent Flake entity exists.

    For other args, please see model properties.
    """
    flake_occurrence_id = cls.GetId(flake_type, build_id, step_ui_name,
                                    test_name)
    build_configuration = BuildConfiguration(
        luci_project=luci_project,
        luci_bucket=luci_bucket,
        luci_builder=luci_builder,
        legacy_master_name=legacy_master_name,
        legacy_build_number=legacy_build_number)

    return cls(
        flake_type=flake_type,
        build_id=build_id,
        step_ui_name=step_ui_name,
        test_name=test_name,
        build_configuration=build_configuration,
        time_happened=time_happened,
        gerrit_cl_id=gerrit_cl_id,
        id=flake_occurrence_id,
        parent=parent_flake_key,
        tags=tags or [])


class LightWeightOccurrence(ndb.Model):
  """Light weight entity to save distinctive information about an occurrence."""
  # The time the flake occurrence happened (test start time).
  time_happened = ndb.DateTimeProperty(required=True)
  # The id of the gerrit cl this occurrence is associated with.
  gerrit_cl_id = ndb.IntegerProperty(required=True)


class CQHiddenFlakeOccurrence(FlakeOccurrence):
  """Tracks flake occurrences that caused test level retries on CQ.

  Because we expecting a large amount of hidden flake occurrences, so we will
  not store each and every such occurrences to data store. Instead, we will
  use one entity to store a group of occurrences if they:
  1. are detected within the same cron job,
  2. are for the same test, meaning they have the same step_ui_name and
    test_name,
  3. are on the same builder.

  Full information of one occurrence in the group will be stored as a sample,
  while others will be consolidated and only save their distinctive information.
  """
  occurrences = ndb.StructuredProperty(LightWeightOccurrence, repeated=True)

  # Time happened of the first occurrence in group.
  first_occurrence_time_happened = ndb.DateTimeProperty()
  # Time happened of the last occurrence in group.
  last_occurrence_time_happened = ndb.DateTimeProperty()

  @classmethod
  def Create(cls,
             flake_type,
             build_id,
             step_ui_name,
             test_name,
             luci_project,
             luci_bucket,
             luci_builder,
             legacy_master_name,
             legacy_build_number,
             time_happened,
             gerrit_cl_id,
             parent_flake_key,
             tags=None):
    occurrence = super(CQHiddenFlakeOccurrence, cls).Create(
        flake_type, build_id, step_ui_name, test_name, luci_project,
        luci_bucket, luci_builder, legacy_master_name, legacy_build_number,
        time_happened, gerrit_cl_id, parent_flake_key, tags)

    occurrence.first_occurrence_time_happened = time_happened
    occurrence.last_occurrence_time_happened = time_happened
    return occurrence
