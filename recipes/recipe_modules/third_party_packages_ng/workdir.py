# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

class Workdir(object):
  """Workdir wrangles all the paths involved with building a package.

  All workdirs are located at:

      [START_DIR]/3pp/wd/${name}/${platform}/${version}

  e.g.

      [START_DIR]/3pp/wd/python/linux-amd64/2.7.13.chromium16

  Each workdir contains the following subdirs:

      WD/
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

  This allows packages to have common scripts, .vpython definitions, etc. which
  are shared across all package definitions in a given repo. The cost of the
  copy is relatively small.
  """
  def __init__(self, api, spec, version):
    self._base = (
      api.path['start_dir'].join(
        '3pp', 'wd', spec.name, spec.platform, version))

  @property
  def base(self):
    """The base of the workdir."""
    return self._base

  @property
  def checkout(self):
    """The checkout where the sources for the package are checked out (the
    result of the `source` phase.)"""
    return self._base.join('checkout')

  @property
  def script_dir_base(self):
    """The directory where all of the package's scripts are copied."""
    return self.checkout.join('.3pp')

  @property
  def tools_prefix(self):
    """The $PREFIX where all of the packages's tools will be installed."""
    return self._base.join('tools_prefix')

  @property
  def deps_prefix(self):
    """The $PREFIX where all of the packages's deps will be installed."""
    return self._base.join('deps_prefix')
