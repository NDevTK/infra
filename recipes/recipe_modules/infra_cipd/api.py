# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import contextlib
import collections

from recipe_engine import recipe_api


class InfraCIPDApi(recipe_api.RecipeApi):
  """API for building packages defined in infra's public and intenral repos.

  Essentially a shim around scripts in
  https://chromium.googlesource.com/infra/infra.git/+/main/build/
  and its internal counterpart.
  """

  def __init__(self, **kwargs):
    super(InfraCIPDApi, self).__init__(**kwargs)
    self._cur_ctx = None  # (path_to_repo, name_prefix)

  @contextlib.contextmanager
  def context(self, path_to_repo, goos=None, goarch=None):
    """Sets context building CIPD packages.

    Arguments:
      path_to_repo (path): path infra or infra_internal repo root dir.
        Expects to find `build/build.py` inside provided dir.
      goos, goarch (str): allows for setting GOOS and GOARCH
        for cross-compiling Go code.

    Doesn't support nesting.
    """
    if self._cur_ctx is not None:  # pragma: no cover
      raise ValueError('Nesting contexts not allowed')
    if bool(goos) != bool(goarch):  # pragma: no cover
      raise ValueError('GOOS and GOARCH must be either both set or both unset')

    env, name_prefix = None, ''
    if goos and goarch:
      env = {'GOOS': goos, 'GOARCH': goarch}
      name_prefix ='[GOOS:%s GOARCH:%s]' % (goos, goarch)
    self._cur_ctx = (path_to_repo, name_prefix)
    try:
      with self.m.context(env=env):
        yield
    finally:
      self._cur_ctx = None

  @property
  def _ctx_path_to_repo(self):
    if self._cur_ctx is None:  # pragma: no cover
      raise Exception('must be run under infra_cipd.context')
    return self._cur_ctx[0]

  @property
  def _ctx_name_prefix(self):
    if self._cur_ctx is None:  # pragma: no cover
      raise Exception('must be run under infra_cipd.context')
    return self._cur_ctx[1]

  def build_without_env_refresh(self, sign_id=None):
    """Builds packages.

    Prevents build.py from refreshing the python ENV.
    """
    args = [
        'vpython3',
        self._ctx_path_to_repo.join('build', 'build.py'),
        '--no-freshen-python-env',
        '--builder',
        self.m.buildbucket.builder_name,
    ]
    if sign_id:
      args.extend(['--signing-identity', sign_id])

    return self.m.step(
        self._ctx_name_prefix + 'cipd - build packages',
        args,
    )

  def test(self):
    """Tests previously built packages integrity."""
    return self.m.step(
        self._ctx_name_prefix+'cipd - test packages integrity',
        ['vpython3', self._ctx_path_to_repo.join('build', 'test_packages.py')],
    )

  def upload(self, tags, step_test_data=None):
    """Uploads previously built packages."""
    args = [
      'vpython3',
      self._ctx_path_to_repo.join('build', 'build.py'),
      '--no-rebuild',
      '--upload',
      '--json-output', self.m.json.output(),
      '--builder', self.m.buildbucket.builder_name,
      '--tags',
    ]
    args.extend(tags)
    try:
      return self.m.step(
          self._ctx_name_prefix+'cipd - upload packages',
          args,
          step_test_data=step_test_data or self.test_api.example_upload,
      )
    finally:
      step_result = self.m.step.active_result
      output = step_result.json.output or {}
      p = step_result.presentation
      for pkg in output.get('succeeded', []):
        info = pkg['info']
        title = '%s %s' % (info['package'], info['instance_id'])
        p.links[title] = info.get(
            'url', 'http://example.com/not-implemented-yet')

  def tags(self, git_repo_url, revision):
    """Returns tags to be attached to uploaded CIPD packages."""
    if self.m.buildbucket.build.number <= 0:
      raise ValueError('buildnumbers must be enabled')
    return [
      'luci_build:%s/%s/%s' % (
        self.m.buildbucket.builder_id.bucket,
        self.m.buildbucket.builder_id.builder,
        self.m.buildbucket.build.number),
      'git_repository:%s' % git_repo_url,
      'git_revision:%s' % revision,
    ]
