# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json
import logging
import re

from gae_libs.http.http_client_appengine import HttpClientAppengine
from infra_api_clients.codereview import cl_info
from infra_api_clients.codereview import codereview
from libs import time_util


class Gerrit(codereview.CodeReview):
  """Stub for implementing Gerrit support."""
  HTTP_CLIENT = HttpClientAppengine(follow_redirects=False)

  def __init__(self, host, settings=None):
    super(Gerrit, self).__init__(host)
    settings = settings or {}
    self.commit_bot_emails = settings.get('commit_bot_emails',
                                          ['commit-bot@chromium.org'])

  def _HandleResponse(self, status_code, content):
    if status_code != 200:
      return None
    # Remove XSSI magic prefix
    if content.startswith(')]}\''):
      content = content[4:]
    return json.loads(content)

  def _AuthenticatedRequest(self, path_parts, payload=None, method='GET',
                            headers=None):
    # Prepend /a/ to make the request authenticated.
    if path_parts[0] != 'a':
      path_parts = ['a'] + list(path_parts)
    url = 'https://%s/%s' % (self._server_hostname, '/'.join(path_parts))
    headers = headers or {}
    # This header tells gerrit to send compact (non-pretty) JSON which is
    # more efficient and encouraged for automated tools.
    headers['Accept'] = 'application/json'
    headers.setdefault('Accept', 'application/json')
    if method == 'GET':
      return self.HTTP_CLIENT.Get(url, params=payload, headers=headers)
    elif method == 'POST':
      return self.HTTP_CLIENT.Post(url, data=payload, headers=headers)
    raise NotImplementedError()  # pragma: no cover

  def _Get(self, path_parts, params=None, headers=None):
    """Makes a simple get to Gerrit's API and parses the json output."""
    return self._HandleResponse(*self._AuthenticatedRequest(
        path_parts, payload=params, headers=headers))

  def _Post(self, path_parts, body=None, headers=None):
    headers = headers or {}
    if body:  # pragma: no branch
      headers['Content-Type'] = 'application/json'
      body = json.dumps(body)
    return self._HandleResponse(*self._AuthenticatedRequest(
        path_parts, payload=body, method='POST', headers=headers))

  def GetCodeReviewUrl(self, change_id):
    return 'https://%s/q/%s' % (self._server_hostname, change_id)

  def PostMessage(self, change_id, message):
    parts = ['changes', change_id, 'revisions', 'current', 'review']
    result = self._Post(parts, body={'message': message})
    return result is not None  # A successful post will return an empty dict.

  def CreateRevert(self, reason, change_id, patchset_id=None):
    parts = ['changes', change_id, 'revert']
    reverting_change = self._Post(parts, body={'message': reason})
    try:
      return reverting_change['change_id']
    except (TypeError, KeyError):
      return None

  def AddReviewers(self, change_id, reviewers, message=None):
    current_reviewers = self.GetClDetails(change_id).reviewers
    try:
      for reviewer in reviewers:
        # reviewer must be an email string.
        assert len(reviewer.split('@')) == 2
        if reviewer in current_reviewers:
          # Only add reviewers not currently assinged to the change.
          continue
        parts =['changes', change_id, 'reviewers']
        response = self._Post(parts, body={'reviewer': reviewer})
        reviewers = response['reviewers']
        if reviewers == []:
          # This might be okay if a user has more than one email.
          logging.warning('Reviewer %s already assigned to cl %s under a '
                          'different email' % (reviewer, change_id))
          continue
        new_reviewer = reviewers[0]['email']
        if new_reviewer != reviewer:
          # This might be okay if a user has more than one email.
          logging.warning('Requested to add %s as reviewer to cl %s but '
                          '%s was added instead.' % (reviewer, change_id,
                                                     new_reviewer))
    except (TypeError, KeyError, IndexError):
      return False
    finally:
      # Post the message even if failed to add reviewers.
      success = not message or self.PostMessage(change_id, message)

    return success

  def GetClDetails(self, change_id):
    # Create cl info based on the url.
    params = [('o', 'CURRENT_REVISION'), ('o', 'CURRENT_COMMIT')]
    change_info = self._Get(['changes', change_id, 'detail'], params=params)
    return self._ParseClInfo(change_info, change_id)

  def _ParseClInfo(self, change_info, change_id):
    if not change_info:  # pragma: no cover
      return None
    result = cl_info.ClInfo(self._server_hostname, change_id)

    result.reviewers = [x['email'] for x in change_info['reviewers'].get(
      'REVIEWER', [])]
    result.cc = [x['email'] for x in change_info['reviewers'].get('CC', [])]
    result.closed = change_info['status'] == 'MERGED'
    result.owner_email = change_info['owner']['email']
    result.subject = change_info['subject']

    # If the status is merged, look at the commit details for the current
    # commit.
    if result.closed:  # pragma: no branch
      current_revision = change_info['current_revision']
      revision_info = change_info['revisions'][current_revision]
      patchset_id = revision_info['_number']
      commit_timestamp = time_util.DatetimeFromString(
          change_info['submitted'])
      result.commits.append(cl_info.Commit(patchset_id, current_revision,
                                           commit_timestamp))
      revision_commit = revision_info['commit']
      # Detect manual commits.
      committer = revision_commit['committer']['email']
      if committer not in self.commit_bot_emails:
        result.AddCqAttempt(patchset_id, committer, commit_timestamp)

      result.description = revision_commit['message']

      # Checks for if the culprit owner has turned off auto revert.
      result.auto_revert_off = codereview.IsAutoRevertOff(result.description)

    # TO FIND COMMIT ATTEMPTS:
    # In messages look for "Patch Set 1: Commit-Queue+2"
    # or "Patch Set 4: Code-Review+1 Commit-Queue+2".
    cq_pattern = re.compile('^Patch Set \d+:( Code-Review..)? Commit-Queue\+2$')
    revert_tag = 'autogenerated:gerrit:revert'
    revert_pattern = re.compile(
        'Created a revert of this change as (?P<change_id>I[a-f\d]{40})')

    for message in change_info['messages']:
      if cq_pattern.match(message['message'].splitlines()[0]):
        patchset_id = message['_revision_number']
        author = message['author']['email']
        timestamp = time_util.DatetimeFromString(message['date'])
        result.AddCqAttempt(patchset_id, author, timestamp)

      # TO FIND REVERT(S):
      if message.get('tag') == revert_tag:
        patchset_id = message['_revision_number']
        author = message['author']['email']
        timestamp = time_util.DatetimeFromString(message['date'])
        reverting_change_id = revert_pattern.match(
          message['message']).group('change_id')
        reverting_cl = self.GetClDetails(reverting_change_id)
        result.reverts.append(cl_info.Revert(patchset_id, reverting_cl, author,
                                             timestamp))
    return result
