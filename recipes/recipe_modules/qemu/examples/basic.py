# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.post_process import DropExpectation, StatusSuccess

DEPS = ['qemu', 'recipe_engine/path']

PYTHON_VERSION_COMPATIBILITY = 'PY3'


def RunSteps(api):
  api.qemu.init('latest')


def GenTests(api):
  yield (api.test('Test qemu init') + api.post_process(StatusSuccess) +
         api.post_process(DropExpectation))
