# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""A helper for bootstrapping Go and Node environments."""

from typing import Optional
import contextlib
import re

from recipe_engine import recipe_api


class BuildEnvApi(recipe_api.RecipeApi):
  """API for bootstrapping Go and Node environments."""

  @contextlib.contextmanager
  def __call__(self,
               root,
               go_version_file: Optional[str] = None,
               nodejs_version_file: Optional[str] = None):
    """A context manager that activates the build environment.

    Used to build code in a standalone git repositories that don't have Go
    or Node.js available via some other mechanism (like via gclient DEPS).

    It reads the Golang version from `<root>/<go_version_file>` and Node.js
    version from `<root>/<nodejs_version_file>`, bootstraps the corresponding
    versions in cache directories, adjusts PATH and other environment variables
    and yields to the user code.

    Let's assume the requested Go version is '1.16.10' and Node.js version
    is '16.13.0', then this call will use following cache directories (notice
    that '.' is replace with '_' since '.' is not allowed in cache names):
      * `go1_16_10`: to install Go under.
      * `gocache`: for Go cache directories.
      * `nodejs16_13_0`: to install Node.js under.
      * `npmcache`: for NPM cache directories.

    For best performance the builder must map these directories as named
    caches using e.g.

        luci.builder(
            ...
            caches = [
                swarming.cache("go1_16_10"),
                swarming.cache("gocache"),
                swarming.cache("nodejs16_13_0"),
                swarming.cache("npmcache"),
            ],
        )

    Args:
      root (Path) - path to the checkout root.
      go_version_file (str) - path within the checkout to a text file with Go
          version to bootstrap or None to skip bootstrapping Go.
      nodejs_version_file (str) - path within the checkout to a text file with
          Node.js version to bootstrap or None to skip bootstrapping Node.js.
    """
    # Validate the version read from the file against a regexp to avoid
    # potential abuse though version like "../evil/".
    #
    # The expected format is e.g. '1.21.0.chromium1'. The last component is
    # optional.
    GOOD_VERSION_RE = r'^\d+\.\d+\.\d+(\.\w+)?$'

    @contextlib.contextmanager
    def bootstrap(tool, module, version_file):
      if version_file:
        version = self.m.file.read_text(
            'read %s' % version_file,
            root.join(version_file),
            test_data='6.6.6.chromium1\n',
        ).strip().lower()
        if not re.match(GOOD_VERSION_RE, version):  # pragma: no cover
          raise ValueError('Bad %s version number %r' % (tool, version))
        install_path = self.m.path['cache'].join(
            '%s%s' % (tool, version.replace('.', '_')))
        with module(version, path=install_path):
          yield
      else:
        yield

    with bootstrap('go', self.m.golang, go_version_file):
      with bootstrap('nodejs', self.m.nodejs, nodejs_version_file):
        yield
