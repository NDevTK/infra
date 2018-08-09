# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from . import gcloud as tpp_gcloud
from . import gsutil as tpp_gsutil
from . import git as tpp_git
from . import python as tpp_python
from . import ninja as tpp_ninja
from . import cmake as tpp_cmake
from . import swig as tpp_swig
from . import go as tpp_go
from . import firebase as tpp_firebase
from . import dep as tpp_dep
from .support_prefix import SupportPrefix

from recipe_engine import recipe_api


UNSET_CROSS_PLATFORM = object()


class ThirdPartyPackagesApi(recipe_api.RecipeApi):

  def __init__(self, *args, **kwargs):
    super(ThirdPartyPackagesApi, self).__init__(*args, **kwargs)
    self._dry_run = True
    self._cross_platform = UNSET_CROSS_PLATFORM
    self._singletons = {}

  @property
  def dry_run(self):
    return self._dry_run
  @dry_run.setter
  def dry_run(self, v):
    self._dry_run = bool(v)

  @property
  def cross_platform(self): # pragma: no cover
    if self._cross_platform is UNSET_CROSS_PLATFORM:
      return None
    return self._cross_platform

  def init_cross_platform(self, cross_platform):
    if self._cross_platform is not UNSET_CROSS_PLATFORM: # pragma: no cover
      raise self.m.step.StepFailure('cross_platform may not be set twice')
    assert isinstance(cross_platform, (str, type(None)))
    if cross_platform is not None:  # catch empty-string
      assert cross_platform in (
        'linux-arm64', 'linux-armv6l', 'linux-mips32', 'linux-mips64'), (
          'unsupported platform %r' % (cross_platform,))
      assert not self.m.platform.is_win, (
        'cross_platform not supported on windows')

    self._cross_platform = cross_platform

    if self._cross_platform:
      self.m.python(
        'setup cross compile for %r' % cross_platform,
        '-m',
        self._dx_args + ['echo', 'test_command']
      )

  @property
  def _dx_args(self):
    """Returns python argument prefix for the current cross-compile
    environment."""
    assert self.cross_platform, "cross_platform is unconfigured"
    return ['infra.tools.dockerbuild', 'run', '--platform', self.cross_platform]

  def _get_singleton(self, cls):
    cur = self._singletons.get(cls)
    if not cur:
      cur = self._singletons[cls] = cls(self)
    return cur

  @property
  def python(self):
    return self._get_singleton(tpp_python.PythonApi)

  @property
  def git(self):
    return self._get_singleton(tpp_git.GitApi)

  @property
  def gcloud(self):
    return self._get_singleton(tpp_gcloud.GcloudApi)

  @property
  def gsutil(self):
    return self._get_singleton(tpp_gsutil.GsutilApi)

  @property
  def ninja(self):
    return self._get_singleton(tpp_ninja.NinjaApi)

  @property
  def cmake(self):
    return self._get_singleton(tpp_cmake.CMakeApi)

  @property
  def swig(self):
    return self._get_singleton(tpp_swig.SwigApi)

  @property
  def go(self):
    return self._get_singleton(tpp_go.GoApi)

  @property
  def firebase(self):
    return self._get_singleton(tpp_firebase.FirebaseApi)

  @property
  def dep(self):
    return self._get_singleton(tpp_dep.DepApi)

  def support_prefix(self, base):
    return SupportPrefix(self, base)

  def get_package_name(self, package_name_prefix):
    return package_name_prefix + self.m.cipd.platform_suffix()

  def ensure_package(self, workdir, repo_url, package_name_prefix, install_fn,
                     tag, version, cipd_install_mode, test_fn=None):
    """Ensures that the specified CIPD package exists."""
    package_name = self.get_package_name(package_name_prefix)

    # Check if the package already exists.
    if self.does_package_exist(package_name, version):
      self.m.python.succeeding_step('Synced', 'Package is up to date.')
      return

    # Fetch source code and build.
    checkout_dir = workdir.join('checkout')
    package_dir = workdir.join('package')
    self.m.git.checkout(
        repo_url, ref='refs/tags/' + tag, dir_path=checkout_dir,
        submodules=False)
    self.m.file.ensure_directory('package_dir', package_dir)

    with self.m.context(cwd=checkout_dir):
      install_fn(package_dir, tag)

    package_file = self.build_package(package_name, workdir, package_dir,
                                      cipd_install_mode)

    if test_fn:
      # Rename our built package just in case the package itself references
      # build paths. This will invalidate those references.
      self.m.file.move(
          'rename package for tests',
          package_dir,
          workdir.join('package.built'))

      with self.m.context(cwd=workdir):
        test_fn(package_file)

    self.register_package(package_file, package_name, version)


  def get_latest_release_tag(self, repo_url, prefix='v'):
    result = None
    result_parsed = None
    tag_prefix = 'refs/tags/'
    for ref in self.m.gitiles.refs(repo_url):
      if not ref.startswith(tag_prefix):
        continue
      t = ref[len(tag_prefix):]

      # Parse version.
      if not t.startswith(prefix):
        continue
      parts = t[len(prefix):].split('.')
      if not all(p.isdigit() for p in parts):
        continue
      parsed = map(int, parts)

      # Is it the latest?
      if result_parsed is None or result_parsed < parsed:
        result = t
        result_parsed = parsed
    return result

  def does_package_exist(self, name, version):
    search = self.m.cipd.search(name, 'version:' + version)
    return bool(search.json.output['result'] and not self.dry_run)

  def build_package(self, name, workdir, root, install_mode):
    package_file = workdir.join('package.cipd')
    self.m.cipd.build(root, package_file, name, install_mode=install_mode)
    return package_file

  def register_package(self, package_file, name, version):
    if not self.dry_run:
      self.m.cipd.register(name, package_file, tags={'version': version},
                           refs=['latest'])
