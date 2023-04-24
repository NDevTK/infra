# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
from . import customization
from . import helper

from PB.recipes.infra.windows_image_builder import windows_iso as winiso
from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
from PB.recipes.infra.windows_image_builder import sources as src_pb
from PB.recipes.infra.windows_image_builder import dest as dest_pb


class WinISOCustomization(customization.Customization):
  """ Windows iso customization support """

  def __init__(self, **kwargs):
    # register all the default args
    super(WinISOCustomization, self).__init__(**kwargs)
    # ensure that we have the correct customization
    assert self.customization().WhichOneof(
        'customization') == 'windows_iso_customization'
    # use custom work dir
    self._name = self.customization().windows_iso_customization.name
    self._workdir = self.m.path['cleanup'].join(self._name, 'workdir')
    self._scratchpad = self.m.path['cleanup'].join(self._name, 'scratchpad')
    self._canon_cust = None
    helper.ensure_dirs(self.m.file, [self._workdir, self._scratchpad])

  def pin_sources(self, ctx):
    """ pins the given config by replacing the sources in the customization

    Args:
      * ctx: dict containing the context for customization
    """
    wic = self.customization().windows_iso_customization
    if wic.base_image.WhichOneof('src'):
      wic.base_image.CopyFrom(self._source.pin(wic.base_image, ctx))
    if wic.boot_image.WhichOneof('src'):
      wic.boot_image.CopyFrom(self._source.pin(wic.boot_image, ctx))
    for x in wic.copy_files:
      x.artifact.CopyFrom(self._source.pin(x.artifact, ctx))

  def download_sources(self):
    """ download_sources downloads the sources in the given config to the
    disk """
    wic = self.customization().windows_iso_customization
    if wic.base_image.WhichOneof('src'):
      self._source.download(wic.base_image)
    if wic.boot_image.WhichOneof('src'):
      self._source.download(wic.boot_image)
    for x in wic.copy_files:
      self._source.download(x.artifact)

  def get_canonical_cfg(self):
    """ get_canonical_cfg returns canonical config after removing the name and
    dests

    Example:
      Given a config

        Customization{
          windows_iso_customization: WinISOImage{
            name: "windows_vanilla_gce"
            base_image: Src{...}
            boot_image: Src{...}
            uploads: [...]
            copy_files: [...]
          }
        }

      Returns config

        Customization{
          windows_iso_customization: WinISOImage{
            name: ""
            base_image: Src{...}
            boot_image: Src{...}
            copy_files: [...]
          }
        }
    """
    if not self._canon_cust:
      wic = self.customization().windows_iso_customization
      self._canon_cust = wib.Customization(
          windows_iso_customization=winiso.WinISOImage(
              base_image=wic.base_image,
              boot_image=wic.boot_image,
              copy_files=wic.copy_files,
          ),)
    return self._canon_cust

  @property
  def outputs(self):
    """ return the output(s) of executing this config. Doesn't guarantee that
    the output(s) exists"""
    if self.get_key():
      output = src_pb.GCSSrc(
          bucket='chrome-gce-images',
          source='WIB-ISO/{}.iso'.format(self.get_key()))
      return [
          dest_pb.Dest(
              gcs_src=output,
              tags={'orig': self._source.get_url(src_pb.Src(gcs_src=output))},
          )
      ]
    return None  # pragma: no cover

  @property
  def inputs(self):
    """ inputs returns the input(s) required for this customization.

    inputs here refer to any external refs that might be required for
    this customization
    """
    inputs = []
    wic = self.customization().windows_iso_customization
    if wic.base_image.WhichOneof('src'):
      inputs.append(wic.base_image)
    if wic.boot_image.WhichOneof('src'):
      inputs.append(wic.boot_image)
    for x in wic.copy_files:
      inputs.append(x.artifact)
    return inputs

  @property
  def context(self):
    """ context returns a dict containing the local_src id mapping to output
    src.
    """
    if not self.outputs:
      return {}  #pragma: nocover
    return {
        '{}-output'.format(self.id): self._source.dest_to_src(self.outputs[0])
    }

  def execute_customization(self):
    """ execute_customization generates the required iso image.
    """
    output = self.customization().windows_iso_customization
    with self.m.step.nest('Windows iso customization {}'.format(output.name)):
      try:
        # Use staging to unpack the iso for modification
        iso_dir = self._workdir.join('staging')
        boot = None
        # Copy base image
        if output.base_image:
          self.copy_base_image(output.base_image, iso_dir)
        # Copy all the modifications
        for cf in output.copy_files:
          self.copy_files_to_image(cf, iso_dir)
        if output.boot_image.WhichOneof('src'):
          src = self._source.download(output.boot_image)
          # Copy the boot_image to /boot dir
          self.m.file.copy('Add {}'.format(src), src, iso_dir.join('boot'))
          boot = iso_dir.join('boot', self.m.path.basename(src))
        output_image = iso_dir.join(output.name + '.iso')
        # package everything into an iso
        self.generate_iso_image(
            output.name, boot=boot, directory=iso_dir, output=output_image)
        # Upload the image
        self._source.upload_package(self.outputs[0], output_image)
        for package in output.uploads:
          package.tags['orig'] = self.outputs[0].tags['orig']
          self._source.upload_package(package, output_image)
      finally:
        # cleanup everything
        with self.m.step.nest('Cleanup customization') as n:
          self.m.file.rmcontents('clean workdir', self._workdir)
          self.m.file.rmcontents('clean scratchpad', self._scratchpad)

  def copy_base_image(self, base_image, iso_staging):
    """ copy_base_image mounts the given iso image and copies the contents to
    given iso_staging dir.

    Args:
      * base_image: sources.Src proto object representing the iso image
      * iso_staging: dir path where we stage the iso to be packaged
    """
    url = self._source.get_url(base_image)
    with self.m.step.nest('Copy {} to staging'.format(url)) as n:
      base_image_path = self._source.get_local_src(base_image)
      loop, mount_loc = self.m.qemu.mount_disk_image(
          self._source.get_local_src(base_image), partitions=None)
      try:
        n.presentation.logs['mount_path'] = mount_loc[0]
        n.presentation.logs['loop'] = loop
        # Copy the base image to the staging dir (iso_dir)
        self.m.file.copytree('Copy base image', mount_loc[0], iso_staging)
        # default permissions are 0555
        self.m.step(
            'Set permissions for base image',
            cmd=['chmod', '0755', '-Rv', iso_staging])
      finally:
        self.m.qemu.unmount_disk_image(loop, partitions=None)

  def copy_files_to_image(self, cf, iso_staging):
    """ copy_files_to_image copies the given file to the image.

    If the given file is an archive, extracts the archive contents. If mount
    is set and there is only one file in the archive or input is just one file
    (that is the image to mount), Mounts the image and then copies the file
    at given source. Otherwise just copies the given file to the destination.

    Args:
      * cf: windows_iso.CopyFile proto object representing the file
      * iso_staging: location to copy the file to
    """
    src = self._source.get_local_src(cf.artifact)
    src_url = self._source.get_url(cf.artifact)
    partitions = None
    dest = iso_staging.join(cf.dest)
    loop = ''
    with self.m.step.nest('Copy {} to staging'.format(src_url)) as n:
      # If src is an archive of sorts
      if str(src).endswith('.zip') or str(src).endswith('.tar'):
        # If this is a archive. Extract the required file to scratchpad
        self.m.archive.extract(
            'Unpack {} to {}'.format(src, self._scratchpad),
            archive_file=src,
            output=self._scratchpad)
        # In the situation that this archive contains an image to extract. Let's
        # figure out the path for the image
        if cf.mount:
          contents = self.m.file.listdir(
              name='Checking the contents of {}'.format(src),
              source=self._scratchpad,
              recursive=False,
              test_data=['image.file'])
          if (len(contents) == 0):  #pragma: nocover
            raise self.m.step.StepFailure('Nothing was extracted to mount')
          elif (len(contents) > 1):  #pragma: nocover
            raise self.m.step.StepFailure('Cannot determine what to mount')
          else:
            image = contents[0]
            partitions = None if str(image).endswith('iso') else [1]
            return self.mount_copy(image, cf.source, dest, partitions)
        else:
          src = self._scratchpad.join(cf.source)
          # Copy the given file as is to the target location
          return self.copy(src, dest)
      if cf.mount:
        # If we are copying from an iso. There are no partitions
        partitions = None if str(src).endswith('iso') else [1]
        return self.mount_copy(src, cf.source, dest, partitions)
      # Append source if given
      if cf.source:
        src = src.join(cf.source)
      # Copy the given file as is to the target location
      return self.copy(src, dest)

  def mount_copy(self, image, source, dest, partitions=None):
    """ mount_copy mounts the give image[partition] and copies the file given
    by source to dest

    Args:
      * image: path to the image
      * source: file/dir to copy in the said image
      * dest: location to copy source to
      * partitions: If the image is a disk image then partitions to mount
    """
    loop, mount_loc = self.m.qemu.mount_disk_image(image, partitions=partitions)
    try:
      # Copy the file from mounted location
      src = mount_loc[0] + '/' + source
      self.copy(src, dest)
    finally:
      self.m.qemu.unmount_disk_image(loop, partitions=partitions)
    return

  def copy(self, src, dest):
    """ copy is a helper function that unifies the action of copying dir or file

    Args:
      * src: path to the file to be copied
      * dest: destination to copy the file to
    """
    if self.m.path.isdir(src):
      self.m.file.copytree('Copy {} to {}'.format(src, dest), src, dest)
    else:
      self.m.file.copy('Copy {} to {}'.format(src, dest), src, dest)

  def generate_iso_image(self, name, boot, directory, output):
    """ generate_iso_image creates an iso image (output) from the given
    directory

    Args:
      * name: The name of the image
      * boot: The bootloader for the iso image
      * directory: The staging directory for the image
      * output: The output image to be generated
    """
    if not boot:
      # If no boot image is given, use the windows default
      boot = 'boot/etfsboot.com'
    # For info regarding el-torito boot image and genisoimage
    # See: https://wiki.osdev.org/El-Torito
    # See: https://wiki.osdev.org/Genisoimage
    cmd = [
        'genisoimage',  # use genisoimage to generate the iso image
        '-b{}'.format(boot),  # use this as the bootloader image
        '-no-emul-boot',  # dont emulate boot image as floppy
        '-boot-load-seg',
        '0',  # load segment address for boot image
        '-boot-load-size',
        '8',  # size of the bios image
        '-iso-level',
        '2',  # Use level 2
        '-J',  # Generate Joliet directory records
        '-l',  # Allow full 31 character filenames
        '-D',  # disable deep directory relocation
        '-N',  # omit version numbers
        '-joliet-long',  # Allow joliet filenames to be 103 unicode chars
        '-allow-limited-size',  # support files larger than 2 GB
        '-relaxed-filenames',  # allow filenames to include all 7-bit ASCII
        '-V',
        '{}'.format(name),  # name for the iso being generated
        '-o',
        output,  # write to this file
        directory  # directory to be used for this image
    ]
    self.m.step('Generate iso image', cmd=cmd)
