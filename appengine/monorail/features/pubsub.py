# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.


"""Task handlers for publishing pubsub notifications of issue changes.
Pubsub event notifications are sent when an issue changes, an issue that is blocking
another issue changes, or a bulk edit is done.  
The topic is projects/monorail/issue-changes"""


from tracker import tracker_views

from googleapiclient.discovery import build

class PublishPubsubIssueChangeTask(notify_helpers.NotifyTaskBase):

  def HandleRequest(self, mr):
    """Process the task to notify users after an issue change.

    Args:
      mr: common information parsed from the HTTP request.

    Returns:
      Nothing.
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
        mr.cnxn, project, issue, ps_topic, ps_project, config, mr.perms)

    return {
        'params': params,
        }

  def _PublishPubsub(
      self, cnxn, project, issue, ps_topic, ps_project, config, perms):

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

