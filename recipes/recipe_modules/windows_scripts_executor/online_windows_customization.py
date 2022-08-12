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

from PB.recipes.infra.windows_image_builder import (online_windows_customization
                                                    as onlinewc)
from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
from PB.recipes.infra.windows_image_builder import sources as src_pb
from PB.recipes.infra.windows_image_builder import dest as dest_pb
from PB.recipes.infra.windows_image_builder import drive as drive_pb
from PB.recipes.infra.windows_image_builder import vm as vm_pb
from PB.recipes.infra.windows_image_builder import actions as act_pb


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
    self._workdir = self.m.path['cleanup'].join(self._name, 'workdir')
    self._scratchpad = self.m.path['cleanup'].join(self._name, 'sp')
    self._canon_cust = None

  def pin_sources(self):
    """ pins the given config by replacing the sources in customization """
    # pin the input images
    owc = self.customization().online_windows_customization
    for boot in owc.online_customizations:
      for drive in boot.vm_config.qemu_vm.drives:
        if drive.input_src.WhichOneof('src'):
          drive.input_src.CopyFrom(self._source.pin(drive.input_src))
      # pin the refs in the actions
      for online_action in boot.online_actions:
        for action in online_action.actions:
          helper.pin_src_from_action(action, self._source)

  def download_sources(self):
    """ download_sources downloads the sources in the given config to disk"""
    # pin the input images
    owc = self.customization().online_windows_customization
    for boot in owc.online_customizations:
      for drive in boot.vm_config.qemu_vm.drives:
        if drive.input_src.WhichOneof('src'):
          self._source.download(drive.input_src)
      # pin the refs in the actions
      for online_action in boot.online_actions:
        for action in online_action.actions:
          self._source.download(helper.get_src_from_action(action))

  def get_canonical_cfg(self):
    """ get_canonical_cfg returns canonical config after removing name and dest
        Example:
          Given a config

            Customization{
              online_windows_customization: OnlineWindowsCustomization{
                name: "win11_vanilla"
                online_customization: [...]
              }
            }

          returns config

            Customization{
              online_windows_customization: OnlineWindowsCustomization{
                name: ""
                online_customization: [...]
              }
            }

    """
    if not self._canon_cust:
      owc = self.customization().online_windows_customization
      # Generate customization without any names or dest refs. This will make
      # customization deterministic to the generated image
      cust = wib.Customization(
          online_windows_customization=onlinewc.OnlineWinCustomization(
              online_customizations=[
                  self.get_canonical_online_customization(x)
                  for x in owc.online_customizations
              ],),)
      self._canon_cust = cust
    return self._canon_cust  # pragma: nocover

  def get_canonical_online_customization(self, cust):
    """ get_canonical_online_customization returns canonical
    OnlineCustomization object.

    Example:
      Given a onlinewc.OnlineCustomization object

      OnlineCustomization{
        name: "install_bootstrap",
        vm_config: vm.VM{...},
        online_actions: [...],
      }

      returns a onlinewc.OnlineCustomization object

      OnlineCustomization{
        vm_config: vm.VM{...},
        online_actions: [...],
      }

    Args:
      * cust: onlinewc.OnlineCustomization proto object
    """
    # convert online_actions to canonical form
    online_actions = [
        act_pb.OnlineAction(
            actions=[helper.get_build_actions(y)
                     for y in x.actions])
        for x in cust.online_actions
    ]
    return onlinewc.OnlineCustomization(
        vm_config=self.get_cannonical_vm_config(cust.vm_config),
        online_actions=online_actions)

  def get_cannonical_vm_config(self, vm_config):
    """ get_canonical_vm_config takes a vm_pb.VM object and returns a canonical
    vm_pb.VM object.
    Example:
      Given a VM config object

      VM{
        qemu_vm: QEMU_VM{
          name: "Win10AARCH64"
          machine: "virt,virtualization=on,highmem=off"
          cpu: "cortex-a57"
          smp: "cpus=4,cores=8"
          memory: 3072
          device: ['nec-usb-xhci','usb-kbd', 'usb-mouse']
          extra_args: ['--accel tcg,thread=multi']
          drives: [...]
        }
      }

      returns a vm_pb.VM object

      VM{
        qemu_vm: QEMU_VM{
          machine: "virt,virtualization=on,highmem=off"
          cpu: "cortex-a57"
          smp: "cpus=4,cores=8"
          memory: 3072
          device: ['nec-usb-xhci','usb-kbd', 'usb-mouse']
          extra_args: ['--accel tcg,thread=multi']
          drives: [...]
        }
      }

    Args:
      * vm_config: vm_pb.VM proto object representing a VM
    """
    cfg = vm_config.qemu_vm
    machine = cfg.machine
    cpu = cfg.cpu
    smp = cfg.smp
    memory = cfg.memory
    device = cfg.device
    extra_args = cfg.extra_args
    drives = [self.get_canonical_drive_config(x) for x in cfg.drives]
    return vm_pb.VM(
        qemu_vm=vm_pb.QEMU_VM(
            machine=machine,
            cpu=cpu,
            smp=smp,
            memory=memory,
            device=device,
            extra_args=extra_args,
            drives=drives))

  def get_canonical_drive_config(self, drive):
    """ get_canonical_drive_config takes a drive_pb.Drive object and returns a
    canonical drive_pb.Drive object

    Example:
      Given a Drive object

      Drive{
        name: "win10_vanilla_iso"
        input_src: Src(...)
        output_dest: Dest(...)
        interface: "none"
        media: "cdrom"
        fmt: "raw"
      }

      returns a drive_pb.Drive object

      Drive{
        input_src: Src(...)
        interface: "none"
        media: "cdrom"
        fmt: "raw"
      }

    Args:
      * drive: drive_pb.Drive proto object representing a drive image.
    """
    return drive_pb.Drive(
        input_src=drive.input_src,
        interface=drive.interface,
        index=drive.index,
        media=drive.media,
        fmt=drive.fmt,
        readonly=drive.readonly,
        size=drive.size,
        filesystem=drive.filesystem)
