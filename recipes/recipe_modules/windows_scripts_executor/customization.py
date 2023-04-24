# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
from . import add_windows_package
from . import add_windows_driver

from PB.recipes.infra.windows_image_builder import windows_image_builder as wib

NAME_SEP = '-'


class Customization(object):
  """ Base customization class. Provides support for pinning and executing
      recipes.
  """

  def __init__(self, image, cust, arch, scripts, configs, module, source):
    """ __init__ copies common module objects to class references. These are
    commonly used for all customizations

    Args:
      * image: wib.Image proto object
      * cust: wib.Customization proto object
      * arch: string representing architecture to build the image for
      * scripts: path to the scripts resource dir
      * configs: dir to store configs in
      * module: module object with all dependencies
      * source: module object for Source from sources.py
    """
    # generate a copy of image
    self._image = wib.Image()
    self._image.CopyFrom(image)
    self._image.ClearField('customizations')
    # generate a copy of customization
    self._customization = wib.Customization()
    self._customization.CopyFrom(cust)
    # Add customization to the image
    self._image.customizations.append(self._customization)
    self._arch = arch
    self._scripts = scripts
    self.m = module
    self._source = source
    self._key = ''
    self._configs = configs
    self._name = ''

  def name(self):
    """ name returns the name of the customization object. This needs to be
    set by the inheriting class"""
    return self._name

  def customization(self):
    """customization returns wib.Customization object representing self"""
    return self._customization

  def image(self):
    """ image returns wib.Image object representing self"""
    return self._image

  @property
  def id(self):
    """ id returns the identifier for this customization"""
    return NAME_SEP.join(
        ['image({})'.format(self.image().name), 'cust({})'.format(self._name)])

  @property
  def type(self):
    """ type returns the customization type for self (customization) """
    return self.customization().WhichOneof('customization')

  @property
  def arch(self):
    """ arch returns the targeted arch for the self (customization) """
    return self.image().arch

  @property
  def needs_build(self):
    """ needs_build returns True if this customization needs to be built.
    False otherwise """
    # Override the check if we need to force the build.
    if self.mode == wib.CustomizationMode.CUST_FORCE_BUILD:
      return True
    for output in self.outputs:
      # If all the outputs exist. We need not build the image
      if output and not self._source.exists(output):
        return True
    return False

  @property
  def mode(self):
    """ mode returns the mode that this customization is set to. """
    return self._customization.mode

  def executable(self, inputs=()):
    """ executable returns true if the customization is executable false
    otherwise.

    executable ensures that at least one of the outputs that is expected from
    this customization doesn't exist (No need to execute the customization if
    all the outputs already exist) and all the inputs required to execute the
    customization are available.

    Args:
      * inputs: list of urls for inputs that will be available at time of
      execution
    """
    if self.needs_build:
      # check if we can build the image
      for ip in self.inputs:
        if self._source.get_url(
            ip) not in inputs and not self._source.exists(ip):
          # still waiting on some input. Cannot build now
          return False
      # we have all the inputs for the image. We can build
      return True
    # image doesn't need building
    return False

  def pinnable(self, ctx):
    """ pinnable determines if the customization is pinnable
    Args:
      * ctx: dict containing the context for pinning

    Returns True if the customization is pinnable. False otherwise
    """
    for ip in self.inputs:
      if ip.WhichOneof('src') == 'local_src' and ip.local_src not in ctx:
        return False
    return True

  def unpinnable(self, ctx):
    """ unpinnable determines if the customization is pinnable. And returns
    srcs that it cannot pin
    Args:
      * ctx: dict containing the context for pinning

    Returns list of srcs that cannot be pinned. Empty list otherwise
    """
    unpins = []
    for ip in self.inputs:
      if ip.WhichOneof('src') == 'local_src' and ip.local_src not in ctx:
        unpins.append(ip)
    return unpins

  def set_key(self, key):
    """ set_key is used to set the identification keys for the customization
    Args:
      * key: string representing the unique key for this customization
    """
    self._key = key

  def get_key(self):
    """ get_key returns the hash associated with the customization
    """
    return self._key

  def execute_script(self, name, command, *args, **kwargs):
    """ Executes the windows powershell script

    Args:
      * name: string representing step name
      * command: string|path representing command to be run
      * args: args to be passed on to the command
      * kwargs: logs and ret_codes, logs ([]str) are list os paths to watch
                and record logs from. ret_codes ([]int) is a list of ints,
                these will be treated as success return codes upon execution
    """
    logs = kwargs['logs']
    ret_codes = kwargs['ret_codes']
    return self.m.powershell(
        name, command, logs=logs, ret_codes=ret_codes, args=list(args))
