# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
from . import helper


GIT_URL = '{}/+/{}/{}'


class GITManager:
  """
    GITManager is used to download required artifacts from git and generate
    a pinned config for the downloaded artifacts. It uses gitiles to pin the
    sources to refs, but downloads them using git. This is done as image
    processor should be able to pin the sources without having to download
    them, as depending on the result of the pinning. We might not need to
    download them.
  """

  def __init__(self, module, cache):
    """ __init__ copies few module objects and cache dir path into class vars

    Args:
      * module: module object with all dependencies
      * cache: path to cache file dir. Files from git will be saved here
    """
    self.m = module
    self._cache = cache
    # cache stored as dict with url as key and pinned src as value
    self._pinning_cache = {}
    # cache stored as dict with url as key and downloaded src as value
    self._downloads_cache = {}
    # set of existing git urls.
    # Note: There is a remote possibility that file might be removed/deleted in
    # git after we record it's existence. This should still be okay as we are
    # checking for existence after pinning. If the file is deleted, it should
    # be part of a new commit.
    self._existence = set()

  def pin_package(self, git_src):
    """ pin_package replaces a volatile ref to deterministic ref in given
    git_src

    Args:
      * git_src: proto GITSrc object representing a volatol
    """
    url = self.get_gitiles_url(git_src)
    if url in self._pinning_cache:
      return self._pinning_cache[url]
    else:
      commits, _ = self.m.gitiles.log(git_src.repo,
                                      git_src.ref + '/' + git_src.src)
      if len(commits) == 0:
        raise Exception("No commits were found")
      # pin the file to the latest available commit
      git_src.ref = commits[0]['commit']
      self._pinning_cache[url] = git_src
      # as we know this file exists
      self._existence.add(url)
      return git_src

  def download_package(self, git_src):
    """ download_package downloads the given package to the downloads dir

    Args:
      * git_src: proto GITSrc object representing the artifact to be downloaded
    """
    with self.m.step.nest(
        'Download {}'.format(self.get_gitiles_url(git_src)), status='last'):
      local_path = self.get_local_src(git_src)
      if not local_path in self._downloads_cache:
        download_path = self._cache.join(git_src.ref)
        self.m.git.checkout(
            step_suffix=git_src.src,
            url=git_src.repo,
            dir_path=download_path,
            ref=git_src.ref,
            file_name=git_src.src)
        self._downloads_cache[local_path] = git_src
      return local_path

  def get_gitiles_url(self, git_src):
    """ get_gitiles_url returns string representing an url for the given source

    Args:
      * git_src: sources.GITSrc object representing an object
    """
    return GIT_URL.format(git_src.repo, git_src.ref, git_src.src)

  def get_local_src(self, git_src):
    """ get_local_src returns the location of the downloaded package in
    disk

    Args:
      * git_src: sources.GITSrc object
    """
    # src is usually given as a unix path. Ensure this works on windows too
    src = git_src.src.split('/')
    # ref can be given as an unix path too.
    ref = git_src.ref.split('/')
    # Add all the path names together
    ref.extend(src)
    return self._cache.join(*ref)

  # TODO(anushruth): Cover this test path
  def exists(self, git_src):  #pragma: no cover
    """ exists returns true if the package exists on gitiles.

    Args:
      *git_src: sources.GITSrc object representing a file in git

    Reurns True if the file exists False otherwise.
    """
    url = self.get_gitiles_url(git_src)
    if url in self._existence:
      return True
    try:
      commits, _ = self.m.gitiles.log(git_src.repo,
                                      git_src.ref + '/' + git_src.src)
      if commits:
        self._existence.add(url)
        return True
    except Exception:
      return False
    return False
