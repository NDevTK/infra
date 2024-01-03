# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import six
from google.appengine.api import datastore_errors
from google.appengine.ext import ndb

if six.PY2:
  from google.appengine.ext.ndb import msgprop
  from protorpc import messages
else:
  from enum import IntEnum

class DependencyRepository(ndb.Model):
  # The source absolute path of the checkout into the root repository.
  # Example: "//third_party/pdfium/" for pdfium in a chromium/src checkout.
  path = ndb.StringProperty(indexed=False, required=True)

  # The Gitiles hostname, e.g. "pdfium.googlesource.com".
  server_host = ndb.StringProperty(indexed=False, required=True)

  # The Gitiles project name, e.g. "pdfium.git".
  project = ndb.StringProperty(indexed=False, required=True)

  # The commit hash of the revision.
  revision = ndb.StringProperty(indexed=False, required=True)

  @property
  def project_url(self):
    return 'https://%s/%s' % (self.server_host, self.project)


def PercentageValidator(_, value):
  """Validates that the total number of lines is greater than 0."""
  if value <= 0:
    raise datastore_errors.BadValueError(
        'total_lines is expected to be greater than 0.')

  return value


class CoveragePercentage(ndb.Model):
  """Represents code coverage percentage metric for a file.

  It is stored as a part of PresubmitCoverageData.
  """

  # The source absolute path of the file. E.g. //base/test.cc.
  path = ndb.StringProperty(indexed=False, required=True)

  # Total number of lines.
  total_lines = ndb.IntegerProperty(
      indexed=False, required=True, validator=PercentageValidator)

  # Number of covered lines.
  covered_lines = ndb.IntegerProperty(indexed=False, required=True)


class CLPatchset(ndb.Model):
  """Represents a CL patchset."""

  # The Gerrit hostname, e.g. "chromium-review.googlesource.com".
  server_host = ndb.StringProperty(indexed=True, required=True)

  # The Gerrit project name, e.g. "chromium/src".
  # Note that project is optional because the other three already uniquely
  # identifies a CL patchset.
  project = ndb.StringProperty(indexed=True, required=False)

  # The Gerrrit change number, e.g. "138000".
  change = ndb.IntegerProperty(indexed=True, required=True)

  # The Gerrit patchset number, e.g. "2".
  patchset = ndb.IntegerProperty(indexed=True, required=True)


if six.PY2:

  class BlockingStatus(messages.Enum):
    """Represents the state machine for blocking low coverage cl logic."""

    # Default. A CL will not be blocked if it has default blocking status
    DEFAULT = 0
    # At least one of the coverage builds failed. Do not block the
    # corresponding CL.
    DONT_BLOCK_BUILDER_FAILURE = 1
    # All coverage builds' data has been processed.
    # CL is awaiting a verdict from the blocking logic.
    READY_FOR_VERDICT = 2
    # Blocking logic has decided to not block the CL.
    VERDICT_NOT_BLOCK = 3
    # Blocking algorithm has decided to block the CL.
    VERDICT_BLOCK = 4
else:

  class BlockingStatus(IntEnum):
    """Represents the state machine for blocking low coverage cl logic."""

    # Default. A CL will not be blocked if it has default blocking status
    DEFAULT = 0
    # At least one of the coverage builds failed. Do not block the
    # corresponding CL.
    DONT_BLOCK_BUILDER_FAILURE = 1
    # All coverage builds' data has been processed.
    # CL is awaiting a verdict from the blocking logic.
    READY_FOR_VERDICT = 2
    # Blocking logic has decided to not block the CL.
    VERDICT_NOT_BLOCK = 3
    # Blocking algorithm has decided to block the CL.
    VERDICT_BLOCK = 4

