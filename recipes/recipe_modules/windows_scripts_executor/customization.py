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
               cipd, git, gcs):
    """ __init__ copies common module objects to class references. These are
        commonly used for all customizations
        Args:
          arch: string representing architecture to build the image for
          scripts: path to the scripts resource dir
          step: module object for recipe_engine/step
          path: module object for recipe_engine/path
          powershell: module object for recipe_modules/powershell
          m_file: module object for recipe_engine/file
          cipd: module object for cipd_manager
          git: module object for git_manager
          gcs: module object for gcs_manager
    """
    self._arch = arch
    self._scripts = scripts
    self._step = step
    self._path = path
    self._powershell = powershell
    self._cipd = cipd
    self._git = git
    self._gcs = gcs
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

  def record_package(self, src):
    """ record_package records the given source for download
        Args:
          src: sources.Src proto representing a file/folder to be used
    """
    if src:
      self._cipd.record_package(src)
      self._gcs.record_package(src)
      self._git.record_package(src)

  def get_local_src(self, src):
    """ get_local_src returns path on the bot to the referenced src
        Args:
          src: sources.Src proto representing a file/folder to be used
    """
    if src.WhichOneof('src') == 'cipd_src':
      return self._cipd.get_local_src(src)
    if src.WhichOneof('src') == 'git_src':
      return self._git.get_local_src(src)
    if src.WhichOneof('src') == 'local_src':  # pragma: no cover
      return src.local_src
    if src.WhichOneof('src') == 'gcs_src':
      return self._gcs.get_local_src(src)
    return ''

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
    src = self.get_local_src(af.src)
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
