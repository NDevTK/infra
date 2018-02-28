# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import datetime
import json
import logging

from google.appengine.api import app_identity
from google.appengine.api import taskqueue
from google.appengine.ext import ndb
import webapp2

from components import auth
from components import decorators
from components import net
from components import pubsub
from components import utils
import bqh

import api_common
import model
import metrics
import v2

_TASK_TAG_PROD = 'prod'
_TASK_TAG_EXPERIMENTAL = 'experimental'


# Mocked in tests.
@ndb.tasklet
def enqueue_pull_task_async(queue, tag, payload):  # pragma: no cover
  # Cannot just return the return value of add_async because it is
  # a non-Future object and does not play nice with `yield fut1, fut2` construct
  yield taskqueue.Task(
      payload=payload,
      tag=tag,
      method='PULL').add_async(queue_name=queue, transactional=True)


def enqueue_bq_export_async(build):  # pragma: no cover
  """Enqueues a pull task to export a completed build to BigQuery."""
  assert ndb.in_transaction()
  assert build
  assert build.status == model.BuildStatus.COMPLETED

  return enqueue_pull_task_async(
      'bq-export',
      _task_tag(build.experimental),
      json.dumps({'id': build.key.id()}))


class CronExportBuilds(webapp2.RequestHandler):  # pragma: no cover
  """Exports builds to a BigQuery table.

  Processes "bq-export" pull queue.
  """

  @decorators.require_cronjob
  def get(self):
    request_deadline = utils.utcnow() + datetime.timedelta(minutes=5)
    experimental = False
    while utils.utcnow() < request_deadline:
      _process_pull_task_batch(experimental)
      experimental = not experimental


def _process_pull_task_batch(experimental):
  # Lease tasks.
  lease_duration = datetime.timedelta(minutes=5)
  lease_deadline = utils.utcnow() + lease_duration
  q = taskqueue.Queue('bq-export')
  tasks = q.lease_tasks_by_tag(
      lease_duration.
      total_seconds(), 1000, tag=_task_tag(experimental))
  if not tasks:
    return

  # Fetch builds for the tasks.
  build_ids = [json.loads(t.payload)['id'] for t in tasks]
  builds = ndb.get_multi([ndb.Key(model.Build, id) for id in build_ids])
  builds_to_export = []
  for bid, b in zip(build_ids, builds):
    if not b:
      logging.error('build %d not found', bid)
    elif b.status != model.BuildStatus.COMPLETED:
      logging.error('build %d is not complete', bid)
    else:
      builds_to_export.append(b)

  row_count = 0
  if builds_to_export:
    logging.debug(
        'processing builds %r', [b.key.id() for b in builds_to_export])
    dataset = 'builds' + ('_experimental' if experimental else '')
    row_count = _export_builds(dataset, builds_to_export, lease_deadline)

  q.delete_tasks(tasks)
  logging.info('inserted %d rows, processed %d tasks', row_count, len(tasks))


def _export_builds(dataset, builds, deadline):
  """Exports builds to BigQuery."""
  # BigQuery API doc:
  # https://cloud.google.com/bigquery/docs/reference/rest/v2/tabledata/insertAll
  rows = []
  for b in builds:
    try:
      rows.append({
        'insertId': str(b.key.id()),
        'json': bqh.message_to_dict(v2.build_to_v2(b)),
      })
    except v2.UnsupportedBuild as ex:
      logging.warning(
          'skipping build %d is not supported by BQ export: %s.',
          b.key.id(), ex)

  logging.info('sending %d rows', len(rows))
  if not rows:
    return 0
  res = net.json_request(
      url=(
        ('https://www.googleapis.com/bigquery/v2/'
          'projects/%s/datasets/%s/tables/completed_beta/insertAll') % (
          app_identity.get_application_id(), dataset)
        ),
      method='POST',
      payload={
        'kind': 'bigquery#tableDataInsertAllRequest',
        'skipInvalidRows': False,
        'ignoreUnknownValues': False,
        'rows': rows,
      },
      scopes=[
        'https://www.googleapis.com/auth/bigquery.insertdata',
        'https://www.googleapis.com/auth/bigquery',
        'https://www.googleapis.com/auth/cloud-platform',
      ],
      # deadline parameter here is duration in seconds.
      deadline=(deadline - utils.utcnow()).total_seconds(),
  )
  return len(rows)


def _task_tag(experimental):
  return 'experimental' if experimental else 'prod'