class LowCoverageBlocking(ndb.Model):
  """Represents the state machine for blocking low coverage cl logic."""

  # Key for the CL Patchset to which this entity belongs to
  cl_patchset = ndb.StructuredProperty(CLPatchset, indexed=True, required=True)

  if six.PY2:
    # Determines if the corresponding patchset may be blocked or not
    blocking_status = msgprop.EnumProperty(
        BlockingStatus, indexed=True, default=BlockingStatus.DEFAULT)
  else:
    blocking_status = ndb.IntegerProperty(
        choices=list(BlockingStatus),
        indexed=True,
        default=BlockingStatus.DEFAULT)
  # List of try builders from whom coverage data is expected to be received
  expected_builders = ndb.StringProperty(repeated=True)

  # List of coverage try builders which ended successfully.
  successful_builders = ndb.StringProperty(repeated=True)

  # List of coverage try builders which were processed successfully by
  # coverage service.
  processed_builders = ndb.StringProperty(repeated=True)

  @classmethod
  def _CreateKey(cls, server_host, change, patchset):
    return ndb.Key(cls, '%s$%s$%s' % (server_host, change, patchset))

  @classmethod
  def Create(cls,
             server_host,
             change,
             patchset,
             blocking_status=BlockingStatus.DEFAULT,
             expected_builders=None,
             successful_builders=None,
             processed_builders=None,
             project=None):
    key = cls._CreateKey(server_host, change, patchset)
    cl_patchset = CLPatchset(
        server_host=server_host,
        project=project,
        change=change,
        patchset=patchset,
    )
    return cls(
        key=key,
        cl_patchset=cl_patchset,
        blocking_status=blocking_status,
        expected_builders=expected_builders or [],
        successful_builders=successful_builders or [],
        processed_builders=processed_builders or [])

  @classmethod
  def Get(cls, server_host, change, patchset):
    return cls.GetAsync(server_host, change, patchset).get_result()

  @classmethod
  def GetAsync(cls, server_host, change, patchset):
    return cls._CreateKey(server_host, change, patchset).get_async()


class PresubmitCoverageData(ndb.Model):
  """Represents the code coverage data of a change during presubmit."""

  # The CL patchset.
  cl_patchset = ndb.StructuredProperty(CLPatchset, indexed=True, required=True)

  # A list of file level coverage data for all the source files modified by
  # the CL.
  data = ndb.JsonProperty(indexed=False, compressed=True, required=False)

  # A list of file level coverage data (unit tests only) for all the source
  # files modified by this CL.
  data_unit = ndb.JsonProperty(indexed=False, compressed=True, required=False)

  # A list of file level coverage data for all the source files modified by the
  # this CL, based on coverage data collected from builders using RTS
  data_rts = ndb.JsonProperty(indexed=False, compressed=True, required=False)

  # A list of file level coverage data (unit tests only) for all the source
  # files modified by this CL, based on coverage data collected from
  # builders using RTS
  data_unit_rts = ndb.JsonProperty(
      indexed=False, compressed=True, required=False)

  # Coverage percentages(overall) of all executable lines of the files.
  absolute_percentages = ndb.LocalStructuredProperty(
      CoveragePercentage, indexed=False, repeated=True)

  # Coverage percentages(overall) of *newly added* and executable lines
  # of the files.
  incremental_percentages = ndb.LocalStructuredProperty(
      CoveragePercentage, indexed=False, repeated=True)

  # Coverage percentages(unit) of all executable lines of the files.
  absolute_percentages_unit = ndb.LocalStructuredProperty(
      CoveragePercentage, indexed=False, repeated=True)

  # Coverage percentages(unit) of *newly added* and executable lines
  # of the files.
  incremental_percentages_unit = ndb.LocalStructuredProperty(
      CoveragePercentage, indexed=False, repeated=True)

  # Coverage percentages(overall) of all executable lines of the files,
  # based on coverage data collected from builders using RTS
  absolute_percentages_rts = ndb.LocalStructuredProperty(
      CoveragePercentage, indexed=False, repeated=True)

  # Coverage percentages(unit) of all executable lines of the files,
  # based on coverage data collected from builders using RTS
  absolute_percentages_unit_rts = ndb.LocalStructuredProperty(
      CoveragePercentage, indexed=False, repeated=True)

  # If assigned, represents the patchset number from which this coverage data is
  # generated, and it specifically refers to the scenario where coverage data
  # are shared between equivalent patchsets, such as trivial-rebase.
  based_on = ndb.IntegerProperty(indexed=True)

  # Timestamp this coverage report got created.
  insert_timestamp = ndb.DateTimeProperty(auto_now_add=True)

  # Timestamp this coverage report was last updated.
  update_timestamp = ndb.DateTimeProperty(auto_now=True)

  # number of times `data` field gets updated
  times_updated = ndb.IntegerProperty(required=False, default=0)

  # number of times `data_unit` field gets updated
  times_updated_unit = ndb.IntegerProperty(required=False, default=0)

  @classmethod
  def _CreateKey(cls, server_host, change, patchset):
    return ndb.Key(cls, '%s$%s$%s' % (server_host, change, patchset))

  @classmethod
  def Create(cls,
             server_host,
             change,
             patchset,
             data=None,
             data_unit=None,
             data_rts=None,
             data_unit_rts=None,
             project=None):
    assert data or data_unit, "Atleast one of data/data_unit must be specified."
    key = cls._CreateKey(server_host, change, patchset)
    cl_patchset = CLPatchset(
        server_host=server_host,
        project=project,
        change=change,
        patchset=patchset)
    return cls(
        key=key,
        cl_patchset=cl_patchset,
        data=data,
        data_unit=data_unit,
        data_rts=data_rts,
        data_unit_rts=data_unit_rts)

  @classmethod
  def Get(cls, server_host, change, patchset):
    return cls.GetAsync(server_host, change, patchset).get_result()

  @classmethod
  def GetAsync(cls, server_host, change, patchset):
    return cls._CreateKey(server_host, change, patchset).get_async()


