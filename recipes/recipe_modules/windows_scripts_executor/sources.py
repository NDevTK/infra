# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
from . import cipd_manager
from . import git_manager
from . import gcs_manager
from . import helper

from PB.recipes.infra.windows_image_builder import sources as src_pb
from PB.recipes.infra.windows_image_builder import dest as dest_pb


class Source:
  """ Source handles all the pinning, downloading, and uploading artifacts to
      Git, GCS and CIPD repositories. See (git|gcs|cipd)_manager.py for
      implementation details on how each is handled.
  """

  def __init__(self, cache, module):
    """ __init__ generates the src managers (git, gcs and cipd) and stores them
    in class variables.

    Args:
      * cache: path to dir that can be used to download artifacts
      * module: module object with all dependencies
    """
    self.m = module
    # dir to store CIPD downloaded packages
    cipd_dir = cache.join('CIPDPkgs')
    # dir to store GIT downloaded packages
    git_dir = cache.join('GITPkgs')
    # dir to store GCS downloaded packages
    gcs_dir = cache.join('GCSPkgs')
    helper.ensure_dirs(self.m.file, [
        cipd_dir, git_dir, gcs_dir,
        gcs_dir.join('chrome-gce-images', 'WIB-WIM')
    ])
    self._cipd = cipd_manager.CIPDManager(module, cipd_dir)
    self._gcs = gcs_manager.GCSManager(module, gcs_dir)
    self._git = git_manager.GITManager(module, git_dir)
    self.cache = cache

  def pin(self, src):
    """ pin pins all the recorded packages to static refs

    Args:
      * src: sources.Src proto object that contains ref to an artifact
    """
    if src and src.WhichOneof('src') == 'git_src':
      src.git_src.CopyFrom(self._git.pin_package(src.git_src))
      return src
    if src and src.WhichOneof('src') == 'gcs_src':
      src.gcs_src.CopyFrom(self._gcs.pin_package(src.gcs_src))
      return src
    if src and src.WhichOneof('src') == 'cipd_src':
      src.cipd_src.CopyFrom(self._cipd.pin_package(src.cipd_src))
      return src

  def download(self, src):
    """ download downloads all the pinned packages to cache on disk

    Args:
      * src: sources.Src proto object that contains ref to an artifact
    """
    if src and src.WhichOneof('src') == 'git_src':
      return self._git.download_package(src.git_src)
    if src and src.WhichOneof('src') == 'gcs_src':
      return self._gcs.download_package(src.gcs_src)
    if src and src.WhichOneof('src') == 'cipd_src':
      return self._cipd.download_package(src.cipd_src)

  def get_local_src(self, src):
    """ get_local_src returns path on the disk that points to the given src ref

    Args:
      * src: sources.Src proto object that is ref to a downloaded artifact
    """
    if src and src.WhichOneof('src') == 'gcs_src':
      return self._gcs.get_local_src(src.gcs_src)
    if src and src.WhichOneof('src') == 'git_src':
      return self._git.get_local_src(src.git_src)
    if src and src.WhichOneof('src') == 'cipd_src':
      return self._cipd.get_local_src(src.cipd_src)
    if src and src.WhichOneof('src') == 'local_src':  # pragma: no cover
      return src.local_src

  def get_rel_src(self, src):
    """ get_rel_src returns relative path on the disk, relative to the cache
    directory.

    Args:
      * src: sources.Src proto object that is a ref to a downloaded artifact
    """
    local_src = self.get_local_src(src)
    return self.m.path.relpath(local_src, self.cache)

  def get_url(self, src):
    """ get_url returns string containing an url referencing the given src

    Args:
      * src: sources.Src proto object that contains ref to an artifact
    """
    if src and src.WhichOneof('src') == 'gcs_src':
      return self._gcs.get_gs_url(src.gcs_src)
    if src and src.WhichOneof('src') == 'cipd_src':
      return self._cipd.get_cipd_url(src.cipd_src)
    if src and src.WhichOneof('src') == 'git_src':  # pragma: no cover
      return self._git.get_gitiles_url(src.git_src)

  def upload_package(self, dest, source):
    """ upload_package uploads a given package to the given destination

    Args:
      * dest: dest_pb.Dest proto object representing the upload to be done
      * source: The contents of the package to be uploaded
    """
    if dest.WhichOneof('dest') == 'gcs_src':
      self._gcs.upload_package(dest, source)
    if dest and dest.WhichOneof('dest') == 'cipd_src':
      self._cipd.upload_package(dest, source)

  def exists(self, src):
    """ exists Returns True if the given src exists

    Args:
      * src: {sources.Src | dest.Dest} proto object representing an artifact
    """
    # TODO(anushruth): add support for git and cipd
    if isinstance(src, src_pb.Src):  #pragma: no cover
      if src.WhichOneof('src') == 'gcs_src':
        return self._gcs.exists(src.gcs_src)
    if isinstance(src, dest_pb.Dest):
      if src.WhichOneof('dest') == 'gcs_src':
        return self._gcs.exists(src.gcs_src)

    return False  # pragma: no cover
