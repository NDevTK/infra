# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import collections
import platform
import subprocess
import sys

_MANYLINUX_ENV = {
    'LDSHARED': '/opt/rh/devtoolset-10/root/usr/bin/gcc -pthread -shared'
}


class Platform(
    collections.namedtuple(
        'Platform',
        (
            # The name of the platform.
            'name',

            # If the platform is "manylinux', the "manylinux" Docker image build
            # name (e.g., "cp38-cp38").
            'manylinux_name',

            # The value to pass to e.g. `./configure --host ...`
            'cross_triple',

            # The Python wheel ABI.
            'wheel_abi',

            # Tuple of Python wheel platforms. Must have at least one.
            #
            # This is used for local wheel naming. Wheels are named universally
            # within CIPD packages. Changing this will not impact CIPD package
            # contents, but will affect the locally generated intermediate wheel
            # names.
            'wheel_plat',

            # The "dockcross" base image (can be None).
            'dockcross_base',

            # The tag to pull for the dockcross_base image
            # (None if not using dockcross).
            'dockcross_tag',

            # The OpenSSL "Configure" script build target.
            'openssl_target',

            # Do Python wheels get packaged on PyPi for this platform?
            'packaged',

            # The name of the CIPD platform to use.
            'cipd_platform',

            # Extra environment variables to set when building wheels on this
            # platform.
            'env',
        ))):

  @property
  def dockcross_base_image(self):
    if not self.dockcross_base:
      return None
    return 'dockcross/%s' % (self.dockcross_base,)

  @property
  def dockcross_image_tag(self):
    return 'infra-dockerbuild-%s' % (self.name,)

  @property
  def universal(self):
    return 'any' in self.wheel_plat


ALL = {
    p.name: p for p in (
        Platform(
            name='linux-armv6-py3',
            manylinux_name=None,
            cross_triple='arm-linux-gnueabihf',
            wheel_abi='cp38',
            wheel_plat=('linux_armv6l', 'linux_armv7l', 'linux_armv8l',
                        'linux_armv9l'),
            dockcross_base='linux-armv6',
            dockcross_tag='20210625-795dd4d',
            openssl_target='linux-armv4',
            packaged=False,
            cipd_platform='linux-armv6l',
            env={},
        ),
        Platform(
            name='linux-arm64-py3',
            manylinux_name=None,
            cross_triple='aarch64-unknown-linux-gnueabi',
            wheel_abi='cp38',
            wheel_plat=('linux_arm64', 'linux_aarch64'),
            dockcross_base='linux-arm64',
            dockcross_tag='20210625-795dd4d',
            openssl_target='linux-aarch64',
            packaged=False,
            cipd_platform='linux-arm64',
            env={},
        ),
        Platform(
            name='manylinux-x64-py3',
            manylinux_name=None,  # Don't use any built-in Python
            cross_triple='x86_64-linux-gnu',
            wheel_abi='cp38',
            wheel_plat=('manylinux2014_x86_64',),
            dockcross_base='manylinux2014-x64',
            dockcross_tag='latest',
            openssl_target='linux-x86_64',
            packaged=True,
            cipd_platform='linux-amd64',
            env=_MANYLINUX_ENV,
        ),
        Platform(
            name='manylinux-x64-py3.9',
            manylinux_name=None,  # Don't use any built-in Python
            cross_triple='x86_64-linux-gnu',
            wheel_abi='cp39',
            wheel_plat=('manylinux2014_x86_64',),
            dockcross_base='manylinux2014-x64',
            dockcross_tag='latest',
            openssl_target='linux-x86_64',
            packaged=True,
            cipd_platform='linux-amd64',
            env=_MANYLINUX_ENV,
        ),
        Platform(
            # TODO: Rename to -py3 to conform to other Python 3 platform names.
            name='mac-x64-cp38',
            manylinux_name=None,
            cross_triple='',
            wheel_abi='cp38',
            wheel_plat=('macosx_10_11_x86_64',),
            dockcross_base=None,
            dockcross_tag=None,
            openssl_target='darwin64-x86_64-cc',
            packaged=True,
            cipd_platform='mac-amd64',
            env={
                # Necessary for some wheels to build. See for instance:
                # https://github.com/giampaolo/psutil/issues/1832
                'ARCHFLAGS': '-arch x86_64',
                'MACOSX_DEPLOYMENT_TARGET': '10.11'
            },
        ),
        Platform(
            name='mac-arm64-cp38',
            manylinux_name=None,
            cross_triple='',
            wheel_abi='cp38',
            wheel_plat=('macosx_11_0_arm64',),
            dockcross_base=None,
            dockcross_tag=None,
            openssl_target='darwin64-arm64-cc',
            # TODO: See whether this can be enabled now that Python 3.8.10
            # supports mac-arm64.
            packaged=False,
            cipd_platform='mac-arm64',
            env={
                # Necessary for some wheels to build. See for instance:
                # https://github.com/giampaolo/psutil/issues/1832
                'ARCHFLAGS': '-arch arm64',
                'MACOSX_DEPLOYMENT_TARGET': '11.0'
            },
        ),
        Platform(
            name='windows-x86-py3',
            manylinux_name=None,
            cross_triple='',
            wheel_abi='cp38',
            wheel_plat=('win32',),
            dockcross_base=None,
            dockcross_tag=None,
            openssl_target='Cygwin-x86',
            packaged=True,
            cipd_platform='windows-386',
            env={},
        ),
        Platform(
            name='windows-x64-py3',
            manylinux_name=None,
            cross_triple='',
            wheel_abi='cp38',
            wheel_plat=('win_amd64',),
            dockcross_base=None,
            dockcross_tag=None,
            openssl_target='Cygwin-x86_64',
            packaged=True,
            cipd_platform='windows-amd64',
            env={},
        ),
        Platform(
            name='universal',
            manylinux_name=None,
            cross_triple='',
            wheel_abi='none',
            wheel_plat=('any',),
            dockcross_base=None,
            dockcross_tag=None,
            openssl_target=None,
            packaged=True,
            cipd_platform=None,
            env={},
        ),
    )
}


