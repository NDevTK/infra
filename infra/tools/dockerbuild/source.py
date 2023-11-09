# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Manages the raw sources needed to build wheels.

A source has a remote (public) address. That file is then downloaded and cached
locally as a CIPD package in the "infra/third_party/source" CIPD tree.

Systems that want to operate on a source reference it by a source constant, or
by a constructor (e.g., Pip).
"""

import collections
import contextlib
import hashlib
import os
import shutil
import subprocess
import tarfile
import tempfile
import zipfile

from . import cipd
from . import concurrency
from . import util


class Source(
    collections.namedtuple(
        'Source',
        (
            'name',  # The name of the source.
            'version',  # The version of the source.
            'download_type',  # Type of download function to use.
            'download_meta',  # Arbitrary metadata to pass to download function.
            'patches',  # Short patch names to apply to the source tree.
            'patch_base',  # Base name of patches, defaults to 'name'.
        ))):

  # A registry of all created Source instances.
  _REGISTRY = {}

  def __new__(cls, *args, **kwargs):
    kwargs.setdefault('patches', ())
    if not kwargs.get('patch_base'):
      kwargs['patch_base'] = kwargs['name']
    src = super(Source, cls).__new__(cls, *args, **kwargs)

    src._patches_hash = None

    # Register source with "_REGISTRY" and enforce that any source with the same
    # (name, version) is defined exactly the same.
    #
    # NOTE: If this expectation is ever violated, we will need to update CIPD
    # source package naming to incorporate the difference.
    key = (src.name, src.version)
    current = cls._REGISTRY.get(key)
    if not current:
      cls._REGISTRY[key] = src
    elif current != src:
      raise ValueError('Incompatible source definitions (%r != %r)' % (
          current, src))

    return src

  @classmethod
  def all(cls):
    return cls._REGISTRY.values()

  @property
  def tag(self):
    """A tag that identifies back to the upstream package."""
    return '%s-%s' % (self.name, self.version)

  @property
  def buildid(self):
    """Uniquely identifies this package & local patches."""
    ret = self.version
    srchash = self.patches_hash
    if srchash:
      ret += '-' + srchash
    return ret

  @property
  def patches_hash(self):
    """Return a hash of all the patches applied to this source."""
    if self._patches_hash is None:
      self._patches_hash = ''
      patches = self.get_patches()
      if patches:
        hash_obj = hashlib.md5()
        for patch in sorted(patches):
          with open(patch, 'rb') as f:
            hash_obj.update(f.read())
        self._patches_hash = hash_obj.hexdigest().lower()

    return self._patches_hash

  def get_patches(self):
    """Return list of patches (full paths) to be applied."""
    return [
        os.path.join(util.PATCHES_DIR,
                     '%s-%s-%s.patch' % (self.patch_base, self.version, x))
        for x in self.patches
    ]


class Repository(object):

  # Map of "download_type" to download function. Mapping will be made
  # later as functions are defined.
  _DOWNLOAD_MAP = {}

  def __init__(self, system, workdir, upload=False, force_download=False):
    self._system = system
    self._root = workdir
    self._upload = upload
    self._force_download = force_download

    # Will be set to True if a source was encountered without a corresponding
    # CIPD package, but not uploaded to CIPD.
    self._missing_sources = False

    # Build our archive suffixes.
    self._archive_suffixes = collections.OrderedDict()
    for suffix in ('.tar.gz', '.tgz', '.tar.bz2'):
      self._archive_suffixes[suffix] = self._unpack_tar_generic
    for suffix in ('.zip',):
      self._archive_suffixes[suffix] = self._unpack_zip_generic

    self.lock = concurrency.KeyedLock()
    self._deployed_packages = set()

  @property
  def missing_sources(self):
    return self._missing_sources

  def _ensure_package_deployed(self, src):
    # TODO: rewrite this to use cipd ensure
    #
    # Lock around accesses to the shared self._root directory, which is a
    # central cache for source packages which may be used between different
    # wheel builds.
    with self.lock.get(src.tag):
      package_dest = os.path.join(self._root, src.tag)
      if src.tag in self._deployed_packages:
        return package_dest
      if os.path.exists(package_dest):
        util.removeall(package_dest)  # Clean up any old state.
      os.makedirs(package_dest)

      # If the package doesn't exist, or if we are forcing a download, create a
      # local package.
      package_path = os.path.join(self._root, '%s.pkg' % (src.tag,))
      have_package = False
      if os.path.isfile(package_path):
        # Package file is on disk, reuse unless we're forcing a download.
        if not self._force_download:
          have_package = True

      # Check if the CIPD package exists.
      package = cipd.Package(
          name=cipd.normalize_package_name('infra/third_party/source/%s' %
                                           (src.name,)),
          tags=('version:%s' % (src.version,),),
          install_mode=cipd.INSTALL_SYMLINK,
          compress_level=cipd.COMPRESS_NONE,
      )

      # By default, assume the cached source package exists and try to download
      # it. If this produces an error we infer that it doesn't exist and go
      # create and upload it. This saves a call to CIPD compared to making an
      # explicit check up-front.
      cipd_exists = True
      if not have_package:
        # Don't even try downloading it if force_download is set.
        if not self._force_download:
          try:
            self._system.cipd.fetch_package(package.name, package.tags[0],
                                            package_path)
          except self._system.SubcommandError as e:
            # The CIPD command line tool returns 1 for all errors, so we're
            # forced to just check its stdout.
            if e.returncode == 1 and ('no such tag' in e.output or
                                      'no such package' in e.output):
              cipd_exists = False
            else:
              raise

        if not cipd_exists or self._force_download:
          self._create_cipd_package(package, src, package_path)

        have_package = True

      # We must have acquired the package at "package_path" by now.
      assert have_package

      # If we built a CIPD package, upload it. This will be fatal if we could
      # not perform the upload; if a user wants to not care about this, set
      # "upload" to False.
      if not cipd_exists:
        if self._upload:
          self._system.cipd.register_package(package_path, package.tags)
          util.LOGGER.info('Uploaded CIPD source package')
        else:
          self._missing_sources = True
          util.LOGGER.warning('Missing CIPD source package, but not uploaded.')

      # Install the CIPD package into our source directory. This is a no-op if
      # it is already installed.
      self._system.cipd.deploy_package(package_path, package_dest)
      self._deployed_packages.add(src.tag)
      return package_dest

  def ensure(self, src, dest_dir, unpack=True, unpack_file_filter=None):
    util.LOGGER.debug('Ensuring source %r', src.tag)

    package_dest = self._ensure_package_deployed(src)

    # The package directory should contain exactly one file.
    package_files = [
        f for f in os.listdir(package_dest) if not f.startswith('.')
    ]
    if len(package_files) != 1:
      raise ValueError('Package contains %d (!= 1) files: %s' %
                       (len(package_files), package_dest))
    package_file = package_files[0]
    package_file_path = os.path.join(package_dest, package_file)

    # The same destination path must not be accessed concurrently, so we
    # do not need to lock around the unpack step.

    # Unpack or copy the source file into the destination path.
    if unpack:
      for suffix, unpack_func in self._archive_suffixes.items():
        if package_file.endswith(suffix):
          # unpack_file_filter is a workaround for python < 3.6 on Windows.
          # Windows only allow path less than 260, which means we need to
          # filter some files from the package if they are not required in the
          # build to avoid triggering the limitation. This can be removed after
          # we migrate to Python 3.6 or later.
          # https://docs.microsoft.com/en-us/windows/win32/fileio/maximum-file-path-limitation
          return self._unpack_archive(package_file_path, dest_dir, unpack_func,
                                      unpack_file_filter)

    # Single file.
    dest = os.path.join(dest_dir, os.path.basename(package_file))
    util.LOGGER.debug('Installing source from [%s] => [%s]', package_file, dest)
    with concurrency.PROCESS_SPAWN_LOCK.shared():
      shutil.copyfile(package_file_path, dest)
    return dest

  def _create_cipd_package(self, package, src, package_path):
    # Download to a temporary file.
    with self._system.temp_subdir(src.tag) as tdir:
      download_dir = util.ensure_directory(tdir, 'download')
      package_dir = util.ensure_directory(tdir, 'package')

      path = os.path.join(download_dir, 'file')
      util.LOGGER.debug('Downloading source to: [%s]', path)
      with concurrency.PROCESS_SPAWN_LOCK.shared():
        with open(path, 'wb') as fd:
          filename = self._DOWNLOAD_MAP[src.download_type](fd,
                                                           src.download_meta)
        # Move the downloaded "file" into the package under its download name
        # and package it.
        os.rename(path, os.path.join(package_dir, filename))

      self._system.cipd.create_package(package, package_dir, package_path)

  def _unpack_archive(self, path, dest_dir, unpack_func, file_filter):
    with self._system.temp_subdir(os.path.basename(
        path)) as tdir, concurrency.PROCESS_SPAWN_LOCK.shared():
      unpack_func(path, tdir, file_filter)

      contents = os.listdir(tdir)
      if len(contents) != 1:
        raise ValueError('Archive contained %d (!= 1) file(s)' % (
            len(contents),))

      archive_base = os.path.join(tdir, contents[0])
      dest = os.path.join(dest_dir, os.path.basename(archive_base))
      os.rename(archive_base, dest)
    return dest

  @staticmethod
  def _unpack_tar_generic(path, dest_dir, file_filter):
    with tarfile.open(path, 'r') as tf:
      tf.extractall(
          dest_dir,
          members=filter(
              (lambda m: file_filter(m.name)) if file_filter else None,
              tf.getmembers(),
          ))

  @staticmethod
  def _unpack_zip_generic(path, dest_dir, file_filter):
    with zipfile.ZipFile(path, 'r') as zf:
      zf.extractall(dest_dir, members=filter(file_filter, zf.namelist()))


def remote_file(name, version, url):
  return Source(
    name=name,
    version=version,
    download_type='url',
    download_meta=url,
  )


def remote_archive(name, version, url):
  return Source(
    name=name,
    version=version,
    download_type='url',
    download_meta=url,
  )


def _download_url(fd, meta):
  url = meta
  return util.download_to(url, fd)
Repository._DOWNLOAD_MAP['url'] = _download_url


def _download_pypi_archive(fd, meta):
  name, version = meta

  url = 'https://pypi.org/pypi/%s/%s/json' % (name, version)
  content = util.download_json(url)
  release = content.get('urls', {})
  if not release:
    raise ValueError('No urls for package %r at version %r' % (name, version))

  entry = None
  for entry in release:
    if entry.get('packagetype') == 'sdist':
      break
  else:
    raise ValueError('No PyPi source distribution for package %r at version '
                     '%r' % (name, version))

  hash_obj = None
  expected_hash = entry.get('md5')
  if expected_hash:
    hash_obj = hashlib.md5()

  url = entry['url']
  util.LOGGER.debug('Downloading package %r @ %r from PyPi: %s',
                    name, version, url)
  filename = util.download_to(url, fd, hash_obj=hash_obj)

  if hash_obj:
    download_hash = hash_obj.hexdigest().lower()
    if download_hash != expected_hash:
      raise ValueError("Download hash %r doesn't match expected hash %r." % (
          download_hash, expected_hash))

  return filename
Repository._DOWNLOAD_MAP['pypi_archive'] = _download_pypi_archive


def _download_local(fd, meta):
  basename = os.path.basename(meta)
  with tarfile.open(mode='w:bz2', fileobj=fd) as tf:
    tf.add(meta, arcname=basename)
  return '%s.tar.bz2' % (basename,)


def local_directory(name, version, path):
  return Source(
      name=name,
      version=version,
      download_type='local_directory',
      download_meta=path)
Repository._DOWNLOAD_MAP['local_directory'] = _download_local


def pypi_sdist(name, version, patches=(), patch_base=None):
  """Defines a Source whose remote data is a PyPi source distribution."""

  return Source(
      name=name,
      version=version,
      download_type='pypi_archive',
      download_meta=(name, version),
      patches=patches,
      patch_base=patch_base,
  )
