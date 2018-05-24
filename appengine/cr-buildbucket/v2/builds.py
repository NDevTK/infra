# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Provides functions specific to v2 builds."""

import logging

from google.protobuf import struct_pb2
from google.protobuf import timestamp_pb2

from . import errors
from proto import build_pb2
from proto import common_pb2
import buildtags
import model


def build_to_v2_partial(build):
  """Converts a model.Build to an incomplete build_pb2.Build.

  The returned build does not include steps.

  May raise errors.UnsupportedBuild or errors.MalformedBuild.
  """
  result_details = build.result_details or {}
  ret = build_pb2.Build(
      id=build.key.id(),
      builder=_get_builder_id(build),
      number=result_details.get('properties', {}).get('buildnumber') or 0,
      created_by=build.created_by.to_bytes(),
      view_url=build.url,
      create_time=_dt2ts(build.create_time),
      start_time=_dt2ts(build.start_time),
      end_time=_dt2ts(build.complete_time),
      update_time=_dt2ts(build.update_time),
      input=build_pb2.Build.Input(
          properties=_dict_to_struct(build.parameters.get('properties')),
          experimental=build.experimental,
      ),
      output=build_pb2.Build.Output(
          properties=_dict_to_struct(result_details.get('properties')),),
      infra=build_pb2.BuildInfra(
          buildbucket=build_pb2.BuildInfra.Buildbucket(canary=build.canary),
          swarming=build_pb2.BuildInfra.Swarming(
              hostname=build.swarming_hostname,
              task_id=build.swarming_task_id,
              task_service_account=build.service_account,
          ),
      ),
  )
  status_to_v2(build, ret)

  task_result = result_details.get('swarming', {}).get('task_result', {})
  for d in task_result.get('bot_dimensions', []):
    for v in d['value']:
      ret.infra.swarming.bot_dimensions.add(key=d['key'], value=v)

  _parse_tags(ret, build.tags)
  return ret


def _parse_tags(dest_msg, tags):
  for t in tags:
    # All builds in the datastore have tags that have a colon.
    k, v = buildtags.parse(t)
    if k == buildtags.BUILDER_KEY:
      # we've already parsed builder from parameters
      # and build creation code guarantees consinstency.
      # Exclude builder tag.
      continue
    if k == buildtags.BUILDSET_KEY:
      m = buildtags.RE_BUILDSET_GITILES_COMMIT.match(v)
      if m:
        dest_msg.input.gitiles_commits.add(
            host=m.group(1),
            project=m.group(2),
            id=m.group(3),
        )
        continue

      m = buildtags.RE_BUILDSET_GERRIT_CL.match(v)
      if m:
        dest_msg.input.gerrit_changes.add(
            host=m.group(1),
            change=int(m.group(2)),
            patchset=int(m.group(3)),
        )
        # TODO(nodir): fetch project from gerrit
        continue
    elif k == buildtags.SWARMING_DIMENSION_KEY:
      if ':' in v:  # pragma: no branch
        k2, v2 = buildtags.parse(v)
        dest_msg.infra.swarming.task_dimensions.add(key=k2, value=v2)

      # This line is actually covered, but pycoverage thinks it is not.
      # It cannot be not covered because the if statement above is covered.
      continue  # pragma: no cover
    elif k == buildtags.SWARMING_TAG_KEY:
      if ':' in v:  # pragma: no branch
        k2, v2 = buildtags.parse(v)
        if k2 == 'priority':
          try:
            dest_msg.infra.swarming.priority = int(v2)
          except ValueError as ex:
            logging.warning('invalid tag %r: %s', t, ex)
        elif k2 == 'buildbucket_template_revision':  # pragma: no branch
          dest_msg.infra.buildbucket.service_config_revision = v2

      # Exclude all "swarming_tag" tags.
      continue
    elif k == buildtags.BUILD_ADDRESS_KEY:
      try:
        _, _, dest_msg.number = buildtags.parse_build_address(v)
        continue
      except ValueError as ex:  # pragma: no cover
        raise errors.MalformedBuild('invalid build address "%s": %s' % (v, ex))
    elif k in ('swarming_hostname', 'swarming_task_id'):
      # These tags are added automatically and are covered by proto fields.
      # Omit them.
      continue

    dest_msg.tags.add(key=k, value=v)


