# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

PYTHON_VERSION_COMPATIBILITY = 'PY2+3'

DEPS = [
    'recipe_engine/path',
    'recipe_engine/step',
    'buildenv',
]


def RunSteps(api):
  with api.buildenv(api.path['start_dir'], 'GO_VERSION', 'NODEJS_VERSION'):
    api.step('full env', ['echo', 'hi'])

  with api.buildenv(api.path['start_dir']):
    api.step('empty env', ['echo', 'hi'])


def GenTests(api):
  yield api.test('full')
