# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Packages the firebase-tools for multiple platforms."""

from . import util

from recipe_engine import recipe_api


PACKAGE_NAME = 'infra/npm/firebase-tools'

# This version suffix serves to distinguish different revisions of gsutil built
# with this recipe.
PACKAGE_VERSION_SUFFIX = ''


class PlatformNotSupported(Exception):
  pass


class FirebaseApi(util.ModuleShim):

  @recipe_api.composite_step
  def package(self):
    if not self.m.platform.is_linux: # no cover
      self.m.python.succeeding_step('Only built on Linux', 'No need.')
      return

    nodejs_dir = self.m.path['start_dir'].join('nodejs')
    self.m.cipd.ensure(nodejs_dir, {
      'infra/nodejs/${platform}': 'version:8.11.3',
    })

    latest_version = self._get_latest_version(nodejs_dir) + PACKAGE_VERSION_SUFFIX

    if self.does_package_exist(PACKAGE_NAME, latest_version):
      self.m.python.succeeding_step('Synced', 'Package is up to date.')
      return

    # Download firebase-tools using npm.
    with self.m.context(cwd=nodejs_dir):
      self.m.step('fetch firebase-tools', [
          nodejs_dir.join('npm'),
          'install',
          'firebase-tools',
        ]
      )

    workdir = nodejs_dir.join('npm', 'node_modules', 'firebase-tools')
    base_dir = workdir.join('firebase-tools')

    # Create CIPD bundle.
    package_file = self.build_package(PACKAGE_NAME, workdir, base_dir, 'copy')
    self.register_package(package_file, PACKAGE_NAME, latest_version)

  def _get_latest_version(self, nodejs_dir):
    with self.m.context(cwd=nodejs_dir):
      return self.m.step('fetch firebase-tools', [
          nodejs_dir.join('npm'),
          'view',
          'firebase-tools',
          'version'
        ],
        stdout=self.m.raw_io.output(),
        step_test_data=lambda: self.m.raw_io.test_api.stream_output('3.19.3', stream='stdout')
      ).stdout
