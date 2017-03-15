# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import logging
import textwrap

from google.appengine.ext import ndb

from common import rotations
from common.pipeline_wrapper import BasePipeline
from infra_api_clients.codereview import codereview_util
from libs import time_util
from model.base_suspected_cl import RevertCL
from model.wf_suspected_cl import WfSuspectedCL
from waterfall import suspected_cl_util


CREATED_BY_FINDIT = 0
CREATED_BY_SHERIFF = 1
ERROR = 2


def _GetLastCommitAndLastRevert(commits, reverts):
  last_commit = sorted(
      commits, key=lambda k: k.timestamp)[-1] if commits else None
  last_revert = sorted(
      reverts, key=lambda k: k.timestamp)[-1] if reverts else None
  return last_commit, last_revert


@ndb.transactional
def _RevertCulprit(repo_name, revision):

  culprit = WfSuspectedCL.Get(repo_name, revision)
  assert culprit

  if culprit.revert_cl:
    # Revert CL is aready created by Findit.
    return CREATED_BY_FINDIT

  culprit.can_be_reverted = True

  # 0. Gets information about this culprit.
  culprit_change_log = (
    suspected_cl_util.GetCulpritChangeLog(repo_name, revision))
  culprit_commit_position = culprit_change_log.commit_position
  culprit_code_review_url = culprit_change_log.code_review_url

  codereview = codereview_util.GetCodeReviewForReview(culprit_code_review_url)
  culprit_change_id = codereview_util.GetChangeIdForReview(
    culprit_code_review_url)

  culprit_cl_info = codereview.GetClDetails(
    culprit_change_id) if codereview and culprit_change_id else None
  if not culprit_cl_info:
    culprit.put()
    logging.error('Failed to get cl_info for %s/%s' % (repo_name, revision))
    return ERROR

  # 1. Checks if a revert CL by sheriff has been created.
  last_commit, last_revert = _GetLastCommitAndLastRevert(
    culprit_cl_info.commits, culprit_cl_info.reverts)

  if not last_commit:
    culprit.put()
    logging.error('Culprit isn"t committed for %s/%s' % (repo_name, revision))
    return ERROR

  if last_revert and last_revert.timestamp > last_commit.timestamp:
    culprit.put()
    return CREATED_BY_SHERIFF

  # 2. Reverts the culprit.
  revert_reason = textwrap.dedent("""
      FYI: Findit identified CL at revision %s as the culprit for
      failures in the build cycles as shown on:
      https://findit-for-me.appspot.com/waterfall/culprit?key=%s""") % (
          culprit_commit_position or revision, culprit.key.urlsafe())

  revert_change_id = codereview.CreateRevert(
    revert_reason, culprit_change_id, last_commit.patchset_id)
  if not revert_change_id:
    logging.error('Revert for culprit %s/%s failed.' % (repo_name, revision))
    culprit.put()
    return ERROR

  # Save revert CL info to culprit
  revert_cl = RevertCL()
  revert_cl.revert_cl_url = codereview.GetCodeReviewUrl(revert_change_id)
  revert_cl.created_time = time_util.GetUTCNow()
  culprit.revert_cl = revert_cl
  culprit.put()

  # 3. Add reviewers.
  reviewers = [culprit_change_log.author.email]
  sheriffs = rotations.current_sheriffs()
  reviewers.extend(sheriffs)

  success = codereview.AddReviewers(revert_change_id, reviewers)

  if not success:
    logging.error('Failed to add reviewers for revert of culprit %s/%s' % (
      repo_name, revision))
    return ERROR


class CreateRevertCLPipeline(BasePipeline):

  # Arguments number differs from overridden method - pylint: disable=W0221
  def run(self, repo_name, revision):
    return _RevertCulprit(repo_name, revision)

