# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

class Workdir(object):
  """Workdir manages all the paths involved with building a package.

  All workdirs (`$WD`) are located at:

      [START_DIR]/3pp/wd/${cipd_pkg_name}/${platform}/${version}

  e.g.

      [START_DIR]/3pp/wd/tools/python/linux-amd64/2.7.13.chromium16
      where cipd_pkg_name is "tools/python", platform is "linux-amd64",
      version is "2.7.13.chromium16"

  Each workdir contains the following subdirs:

      $WD/
        checkout/      # package sources
          .3pp/        # ALL package scripts copied here
        tools_prefix/  # build tools are installed
        deps_prefix/   # deps are installed
        out/           # output of the build

  The checkout/.3pp directory contains a copy of ALL the package definition
  folders and their contents for where the package was defined. e.g. if `python`
  is located at:

     some/repo/third_party_packages/python/3pp.pb

  Then WD/checkout/.3pp is a copy of:

     some/repo/third_party_packages/

  This allows packages to have common scripts, .vpython3 definitions, etc. which
  are shared across all package definitions in a given repo. The cost of the
  copy is relatively small.
  """
  def __init__(self, api, spec, version):
    paths = (['3pp', 'wd'] + spec.cipd_pkg_name.split('/') +
             [spec.platform, version])
    self._base = api.path.start_dir.joinpath(*paths)

  @property
  def base(self):
    """The base of the workdir."""
    return self._base

  @property
  def checkout(self):
    """The path where the sources for the package are checked out (the
    result of the `source` phase.)"""
    return self._base / 'checkout'

  @property
  def verify(self):
    """The path where the verification of the package takes place. The test
    script is invoked within this directory and can use it how it pleases."""
    return self._base / 'verify'

  @property
  def bin_tools(self):
    """The path where the recipe can install tools (like `cipd`) to put in $PATH
    for the docker environment."""
    return self._base / 'bin_tools'

  @property
  def script_dir_base(self):
    """The directory where ALL of the package scripts are copied."""
    return self.checkout / '.3pp'

  def script_dir(self, spec):
    """The directory where stores the scripts of a given spec (ResolvedSpec).

    Returns the Path to the spec scripts.
    """
    return self.script_dir_base / spec.cipd_pkg_name

  @property
  def tools_prefix(self):
    """The $PREFIX where all of the packages's tools will be installed."""
    return self._base / 'tools_prefix'

  @property
  def deps_prefix(self):
    """The $PREFIX where all of the packages's deps will be installed."""
    return self._base / 'deps_prefix'

  @property
  def output_prefix(self):
    """The $PREFIX which contains the contents of the built package."""
    return self._base / 'out'
