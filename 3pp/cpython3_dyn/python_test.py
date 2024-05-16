# Copyright 2018 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import os
import sys
import shutil
import subprocess
import tempfile
import unittest

PYTHON_TEST_CIPD_PACKAGE = None


class TestPython(unittest.TestCase):

  # Public repository that uses HTTPS.
  HTTPS_REPO_URL = 'https://chromium.googlesource.com/infra/infra'

  @classmethod
  def setUpClass(cls):
    cls._is_windows = os.name == 'nt'
    cls._exe_suffix = '.exe' if cls._is_windows else ''

    cls.tdir = tempfile.mkdtemp(dir=os.getcwd(), suffix='test_python')

    cls.pkg_dir = os.path.join(cls.tdir, 'install')
    subprocess.check_call(
        ['cipd', 'pkg-deploy', PYTHON_TEST_CIPD_PACKAGE, '-root', cls.pkg_dir],
        shell=cls._is_windows)
    cls.python = os.path.join(cls.pkg_dir, 'bin', 'python3' + cls._exe_suffix)

  @classmethod
  def tearDownClass(cls):
    shutil.rmtree(cls.tdir, ignore_errors=True)

  def setUp(self):
    self.test_tdir = tempfile.mkdtemp(dir=self.tdir)
    self.env = os.environ.copy()

  def _write_file(self, content):
    fd = None
    try:
      fileno, path = tempfile.mkstemp(dir=self.test_tdir)
      fd = os.fdopen(fileno, 'w')
      fd.write(content)
      return path
    finally:
      if fd:
        fd.close()

  def test_version(self):
    output = subprocess.check_output([self.python, '-VV'],
                                     stderr=subprocess.STDOUT, env={'LD_LIBRARY_PATH': '/work/checkout'}).decode('utf-8')
    name, version, buildinfo = output.splitlines()[0].split(' ', maxsplit=2)
    self.assertEqual(name, 'Python')
    self.assertEqual(version, os.environ['_3PP_VERSION'])

    # On windows we don't append the patch version because we actually bundle
    # the official python release, so don't have an opportunity to change the
    # version string.
    if '_3PP_PATCH_VERSION' in os.environ and not self._is_windows:
      self.assertIn(f'({os.environ["_3PP_PATCH_VERSION"]},', buildinfo)

  def test_package_import(self):
    for pkg in ('ctypes', 'ssl', 'io', 'binascii', 'hashlib', 'sqlite3'):
      script = 'import %s; print(%s)' % (pkg, pkg)
      rv = subprocess.call([self.python, '-c', script], env={'LD_LIBRARY_PATH': '/work/checkout'})
      self.assertEqual(rv, 0, 'Could not import %r.' % (pkg,))

  def test_use_https(self):
    script = 'import urllib.request; print(urllib.request.urlopen("%s"))' % (
        self.HTTPS_REPO_URL)
    rv = subprocess.call([self.python, '-c', script], env={'LD_LIBRARY_PATH': '/work/checkout'})
    self.assertEqual(rv, 0)

  def test_no_version_script_in_sysconfig(self):
    # On Linux, we use a linker version script to restrict the exported
    # symbols. Verify that this has not leaked into the build flags that
    # will be used by Python wheels.
    script = ('import sysconfig\n'
              'for k, v in sysconfig.get_config_vars().items():\n'
              '  if (isinstance(v, str) and not k.endswith("_NODIST")\n'
              '      and k not in ("PY_CORE_LDFLAGS", "BLDSHARED")):\n'
              '    assert "version-script" not in v, (\n'
              '      "Found unexpected version-script in %s: %s" % (k, v))')
    rv = subprocess.call([self.python, '-c', script], env={'LD_LIBRARY_PATH': '/work/checkout'})
    self.assertEqual(rv, 0)


if __name__ == '__main__':
  platform = os.environ['_3PP_PLATFORM']
  tool_platform = os.environ['_3PP_TOOL_PLATFORM']
  if 'windows' not in platform and platform != tool_platform:
    print('SKIPPING TESTS')
    print('  platform:', platform)
    print('  tool_platform:', tool_platform)
    sys.exit(0)

  PYTHON_TEST_CIPD_PACKAGE = sys.argv[1]
  sys.argv.pop(1)
  unittest.main()
