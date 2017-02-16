#!/usr/bin/env python
# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import logging
import os
import platform
import subprocess
import sys


# The path to the "infra/bootstrap/" directory.
BOOTSTRAP_DIR = os.path.dirname(os.path.abspath(__file__))
# The path to the "infra/" directory.
ROOT = os.path.dirname(BOOTSTRAP_DIR)
# The path where CIPD install lists are stored.
CIPD_LIST_DIR = os.path.join(BOOTSTRAP_DIR, 'cipd')
# Default sysroot install root.
DEFAULT_INSTALL_ROOT = os.path.join(ROOT, 'cipd')
# For Windows.
EXE_SFX = '.exe' if sys.platform == 'win32' else ''
# Default CIPD server url
DEFAULT_SERVER_URL='https://chrome-infra-packages.appspot.com'

# Map of CIPD configuration based on the current architecture/platform. If a
# platform is not listed here, the bootstrap will be a no-op.
#
# This is keyed on the platform's (system, machine).
ARCH_CONFIG_MAP = {
  ('Linux', 'x86_64'): {
    'cipd_install_list': 'cipd_linux_amd64.txt',
  },
  ('Linux', 'x86'): {
    'cipd_install_list': None,
  },
  ('Darwin', 'x86_64'): {
    'cipd_install_list': 'cipd_mac_amd64.txt',
  },
  ('Windows', 'x86_64'): {
    'cipd_install_list': None,
  },
  ('Windows', 'x86'): {
    'cipd_install_list': None,
  },
}


def get_platform_config():
  key = get_platform()
  return key, ARCH_CONFIG_MAP.get(key)


def get_platform():
  machine = platform.machine().lower()
  system = platform.system()
  machine = ({
    'amd64': 'x86_64',
    'i686': 'x86',
  }).get(machine, machine)
  if (machine == 'x86_64' and system == 'Linux' and
      sys.maxsize == (2 ** 31) - 1):
    # This is 32bit python on 64bit CPU on linux, which probably means the
    # entire userland is 32bit and thus we should play along and install 32bit
    # packages.
    machine = 'x86'
  return system, machine


def ensure_directory(path):
  # Ensure the parent directory exists.
  if os.path.isdir(path):
    return
  if os.path.exists(path):
    raise ValueError("Target file's directory [%s] exists, but is not a "
                     "directory." % (path,))
  logging.debug('Creating directory: [%s]', path)
  os.makedirs(path)


def execute(*cmd):
  if not logging.getLogger().isEnabledFor(logging.DEBUG):
    code = subprocess.call(cmd)
  else:
    # Execute the process, passing STDOUT/STDERR through our logger.
    proc = subprocess.Popen(cmd, stdout=subprocess.PIPE,
                            stderr=subprocess.STDOUT)
    for line in proc.stdout:
      logging.debug('%s: %s', cmd[0], line.rstrip())
    code = proc.wait()
  if code:
    logging.error('Process failed with exit code: %d', code)
  return code


class CipdError(Exception):
  """Raised by install_cipd_client on fatal error."""


def cipd_ensure(cipd_backend_url, list_path, root):
  assert os.path.isfile(list_path)
  assert os.path.isdir(root)
  logging.debug('Installing CIPD packages from [%s] to [%s]', list_path, root)
  args = [
    'ensure',
    '-ensure-file', list_path,
    '-root', root,
    '-service-url', cipd_backend_url
  ]
  if execute('cipd'+EXE_SFX, *args):
    raise CipdError('Failed to execute CIPD client: %s', ' '.join(args))


def main(argv):
  parser = argparse.ArgumentParser('Installs CIPD bootstrap packages.')
  parser.add_argument('-v', '--verbose', action='count', default=0,
      help='Increase logging verbosity. Can be specified multiple times.')
  parser.add_argument('--cipd-backend-url', metavar='URL',
      default=DEFAULT_SERVER_URL,
      help='Specify the CIPD backend URL (default is %(default)s)')
  parser.add_argument('-d', '--cipd-root-dir', metavar='PATH',
      default=DEFAULT_INSTALL_ROOT,
      help='Specify the root CIPD package installation directory.')

  opts = parser.parse_args(argv)

  # Setup logging verbosity.
  if opts.verbose == 0:
    level = logging.WARNING
  elif opts.verbose == 1:
    level = logging.INFO
  else:
    level = logging.DEBUG
  logging.getLogger().setLevel(level)

  # Make sure our root directory exists.
  root = os.path.abspath(opts.cipd_root_dir)
  ensure_directory(root)

  platform_key, config = get_platform_config()
  if not config:
    logging.info('No bootstrap configuration for platform [%s].', platform_key)
    return 0

  # Install the CIPD list for this configuration.
  cipd_install_list = config.get('cipd_install_list')
  if cipd_install_list:
    cipd_ensure(opts.cipd_backend_url,
                os.path.join(CIPD_LIST_DIR, cipd_install_list), root)
  return 0


if __name__ == '__main__':
  logging.basicConfig()
  logging.getLogger().setLevel(logging.INFO)
  sys.exit(main(sys.argv[1:]))
