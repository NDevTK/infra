# Copyright 2022 The Chromium Authors. All rights reserved.
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
from PB.go.chromium.org.luci.buildbucket.proto import common


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

  def pin_sources(self, ctx):
    """ pins the given config by replacing the sources in customization
    Args:
      * ctx: dict containing the context for the customization
    """
    # pin the input images
    owc = self.customization().online_windows_customization
    for boot in owc.online_customizations:
      for drive in boot.vm_config.qemu_vm.drives:
        if drive.input_src.WhichOneof('src'):
          drive.input_src.CopyFrom(self._source.pin(drive.input_src, ctx))
      # pin the refs in the actions
      for online_action in boot.online_actions:
        for action in online_action.actions:
          helper.pin_src_from_action(action, self._source, ctx)
    return ctx

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
          srcs = helper.get_src_from_action(action)
          for src in srcs:
            self._source.download(src)

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

  def remove_upload_dests(self):
    """ remove_upload_dests removes the upload_dests specified by a config.
    This is meant to be used by the try builder to avoid uploading to prod
    locations from a try job.
    Note: Run before pinning is done on the customization
    """
    owc = self.customization().online_windows_customization
    for boot in owc.online_customizations:
      for drive in boot.vm_config.qemu_vm.drives:
        if drive.output_dests:
          drive.output_dests.clear()

  @property
  def outputs(self):
    """ return the output of executing this config. Doesn't guarantee that the
    output exists. The outputs are recorded only for those drives that have
    a dest. All other drives are discarded after execution"""
    outputs = []
    if self.get_key():
      owc = self.customization().online_windows_customization
      for boot in owc.online_customizations:
        for drive in boot.vm_config.qemu_vm.drives:
          if drive.output_dests:
            # If there is a user defined upload for this, ensure cache upload
            ctx = self.update_drive_context(boot.name, drive, {})
            for key, op in ctx.items():
              self.ensure_cache_upload(drive, op)
            outputs.extend(drive.output_dests)
    return outputs

  @property
  def inputs(self):  # pragma: no cover
    """ inputs returns the input(s) required for this customization.

    inputs here refer to any external refs that might be required for this
    customization
    """
    inputs = []
    owc = self.customization().online_windows_customization
    for boot in owc.online_customizations:
      for drive in boot.vm_config.qemu_vm.drives:
        if drive.input_src.WhichOneof('src'):
          inputs.append(drive.input_src)
    return inputs

  @property
  def context(self):
    """ context returns a dict containing the map to image id to output dest
    """
    outputs = {}
    if self.get_key():
      owc = self.customization().online_windows_customization
      for boot in owc.online_customizations:
        for drive in boot.vm_config.qemu_vm.drives:
          if not drive.readonly:
            outputs = self.update_drive_context(boot.name, drive, outputs)
    return outputs

  def update_drive_context(self, oc_name, drive, ctx):
    """ update_drive_context updates the context with the output ref

    Args:
      * oc_name: online customization (boot) name
      * drive: drives proto object representing a drive
      * ctx: context dictionary that needs to be updated
    Returns updated context
    """
    f_name = 'boot({})-drive({})-output'.format(oc_name, drive.name)
    key = '{}-{}'.format(self.id, f_name)
    location = 'WIB-ONLINE-CACHE/{}-{}.zip'
    if self.tryrun:
      location = 'WIB-ONLINE-CACHE-TRY/{}-{}.zip'  # pragma: nocover
    ctx[key] = src_pb.Src(
        gcs_src=src_pb.GCSSrc(
            bucket='chrome-gce-images',
            source=location.format(self.get_key(), f_name)))
    return ctx

  def inject_cache_upload(self, inputs):
    """ inject_cache_upload injects the outputs that we need to upload.

    Ensures that there is at lease one cache upload if the output is referenced
    in any of the inputs given.

    Args:
      * inputs: List of local_src inputs that can ref drives
    """
    if self.get_key():
      owc = self.customization().online_windows_customization
      for boot in owc.online_customizations:
        for drive in boot.vm_config.qemu_vm.drives:
          ctx = self.update_drive_context(boot.name, drive, {})
          # unpack the ctx. It should only contain one entry
          for key, op in ctx.items():
            # Refs to this drive will be the key string, if it was referenced
            # somewhere. If this ref exists then upload to cache
            if key in inputs:
              self.ensure_cache_upload(drive, op)

  def ensure_cache_upload(self, drive, cache):
    """ ensure_cache_upload ensures that at least one upload is scheduled for
    uploading the results to GCS cache

    Args:
      * drive: drive_pb.Drive proto object that points to a virtual disk/cdrom
      * cache: upload path to upload the said drive to in GCS cache
    """
    cache_url = self._source.get_url(cache)
    tags = {'orig': cache_url}
    # We only need one copy to upload the drive.
    for dest in drive.output_dests:
      if self._source.get_url(dest) == cache_url:
        # Looks like we already scheduled one. Return
        return  # pragma: nocover
    # Add a cache upload to the config
    drive.output_dests.append(dest_pb.Dest(gcs_src=cache.gcs_src, tags=tags))
    return

  def process_disks(self, drive, include=None):
    ''' process_disks processes the disk and prepares them to be used on a VM.

    Copies the given disk/cdrom image to staging directory. Creates new
    empty/non-empty disks if required.

    Args:
      * drive: Drive proto object
      * include: dict containing the files to be copied
    '''
    if not self.m.path.exists(self.m.qemu.disks.join(drive.name)):
      if not drive.input_src.WhichOneof('src'):
        # create a new drive and copy the files to it.
        self.m.qemu.create_disk(
            disk_name=drive.name,
            fs_format=drive.filesystem,
            min_size=drive.size,
            include=include)
      else:
        disk_image = self._source.get_local_src(drive.input_src)
        disk_folder = self.m.qemu.disks
        if str(disk_image).endswith('zip') or str(disk_image).endswith('tar'):
          self.m.archive.extract(
              'Unpack {} to staging dir'.format(drive.name),
              disk_image,
              disk_folder,
              include_files=[drive.name],
              archive_type='zip')
        else:
          # Everything else, just link them to disk dir (isos, flash image,...)
          self.m.file.symlink('Link {} to {}'.format(disk_image, disk_folder),
                              disk_image, disk_folder.join(drive.name))

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
          srcs = helper.get_src_from_action(action)
          for src in srcs:
            if src.WhichOneof('src'):
              local_src = self._source.get_local_src(src)
              deps[local_src] = self._source.get_rel_src(src)
      if len(deps) > 0:
        deps_disk = drive_pb.Drive(
            name='DEPS',
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
      # Wait for boot up by attempting to query for date
      self.execute_powershell(
          'Wait for boot up', ctx={}, expr='Get-Date', timeout=boot_time)
      return qemu_vm

  def shutdown_vm(self, vm_name):
    ''' shutdown_vm sends `Stop-Computer` signal to the powershell session

    Args:
      * vm_name: name of the vm
    '''
    try:
      self.execute_powershell(
          'Shutdown {}'.format(vm_name), ctx={}, expr='Stop-Computer')
      # The command was sent successfully. VM must be shutting down
      return True
    except Exception:
      # catch the step failure. This probably happened because the vm is
      # down. Return none if the shutdown attempt fails
      return False

  def safely_shutdown_vm(self, oc):
    ''' safely_shutdown_vm attempts to shutdown the vm safely.

    There are 3 ways to stop a vm
    * shutdown_vm: This attempts to send a `Stop-Computer` powershell command.
    This is same as clicking on shutdown in windows.
    * powerdown_vm: This attempts to mimic the powerbutton on the system being
    pressed. Sometimes this is ignored by the OS.
    * quit_vm: This is basically a kill signal sent to QEMU. QEMU honors this
    by killing the VM and the OS is not shutdown safely.

    We first attempt to shutdown_vm. If that fails we do powerdown_vm. If the
    vm is still up then kill it.

    Args:
      * oc: OnlineCustomization proto object
    '''
    vm_name = oc.vm_config.qemu_vm.name

    with self.m.step.nest('Shutting down {}'.format(vm_name), status='last'):
      # Give 5 minutes for the VM to quit if shutdown time is not given.
      shutdown_time = 300
      if oc.win_vm_config and oc.win_vm_config.shutdown_time > 0:
        shutdown_time = oc.win_vm_config.shutdown_time

      # Try to shutdown the vm through powershell
      if not self.shutdown_vm(vm_name):
        # Powershell session must be down. Try sending the powerdown signal
        self.m.qemu.powerdown_vm(vm_name)

      # wait for shutdown to complete
      self.m.step('Wait for vm to stop', ['sleep', '{}'.format(shutdown_time)])

      # TODO(anushruth): Poll for vm status instead of sleeping for given time
      # check the vm status. If we receive
      # {
      #   "return": {
      #   "Error": "[Errno 111] Connection refused"
      #   }
      # }
      # It means that VM is down
      resp = self.m.qemu.vm_status(vm_name)
      if 'return' in resp and 'Error' in resp['return'] and \
          'Connection refused' in resp['return']['Error']:
        # The VM is already down. Return True
        return True

      # If we reach to this point. Then VM is not gonna shutdown. Kill it
      self.m.qemu.quit_vm(vm_name)

      # Raise an error as we couldn't terminate safely
      raise self.m.step.StepFailure(
          'Unable to shutdown vm {}. Force killed'.format(vm_name))

  def upload_disks(self, oc):
    """ upload_disks compresses and then uploads the disk image.

    Ideally this should be used minimally to avoid network traffic. But in
    situations where this is unavoidable. We can upload the disk for use in
    another build. Unlike offline builder. This only uploads the disk if
    specified by the config. The disk images are compressed before upload.

    Args:
      oc: onlinewc.OnlineCustomization, this should have already executed and
      the VM terminated when this is called.
    """
    for drive in oc.vm_config.qemu_vm.drives:
      if drive.output_dests:
        pkg = self.m.qemu.disks.join(drive.name)
        # compress disk images as they are pretty big
        compressed_pkg = self.m.qemu.disks.join('{}.zip'.format(drive.name))
        (self.m.archive.package(self.m.qemu.disks).with_file(pkg).archive(
            'Archive {} for upload'.format(drive.name), compressed_pkg))
        # Upload a cached copy first
        ctx = self.update_drive_context(oc.name, drive, {})
        url = ''
        # there is only one value here
        for src in ctx.values():
          url = self._source.get_url(src)
          self._source.upload_package(
              dest_pb.Dest(gcs_src=src.gcs_src, tags={'orig': url}),
              compressed_pkg)
        for dest in drive.output_dests:
          # Avoid reuploading cached copy
          if self._source.get_url(dest) != url:
            if url:
              # update the tags
              dest.tags['orig'] = url
            self._source.upload_package(dest, compressed_pkg)
        # delete the compressed disk image
        self.m.file.remove(
            'Delete compressed {} after upload'.format(drive.name),
            compressed_pkg)

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
        # Clear out the disks, free up space
        with self.m.step.nest('Cleanup after {}'.format(owc.name)):
          self.m.qemu.cleanup_disks()
          self.m.file.rmcontents('clean workdir', self._workdir)
          self.m.file.rmcontents('clean scratchpad', self._scratchpad)

  def execute_online_customization(self, oc):
    ''' execute_online_customization performs all the required initialization,
    Boots the VM, waits for the VM to boot and then executes all the given
    actions

    Args:
      * oc: an OnlineCustomization proto object containing the data
    '''
    with self.m.step.nest('Execute online customization {}'.format(
        oc.name)) as s:
      # If set, attempt to upload the disks
      attempt_upload = True
      # Boot up the vm
      self.start_qemu(oc)
      try:
        for online_action in oc.online_actions:
          with self.m.step.nest('Execute online action {}'.format(
              online_action.name)):
            for action in online_action.actions:
              self.execute_action(action, oc.win_vm_config.context)
      except Exception as e:
        s.presentation.logs['Error'] = str(e)
        # Don't upload these disks.
        attempt_upload = False
        if self.mode == wib.CustomizationMode.CUST_DEBUG:
          # If its debug mode then we just sleep for boot_time. This will let
          # you debug by forwarding port 5900 and using a vnc viewer like
          # remmina. You can also forward port 4445 and serial_over_tcp_port.py
          # can be used to get powershell that the recipe engine is using.
          # Don't worry about boot_time, because we cancel the wait as soon as
          # the PSOverCom.ps1 kicks in.
          debug_time = oc.win_vm_config.boot_time
          self.m.step(
              'Debug sleep for {} seconds'.format(debug_time),
              cmd=['sleep', debug_time])
        raise e
      finally:
        with self.m.step.nest('Deinit online customization {}'.format(oc.name)):
          self.safely_shutdown_vm(oc)
          if attempt_upload:
            # upload the disk(s) if required
            self.upload_disks(oc)
          # Delete the dependency disk. We don't need it anymore
          if self.m.path.exists(
              self.m.qemu.disks.join("DEPS")):  #pragma: nocover
            self.m.file.remove("Delete deps disk",
                               self.m.qemu.disks.join("DEPS"))

  def execute_action(self, action, ctx):
    ''' execute_action runs the given action in the given context

    Args:
      * action: actions.Action proto object representing the action to be
      performed
      * ctx: dict representing a global context.
    '''
    a = action.WhichOneof('action')
    if a == 'add_file':
      return self.add_file(action.add_file, ctx, action.timeout)
    if a == 'powershell_expr':
      return self.powershell_expr(action.powershell_expr, ctx, action.timeout)
    raise self.m.step.StepFailure(
        'Executing {} not supported yet'.format(a))  # pragma: nocover

  def add_file(self, add_file, ctx, timeout):
    ''' add_file copy a file from remote to local destination

    Args:
      * add_file: actions.AddFile proto object representing the operation
      * ctx: global context for the operation
      * timeout: maximum time this copy operation is expected to take
    '''
    rel_src = self._source.get_rel_src(add_file.src)
    local_src = self._source.get_local_src(add_file.src)
    # src_file contains the file/dir name to be copied
    src_file = '*'
    if not self.m.path.isdir(local_src):
      # if the src is a file then src is the dir name and src_file is filename
      src_file = self.m.path.basename(rel_src)
      rel_src = self.m.path.dirname(rel_src)
    # powershell expression to copy the artifacts. Using robocopy
    expr = 'robocopy ${DRIVE_DEPS}\\' + '{} {} {} /e'.format(
        helper.conv_to_win_path(rel_src), add_file.dst, src_file)
    self.execute_powershell(
        'Add File: {}'.format(add_file.name),
        ctx,
        expr,
        cont=False,
        timeout=timeout,
        retcode=(0, 1, 2, 3))

  def powershell_expr(self, pwsh_expr, ctx, timeout):
    ''' powershell_expr runs a given powershell expression and executes it

    Args:
      * pwsh_expr: action.PowershellExpr proto object containing the
      expression to be executed
      * ctx: Context to be set before executing the expression
      * timeout: timeout in seconds for the given expression
    '''
    # copy the global ctx
    ps_ctx = {}
    if ctx:
      for var, val in ctx.items():
        ps_ctx[var] = val

    # add all the srcs as context for the expression
    if pwsh_expr.srcs:
      for var, src in pwsh_expr.srcs.items():
        win_src = helper.conv_to_win_path(self._source.get_rel_src(src))
        ps_ctx[var] = '${DRIVE_DEPS}:\\' + win_src

    r_codes = pwsh_expr.return_codes
    # Default successful return code is 0
    r_codes = (0,) if not r_codes else r_codes

    self.execute_powershell(
        pwsh_expr.name,
        ps_ctx,
        pwsh_expr.expr,
        cont=pwsh_expr.continue_ctx,
        logs=pwsh_expr.logs,
        timeout=timeout,
        retcode=r_codes,
        ignore_timeout=pwsh_expr.ignore_timeout)

  def execute_powershell(self,
                         name,
                         ctx,
                         expr,
                         logs=(),
                         cont=False,
                         timeout=300,
                         retcode=(0,),
                         ignore_timeout=False):
    ''' execute_powershell runs the given powershell expression on the vm.

    If cont is true, the session is kept alive. This means the next expression
    will be run in the same context and can use the results of the last
    expression

    Args:
      * name: name of the step
      * ctx: context dictionary. This is a set of key value pairs that can be
      used in the expression
      * expr: powershell expression to be executed
      * logs: optional log files to read
      * cont: If true the session is kept alive. If false exit the session
      after execution.
      * timeout: time in seconds to wait for the expression to execute.
      * retcode: iterable containing all possible return codes to treat as
      success
      * ignore_timeout: if set, no exception is thrown on timeout
    '''
    # use the serial_port_over_tcp script to execute the expression
    cmd = [
        'python3',
        self._scripts('serial_port_over_tcp.py'), '-s', 'localhost:4445'
    ]
    # add all the context to the expression
    for k, v in ctx.items():
      cmd += ['-l', '{}="{}"'.format(k, v)]
    # add logs to read back if any
    for log in logs:
      cmd += ['-L', log]  # pragma: nocover
    if timeout:
      cmd += ['-t', timeout]  # pragma: nocover
    # add the expression to the script
    cmd += ['-e', expr]
    # continue session if required
    if cont:
      cmd += ['-c']  # pragma: nocover
    res = self.m.step(
        'Powershell> {}'.format(name),
        cmd=cmd,
        stdout=self.m.raw_io.output(),
        timeout=(timeout + 300)  # Wait for 5 more minutes more
    )
    res.presentation.logs['RawResp'] = res.stdout
    ret = {}
    try:
      # parse the json file
      ret = self.m.json.loads(res.stdout)
    except Exception as e:  # pragma: nocover
      res.presentation.logs['Error'] = 'Failed to parse: ' + str(e)
      raise e
    # Update the step presentation
    if 'Logs' in ret and ret['Logs']:  # pragma: nocover
      for log_file, log in ret['Logs'].items():
        res.presentation.logs[log_file] = log
    if 'Output' in ret and ret['Output']:
      res.presentation.logs['STDOUT'] = ret['Output']
    if 'Error' in ret and ret['Error']:
      res.presentation.logs['Error'] = ret['Error']
      if 'Timeout' in ret['Error'] and not ret['Success']:
        if not ignore_timeout:
          raise self.m.step.StepFailure('Timeout')
        # This was expected. Just return
        return res
    # Throw error if the expression failed to execute
    if 'Success' in ret and not ret['Success']:
      raise self.m.step.StepFailure('Error in execution. Check stdout, stderr')
    # Throw error if return code is not what we expect. Ignore success
    if 'RetCode' in ret and int(ret['RetCode']) not in retcode:
      err = 'Expression returned {} Expecting {}. Check stdout, stderr'
      raise self.m.step.StepFailure(err.format(ret['RetCode'], retcode))
