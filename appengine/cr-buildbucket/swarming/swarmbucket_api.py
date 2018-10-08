# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json

from google.appengine.ext import ndb
from protorpc import messages
from protorpc import message_types
from protorpc import remote
import endpoints

from components import auth
from components import utils
import gae_ts_mon

from . import flatten_swarmingcfg
from . import swarming
from . import swarmingcfg
import api
import api_common
import config
import errors
import sequence
import user


def swarmbucket_api_method(
    request_message_class, response_message_class, **kwargs
):
  """Defines a swarmbucket API method."""

  endpoints_decorator = auth.endpoints_method(
      request_message_class, response_message_class, **kwargs
  )

  def decorator(fn):
    fn = auth.public(fn)
    fn = endpoints_decorator(fn)

    ts_mon_time = lambda: utils.datetime_to_timestamp(utils.utcnow()) / 1e6
    fn = gae_ts_mon.instrument_endpoint(time_fn=ts_mon_time)(fn)
    # ndb.toplevel must be the last one.
    # See also the comment in endpoint decorator in api.py
    return ndb.toplevel(fn)

  return decorator


class BuilderMessage(messages.Message):
  name = messages.StringField(1)
  category = messages.StringField(2)
  properties_json = messages.StringField(3)
  swarming_dimensions = messages.StringField(4, repeated=True)


class BucketMessage(messages.Message):
  # Bucket name. Unique per buildbucket instance.
  name = messages.StringField(1)
  builders = messages.MessageField(BuilderMessage, 2, repeated=True)
  swarming_hostname = messages.StringField(3)


class GetBuildersResponseMessage(messages.Message):
  buckets = messages.MessageField(BucketMessage, 1, repeated=True)


class GetTaskDefinitionRequestMessage(messages.Message):
  # A build creation request. Buildbucket will not create the build and won't
  # allocate a build number, but will return a definition of the swarming task
  # that would be created for the build. Build id will be 1 and build number (if
  # configured) will be 0.
  build_request = messages.MessageField(api.PutRequestMessage, 1, required=True)


class GetTaskDefinitionResponseMessage(messages.Message):
  # A definition of the swarming task that would be created for the specified
  # build.
  task_definition = messages.StringField(1)

  # The swarming host that we would send this task request to.
  swarming_host = messages.StringField(2)


class SetNextBuildNumberRequest(messages.Message):
  bucket = messages.StringField(1, required=True)
  builder = messages.StringField(2, required=True)
  next_number = messages.IntegerField(3, required=True)


@auth.endpoints_api(
    name='swarmbucket', version='v1', title='Buildbucket-Swarming integration'
)
class SwarmbucketApi(remote.Service):
  """API specific to swarmbucket."""

  @swarmbucket_api_method(
      endpoints.ResourceContainer(
          message_types.VoidMessage,
          bucket=messages.StringField(1, repeated=True),
      ),
      GetBuildersResponseMessage,
      path='builders',
      http_method='GET'
  )
  def get_builders(self, request):
    """Returns defined swarmbucket builders.

    Returns legacy bucket names, e.g. "luci.chromium.try", not "chromium/try".

    Can be used to discover builders.
    """
    if len(request.bucket) > 100:
      raise endpoints.BadRequestException(
          'Number of buckets cannot be greater than 100'
      )
    if request.bucket:
      # Buckets were specified explicitly.
      bucket_ids = map(api_common.parse_luci_bucket, request.bucket)
      bucket_ids = [bid for bid in bucket_ids if bid]
      # Filter out inaccessible ones.
      bucket_ids = [
          bid for bid, can in
          utils.async_apply(bucket_ids, user.can_access_bucket_async) if can
      ]
    else:
      # Buckets were not specified explicitly.
      # Use the available ones.
      bucket_ids = (
          user.get_accessible_buckets_async(legacy_mode=False).get_result()
      )
      # bucket_ids is None => all buckets are available.

    res = GetBuildersResponseMessage()
    buckets = (
        config.get_buckets_async(bucket_ids, legacy_mode=False).get_result()
    )
    for bucket_id, cfg in buckets.iteritems():
      if cfg and not cfg.swarming.builders:
        continue

      def to_dims(b):
        return flatten_swarmingcfg.format_dimensions(
            swarmingcfg.read_dimensions(b)
        )

      res.buckets.append(
          BucketMessage(
              name=api_common.format_luci_bucket(bucket_id),
              builders=[
                  BuilderMessage(
                      name=builder.name,
                      category=builder.category,
                      properties_json=json.dumps(
                          flatten_swarmingcfg.read_properties(builder.recipe)
                      ),
                      swarming_dimensions=to_dims(builder)
                  ) for builder in cfg.swarming.builders
              ],
              swarming_hostname=cfg.swarming.hostname,
          )
      )
    return res

  @swarmbucket_api_method(
      GetTaskDefinitionRequestMessage, GetTaskDefinitionResponseMessage
  )
  def get_task_def(self, request):
    """Returns a swarming task definition for a build request."""
    try:
      build_request = api.put_request_message_to_build_request(
          request.build_request
      )
      build_request = build_request.normalize()

      identity = auth.get_current_identity()
      if not user.can_view_build_async(build_request).get_result():
        raise endpoints.ForbiddenException(
            '%s cannot view builds in bucket %s' %
            (identity, build_request.bucket)
        )

      build = build_request.create_build(1, identity, utils.utcnow())
      task_def = swarming.prepare_task_def_async(
          build, build_number=0, fake_build=True
      ).get_result()
      task_def_json = json.dumps(task_def)

      return GetTaskDefinitionResponseMessage(
          task_definition=task_def_json,
          swarming_host=build.swarming_hostname,
      )
    except errors.InvalidInputError as ex:
      raise endpoints.BadRequestException(
          'invalid build request: %s' % ex.message
      )
    except errors.BuilderNotFoundError as ex:
      raise endpoints.NotFoundException(ex.message)

  @swarmbucket_api_method(SetNextBuildNumberRequest, message_types.VoidMessage)
  def set_next_build_number(self, request):
    """Sets the build number that will be used for the next build."""
    if not user.can_set_next_number_async(request.bucket).get_result():
      raise endpoints.ForbiddenException('access denied')
    _, bucket = config.get_bucket(request.bucket)

    if not any(b.name == request.builder for b in bucket.swarming.builders):
      raise endpoints.BadRequestException(
          'builder "%s" not found in bucket "%s"' %
          (request.builder, request.bucket)
      )

    seq_name = sequence.builder_seq_name(request.bucket, request.builder)
    try:
      sequence.set_next(seq_name, request.next_number)
    except ValueError as ex:
      raise endpoints.BadRequestException(str(ex))
    return message_types.VoidMessage()
