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

COPYPE = 'Copy-PE.ps1'
ADDFILE = 'robocopy'

# map arch to microsofts version of arch
# https://docs.microsoft.com/en-us/windows-hardware/manufacture/desktop/copype-command-line-options?#copype-command-line-options-1
COPYPE_ARCH = {'amd64': 'amd64',
               'x86': 'x86',
               'aarch64': 'arm64'}


class OfflineWinPECustomization(customization.Customization):
  """ WinPE based customization support """

  def __init__(self, **kwargs):
    """ __init__ generates a ref for the given customization
    """
    super(OfflineWinPECustomization, self).__init__(**kwargs)
    # ensure that the customization is of the correct type
    assert self.customization().WhichOneof(
        'customization') == 'offline_winpe_customization'
    # use a custom work dir
    self._name = self.customization().offline_winpe_customization.name
    self._workdir = self.m.path['cleanup'].join(self._name, 'workdir')
    self._scratchpad = self.m.path['cleanup'].join(self._name, 'sp')
    self._canon_cust = None
    helper.ensure_dirs(self.m.file, [self._workdir])

  def pin_sources(self, ctx):
    """ pins the given config by replacing the sources in customization

    Args:
      * ctx: dict containing the context for the customization
    """
    wpec = self._customization.offline_winpe_customization
    if wpec.image_src.WhichOneof('src'):
      wpec.image_src.CopyFrom(self._source.pin(wpec.image_src, ctx))
    for off_action in wpec.offline_customization:
      for action in off_action.actions:
        helper.pin_src_from_action(action, self._source, ctx)
    if self.tryrun:
      wpec.image_dests.clear()  # pragma: nocover

  def download_sources(self):
    """ download_sources downloads the sources in the given config to disk"""
    wpec = self._customization.offline_winpe_customization
    self._source.download(wpec.image_src)
    for off_action in wpec.offline_customization:
      for action in off_action.actions:
        srcs = helper.get_src_from_action(action)
        for src in srcs:
          self._source.download(src)

  def get_canonical_cfg(self):
    """ get_canonical_cfg returns canonical config after removing name and dest

    Example:
      Given a config

        Customization{
          offline_winpe_customization: OfflineWinPECustomization{
            name: "winpe_gce_vanilla"
            image_src: Src{...}
            image_dests: [...]
            offline_customization: [...]
          }
        }

      returns config

        Customization{
          offline_winpe_customization: OfflineWinPECustomization{
            name: ""
            image_src: Src{...}
            offline_customization: [...]
          }
        }
    """
    if not self._canon_cust:
      wpec = self._customization.offline_winpe_customization
      # Generate customization without any names or dest refs. This will make
      # customization deterministic to the generated image
      cust = wib.Customization(
          offline_winpe_customization=winpe.OfflineWinPECustomization(
              image_src=wpec.image_src,
              offline_customization=[
                  helper.get_build_offline_customization(c)
                  for c in wpec.offline_customization
              ],
          ),)
      self._canon_cust = cust
    return self._canon_cust

  @property
  def outputs(self):
    """ return the output(s) of executing this config. Doesn't guarantee that
    the output(s) exists."""
    if self.get_key():
      location = 'WIB-WIM/{}.zip'
      if self.tryrun:
        location = 'WIB-WIM-TRY/{}.zip'  # pragma: nocover
      output = src_pb.GCSSrc(
          bucket='chrome-gce-images', source=location.format(self.get_key()))
      return [
          dest_pb.Dest(
              gcs_src=output,
              tags={'orig': self._source.get_url(src_pb.Src(gcs_src=output))},
          )
      ]
    return None  # pragma: no cover

  @property
  def inputs(self):  # pragma: no cover
    """ inputs returns the input(s) required for this customization.

    inputs here refer to any external refs that might be required for this
    customization
    """
    inputs = []
    wpec = self._customization.offline_winpe_customization
    if wpec.image_src.WhichOneof('src'):
      # Add the image src required to the list
      inputs.append(wpec.image_src)
      for off_action in wpec.offline_customization:
        for action in off_action.actions:
          # Add all the srcs from actions to the list
          inputs.extend(helper.get_src_from_action(action))
    return inputs

  @property
  def context(self):
    """ context returns a dict containing the local_src id mapping to output
    src.
    """
    return {
        '{}-output'.format(self.id): self._source.dest_to_src(self.outputs[0])
    }

  def execute_customization(self):
    """ execute_customization initializes the winpe image, runs the given
    actions and repackages the image and uploads the result to GCS"""
    wpec = self._customization.offline_winpe_customization
    if wpec and len(wpec.offline_customization) > 0:
      with self.m.step.nest('offline winpe customization ' + wpec.name):
        self.init_win_pe_image(self._arch, wpec.image_src)
        try:
          for action in wpec.offline_customization:
            self.perform_winpe_actions(action)
        except Exception:
          # Unmount the image and discard changes on failure
          self.deinit_win_pe_image(save=False)
          raise
        else:
          self.deinit_win_pe_image()

  def init_win_pe_image(self, arch, image, index=1):
    """ init_win_pe_image initializes the source image (if given) by mounting
    it to dest

    Args:
      * arch: string representing architecture of the image
      * image: sources.Src object ref an image to be modified
      * dest: destination to upload artifacts to
      * index: index of the image to be mounted
    """
    with self.m.step.nest('Init WinPE image modification ' + arch + ' in ' +
                          str(self._workdir)):
      # Path to boot.wim. This is where COPY-PE generates the image
      wim_path = self._workdir.join('media', 'sources', 'boot.wim')
      # Use WhichOneOf to test for emptiness
      # https://developers.google.com/protocol-buffers/docs/reference/python-generated#oneof
      if not image.WhichOneof('src'):
        # gen a winpe arch dir for the given arch
        self.m.powershell(
            'Gen WinPE media for {}'.format(arch),
            self._scripts('WindowsPowerShell\Scripts').join('Copy-PE.ps1'),
            args=[
                '-WinPeArch', COPYPE_ARCH[arch], '-Destination',
                str(self._workdir)
            ])
      else:
        image_path = self._source.get_local_src(image)
        if str(image_path).endswith('.zip'):
          # unzip the given image
          self.m.archive.extract(
              'Unpack {}'.format(self._source.get_url(image)),
              self._source.get_local_src(image), self._workdir)
        else:
          # Path to copy the image to
          wim_path = self._workdir.join(self.m.path.basename(image_path))
          # image was from remote source. Copy to workdir
          self.m.file.copy(
              'Copy {} to workdir'.format(self._source.get_url(image)),
              image_path, wim_path)
          # I can't figure out a way to get this to work with cipd. Somehow the
          # wim gets set to ReadOnly. I tried changing the permissions before
          # uploading to cipd. But, It was still set ReadOnly.
          self.m.powershell(
              'Set RW access',
              'Set-ItemProperty',
              args=[
                  '-Path', wim_path, '-Name', 'IsReadOnly', '-Value', '$false'
              ])
      # ensure that the destination exists
      dest = self._workdir.join('mount')
      self.m.file.ensure_directory('Ensure mount point', dest)
      # Mount the boot.wim to mount dir for modification
      mount_wim.mount_win_wim(self.m.powershell, dest, wim_path, index,
                              self.m.path['cleanup'])

  def deinit_win_pe_image(self, save=True):
    """ deinit_win_pe_image unmounts the winpe image and saves/discards changes
    to it

    Args:
      * save: bool to determine if we need to save the changes to this image.
    """
    with self.m.step.nest('Deinit WinPE image modification'):
      if save:
        # copy the config used for building the image
        source = self._configs.join('{}.cfg'.format(self.get_key()))
        self.execute_script(
            'Add cfg {}'.format(source),
            ADDFILE,
            self._configs,
            self._workdir.join('mount'),
            '{}.cfg'.format(self.get_key()),
            logs=None,
            ret_codes=[0, 1])
      unmount_wim.unmount_win_wim(
          self.m.powershell,
          self._workdir.join('mount'),
          self._scratchpad,
          save=save)
      if save:
        with self.m.step.nest('Upload the output of {}'.format(self.name())):
          # There is only one output for offine winpe build
          def_dest = self.outputs[0]
          # upload the output to default bucket for offline_winpe_customization
          self._source.upload_package(def_dest, self._workdir)
          # upload to any custom destinations that might be given
          cust = self._customization.offline_winpe_customization
          for image_dest in cust.image_dests:
            # update the link to the original upload.
            image_dest.tags['orig'] = def_dest.tags['orig']
            self._source.upload_package(image_dest, self._workdir)

  def perform_winpe_action(self, action):
    """ perform_winpe_action Performs the given action

    Args:
      * action: actions.Action proto object that specifies an action to be
      performed
    """
    a = action.WhichOneof('action')
    if a == 'add_file':
      return self.add_file(action.add_file)

    if a == 'add_windows_package':
      src = self._source.get_local_src(action.add_windows_package.src)
      return self.add_windows_package(action.add_windows_package, src)

    if a == 'add_windows_driver':
      src = self._source.get_local_src(action.add_windows_driver.src)
      return self.add_windows_driver(action.add_windows_driver, src)

    if a == 'edit_offline_registry':
      return regedit.edit_offline_registry(
          self.m.powershell, self._scripts('WindowsPowerShell\Scripts'),
          action.edit_offline_registry, self._workdir.join('mount'))

  def perform_winpe_actions(self, offline_action):
    """ perform_winpe_actions Performs the given offline_action

    Args:
      * offline_action: actions.OfflineAction proto object that needs to be
      executed
    """
    for a in offline_action.actions:
      self.perform_winpe_action(a)

  def add_windows_package(self, awp, src):
    """ add_windows_package runs Add-WindowsPackage command in powershell.
    https://docs.microsoft.com/en-us/powershell/module/dism/add-windowspackage?view=windowsserver2019-ps

    Args:
      * awp: actions.AddWindowsPackage proto object
      * src: Path to the package on bot disk
    """
    add_windows_package.install_package(
        self.m.powershell, self._scripts('WindowsPowerShell\Scripts'), awp, src,
        self._workdir.join('mount'), self._scratchpad)

  def add_file(self, af):
    """ add_file runs Copy-Item in Powershell to copy the given file to image.
    https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.management/copy-item?view=powershell-5.1

    Args:
      * af: actions.AddFile proto object
    """
    # src contains the path for the src dir
    src = self._source.get_local_src(af.src)
    # src_file contains the file/dir name to be copied
    src_file = '*'
    if not self.m.path.isdir(src):
      # if the src is a file then src is the dir name and src_file is filename
      src_file = self.m.path.basename(src)
      src = self.m.path.dirname(src)
    # destination to copy the file to
    dest = '"{}"'.format(self._workdir.join('mount', af.dst))
    self.execute_script(
        'Add file {}'.format(self._source.get_url(af.src)),
        ADDFILE,
        src,
        dest,
        src_file,
        '/e',
        logs=None,
        ret_codes=[0, 1, 2, 3])

  def add_windows_driver(self, awd, src):
    """ add_windows_driver runs Add-WindowsDriver command in powershell.
    https://docs.microsoft.com/en-us/powershell/module/dism/add-windowsdriver?view=windowsserver2019-ps

    Args:
      * awd: actions.AddWindowsDriver proto object
      * src: Path to the driver on bot disk
    """
    add_windows_driver.install_driver(
        self.m.powershell, self._scripts('WindowsPowerShell\Scripts'), awd, src,
        self._workdir.join('mount'), self._scratchpad)
