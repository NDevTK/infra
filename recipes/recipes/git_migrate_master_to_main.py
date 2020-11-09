# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.recipe_api import Property

DEPS = [
    'depot_tools/gerrit',
    'recipe_engine/properties',
]

PROPERTIES = {
    'repos': Property(default=[]),
}


def RunSteps(api, repos):
  for host, project in repos:
    api.gerrit.move_changes(
        host,
        project,
        'master',  # from
        'main',  # to
    )


def GenTests(api):
  yield api.test('empty') + api.properties(repos=[])
  yield api.test('basic') + api.properties(
      repos=[('https://chromium.googlesource.com/', 'foo')])
