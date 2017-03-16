# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import logging
import textwrap

from google.appengine.ext import ndb

from common import constants
from common import rotations
from common.pipeline_wrapper import BasePipeline
from infra_api_clients.codereview import codereview_util
from libs import time_util
from model import analysis_status as status
from model.base_suspected_cl import RevertCL
from model.wf_suspected_cl import WfSuspectedCL
from waterfall import suspected_cl_util


CREATED_BY_FINDIT = 0
CREATED_BY_SHERIFF = 1
ERROR = 2


def _GetCulpritPatchsetId(revision, commits):
  for commit in commits:
    if commit.revision == revision:  # pragma: no branch
      return commit.patchset_id
  return None  # pragma: no cover


def _GetsExistingRevertCL(revision, commits, reverts):
  culprit_patchset_id = _GetCulpritPatchsetId(revision, commits)
  if not culprit_patchset_id:  # pragma: no cover
    return None

  reverts_for_culprit = []
  for revert in reverts:
    if revert.patchset_id == culprit_patchset_id:  # pragma: no branch
      reverts_for_culprit.append(revert)
  return reverts_for_culprit


@ndb.transactional
def _RevertCulprit(repo_name, revision):

  culprit = WfSuspectedCL.Get(repo_name, revision)
  assert culprit

  if culprit.revert_cl and culprit.revert_status == status.COMPLETED:
    return CREATED_BY_FINDIT

  culprit.should_be_reverted = True

  # 0. Gets information about this culprit.
  culprit_commit_position, culprit_code_review_url = (
      suspected_cl_util.GetCulpritInfo(repo_name, revision))

  codereview = codereview_util.GetCodeReviewForReview(culprit_code_review_url)
  culprit_change_id = codereview_util.GetChangeIdForReview(
    culprit_code_review_url)

  if not codereview or not culprit_change_id:  # pragma: no cover
    culprit.put()
    logging.error('Failed to get change id for %s/%s' % (repo_name, revision))
    return ERROR

  culprit_cl_info = codereview.GetClDetails(
    culprit_change_id) if codereview and culprit_change_id else None
  if not culprit_cl_info:  # pragma: no cover
    culprit.put()
    logging.error('Failed to get cl_info for %s/%s' % (repo_name, revision))
    return ERROR

  # 1. Checks if a revert CL by sheriff has been created.
  reverts = _GetsExistingRevertCL(
      revision, culprit_cl_info.commits, culprit_cl_info.reverts)

  if reverts is None:  # pragma: no cover
    # if no reverts, reverts should be [], only when some error happens it will
    # be None.
    culprit.put()
    logging.error('Failed to find patchset_id for %s/%s' % (
        repo_name, revision))
    return ERROR

  findit_revert = None
  for revert in reverts:
    if revert.reverting_user_email == constants.DEFAULT_SERVICE_ACCOUNT:
      findit_revert = revert
      break

  if reverts and not findit_revert:
    # Sheriff(s) created the revert CL(s).
    culprit.put()
    return CREATED_BY_SHERIFF

  # 2. Reverts the culprit.
  culprit.revert_status = culprit.revert_status or status.RUNNING

  revert_change_id = codereview_util.GetChangeIdForReview(
      findit_revert.reverting_cl.url) if findit_revert else None

  if not findit_revert:
    revert_reason = textwrap.dedent("""
        Findit identified CL at revision %s as the culprit for
        failures in the build cycles as shown on:
        https://findit-for-me.appspot.com/waterfall/culprit?key=%s""") % (
            culprit_commit_position or revision, culprit.key.urlsafe())

    revert_change_id = codereview.CreateRevert(
      revert_reason, culprit_change_id, _GetCulpritPatchsetId(
          revision, culprit_cl_info.commits))
    if not revert_change_id:  # pragma: no cover
      logging.error('Revert for culprit %s/%s failed.' % (repo_name, revision))
      culprit.put()
      return ERROR


  # Save revert CL info and notification info to culprit.
  if not culprit.revert_cl:
    revert_cl = RevertCL()
    revert_cl.revert_cl_url = codereview.GetCodeReviewUrl(revert_change_id)
    revert_cl.created_time = time_util.GetUTCNow()
    culprit.revert_cl = revert_cl
    culprit.cr_notification_time = time_util.GetUTCNow()
    culprit.cr_notification_status = status.COMPLETED
    culprit.put()

  # 3. Add reviewers.
  reviewers = rotations.current_sheriffs()
  success = codereview.AddReviewers(revert_change_id, reviewers)

  if not success:  # pragma: no cover
    logging.error('Failed to add reviewers for revert of culprit %s/%s' % (
      repo_name, revision))
    return ERROR

  culprit.revert_status = status.COMPLETED
  culprit.put()
  return CREATED_BY_FINDIT


class CreateRevertCLPipeline(BasePipeline):

  # Arguments number differs from overridden method - pylint: disable=W0221
  def run(self, repo_name, revision):
    return _RevertCulprit(repo_name, revision)