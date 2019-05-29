# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""Task handlers for publishing pubsub notifications of issue changes.
Pubsub event notifications are sent when an issue changes, an issue that is blocking
another issue changes, or a bulk edit is done.  
The topic is projects/monorail/issue-changes"""


import collections
import json
import logging
import os

from third_party import ezt

from google.appengine.api import taskqueue
from google.appengine.api import urlfetch
from google.appengine.runtime import apiproxy_errors

import settings
from features import autolink
from features import notify_helpers
from features import notify_reasons
from framework import authdata
from framework import emailfmt
from framework import exceptions
from framework import framework_bizobj
from framework import framework_helpers
from framework import framework_views
from framework import jsonfeed
from framework import monorailrequest
from framework import permissions
from framework import template_helpers
from framework import urls
from tracker import tracker_bizobj
from tracker import tracker_helpers
from tracker import tracker_views
from proto import tracker_pb2

from googleapiclient.discovery import build

class PublishPubsubIssueChangeTask(notify_helpers.NotifyTaskBase):
  """JSON servlet that notifies appropriate users after an issue change."""

  def HandleRequest(self, mr):
    """Process the task to notify users after an issue change.

    Args:
      mr: common information parsed from the HTTP request.

    Returns:
      Results dictionary in JSON format which is useful just for debugging.
      The main goal is the side-effect of sending emails.
    """
    issue_id = mr.GetPositiveIntParam('issue_id')
    if not issue_id:
      return {
          'params': {},
          'notified': [],
          'message': 'Cannot proceed without a valid issue ID.',
      }
    commenter_id = mr.GetPositiveIntParam('commenter_id')
    seq_num = mr.seq
    omit_ids = [commenter_id]
    old_owner_id = mr.GetPositiveIntParam('old_owner_id')
    comment_id = mr.GetPositiveIntParam('comment_id')
    params = dict(
        issue_id=issue_id, commenter_id=commenter_id,
        seq_num=seq_num, old_owner_id=old_owner_id,
        omit_ids=omit_ids, comment_id=comment_id)

    logging.info('issue change params are %r', params)
    # TODO(jrobbins): Re-enable the issue cache for notifications after
    # the stale issue defect (monorail:2514) is 100% resolved.
    issue = self.services.issue.GetIssue(mr.cnxn, issue_id, use_cache=False)
    project = self.services.project.GetProject(mr.cnxn, issue.project_id)
    config = self.services.config.GetProjectConfig(mr.cnxn, issue.project_id)


    all_comments = self.services.issue.GetCommentsForIssue(
        mr.cnxn, issue.issue_id)
    if comment_id:
      logging.info('Looking up comment by comment_id')
      for c in all_comments:
        if c.id == comment_id:
          comment = c
          logging.info('Comment was found by comment_id')
          break
      else:
        raise ValueError('Comment %r was not found' % comment_id)
    else:
      logging.info('Looking up comment by seq_num')
      comment = all_comments[seq_num]

    # Make followup tasks to send pubsub notification
    tasks = []
    tasks = self._PublishPubsub(
        mr.cnxn, project, issue, config, ps_topic, ps_project, mr.perms)

    return {
        'params': params,
        }

  def _PublishPubsub(
      self, cnxn, project, issue, ps_topic, ps_project, config):

    pubsub_data = {
        # Pass open_related and closed_related into this method and to
        # the issue view so that we can show it on new issue email.
        'issue': tracker_views.IssueView(issue, users_by_id, config)
        }
    pubsub_token = "123456"
    pubsub_topic = ps_topic

    service = build('pubsub', 'v1')
    topic_path = 'projects/{project_id}/topics/{topic}'.format(
        project_id=ps_project,
        topic=pubsub_topic
    )
    #Timestamp var is based on PubsubMessage Timestamp documentation:
    #https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
    service.projects().topics().publish(
        topic=topic_path, body={
          "messages": [{"issue_id":issue.local_id,
          "project_id":topic_path.project_id,}]
          #TODO: Add timestamp variable of when issue was updated.
          #"timestamp": publishTime}]
        }).execute()