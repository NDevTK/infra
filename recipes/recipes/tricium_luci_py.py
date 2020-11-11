# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine import post_process

DEPS = [
    'infra_checkout',
    'depot_tools/bot_update',
    'depot_tools/gclient',
    'depot_tools/gerrit',
    'depot_tools/tryserver',
    'recipe_engine/buildbucket',
    'recipe_engine/json',
    'recipe_engine/path',
    'recipe_engine/platform',
    'recipe_engine/tricium',
]


def RunSteps(api):
  """This recipe runs legacy analyzers for the luci-py repo."""
  assert api.platform.is_linux and api.platform.bits == 64
  # We want line numbers for the file as it is in the CL, not rebased.
  # gerrit_no_rebase_patch_ref prevents rebasing.
  checkout = api.infra_checkout.checkout(
      'luci_py', patch_root='infra/luci', gerrit_no_rebase_patch_ref=True)
  checkout.gclient_runhooks()
  commit_message = api.gerrit.get_change_description(
      'https://%s' % api.tryserver.gerrit_change.host,
      api.tryserver.gerrit_change.change, api.tryserver.gerrit_change.patchset)
  input_dir = api.path['checkout'].join('luci')
  analyzers = [
      api.tricium.analyzers.SPACEY,
      api.tricium.analyzers.SPELLCHECKER,
      # TODO(qyearsley): Enable (custom) pylint.
  ]
  # Get affected files.
  affected_files = [
      f for f in checkout.get_changed_files()
      if api.path.exists(input_dir.join(f))
  ]
  api.tricium.run_legacy(analyzers, input_dir, affected_files, commit_message)


def GenTests(api):

  def test_with_patch(name, affected_files):
    test = api.test(
        name,
        api.platform('linux', 64),
        api.buildbucket.try_build(
            project='infra',
            bucket='try',
            builder='tricium-luci-py',
            git_repo='https://chromium.googlesource.com/infra/luci/luci-py') +
        api.override_step_data(
            'gerrit changes',
            api.json.output([{
                'revisions': {
                    'aaaa': {
                        '_number': 7,
                        'commit': {
                            'author': {
                                'email': 'user@a.com'
                            },
                            'message': 'my commit msg',
                        }
                    }
                }
            }])),
    )
    existing_files = [
        api.path['cache'].join('builder', 'luci', x) for x in affected_files
    ]
    test += api.path.exists(*existing_files)
    return test

  yield test_with_patch('one_file', ['README.md']) + api.post_check(
      post_process.StatusSuccess) + api.post_process(
          post_process.DropExpectation)