class GitilesCommit(ndb.Model):
  """Represents a Gitiles commit."""

  # The Gitiles hostname, e.g. "chromium.googlesource.com".
  server_host = ndb.StringProperty(indexed=True, required=True)

  # The Gitiles project name, e.g. "chromium/src".
  project = ndb.StringProperty(indexed=True, required=True)

  # The giiles ref, e.g. "refs/heads/master".
  # NOT a branch name: if specified, must start with "refs/".
  ref = ndb.StringProperty(indexed=True, required=True)

  # The commit hash of the revision.
  revision = ndb.StringProperty(indexed=True, required=True)


class CoverageReportModifier(ndb.Model):
  """Represents a filter setting  used to generate custom coverage reports."""

  # The Gitiles hostname, e.g. "chromium.googlesource.com".
  server_host = ndb.StringProperty(
      indexed=True, default='chromium.googlesource.com', required=True)

  # The Gitiles project name, e.g. "chromium/src.git".
  project = ndb.StringProperty(
      indexed=True, default='chromium/src', required=True)

  # Controls whether custom reports are to be generated for this modifier or not
  is_active = ndb.BooleanProperty(indexed=True, default=True, required=True)

  # Gerrit hashtag to uniquely identify a feature.
  gerrit_hashtag = ndb.StringProperty(indexed=True)

  # email of a chromium contributor
  author = ndb.StringProperty(indexed=True)

  # Reference commit to generate coverage reports past a checkpoint.
  reference_commit = ndb.StringProperty(indexed=True)

  # Timestamp when the reference commit was committed.
  reference_commit_timestamp = ndb.DateTimeProperty(indexed=True)

  # Timestamp this modifier got created.
  insert_timestamp = ndb.DateTimeProperty(auto_now_add=True)

  # Timestamp this modifier was last updated.
  update_timestamp = ndb.DateTimeProperty(auto_now=True)

  @classmethod
  def Get(cls, modifier_id):
    return ndb.Key(cls, int(modifier_id)).get()

  @classmethod
  def InsertModifierForAuthorIfNeeded(cls,
                                      server_host,
                                      project,
                                      author,
                                      is_active=True):
    query = CoverageReportModifier.query(
        CoverageReportModifier.server_host == server_host,
        CoverageReportModifier.project == project,
        CoverageReportModifier.author == author)
    modifier = query.get()
    if not modifier:
      modifier = cls(
          server_host=server_host,
          project=project,
          author=author,
          is_active=is_active)
      modifier.put()
    return modifier


