# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
from PB.recipes.infra.windows_image_builder import sources
from . import helper


class GCSManager:
  """
    GCSManager is used to download required artifacts from google cloud storage
    and generate a pinned config for the downloaded artifacts. Also supports
    uploading artifacts to GCS.
  """

  def __init__(self, module, cache):
    """ __init__ copies few module objects and cache dir path into class vars

    Args:
      * module: module object with all dependencies
      * cache: path to cache file dir. Files from gcs will be saved here
    """
    self.m = module
    self._cache = cache
    # dict mapping unpinned url to pinned package
    self._pinned_srcs = {}
    # dict mapping downloaded package paths to package
    self._downloaded_srcs = {}
    # set containing all the existing packages
    self._existence = set()

  def pin_package(self, gcs_src):
    """ pin_package takes a gcs_src and returns a pinned gcs_src

    Args:
      * gcs_src: sources.GCSSrc type referring to a package
    """
    pkg_url = self.get_gs_url(gcs_src)
    if pkg_url in self._pinned_srcs:
      return self._pinned_srcs[pkg_url]  # pragma: no cover
    else:
      pin_url = self.get_orig(pkg_url)
      if pin_url:
        # This package was linked to another
        b, s = self.get_bucket_source(pin_url)
        gcs_src.bucket = b
        gcs_src.source = s
        # Add pkg_url as that's the only one we know that exists. Not the
        # resolved pin
        self._existence.add(pkg_url)
      self._pinned_srcs[pkg_url] = gcs_src
      return gcs_src

  def get_orig(self, url):
    """ get_orig goes through the metadata to determine original object and
    returns url for the original GCS object. See upload_packages

    Args:
      * url: string representing url that describes a gcs object
    """
    res = self.m.gsutil.stat(
        url,
        name='stat {}'.format(url),
        stdout=self.m.raw_io.output(),
        ok_ret='any')
    ret_code = res.exc_result.retcode
    if ret_code == 0:
      text = res.stdout.decode('utf-8')
      # return the given url if not pinned
      orig_url = url
      for line in text.split('\n'):
        if 'orig:' in line:
          orig_url = line.replace('orig:', '').strip()
      return orig_url
    return ''

  # TODO(anushruth): Cover this test path
  def exists(self, src):  #pragma: no cover
    """ exists returns True if the given ref exists on GCS

    Args:
      * src: sources.Src proto object to check for existence
    """
    url = self.get_gs_url(src)
    if url in self._existence:
      return True
    src_exists = self.get_orig(url) != ''
    if src_exists:
      self._existence.add(url)
    return src_exists

  def download_package(self, gcs_src):
    """ download_package downloads the given package if required and returns
    local_path to the package.

    Args:
      * gcs_src: source.GCSSrc object referencing a package
    """
    local_path = self.get_local_src(gcs_src)
    if not local_path in self._downloaded_srcs:
      self.m.gsutil.download(
          gcs_src.bucket,
          gcs_src.source,
          local_path,
          name='download {}'.format(self.get_gs_url(gcs_src)))
      self._downloaded_srcs[local_path] = gcs_src
    return local_path

  def get_local_src(self, gcs_src):
    """ get_local_src returns the path to the source on disk

    Args:
      * gcs_src: sources.GCSSrc proto object referencing an artifact in GCS
    """
    # source is usually given as a unix path. Ensure this works on windows
    source = gcs_src.source.split('/')
    return self._cache.join(gcs_src.bucket, *source)

  def get_gs_url(self, gcs_src):
    """ get_gs_url returns the gcs url for the given gcs src

    Args:
      * gcs_src: sources.GCSSrc proto object referencing an artifact in GCS
    """
    return 'gs://{}/{}'.format(gcs_src.bucket, gcs_src.source)

  def get_bucket_source(self, url):
    """ get_bucket_source returns bucket and source given gcs url

    Args:
      * url: gcs url representing a file on GCS
    """
    bs = url.replace('gs://', '')
    tokens = bs.split('/')
    bucket = tokens[0]
    source = bs.replace(bucket + '/', '')
    return bucket, source

  def upload_package(self, dest, source):
    """ upload_package uploads the contents of source on disk to dest.

    Args:
      * dest: dest.Dest proto object representing an upload location
      * source: path to the package on disk to be uploaded
    """
    self.stage_upload(dest, source)
    # upload the package to gcs
    self.m.gsutil.upload(
        self.get_local_src(dest.gcs_src),
        dest.gcs_src.bucket,
        dest.gcs_src.source,
        metadata=dest.tags,
        name='upload {}'.format(self.get_gs_url(dest.gcs_src)))

  def stage_upload(self, dest, source):
    """ stage_upload copies the contents to be uploaded to local cache.
    Doing this allows not having to do the upload, and still make the file
    available for downstream customizations.

    Args:
      * dest: dest.Dest proto object representing an upload location
      * source: path to the package on disk to be uploaded
    """
    local_path = self.get_local_src(dest.gcs_src)
    if local_path in self._downloaded_srcs:  # pragma: no cover
      # Staging already done.
      return
    if not self.m.path.exists(source):  # pragma: no cover
      raise self.m.step.StepFailure('Cannot find {}'.format(source))
    up_path = source
    if self.m.path.isdir(source):
      # package the dir to zip
      up_path = source.join('gcs.zip')
      self.m.archive.package(source).archive(
          'Package {} for upload'.format(source), up_path)
    try:
      # ensure that the folder exists
      self.m.file.ensure_directory(
          'Ensure cache path {}'.format(local_path),
          dest=self.m.path.dirname(local_path))
      # copy the package to cache
      self.m.file.copy('Copy {} to cache'.format(up_path), up_path, local_path)
      # Record the local copy to avoid downloads
      self._downloaded_srcs[local_path] = dest.gcs_src
      # Record existence of the src too
      self._existence.add(self.get_gs_url(dest.gcs_src))
    finally:
      if self.m.path.isdir(source):
        self.m.file.remove('Delete {}'.format(up_path), up_path)
