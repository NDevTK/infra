# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Utils for redirect."""

PROJECT_REDIRECT_MAP = {
    'pigweed': 'https://issues.pigweed.dev/',
    'git': 'https://git.issues.gerritcodereview.com',
}


def GetRedirectURL(project_name):
  return PROJECT_REDIRECT_MAP.get(project_name, None)
