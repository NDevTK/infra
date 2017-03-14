# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.


class CodeReview(object):  # pragma: no cover.
  """Abstract class to interact with code review."""

  def __init__(self, server_hostname):
    """
    Args:
      server_hostname (str): The hostname of the codereview server, eg:
          codereview.chromium.org or chromium-review.googlesource.com.
    """
    self._server_hostname = server_hostname

  def PostMessage(self, change_id, message):
    """Posts the given message to the CL codereview of the given change id.

    Args:
      change_id (str or int): The change id of the CL on Gerrit or the issue
          number of the CL on Rietveld.
      message(str): The message to be posted to the codereview.
    """
    raise NotImplementedError()

  def AddReviewers(self, codereview_url, reviewers, message=None):
    """Adds a list of users to the CL of the specified url as reviewers.

    Args:
      codereview_url(str): The url to a CL codereview.
      reviewers (list of str): The users to be added as reviewrs
      message(str): (optional) The message to be posted to the codereview.

    Returns:
      A boolean indicating success.
    """
    raise NotImplementedError()
