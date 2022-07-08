# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
from . import customization
from . import helper
from . import mount_wim
from . import unmount_wim
from . import regedit
from . import add_windows_package
from . import add_windows_driver

from PB.recipes.infra.windows_image_builder import (offline_winpe_customization
                                                    as winpe)
from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
from PB.recipes.infra.windows_image_builder import sources as src_pb
from PB.recipes.infra.windows_image_builder import dest as dest_pb


class OnlineWindowsCustomization(customization.Customization):
  """ Online windows customization support """

  def __init__(self, **kwargs):
    """ __init__ generates a ref for the given customization
    """
    super(OnlineWindowsCustomization, self).__init__(**kwargs)
    # ensure that the customization is of the correct type
    assert self.customization().WhichOneof(
        'customization') == 'online_windows_customization'
    self._name = self.customization().online_windows_customization.name
    self._workdir = self._path['cleanup'].join(self._name, 'workdir')
    self._scratchpad = self._path['cleanup'].join(self._name, 'sp')
    self._canon_cust = None

  def pin_sources(self):
    """ pins the given config by replacing the sources in customization """
    # pin the input images
    owc = self.customization().online_windows_customization
    for boot in owc.online_customizations:
      for drive in boot.vm_config.qemu.drives:
        if drive.input_src.WhichOneof('src'):
          drive.input_src.CopyFrom(self._source.pin(drive.input_src))
      # pin the refs in the actions
      for online_action in boot.online_actions:
        for action in online_action.actions:
          helper.pin_src_from_action(action, self._source)
