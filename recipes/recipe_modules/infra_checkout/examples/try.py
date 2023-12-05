# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.post_process import DropExpectation

from PB.go.chromium.org.luci.buildbucket.proto.common import GerritChange

PYTHON_VERSION_COMPATIBILITY = 'PY2+3'

DEPS = [
    'infra_checkout',
    'recipe_engine/buildbucket',
    'recipe_engine/file',
    'recipe_engine/platform',
    'recipe_engine/raw_io',
    'depot_tools/gerrit',
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
      api.infra_checkout.apply_golangci_lint(co, 'go/src/infra/')

  change = GerritChange(
      host='host', project='infra/infra', change=1234, patchset=5)
  overrides = api.infra_checkout.get_footer_infra_deps_overrides(
      change,
      step_test_data='chickens\ntry-infra: 987f\n'
      'try-infra_internal: 123abc\n'
      'try-.: HEAD\n'
      'try-third_party: lkj123')
  assert overrides == {'infra': '987f', 'infra_internal': '123abc', '.': 'HEAD'}


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
         api.step_data('read go/src/infra/.go-lintable',
                       api.file.read_text('[section]\npaths = .\n')) +
         api.post_process(DropExpectation))