def _get_builder_id(build):
  builder = (build.parameters or {}).get(model.BUILDER_PARAMETER)
  if not builder:
    raise errors.UnsupportedBuild(
        'does not have %s parameter' % model.BUILDER_PARAMETER)

  bucket = build.bucket
  # in V2, we drop "luci.{project}." prefix.
  luci_prefix = 'luci.%s.' % build.project
  if bucket.startswith(luci_prefix):
    bucket = bucket[len(luci_prefix):]
  return build_pb2.Builder.ID(
      project=build.project,
      bucket=bucket,
      builder=builder,
  )


def status_to_v2(src, dest):
  """Converts a V1 status to V2 status.

  Args:
    src: a model.Build, source of V1 status.
    dest: a build_pb2.Build, destination of V2 status. Its status and
      all of status_reason fields will be mutated.
  """
  dest.status = common_pb2.STATUS_UNSPECIFIED
  dest.ClearField('infra_failure_reason')
  dest.ClearField('cancel_reason')

  if src.status == model.BuildStatus.SCHEDULED:
    dest.status = common_pb2.SCHEDULED
  elif src.status == model.BuildStatus.STARTED:
    dest.status = common_pb2.STARTED
  elif src.status == model.BuildStatus.COMPLETED:  # pragma: no branch
    if src.result == model.BuildResult.SUCCESS:
      dest.status = common_pb2.SUCCESS
    elif src.result == model.BuildResult.FAILURE:
      if src.failure_reason == model.FailureReason.BUILD_FAILURE:
        dest.status = common_pb2.FAILURE
      else:
        dest.status = common_pb2.INFRA_FAILURE
        dest.infra_failure_reason.resource_exhaustion = False
    elif src.result == model.BuildResult.CANCELED:  # pragma: no branch
      if src.cancelation_reason == model.CancelationReason.CANCELED_EXPLICITLY:
        dest.status = common_pb2.CANCELED
        # V1 doesn't provide any cancel details.
      elif src.cancelation_reason == model.CancelationReason.TIMEOUT:
        # V1 timeout is V2 infra failure with resource exhaustion.
        dest.status = common_pb2.INFRA_FAILURE
        dest.infra_failure_reason.resource_exhaustion = True

  if dest.status == common_pb2.STATUS_UNSPECIFIED:  # pragma: no cover
    raise errors.MalformedBuild('invalid status in src %d' % src.key.id())


def status_to_v1(src, dest):
  """Converts a V1 status to V2 status.

  Args:
    src: a build_pb2.Build, source of V2 status.
    dest: a model.Build, destination of V1 status. Its status, result,
      failure_reason and cancelation_reason will be set.
  """
  dest.status = None
  dest.result = None
  dest.failure_reason = None
  dest.cancelation_reason = None

  if src.status == common_pb2.SCHEDULED:
    dest.status = model.BuildStatus.SCHEDULED
  elif src.status == common_pb2.STARTED:
    dest.status = model.BuildStatus.STARTED
  elif src.status == common_pb2.SUCCESS:
    dest.status = model.BuildStatus.COMPLETED
    dest.result = model.BuildResult.SUCCESS
  elif src.status == common_pb2.FAILURE:
    dest.status = model.BuildStatus.COMPLETED
    dest.result = model.BuildResult.FAILURE
    dest.failure_reason = model.FailureReason.BUILD_FAILURE
  elif src.status == common_pb2.INFRA_FAILURE:
    dest.status = model.BuildStatus.COMPLETED
    if src.infra_failure_reason.resource_exhaustion:
      # In python implementation, V2 resource exhaustion is V1 timeout.
      dest.result = model.BuildResult.CANCELED
      dest.cancelation_reason = model.CancelationReason.TIMEOUT
    else:
      dest.result = model.BuildResult.FAILURE
      dest.failure_reason = model.FailureReason.INFRA_FAILURE
  elif src.status == common_pb2.CANCELED:
    dest.status = model.BuildStatus.COMPLETED
    dest.result = model.BuildResult.CANCELED
    dest.cancelation_reason = model.CancelationReason.CANCELED_EXPLICITLY

  if dest.status is None:  # pragma: no cover
    raise errors.MalformedBuild('invalid status in src %d' % src.id)


def _dict_to_struct(d):
  if d is None:
    return None
  s = struct_pb2.Struct()
  s.update(d)
  return s


def _dt2ts(dt):
  if dt is None:
    return None
  ts = timestamp_pb2.Timestamp()
  ts.FromDatetime(dt)
  return ts