class PostsubmitReport(ndb.Model):
  """Represents a postsubmit code coverage report."""

  # The Gitiles commit.
  gitiles_commit = ndb.StructuredProperty(
      GitilesCommit, indexed=True, required=True)

  # An optional increasing numeric number assigned to each commit.
  commit_position = ndb.IntegerProperty(indexed=True, required=False)

  # Timestamp when the commit was committed.
  commit_timestamp = ndb.DateTimeProperty(indexed=True, required=True)

  # TODO(crbug.com/939443): Make it required once data are backfilled.
  # Name of the luci builder that generates the data.
  bucket = ndb.StringProperty(indexed=True, required=False)
  builder = ndb.StringProperty(indexed=True, required=False)

  # Manifest of all the code checkouts when the coverage report is generated.
  # In descending order by the length of the relative path in the root checkout.
  manifest = ndb.LocalStructuredProperty(
      DependencyRepository, repeated=True, indexed=False)

  # The top level coverage metric of the report.
  # For Clang based languages, the format is a list of 3 dictionaries
  # corresponds to 'line', 'function' and 'region' respectively, and each dict
  # has format: {'covered': 9526650, 'total': 12699841, 'name': u'|name|'}
  summary_metrics = ndb.JsonProperty(indexed=False, required=True)

  # The build id that uniquely identifies the build.
  build_id = ndb.IntegerProperty(indexed=False, required=True)

  # Used to control if a report is visible to the users, and the main use case
  # is to quanrantine a 'bad' report. All the reports are visible to admins.
  visible = ndb.BooleanProperty(indexed=True, default=False, required=True)

  # TODO(crbug.com/1237114): Mark as required once data are backfilled.
  # Id of the associated coverage report modifier, 0 otherwise
  # For e.g. a usual full codebase PostSubmitReport would
  # have modifier_id as 0.
  modifier_id = ndb.IntegerProperty(indexed=True, required=False)

  @classmethod
  def _CreateKey(cls, server_host, project, ref, revision, bucket, builder,
                 modifier_id):
    return ndb.Key(
        cls, '%s$%s$%s$%s$%s$%s$%s' % (server_host, project, ref, revision,
                                       bucket, builder, str(modifier_id)))

  @classmethod
  def Create(cls,
             server_host,
             project,
             ref,
             revision,
             bucket,
             builder,
             commit_timestamp,
             manifest,
             summary_metrics,
             build_id,
             visible,
             commit_position=None,
             modifier_id=0):
    key = cls._CreateKey(server_host, project, ref, revision, bucket, builder,
                         modifier_id)
    gitiles_commit = GitilesCommit(
        server_host=server_host, project=project, ref=ref, revision=revision)
    return cls(
        key=key,
        gitiles_commit=gitiles_commit,
        bucket=bucket,
        builder=builder,
        commit_position=commit_position,
        commit_timestamp=commit_timestamp,
        manifest=manifest,
        summary_metrics=summary_metrics,
        build_id=build_id,
        visible=visible,
        modifier_id=modifier_id)

  @classmethod
  def Get(cls,
          server_host,
          project,
          ref,
          revision,
          bucket,
          builder,
          modifier_id=0):
    entity_v3 = cls._CreateKey(server_host, project, ref, revision, bucket,
                               builder, modifier_id).get()
    if entity_v3:
      return entity_v3

    # TODO(crbug.com/1237114): Remove following code once data are backfilled.
    entity_v2 = ndb.Key(
        cls, '%s$%s$%s$%s$%s$%s' %
        (server_host, project, ref, revision, bucket, builder)).get()
    if entity_v2:
      return entity_v2

    # TODO(crbug.com/939443): Remove following code once data are backfilled.
    legacy_key_v1 = ndb.Key(
        cls, '%s$%s$%s$%s' % (server_host, project, ref, revision))
    return legacy_key_v1.get()


