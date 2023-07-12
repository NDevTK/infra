# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
from . import cipd_manager
from . import git_manager
from . import gcs_manager
from . import helper

from PB.recipes.infra.windows_image_builder import sources as src_pb
from PB.recipes.infra.windows_image_builder import dest as dest_pb


class SourceException(Exception):
  pass


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

  def pin(self, src, ctx):
    """ pin pins all the recorded packages to static refs

    Args:
      * src: sources.Src proto object that contains ref to an artifact
      * ctx: dict containing the context for local_src
    """
    try:
      if src and src.WhichOneof('src') == 'git_src':
        src.git_src.CopyFrom(self._git.pin_package(src.git_src))
        return src
      if src and src.WhichOneof('src') == 'gcs_src':
        src.gcs_src.CopyFrom(self._gcs.pin_package(src.gcs_src))
        return src
      if src and src.WhichOneof('src') == 'cipd_src':
        src.cipd_src.CopyFrom(self._cipd.pin_package(src.cipd_src))
        return src
    except Exception as e:
      raise SourceException('Cannot resolve {}: {}'.format(
          self.get_url(src), e))
    if src and src.WhichOneof('src') == 'local_src':  # pragma: no cover
      if src.local_src in ctx:
        src.CopyFrom(ctx[src.local_src])
        return src
      else:
        raise SourceException('Cannot resolve {}'.format(src.local_src))

  def download(self, src):
    """ download downloads all the pinned packages to cache on disk

    Args:
      * src: sources.Src proto object that contains ref to an artifact
    """
    try:
      if src and src.WhichOneof('src') == 'git_src':
        return self._git.download_package(src.git_src)
      if src and src.WhichOneof('src') == 'gcs_src':
        return self._gcs.download_package(src.gcs_src)
      if src and src.WhichOneof('src') == 'cipd_src':
        return self._cipd.download_package(src.cipd_src)
    except Exception as e:
      raise SourceException('Cannot download {}: {}'.format(
          self.get_url(src), e))

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

  # TODO(anushruth): Cover this test path
  def get_url(self, artifact):  #pragma: no cover
    """ get_url returns string containing an url referencing the given src

    Args:
      * artifact: sources.Src|dest.Dest proto object that contains ref to an
      artifact
    """
    if isinstance(artifact, dest_pb.Dest):
      if artifact and artifact.WhichOneof('dest') == 'gcs_src':
        return self._gcs.get_gs_url(artifact.gcs_src)
      if artifact and artifact.WhichOneof('dest') == 'cipd_src':
        return self._cipd.get_cipd_url(artifact.cipd_src)
    if isinstance(artifact, src_pb.Src):
      if artifact and artifact.WhichOneof('src') == 'gcs_src':
        return self._gcs.get_gs_url(artifact.gcs_src)
      if artifact and artifact.WhichOneof('src') == 'cipd_src':
        return self._cipd.get_cipd_url(artifact.cipd_src)
      if artifact and artifact.WhichOneof('src') == 'git_src':
        return self._git.get_gitiles_url(artifact.git_src)

    raise SourceException('Cannot get url for {}'.format(artifact))

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

  # TODO(anushruth): Cover this test path
  def exists(self, src):  # pragma: no cover
    """ exists Returns True if the given src exists

    Args:
      * src: {sources.Src | dest.Dest} proto object representing an artifact
    """
    if isinstance(src, src_pb.Src):
      if src.WhichOneof('src') == 'gcs_src':
        return self._gcs.exists(src.gcs_src)
      if src.WhichOneof('src') == 'cipd_src':
        return self._cipd.exists(src.cipd_src)
      if src.WhichOneof('src') == 'git_src':
        return self._git.exists(src.git_src)
    if isinstance(src, dest_pb.Dest):
      if src.WhichOneof('dest') == 'gcs_src':
        return self._gcs.exists(src.gcs_src)
      if src.WhichOneof('dest') == 'cipd_src':
        return self._cipd.exists(src.cipd_src, tags=src.tags)

    raise SourceException('Cannot determine if {} exists'.format(
        self.get_url(src)))

  def dest_to_src(self, dest):  # pragma: no cover
    """ dest_to_src returns a src_pb.Src object from dest.Dest object.

    Args:
      * dest: dest.Dest object representing an upload
    """
    if dest.WhichOneof('dest') == 'gcs_src':
      return src_pb.Src(gcs_src=dest.gcs_src)
    if dest.WhichOneof('dest') == 'cipd_src':  # pragma: no cover
      return src_pb.Src(cipd_src=dest.cipd_src)
