#!/usr/bin/env python3
# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Tests generated CIPD packages.

Supposed to be invoked after build.py has run. Uses packages from out/*.cipd and
tests from tests/*.py.

Assumes cipd client is built in out/.cipd_client/cipd_* (true after build.py has
run).
"""

import argparse
import os
import re
import shutil
import subprocess
import sys
import tempfile


# Root of infra.git repository.
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# .exe on Windows.
EXE_SUFFIX = '.exe' if sys.platform == 'win32' else ''


class TestException(Exception):
  pass


def print_title(title):
  """Pretty prints a banner to stdout."""
  sys.stdout.flush()
  print()
  print('-' * 80)
  print(title)
  print('-' * 80)


def get_docstring(test_script):
  """Hacky way to grab a first line of a module docstring using regexps."""
  with open(test_script, 'rt') as f:
    text = f.read()
  m = re.match(r'^.*"""(.*?)"""', text, re.DOTALL)
  if not m:
    return None
  return m.group(1).strip().splitlines()[0]


def find_cipd():
  """Finds a CIPD client to use in PATH."""
  exts = ('.exe', '.bat') if sys.platform == 'win32' else ('',)
  for p in os.environ.get('PATH', '').split(os.pathsep):
    base = os.path.join(p, 'cipd')
    for ext in exts:
      candidate = base + ext
      if os.path.isfile(candidate):
        return candidate
  return 'cipd' + EXE_SUFFIX


def run_test(cipd_client, package, work_dir, test_script):
  """Extracts a package to a dir and runs test_script with cwd == work_dir."""
  print_title('Deploying %s' % os.path.basename(package))
  cmd_line = ['cipd', 'pkg-deploy', '-root', work_dir, package]
  print(' '.join(cmd_line))
  if subprocess.call(args=cmd_line, executable=cipd_client):
    raise TestException('Failed to install %s, see logs' % package)

  print_title(get_docstring(test_script) or 'Running tests...')
  cmd_line = [sys.executable, test_script]
  print('%s in %s' % (' '.join(cmd_line), work_dir))
  env = os.environ.copy()
  env.pop('PYTHONPATH', None)
  ret = subprocess.call(
      args=cmd_line, executable=sys.executable, env=env, cwd=work_dir)
  if ret:
    raise TestException('Non zero exit code (%d)' % ret)


def run(
    package_out_dir,
    package_tests_dir,
    work_dir,
    packages):
  """Deploys build *.cipd package locally and runs tests against them.

  Used to verify the packaged code works when installed as CIPD package.

  Args:
    package_out_dir: where to search for built packages.
    work_dir: where to install/update packages into.
    packages: names of *.cipd files in package_out_dir or [] for all.

  Returns:
    0 on success, 1 or error.
  """
  # Discover what to test.
  paths = []
  if not packages:
    # Enumerate all known tests in tests/*.py and filter them based on
    # availability of corresponding *.cipd package in package_out_dir. It will
    # skip any cross-compiled packages, since they have additional '+<platform>'
    # suffix in the package file name.
    for test in os.listdir(package_tests_dir):
      if not test.endswith('.py'):
        continue
      pkg_file = os.path.join(
          package_out_dir, os.path.splitext(test)[0] + '.cipd')
      if os.path.exists(pkg_file):
        paths.append(pkg_file)
  else:
    for name in packages:
      abs_path = os.path.join(package_out_dir, name)
      if not os.path.isfile(abs_path):
        raise TestException('No such package file: %s' % name)
      paths.append(abs_path)
  paths = sorted(paths)
  if not paths:
    print('Nothing to test.')
    return 0

  cipd_client = find_cipd()
  if not cipd_client:
    return 1

  # Run all tests sequentially. Most of the are extra fast.
  nuke_temp = False
  if not work_dir:
    work_dir = tempfile.mkdtemp(suffix='cipd_test')
    nuke_temp = True
  work_dir = os.path.abspath(work_dir)
  try:
    fail = False
    for path in paths:
      name = os.path.splitext(os.path.basename(path))[0]
      test_script = os.path.join(package_tests_dir, '%s.py' % name)
      if not os.path.isfile(test_script):
        print('Skipping tests for %s - no such file: %s' % (name, test_script))
        continue
      try:
        run_test(
            cipd_client=cipd_client,
            package=path,
            work_dir=os.path.join(work_dir, name),
            test_script=test_script)
        print('')
        print('PASS')
      except TestException as exc:
        print('')
        print('FAILED! ' * 10)
        print('Tests for %s failed: %s' % (name, exc))
        fail = True
    return 1 if fail else 0
  finally:
    if nuke_temp:
      try:
        shutil.rmtree(work_dir, ignore_errors=True)
      except OSError as exc:
        print('Failed to delete %s: %s' % (work_dir, exc))


def main(
    args,
    package_out_dir=os.path.join(ROOT, 'build', 'out'),
    package_tests_dir=os.path.join(ROOT, 'build', 'tests')):
  parser = argparse.ArgumentParser(description='Tests infra CIPD packages')
  parser.add_argument(
      'packages', metavar='NAME', type=str, nargs='*',
      help='name of a built package file in build/out/* to deploy and test')
  parser.add_argument(
      '--work-dir', metavar='DIR', dest='work_dir',
      help='directory to deploy packages into (temporary dir by default)')
  args = parser.parse_args(args)
  return run(
      package_out_dir,
      package_tests_dir,
      args.work_dir,
      [n + '.cipd' if not n.endswith('.cipd') else n for n in args.packages])


if __name__ == '__main__':
  sys.exit(main(sys.argv[1:]))
