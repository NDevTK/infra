# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import traceback


def process_page(request):
  """Fetch, process and upload Rietveld pages.

  Defines a cloud function to fetch pages from an existing Rietveld instance,
  process it to remove dynamic content, and upload the resulting page to Google
  Storage.

  Args:
    request: A dict containing entries for 'path', 'type' and 'private'.
      path: The page to fetch, e.g. '/1234/patchset/5'
      type: One of 'Issue', 'PatchSet' or 'Patch'.
      private: Whether this page is from a private Rietveld issue.
  """
  params = request.get_json(force=True, silent=True)
  path = params['path']
  page_type = params['type']
  private = params['private']

  pages.process_page(path, page_type, private)

  return ''

