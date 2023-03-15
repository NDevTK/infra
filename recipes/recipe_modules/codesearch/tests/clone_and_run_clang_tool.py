# Copyright 2017 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.post_process import (DropExpectation, StatusSuccess,
                                        StepSuccess, StepWarning)

DEPS = [
    'codesearch',
    'recipe_engine/path',
]


def RunSteps(api):
  api.codesearch.set_config('chromium', PROJECT='chromium')
  api.codesearch.clone_clang_tools(api.path['cache'])
  api.codesearch.run_clang_tool(clang_dir=None, run_dirs=[api.path['cache']])


def GenTests(api):
  yield api.test(
      'basic',
      api.post_process(StepSuccess, 'remove previous instance of clang tools'),
      api.post_process(StepSuccess, 'download translation_unit clang tool'),
      api.post_process(StepSuccess, 'run translation_unit clang tool'),
      api.post_process(StatusSuccess),
      api.post_process(DropExpectation),
  )

  yield api.test(
      'run_translation_unit_clang_tool_failed',
      api.step_data('run translation_unit clang tool', retcode=1),
      api.post_process(StepSuccess, 'remove previous instance of clang tools'),
      api.post_process(StepSuccess, 'download translation_unit clang tool'),
      api.post_process(StepWarning, 'run translation_unit clang tool'),
      api.post_process(StatusSuccess),
      api.post_process(DropExpectation),
  )
