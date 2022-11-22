# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import contextlib
import datetime
import errno
import logging
import os
import shutil
import stat
import tempfile
import urllib.parse
import urllib.request

import requests
import requests.exceptions

from . import concurrency

PKG_DIR = os.path.abspath(os.path.dirname(__file__))
RESOURCES_DIR = os.path.join(PKG_DIR, 'resources')
PATCHES_DIR = os.path.join(PKG_DIR, 'patches')

DOWNLOAD_CHUNK_SIZE = 4 * 1024 * 1024

LOGGER = logging.getLogger('dockerbuild')


def ensure_directory(*parts):
  name = os.path.join(*parts)
  try:
    os.makedirs(name)
  except OSError as e:
    # NOTE: Python 3 provides the 'exist_ok' keyword argument for os.makedirs,
    # which we can use instead of catching this exception. Or perhaps
    # pathlib.Path.mkdir.
    if e.errno != errno.EEXIST:
      raise

  return name


def copy_to(src, dest_dir):
  dst = os.path.join(dest_dir, os.path.basename(src))

  LOGGER.debug('Copy %r => %r', src, dst)

  with concurrency.PROCESS_SPAWN_LOCK.shared():
    if os.path.isdir(src):
      shutil.copytree(src, dst, symlinks=True)
    else:
      shutil.copy(src, dst)
  return dst


def removeall(path):
  # If removal fails, try changing mode and trying again. This is necessary for
  # deleting readonly files on Windows.
  # See https://docs.python.org/3.8/library/shutil.html#rmtree-example
  def remove_readonly(f, path, _):
    os.chmod(path, stat.S_IWRITE)
    f(path)

  if os.path.isfile(path):
    os.remove(path)
  else:
    shutil.rmtree(path, onerror=remove_readonly)


class NamedAnonymousFile(object):
  def __init__(self, fd, name):
    self._fd = fd
    self._name = name

  @property
  def name(self):
    return self._name

  def __getattr__(self, key):
    return getattr(self._fd, key)


@contextlib.contextmanager
def tempdir(parent, suffix=''):
  """contextmanager that creates a tempdir and deletes it afterwards.

  Generally, do not use this function; instead, use runtime.System's
  "temp_subdir", which implements common behavior expectations.
  """
  tdir = tempfile.mkdtemp(dir=parent, suffix=suffix)
  try:
    yield tdir
  finally:
    removeall(tdir)


@contextlib.contextmanager
def anonfile(base, prefix='tmp', suffix='', text=False):
  fd, path = tempfile.mkstemp(suffix=suffix, prefix=prefix, dir=base, text=text)
  fd = os.fdopen(fd, 'w')
  try:
    yield NamedAnonymousFile(fd, path)
  finally:
    fd.close()


def resource_path(name):
  return os.path.join(RESOURCES_DIR, name)


def resource_install(name, dest_dir):
  dest = os.path.join(dest_dir, name)
  shutil.copyfile(resource_path(name), dest)
  return dest


def download_to(url, dst_fd, hash_obj=None):
  """Downloads the specified URL, writing it to "dst_fd". Returns the
  specified hash.

  If "hash_obj" is None, no hash will be generated. Otherwise, it should be a
  hashlib instance that will be updated with the downloaded file contents.

  Returns (str): The name of the file that was downloaded (end of the URL).
  """
  def _download_hash_chunks(chunks):
    for chunk in chunks:
      if hash_obj:
        hash_obj.update(chunk)
      dst_fd.write(chunk)

  try:
    with requests.Session() as s:
      r = s.get(url, verify=True)
      _download_hash_chunks(
          r.iter_content(chunk_size=DOWNLOAD_CHUNK_SIZE))
  except requests.exceptions.InvalidSchema:
    # "requests" can't handle this schema (e.g., "ftp://"), use urllib :(
    fd = None
    try:
      LOGGER.debug('Downloading via "urllib": %s', url)
      fd = urllib.request.urlopen(url)

      def _chunk_gen():
        while True:
          data = fd.read(DOWNLOAD_CHUNK_SIZE)
          if not data:
            return
          yield data
      _download_hash_chunks(_chunk_gen())
    finally:
      if fd:
        fd.close()

  parsed_url = urllib.parse.urlparse(url)
  return parsed_url.path.rsplit('/', 1)[-1]



def download_json(url):
  return requests.get(url, verify=True).json()


class Timer(object):

  def __init__(self):
    self._start_time = datetime.datetime.now()
    self._end_time = None

  def stop(self):
    if not self._end_time:
      self._end_time = datetime.datetime.now()

  @property
  def delta(self):
    assert self._end_time, 'Timer is still running!'
    return self._end_time - self._start_time

  @classmethod
  @contextlib.contextmanager
  def run(cls):
    t = cls()
    try:
      yield t
    finally:
      t.stop()


def check_run(system, dx, work_root, cmd, cwd=None, env=None, **kwargs):
  """Runs a command |cmd|.

  Args:
    system (runtime.System): The System instance.
    dx (dockcross.Image or None): The DockCross image to use. If None, the
        command will be run on the local system.
    work_root (str): The work root directory. If |dx| is not None, this will
        be the directory mounted as "/work" in the Docker environment.
    cmd (list): The command to run. Any components that are paths beginning
        with |work_root| will be automatically made relative to |work_root|.
    cwd (str or None): The working directory for the command. If None,
        |work_root| will be used. Otherwise, |cwd| must be a subdirectory of
        |work_root|.
    env (dict or None): Extra environment variables (will be applied to current
        env with dict.update)
    """
  if dx is None:
    return system.check_run(cmd, cwd=cwd or work_root, env=env, **kwargs)
  return dx.check_run(work_root, cmd, cwd=cwd, env=env, **kwargs)


def check_run_script(system, dx, work_root, script, args=None, cwd=None,
                     env=None):
  """Runs a script, |script|.

  An anonymous file will be created under |work_root| holding the specified
  script.

  Args:
    script (list): A list of script lines to execute.
    See "check_run" for full argument definition.
  """
  with anonfile(work_root, text=True) as fd:
    for line in script:
      fd.write(line)
      fd.write('\n')
  os.chmod(fd.name, 0o755)

  LOGGER.debug('Running script (path=%s): %s', fd.name, script)
  cmd = [fd.name]
  if args:
    cmd.extend(args)
  return check_run(system, dx, work_root, cmd, cwd=cwd, env=env)
