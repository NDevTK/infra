# Copyright 2017 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from PB.recipe_modules.infra.codesearch.tests.properties import CloneAndRunClangToolProps
from recipe_engine.post_process import (DropExpectation, StatusSuccess,
                                        StepCommandContains,
                                        StepCommandDoesNotContain, StepSuccess,
                                        StepWarning)

DEPS = [
    'codesearch',
    'recipe_engine/path',
    'recipe_engine/properties',
]

PROPERTIES = CloneAndRunClangToolProps


def RunSteps(api, properties):
  api.path.checkout_dir = api.path.cache_dir.join('builder', 'src')
  api.codesearch.set_config('chromium', PROJECT='chromium')
  api.codesearch.clone_clang_tools(api.path.cache_dir)

  target_architecture: Optional[str] = None
  if properties.HasField("target_architecture"):
    target_architecture = properties.target_architecture
  api.codesearch.run_clang_tool(
      clang_dir=None,
      run_dirs=[api.path.cache_dir],
      target_architecture=target_architecture)


def GenTests(api):
  yield api.test(
      'basic',
      api.post_process(StepSuccess, 'remove previous instance of clang tools'),
      api.post_process(StepSuccess, 'download translation_unit clang tool'),
      api.post_process(StepSuccess, 'run translation_unit clang tool'),
      api.post_process(StepCommandDoesNotContain,
                       'run translation_unit clang tool', ['-arch']),
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

  yield api.test(
      'target_architecture_arm64',
      api.properties(target_architecture='arm64'),
      api.post_process(StepCommandContains, 'run translation_unit clang tool',
                       ['--tool-arg=--extra-arg=--target=arm64']),
      api.post_process(StatusSuccess),
      api.post_process(DropExpectation),
  )
