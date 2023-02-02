# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import itertools
import os

from .build_types import Spec
from .builder import BuildPackageFromPyPiWheel, StageWheelForPackage
from .builder import SetupPythonPackages
from .wheel_wheel import SourceOrPrebuilt

from . import source
from . import util


class Mpi4py(SourceOrPrebuilt):

  def __init__(self, name, version, mpich_version, **kwargs):
    """Specialized wheel builder for the "cryptography" package.

    Args:
      name (str): The wheel name.
      version (str): The wheel version.
      mpich_version (str): The version of the mpich 3pp package to compile
          with (for non-packaged platforms).
    """
    self._mpich_version = mpich_version
    super(Mpi4py, self).__init__(name, version, **kwargs)

  def build_fn(self, system, wheel, output_dir):
    if wheel.plat.name in self._packaged:
      return BuildPackageFromPyPiWheel(system, wheel, output_dir)

    dx = system.dockcross_image(wheel.plat)

    with system.temp_subdir('%s_%s' % wheel.spec.tuple) as tdir:
      build_dir = system.repo.ensure(self._pypi_src, tdir)

      # TODO: Refactor this to use BuildPackageFromSource with tpp_libs.
      mpich_pkg = ('infra/3pp/static_libs/mpich/%s' % wheel.plat.cipd_platform)
      pkg_dir = os.path.join(build_dir, mpich_pkg + '_cipd')
      system.cipd.init(pkg_dir)
      system.cipd.install(mpich_pkg, self._mpich_version, pkg_dir)

      # Add MPICH CFLAGS and LDFLAGS to the environment. We can't easily
      # use the 'mpicc' wrapper to do this as it's hardcoded to a specific
      # install location.
      cflags = '-I' + os.path.join(pkg_dir, 'include')
      ldflags = '-L' + os.path.join(pkg_dir, 'lib')
      if wheel.plat.name.startswith('mac'):
        ldflags += ' -framework OpenCL'
      env = {'CFLAGS': cflags, 'LDFLAGS': ldflags}

      for patch in self._pypi_src.get_patches():
        util.LOGGER.info('Applying patch %s', os.path.basename(patch))
        cmd = ['patch', '-p1', '--quiet', '-i', patch]
        subprocess.check_call(cmd, cwd=build_dir)

      if wheel.plat.name.startswith('mac'):
        mpi_libs = ['mpi', 'pmpi', 'm', 'pthread']
      elif 'linux' in wheel.plat.name:
        mpi_libs = ['mpi', 'm', 'pthread', 'rt']
      else:
        raise NotImplementedError('Implement mpi_libs for %s' % wheel.plat.name)

      build_ext_options = ['--libraries=%s' % ' '.join(mpi_libs)]
      if wheel.plat.dockcross_base:
        build_ext_options.extend([
            '--include-dirs',
            '/usr/cross/include',
            '--library-dirs',
            '/usr/cross/lib',
        ])

      interpreter, extra_env = SetupPythonPackages(system, wheel, tdir, tdir)
      cmd = [
          interpreter,
          '-m',
          'pip',
          'wheel',
          '--no-deps',
          '--only-binary=:all:',
          '--wheel-dir',
          tdir,
      ]
      for opt in itertools.chain(['build_ext'], build_ext_options):
        cmd.extend(['--global-option', opt])
      cmd.append('.')

      if dx.platform:
        extra_env.update({
            'host_alias': dx.platform.cross_triple,
        })
      if env:
        extra_env.update(env)

      util.check_run(system, dx, tdir, cmd, cwd=build_dir, env=extra_env)

      StageWheelForPackage(system, tdir, wheel, output_dir)