class SummaryCoverageData(ndb.Model):
  """Represents the code coverage data of a directory or a component."""

  # The Gitiles commit.
  gitiles_commit = ndb.StructuredProperty(
      GitilesCommit, indexed=True, required=True)

  # Type of the summary coverage data.
  data_type = ndb.StringProperty(
      indexed=True, choices=['dirs', 'components'], required=True)

  # Source absolute path to the file or path to the components. E.g:
  # // or /media/cast/net/rtp/frame_buffer.cc for directories.
  # >> or Blink>Fonts for components.
  path = ndb.StringProperty(indexed=True, required=True)

  # TODO(crbug.com/939443): Make them required once data are backfilled.
  # Name of the luci builder that generates the data.
  bucket = ndb.StringProperty(indexed=True, required=False)
  builder = ndb.StringProperty(indexed=True, required=False)

  # Coverage data for a directory or a component.
  data = ndb.JsonProperty(indexed=False, compressed=True, required=True)

  # TODO(crbug.com/1237114): Mark as required once data are backfilled.
  # Id of the associated coverage report modifier, 0 otherwise
  # For e.g. a summary entity corresponding to usual full codebase
  # PostSubmitReport would have modifier_id as 0.
  modifier_id = ndb.IntegerProperty(indexed=True, required=False)

  @classmethod
  def _CreateKey(cls, server_host, project, ref, revision, data_type, path,
                 bucket, builder):
    return ndb.Key(
        cls, '%s$%s$%s$%s$%s$%s$%s$%s' %
        (server_host, project, ref, revision, data_type, path, bucket, builder))

  @classmethod
  def _CreateKey(cls, server_host, project, ref, revision, data_type, path,
                 bucket, builder, modifier_id):
    return ndb.Key(
        cls, '%s$%s$%s$%s$%s$%s$%s$%s$%s' %
        (server_host, project, ref, revision, data_type, path, bucket, builder,
         modifier_id))

  @classmethod
  def Create(cls,
             server_host,
             project,
             ref,
             revision,
             data_type,
             path,
             bucket,
             builder,
             data,
             modifier_id=0):
    if data_type == 'dirs':
      assert path.startswith('//'), 'Directory path must start with //'

    key = cls._CreateKey(server_host, project, ref, revision, data_type, path,
                         bucket, builder, modifier_id)
    gitiles_commit = GitilesCommit(
        server_host=server_host, project=project, ref=ref, revision=revision)
    return cls(
        key=key,
        gitiles_commit=gitiles_commit,
        data_type=data_type,
        path=path,
        bucket=bucket,
        builder=builder,
        data=data,
        modifier_id=modifier_id)

  @classmethod
  def Get(cls,
          server_host,
          project,
          ref,
          revision,
          data_type,
          path,
          bucket,
          builder,
          modifier_id=0):
    entity_v3 = cls._CreateKey(server_host, project, ref, revision, data_type,
                               path, bucket, builder, modifier_id).get()
    if entity_v3:
      return entity_v3

    # TODO(crbug.com/1237114): Remove following code once data are backfilled.
    entity_v2 = ndb.Key(
        cls,
        '%s$%s$%s$%s$%s$%s$%s$%s' % (server_host, project, ref, revision,
                                     data_type, path, bucket, builder)).get()
    if entity_v2:
      return entity_v2

    # TODO(crbug.com/939443): Remove following code once data are backfilled.
    legacy_key_v1 = ndb.Key(
        cls, '%s$%s$%s$%s$%s$%s' %
        (server_host, project, ref, revision, data_type, path))
    return legacy_key_v1.get()