# Detect whether we're on an ARM64 Mac running emulated x86_64 Python.
# In this situation, we still consider ARM64 the native platform.
def _CheckTranslated():
  if sys.platform != 'darwin':
    return False

  try:
    output = subprocess.check_output(
        ["/usr/sbin/sysctl", "-n", "sysctl.proc_translated"], text=True)
    return output[0] == '1'
  except subprocess.CalledProcessError:
    # The call will fail on x86_64 Macs.
    return False


NAMES = sorted(ALL.keys())
PACKAGED = [p for p in ALL.values() if p.packaged]
ALL_LINUX = [p.name for p in ALL.values() if 'linux' in p.name]
UNIVERSAL = [p.name for p in ALL.values() if 'universal' in p.name]
_IS_TRANSLATED = _CheckTranslated()


def NativeMachine():
  machine = platform.machine()
  if sys.platform == 'darwin' and machine == 'x86_64' and _IS_TRANSLATED:
    machine = 'arm64'
  return machine


def NativePlatforms():
  # Every supported OS can build universal wheels.
  plats = [ALL[u] for u in UNIVERSAL]

  # Identify our native platforms.
  if sys.platform == 'darwin':
    arch = {'x86_64': 'x64', 'arm64': 'arm64'}[NativeMachine()]
    plat_name = 'mac-%s' % arch
    return plats + [ALL[plat_name + '-cp38']]
  elif sys.platform == 'win32':
    return plats + [ALL['windows-x86-py3'], ALL['windows-x64-py3']]
  elif sys.platform.startswith('linux'):
    # Linux platforms are built with docker, so Linux doesn't support any
    # non-universal platforms natively.
    return plats
  raise ValueError('Cannot identify native image for %r-%r.' %
                   (sys.platform, NativeMachine()))
