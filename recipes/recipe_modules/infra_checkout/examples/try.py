# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.post_process import DropExpectation

PYTHON_VERSION_COMPATIBILITY = 'PY2+3'

DEPS = [
    'infra_checkout',
    'recipe_engine/buildbucket',
    'recipe_engine/platform',
    'recipe_engine/raw_io',
]

def RunSteps(api):
  co = api.infra_checkout.checkout(
      gclient_config_name='infra', patch_root='infra')
  co.commit_change()
  co.get_changed_files()
  co.get_changed_files(diff_filter='A')
  if api.platform.is_linux:
    with co.go_env():
      co.run_presubmit()
      api.infra_checkout.apply_golangci_lint(co)


def GenTests(api):
  def diff(*files):
    return api.step_data('get change list',
                         api.raw_io.stream_output_text('\n'.join(files)))

  for plat in ('linux', 'mac', 'win'):
    yield (
        api.test(plat) +
        api.platform(plat, 64) +
        api.buildbucket.try_build(
            project='infra',
            bucket='try',
            builder='presubmit',
            git_repo='https://chromium.googlesource.com/infra/infra',
        ) +
        # Simulate too many files on Mac.
        diff(*['file_%d' % i for i in range(1000 if plat == 'mac' else 2)])
    )

  yield (api.test('golangci-lint') + api.buildbucket.try_build(
      project='infra',
      bucket='try',
      builder='presubmit',
      git_repo='https://chromium.googlesource.com/infra/infra',
  ) + api.step_data('get change list (3)',
                    api.raw_io.stream_output_text('go/src/infra/main.go')) +
         api.post_process(DropExpectation))
