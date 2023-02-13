# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import collections
import contextlib

from recipe_engine import recipe_api

class InfraCheckoutApi(recipe_api.RecipeApi):
  """Stateless API for using public infra gclient checkout."""

  def checkout(self, gclient_config_name,
               patch_root=None,
               path=None,
               internal=False,
               generate_env_with_system_python=False,
               go_version_variant=None,
               **kwargs):
    """Fetches infra gclient checkout into a given path OR named_cache.

    Arguments:
      * gclient_config_name (string) - name of gclient config.
      * patch_root (path or string) - path **inside** infra checkout to git repo
        in which to apply the patch. For example, 'infra/luci' for luci-py repo.
        If None (default), no patches will be applied.
      * path (path or string) - path to where to create/update infra checkout.
        If None (default) - path is cache with customizable name (see below).
      * internal (bool) - by default, False, meaning infra gclient checkout
          layout is assumed, else infra_internal.
          This has an effect on named_cache default and inside which repo's
          go corner the ./go/env.py command is run.
      * generate_env_with_system_python uses the bot "infra_system" python to
        generate infra.git's ENV. This is needed for bots which build the
        "infra/infra_python/${platform}" CIPD packages because they incorporate
        the checkout's VirtualEnv inside the package. This, in turn, results in
        the CIPD package containing absolute paths to the Python that was used
        to create it. In order to enable this madness to work, we ensure that
        the Python is a system Python, which resides at a fixed path. No effect
        on arm64 because the arm64 bots have no such python available.
      * go_version_variant can be set go "legacy" or "bleeding_edge" to force
        the builder to use a non-default Go version. What exact Go versions
        correspond to "legacy" and "bleeding_edge" and default is defined in
        bootstrap.py in infra.git.
      * kwargs - passed as is to bot_update.ensure_checkout.

    Returns:
      a Checkout object with commands for common actions on infra checkout.
    """
    assert gclient_config_name, gclient_config_name

    with self.m.context(cwd=self.m.path['start_dir']):
      if self.m.platform.is_win:
        # Need to enable support for symlinks on Windows for unittest in luci-py
        self.m.git(
            'config', '--global', 'core.symlinks', 'true', name='set symlinks')

    path = path or self.m.path['cache'].join('builder')
    self.m.file.ensure_directory('ensure builder dir', path)

    # arm64 bots don't have this system python stuff
    if generate_env_with_system_python and (
        self.m.platform.arch == 'arm' and self.m.platform.bits == 64):
      generate_env_with_system_python = False

    with self.m.context(cwd=path):
      self.m.gclient.set_config(gclient_config_name)
      if generate_env_with_system_python:
        sys_py = self.m.path.join(self.m.infra_system.sys_bin_path, 'python')
        if self.m.platform.is_win:
          sys_py += '.exe'
        self.m.gclient.c.solutions[0].custom_vars['infra_env_python'] = sys_py

      bot_update_step = self.m.bot_update.ensure_checkout(
          patch_root=patch_root, **kwargs)

    env_with_override = {
        'INFRA_GO_SKIP_TOOLS_INSTALL': '1',
        'GOFLAGS': '-mod=readonly',
    }
    if go_version_variant:
      env_with_override['INFRA_GO_VERSION_VARIANT'] = go_version_variant

    class Checkout(object):
      def __init__(self, m):
        self.m = m
        self._go_env = None
        self._go_env_prefixes = None
        self._go_env_suffixes = None
        self._committed = False

      @property
      def path(self):
        return path

      @property
      def bot_update_step(self):
        return bot_update_step

      @property
      def patch_root_path(self):
        assert patch_root
        return path.join(patch_root)

      def commit_change(self):
        assert patch_root
        with self.m.context(cwd=path.join(patch_root)):
          self.m.git(
              '-c', 'user.email=commit-bot@chromium.org',
              '-c', 'user.name=The Commit Bot',
              'commit', '-a', '-m', 'Committed patch',
              name='commit git patch')
        self._committed = True

      def get_changed_files(self, diff_filter=None):
        """Lists files changed in the patch.

        This assumes that commit_change() has been called.

        Returns:
          A set of relative paths (strings) of changed files,
          including added, modified and deleted file paths.
        """
        assert patch_root
        assert self._committed
        # Grab a list of changed files.
        git_args = ['diff', '--name-only', 'HEAD~', 'HEAD']
        if diff_filter:
          git_args.extend(['--diff-filter', diff_filter])
        with self.m.context(cwd=path.join(patch_root)):
          result = self.m.git(
              *git_args,
              name='get change list',
              stdout=self.m.raw_io.output_text())
        files = result.stdout.splitlines()
        if len(files) < 50:
          result.presentation.logs['change list'] = sorted(files)
        else:
          result.presentation.logs['change list is too long'] = (
              '%d files' % len(files))
        return set(files)

      @staticmethod
      def gclient_runhooks():
        with self.m.context(cwd=path, env=env_with_override):
          self.m.gclient.runhooks()

      @contextlib.contextmanager
      def go_env(self):
        name = 'infra_internal' if internal else 'infra'
        self._ensure_go_env()
        with self.m.context(
            cwd=self.path.join(name, 'go', 'src', name),
            env=self._go_env,
            env_prefixes=self._go_env_prefixes,
            env_suffixes=self._go_env_suffixes):
          yield

      def _ensure_go_env(self):
        if self._go_env is not None:
          return  # already did this

        with self.m.context(cwd=self.path, env=env_with_override):
          where = 'infra_internal' if internal else 'infra'
          bootstrap = 'bootstrap_internal.py' if internal else 'bootstrap.py'
          step = self.m.step(
              'init infra go env',
              ['python3', path.join(where, 'go', bootstrap), self.m.json.output()],
              infra_step=True,
              step_test_data=lambda: self.m.json.test_api.output({
                  'go_version': '1.66.6',
                  'env': {'GOROOT': str(path.join('golang', 'go'))},
                  'env_prefixes': {
                      'PATH': [str(path.join('golang', 'go'))],
                  },
                  'env_suffixes': {
                      'PATH': [str(path.join(where, 'go', 'bin'))],
                  },
              }))

        out = step.json.output
        step.presentation.step_text += 'Using go %s' % (out.get('go_version'),)

        self._go_env = env_with_override.copy()
        self._go_env.update(out['env'])
        self._go_env_prefixes = out['env_prefixes']
        self._go_env_suffixes = out['env_suffixes']

      @staticmethod
      def run_presubmit():
        assert patch_root
        revs = self.m.bot_update.get_project_revision_properties(patch_root)
        upstream = bot_update_step.json.output['properties'].get(revs[0])
        gerrit_change = self.m.buildbucket.build.input.gerrit_changes[0]
        with self.m.context(env={'PRESUBMIT_BUILDER': '1'}):
          return self.m.step(
              'presubmit',
              ['vpython3', self.m.presubmit.presubmit_support_path,
                  '--root', path.join(patch_root),
                  '--commit',
                  '--verbose', '--verbose',
                  '--issue', gerrit_change.change,
                  '--patchset', gerrit_change.patchset,
                  '--gerrit_url', 'https://' + gerrit_change.host,
                  '--gerrit_fetch',
                  '--upstream', upstream,
                  '--skip_canned', 'CheckTreeIsOpen',
              ])

    return Checkout(self.m)

  def apply_golangci_lint(self, co, path_to_go_modules=''):
    """Apply goalngci-lint to existing diffs and emit lint warnings via tricium.
    """
    go_files = sorted(
        set([
            (self.m.path.dirname(f[len(path_to_go_modules):]) or '.') + '/...'
            # Set --diff-filter to exclude deleted/renamed files.
            # https://git-scm.com/docs/git-diff#Documentation/git-diff.txt---diff-filterACDMRTUXB82308203
            for f in co.get_changed_files(diff_filter='ACMTR')
            if f.endswith('.go') and f.startswith(path_to_go_modules)
        ]))

    if not go_files:
      return  # pragma: no cover

    # https://chrome-infra-packages.appspot.com/p/infra/3pp/tools/golangci-lint
    linter = self.m.cipd.ensure_tool(
        'infra/3pp/tools/golangci-lint/${platform}', 'version:2@1.50.0')
    result = self.m.step(
        'run golangci-lint',
        [
            linter,
            'run',
            '--out-format=json',
            '--issues-exit-code=0',
            '--timeout=5m',
        ] + go_files,
        step_test_data=lambda: self.m.json.test_api.output_stream({
            'Issues': [{
                'FromLinter': 'deadcode',
                'Text': '`foo` is unused',
                'Severity': '',
                'SourceLines': ['func foo() {}'],
                'Pos': {
                    'Filename': 'client/cmd/isolate/lib/batch_archive.go',
                    'Offset': 7960,
                    'Line': 250,
                    'Column': 6
                },
                'HunkPos': 4,
                'ExpectedNoLintLinter': ''
            }],
        }),
        stdout=self.m.json.output())

    for issue in result.stdout.get('Issues') or ():
      pos = issue['Pos']
      line = pos['Line']
      self.m.tricium.add_comment(
          'golangci-lint (%s)' % issue['FromLinter'],
          issue['Text'],
          self.m.path.join(path_to_go_modules, pos['Filename']),
          start_line=line,
          end_line=line,
          # Gerrit (and Tricium, as a pass-through proxy) requires robot
          # comments to have valid start/end position.
          # TODO(crbug/1239584): provide accurate start/end position.
          start_char=0,
          end_char=0,
      )

    self.m.tricium.write_comments()
