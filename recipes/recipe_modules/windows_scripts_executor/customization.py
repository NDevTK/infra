# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

COPYPE = 'Copy-PE.ps1'
ADDFILE = 'Copy-Item'

from . import add_windows_package


class Customization(object):
  """ Base customization class. Provides support for pinning and executing
      recipes.
  """

  def __init__(self, arch, scripts, configs, step, path, powershell, m_file,
               source):
    """ __init__ copies common module objects to class references. These are
        commonly used for all customizations
        Args:
          arch: string representing architecture to build the image for
          scripts: path to the scripts resource dir
          step: module object for recipe_engine/step
          path: module object for recipe_engine/path
          powershell: module object for recipe_modules/powershell
          m_file: module object for recipe_engine/file
          source: module object for Source from sources.py
    """
    self._arch = arch
    self._scripts = scripts
    self._step = step
    self._path = path
    self._powershell = powershell
    self._source = source
    self._file = m_file
    self._key = ''
    self._configs = configs
    self._name = ''

  def name(self):
    """ name returns the name of the customization object. This needs to be
        set by the inheriting class"""
    return self._name

  def set_key(self, key):
    """ set_key is used to set the identification keys for the customization
        Args:
          key: string representing the unique key for this customization
    """
    self._key = key

  def add_windows_package(self, awp, src):
    """ add_windows_package runs Add-WindowsPackage command in powershell.
        https://docs.microsoft.com/en-us/powershell/module/dism/add-windowspackage?view=windowsserver2019-ps
        Args:
          awp: actions.AddWindowsPackage proto object
          src: Path to the package on bot disk
    """
    add_windows_package.install_package(self._powershell, awp, src,
                                        self._workdir.join('mount'),
                                        self._scratchpad)

  def add_file(self, af):
    """ add_file runs Copy-Item in Powershell to copy the given file to image.
        https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.management/copy-item?view=powershell-5.1
        Args:
          af: actions.AddFile proto object
    """
    src = self._source.get_local_src(af.src)
    if self._path.isdir(src):
      src.join('*')  # pragma: no cover
    self.execute_script('Add file {}'.format(src), ADDFILE, None, '-Path', src,
                        '-Recurse', '-Force', '-Destination',
                        self._workdir.join('mount', af.dst))

  def execute_script(self, name, command, logs=None, *args):
    """ Executes the windows powershell script
        Args:
          name: string representing step name
          command: string|path representing command to be run
          logs: list of strings representing log files/folder to be read
          args: args to be passed on to the command
    """
    return self._powershell(name, command, logs=logs, args=list(args))
