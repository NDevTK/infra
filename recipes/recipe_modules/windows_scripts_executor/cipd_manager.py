# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
from . import helper


CIPD_URL = 'https://chrome-infra-packages.appspot.com/p/{}/{}/+/{}'


class CIPDManager:
  """
  CIPDManager is used to parse through the configs and download all the
  cipd packages. It also modifies the config such that same packages will be
  downloaded next time the modified config is used.
  """

  def __init__(self, module, cache):
    """ __init__ copies common module objects and cache dir for downloading
    cipd artifacts to

    Args:
      * module: module object with all dependencies
      * cache: path to cache dir. CIPD artifacts will be downloaded to this dir
    """
    self.m = module
    # cache dir to be used to download the packages to
    self._cache = cache
    # dict mapping url to pinned srcs
    self._pinning_cache = {}
    # dict mapping local path to downloaded srcs
    self._downloads_cache = {}
    self._existence = set()

  def pin_package(self, cipd_src):
    """ pin_package pins the given package to instance identifier

    Args:
      * cipd_src: sources.CIPDSrc object representing an cipd package
    """
    unpinned_url = self.get_cipd_url(cipd_src)
    if not unpinned_url in self._pinning_cache:
      desc = self.m.cipd.describe(
          '/'.join([cipd_src.package, cipd_src.platform]), cipd_src.refs)
      # update the refs with the corresponding instance id
      cipd_src.refs = desc.pin.instance_id
      self._pinning_cache[unpinned_url] = cipd_src
      self._existence.add(self.get_cipd_url(cipd_src))
    # update the refs to the given src. This might be a different src from the
    # one that did the pinning (filename might be included/not included)
    pinned_ref = self._pinning_cache[unpinned_url].refs
    # update the ref
    cipd_src.refs = pinned_ref
    # return the updated cipd_src
    return cipd_src

  def download_package(self, cipd_src):
    """ download_package downloads the given package to disk if required.
    Returns local path on the disk

    Args:
      * cipd_src: sources.CIPDSrc object representing a cipd package
    """
    local_path = self.get_local_src(cipd_src)
    if not local_path in self._downloads_cache:
      e_file = self.m.cipd.EnsureFile()
      # cipd expects unix path
      loc = '/'.join([cipd_src.package, cipd_src.platform])
      self.add_src_to_ensurefile(cipd_src, loc, e_file)
      # add refs to root dir. This will ensure that the root dirs are unique.
      # cipd will delete older files if the root dir is same
      self.m.cipd.ensure(
          self._cache.join(cipd_src.refs),
          e_file,
          name="Download {}".format(self.get_cipd_url(cipd_src)))
      self._downloads_cache[local_path] = cipd_src
    return local_path

  def add_src_to_ensurefile(self, cipd_src, loc, ensure_file):
    """ add_src_to_ensurefile adds the given src to cipd ensurefile.

    Args:
      * cipd_src: sources.CIPDSrc proto object
      * loc: path to download the cipd artifact to
      * ensure_file: CIPD EnsureFile object. Used for downloading multiple
        instances in parallel
    """
    # Generate the complete package name
    pname = '/'.join([cipd_src.package, cipd_src.platform])
    # Add the package to the ensure file
    ensure_file.add_package(str(pname), str(cipd_src.refs), str(loc))

  def get_local_src(self, cipd_src):
    """ get_local_src returns local path to given cipd_src file

    Args:
      * cipd_src: sources.CIPDSrc object representing cipd package
    """
    # package is always provided with unix path. Ensure this will work on both
    # windows and linux.
    package = cipd_src.package.split('/')
    if cipd_src.platform:
      # platform is typically added to the package name in cipd.
      package.append(cipd_src.platform)
    if cipd_src.filename:
      # filename is typically added to the package name in cipd.
      package.append(cipd_src.filename)
    # return the deref
    return self._cache.join(cipd_src.refs, *package)

  def get_cipd_url(self, cipd_src):
    """ get_url returns string containing an url referencing the given src

    Args:
      * cipd_src: sources.CIPDSrc object representing an object
    """
    return CIPD_URL.format(cipd_src.package, cipd_src.platform, cipd_src.refs)

  def upload_package(self, dest, source):
    """ upload_package uploads the given package at source to dest.

    Args:
      * dest: dest.Dest proto object representing the object to be uploaded
      * source: path on the disk to the package to be uploaded
    """
    if self.m.path.exists(source):
      root, filename = self.m.path.split(source)
      name = '{}/{}'.format(dest.cipd_src.package, dest.cipd_src.platform)
      pkg = self.m.cipd.PackageDefinition(name, root)
      if self.m.path.isdir(source):
        pkg.add_dir(root.join(filename))
      else:
        pkg.add_file(root.join(filename))  # pragma: no cover
      self.m.cipd.create_from_pkg(
          pkg,
          refs=[dest.cipd_src.refs],
          tags=dest.tags,
          compression_level=0,
          verification_timeout='30m')

  # TODO(anushruth): Cover this test path
  def exists(self, cipd_src):  #pragma: no cover
    """ exists returns true if the package exists on cipd backend

    Args:
      * cipd_src: sources.CIPDSrc object representing a cipd package
    """
    url = self.get_cipd_url(cipd_src)
    if url in self._existence:
      return True
    try:
      self.m.cipd.describe(
          package_name='{}/{}'.format(cipd_src.package, cipd_src.platform),
          version=cipd_src.refs)
      self._existence.add(url)
      return True
    except Exception:
      return False
