# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.post_process import DropExpectation, StatusFailure
from recipe_engine.recipe_api import Property

DEPS = ['qemu', 'recipe_engine/path']

PYTHON_VERSION_COMPATIBILITY = 'PY3'


def RunSteps(api):
  api.qemu.init('latest')
  # fail because fat64 doesn't exist (or does it?)
  api.qemu.create_disk('fat_disk', 'fat64', 2048)


def GenTests(api):
  yield (api.test('Test create_disk fail') + api.post_process(StatusFailure) +
         api.post_process(DropExpectation))