class FileCoverageData(ndb.Model):
  """Represents the code coverage data of a single file.

  File can be from a dependency checkout, and it can be a generated file instead
  of a source file checked into the repo.
  """

  # The Gitiles commit.
  gitiles_commit = ndb.StructuredProperty(
      GitilesCommit, indexed=True, required=True)

  # Source absolute file path.
  path = ndb.StringProperty(indexed=True, required=True)

  # TODO(crbug.com/939443): Make it required once data are backfilled.
  # Name of the luci builder that generates the data.
  bucket = ndb.StringProperty(indexed=True, required=False)
  builder = ndb.StringProperty(indexed=True, required=False)

  # Coverage data for a single file.
  # Json structure corresponds to File proto at
  # https://chromium.googlesource.com/infra/infra/+/refs/heads/main/appengine/findit/model/proto/code_coverage.proto
  data = ndb.JsonProperty(indexed=False, compressed=True, required=True)

  # TODO(crbug.com/1237114): Mark as required once data are backfilled.
  # Id of the associated coverage report modifier, 0 otherwise
  # For e.g. FileCoverageData corresponding to default PostSubmitReport would
  # have modifier_id as 0.
  modifier_id = ndb.IntegerProperty(indexed=True, required=False)

  @classmethod
  def _CreateKey(cls, server_host, project, ref, revision, path, bucket,
                 builder, modifier_id):
    return ndb.Key(
        cls,
        '%s$%s$%s$%s$%s$%s$%s$%s' % (server_host, project, ref, revision, path,
                                     bucket, builder, str(modifier_id)))

  @classmethod
  def Create(cls,
             server_host,
             project,
             ref,
             revision,
             path,
             bucket,
             builder,
             data,
             modifier_id=0):
    assert path.startswith('//'), 'File path must start with "//"'

    key = cls._CreateKey(server_host, project, ref, revision, path, bucket,
                         builder, modifier_id)
    gitiles_commit = GitilesCommit(
        server_host=server_host, project=project, ref=ref, revision=revision)
    return cls(
        key=key,
        gitiles_commit=gitiles_commit,
        path=path,
        bucket=bucket,
        builder=builder,
        modifier_id=modifier_id,
        data=data)

  @classmethod
  def Get(cls,
          server_host,
          project,
          ref,
          revision,
          path,
          bucket,
          builder,
          modifier_id=0):
    entity_v3 = cls._CreateKey(server_host, project, ref, revision, path,
                               bucket, builder, modifier_id).get()
    if entity_v3:
      return entity_v3

    # TODO(crbug.com/1237114): Remove following code once data are backfilled.
    entity_v2 = ndb.Key(
        cls, '%s$%s$%s$%s$%s$%s$%s' %
        (server_host, project, ref, revision, path, bucket, builder)).get()
    if entity_v2:
      return entity_v2

    # TODO(crbug.com/939443): Remove following code once data are backfilled.
    legacy_key_v1 = ndb.Key(
        cls, '%s$%s$%s$%s$%s' % (server_host, project, ref, revision, path))
    return legacy_key_v1.get()

  @classmethod
  def GetModifiedLineCoverage(cls, file_coverage, lines_of_interest):
    """Returns line coverage metrics for interesting lines in a file.

    This function returns a modified `data` field in FileCoverageData,
    where only coverage data for lines_of_interest in a file is retained.

    Args:
      file_coverage (FileCoverageData): File coverage report.
      lines_of_interest (set): Set of lines whose coverage is to be retained.

    Returns:
      A dict containing coverage info dropped for all lines except
      lines_of_interest. Returns None if there are no lines with coverage info.
    """

    total = 0
    covered = 0
    # add a dummy range to simplify logic
    modified_line_ranges = [{'first': -1, 'last': -1}]
    for line_range in file_coverage.data['lines']:
      for line_num in range(line_range['first'], line_range['last'] + 1):
        if line_num in lines_of_interest:
          total += 1
          if line_num == modified_line_ranges[-1]['last'] + 1 and line_range[
              'count'] == modified_line_ranges[-1]['count']:
            # Append to the last interesting line range if line numbers are
            # continuous and they share the same execution count
            modified_line_ranges[-1]['last'] += 1
          else:
            # Line range gets broken by an unmodified line
            # or new line range with a different execution count is encountered
            modified_line_ranges.append({
                'first': line_num,
                'last': line_num,
                'count': line_range['count']
            })
          if line_range['count'] != 0:
            covered += 1
    if total > 0:
      data = {
          'path': file_coverage.path,
          'lines': modified_line_ranges[1:],
          'summaries': [{
              'name': 'line',
              'total': total,
              'covered': covered
          }],
          'revision': file_coverage.gitiles_commit.revision
      }
      return data
