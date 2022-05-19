# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine import recipe_api
from recipe_engine.recipe_api import Property

#TODO(anushruth): Move the emulator to a proper cipd location
QEMU_PKG = 'experimental/anushruth_at_google_com/emulators/qemu/linux-amd64'


class QEMUAPI(recipe_api.RecipeApi):
  """ API to manage qemu VMs """

  def __init__(self, *args, **kwargs):
    super(QEMUAPI, self).__init__(*args, **kwargs)
    self._install_dir = ''

  def init(self, version):
    """ Initialize the module, ensure that qemu exists on the system """
    # create a directory to store qemu tools
    self._install_dir = self.m.path['cache'].join('qemu')
    # download the binaries to the install directory
    e = self.m.cipd.EnsureFile()
    e.add_package(QEMU_PKG, version)
    self.m.cipd.ensure(self._install_dir, e, name="Download qemu")
