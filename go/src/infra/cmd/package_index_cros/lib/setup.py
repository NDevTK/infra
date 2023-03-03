# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import os
from typing import List

from chromite.lib import constants
from chromite.lib import git
from chromite.lib import path_util
from chromite.lib import repo_util


class Setup:
  """
  POD to keep all data related to a setup:
    * board
    * cros_dir: absolute path to chromeos checkout root dir
    * chroot_dir: absolute path to chroot dir
    * src_dir: absolute path to src
    * manifest: manifest handler
  """

  def __init__(self,
               board: str,
               *,
               skip_packages: List[str] = None,
               with_tests: bool = False,
               chroot_dir: str = None):
    self.board = board

    checkout_info = path_util.DetermineCheckout()
    if checkout_info.type != path_util.CHECKOUT_TYPE_REPO:
      raise repo_util.NotInRepoError(
          'Script is executed outside of ChromeOS checkout')

    self.cros_dir = checkout_info.root
    if chroot_dir:
      self.chroot_dir = os.path.realpath(chroot_dir)
      assert (
          not self.chroot_dir.startswith(self.cros_dir) or
          self.chroot_dir == constants.DEFAULT_CHROOT_DIR), (
              f"Custom chroot dir inside {self.cros_dir} is not supported, and "
              f"chromite resolves it to {constants.DEFAULT_CHROOT_DIR}.")
    else:
      self.chroot_dir = path_util.FromChrootPath('/', self.cros_dir)
    self.board_dir = os.path.join(self.chroot_dir, 'build', self.board)
    self.src_dir = os.path.join(self.cros_dir, 'src')
    self.platform2_dir = os.path.join(self.src_dir, 'platform2')

    # List of dirs that might not exist and can be ignored during path fix.
    self.ignorable_dirs = [
        os.path.join(self.board_dir, 'usr', 'include', 'chromeos', 'libica'),
        os.path.join(self.board_dir, 'usr', 'include', 'chromeos', 'libsoda'),
        os.path.join(self.board_dir, 'usr', 'include', 'u2f', 'client'),
        os.path.join(self.board_dir, 'usr', 'share', 'dbus-1'),
        os.path.join(self.board_dir, 'usr', 'share', 'proto'),
        os.path.join(self.chroot_dir, 'build', 'share'),
        os.path.join(self.chroot_dir, 'usr', 'include', 'android'),
        os.path.join(self.chroot_dir, 'usr', 'include', 'cros-camera'),
        os.path.join(self.chroot_dir, 'usr', 'lib64', 'shill'),
        os.path.join(self.chroot_dir, 'usr', 'libexec', 'ipsec'),
        os.path.join(self.chroot_dir, 'usr', 'libexec', 'l2tpipsec_vpn'),
        os.path.join(self.chroot_dir, 'usr', 'share', 'cros-camera'),
    ]

    self.skip_packages = skip_packages
    self.with_tests = with_tests

  @property
  def manifest(self) -> git.ManifestCheckout:
    return git.ManifestCheckout.Cached(self.cros_dir)
