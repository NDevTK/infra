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

  def process_disks(self, drive, include=None):
    ''' process_disks processes the disk and prepares them to be used on a VM.

    Copies the given disk/cdrom image to staging directory. Creates new
    empty/non-empty disks if required.

    Args:
      * drive: Drive proto object
      * include: dict containing the files to be copied
    '''
    if not drive.input_src.WhichOneof('src'):
      # create a new drive and copy the files to it.
      self.m.qemu.create_disk(
          disk_name=drive.name,
          fs_format=drive.filesystem,
          min_size=drive.size,
          include=include)
    else:
      disk_image = self._source.get_local_src(drive.input_src)
      if not str(disk_image).endswith(drive.name):
        # add file name if it doesn't have it
        disk_image = disk_image.join(drive.name)
      # copy the disk image to workdir
      self.m.file.copy('Staging disk image {}'.format(drive.name), disk_image,
                       self.m.qemu.disks.join(drive.name))

  def start_qemu(self, oc):
    ''' start_qemu starts a qemu vm with given config and drives.

    Args:
      * qemu_vm: vm config for a qemu vm to be run
      * drives: list of drives to be attached to the vm
    '''
    with self.m.step.nest('Boot {}'.format(oc.vm_config.qemu_vm.name)):
      # make a copy of the config
      qemu_vm = vm_pb.QEMU_VM()
      qemu_vm.CopyFrom(oc.vm_config.qemu_vm)

      # initialize qemu version
      self.m.qemu.init(qemu_vm.version)

      # process the disks
      for drive in qemu_vm.drives:
        self.process_disks(drive)

      # dependencies to run the VM/Online customization
      deps = {}
      # create dependency disk
      for online_action in oc.online_actions:
        for action in online_action.actions:
          src = helper.get_src_from_action(action)
          if src.WhichOneof('src'):
            local_src = self._source.get_local_src(src)
            deps[local_src] = self._source.get_rel_src(src)
      if len(deps) > 0:
        deps_disk = drive_pb.Drive(
            name='deps.img',
            interface='none',
            media='disk',
            filesystem='fat',
            fmt='raw')
        self.process_disks(deps_disk, deps)
        # Add the dependency list
        qemu_vm.drives.append(deps_disk)
      host_arch = self.m.platform.arch
      host_bits = self.m.platform.bits
      arch = self._arch
      kvm = False
      if host_arch == 'intel' and host_bits == 64 and arch == 'amd64':
        # If we are running on an intel 64 bit system and starting an
        # amd64 or x86_64 vm. Use kvm
        kvm = True
      if host_arch == 'intel' and host_bits == 32 and arch == 'x86':
        # If we are running on an intel 64 bit system and starting an
        # x86 vm. Use kvm
        kvm = True
      if host_arch == 'arm' and host_bits == 64 and arch == 'aarch64':
        # If we are running on arm 64 bit system and targeting a aarch64 VM
        # then use kvm
        kvm = True
      self.m.qemu.start_vm(self._arch, qemu_vm, kvm=kvm)
      # Wait for 5 minutes (300 secs) until the VM boots up
      boot_time = 300
      if oc.win_vm_config:
        # if custom time is specified, Sleep for that amount
        if oc.win_vm_config.boot_time > 0:
          boot_time = oc.win_vm_config.boot_time
      # Wait for the vm to boot up
      self.m.step('Wait for the vm to boot up',
                  ['sleep', '{}'.format(boot_time)])
      return qemu_vm

  def execute_customization(self):
    ''' execute_customization runs all the online customizations included in
    the given customization.
    '''
    owc = self.customization().online_windows_customization
    if owc and len(owc.online_customizations) > 0:
      with self.m.step.nest('Execute online windows customization {}'.format(
          owc.name)):
        for oc in owc.online_customizations:
          self.execute_online_customization(oc)

  def execute_online_customization(self, oc):
    ''' execute_online_customization performs all the required initialization,
    Boots the VM, waits for the VM to boot and then executes all the given
    actions

    Args:
      * oc: an OnlineCustomization proto object containing the data
    '''
    with self.m.step.nest('Execute online customization {}'.format(oc.name)):
      # Boot up the vm
      qemu_vm = self.start_qemu(oc)
      # TODO(anushruth): Execute all the actions on the VM.
      return qemu_vm
