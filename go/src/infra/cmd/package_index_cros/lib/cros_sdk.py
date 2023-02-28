# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from typing import List

from chromite.lib import cros_build_lib

from .constants import PRINT_DEPS_SCRIPT_PATH
from .constants import PACKAGES_FAILING_TESTS
from .logger import g_logger
from .path_handler import PathHandler
from .setup import Setup


class CrosSdk:
  """Handles requests to cros_sdk."""

  def _Exec(self,
            cmd: List[str],
            *,
            capture_output: bool = False,
            with_sudo: bool = False) -> cros_build_lib.CompletedProcess:
    shell = True
    if isinstance(cmd, List):
      shell = False

    encoding = 'utf-8' if capture_output else None

    g_logger.debug("Executing: '%s'", cmd)
    run_func = cros_build_lib.sudo_run if with_sudo else cros_build_lib.run
    res = run_func(
        cmd,
        enter_chroot=True,
        shell=shell,
        capture_output=capture_output,
        encoding=encoding,
        check=True,
        print_cmd=False,
        chroot_args=['--chroot', self.setup.chroot_dir])
    return res

  def __init__(self, setup: Setup):
    self.setup = setup

  def StartWorkonPackages(self, package_names: List[str]) -> None:
    """
    Runs cros_workon start with given packages.

    Raises:
      * cros_build_lib.CompletedProcess if command failed.
    """
    cmd = ['cros_workon', f"--board={self.setup.board}", "start"
          ] + package_names
    self._Exec(cmd)

  def StopWorkonPackages(self, package_names: List[str]) -> None:
    """
    Runs cros_workon stop with given packages.

    Raises:
      * cros_build_lib.CompletedProcess if command failed.
    """
    cmd = ['cros_workon', f"--board={self.setup.board}", "stop"] + package_names
    self._Exec(cmd)

  def StartWorkonPackagesSafe(self, package_names: List[str]):
    """
    Safe version of start/stop workon. Preserves state of already workon
    packages.

    Returns a handler to stop workon with 'with' statement.

    Raises:
      * cros_build_lib.CompletedProcess if command failed.
    """
    check_workon_cmd = ['cros_workon', f"--board={self.setup.board}", "list"]

    output = self._Exec(check_workon_cmd, capture_output=True).stdout
    workon_packages = set(output.split('\n'))
    non_workon_packages = [p for p in package_names if not p in workon_packages]

    g_logger.debug('Packages to start workon: %s', non_workon_packages)
    g_logger.debug('Packages already workon: %s',
                   [p for p in package_names if p in workon_packages])

    class WorkonRaii:

      def __init__(self, cros_sdk: CrosSdk, packages_to_workon: List[str]):
        self.cros_sdk = cros_sdk
        self.packages_to_workon = packages_to_workon

      def __enter__(self) -> 'WorkonRaii':
        if self.packages_to_workon and len(self.packages_to_workon) != 0:
          self.cros_sdk.StartWorkonPackages(self.packages_to_workon)
        return self

      def __exit__(self, type, value, traceback) -> None:
        if self.packages_to_workon and len(self.packages_to_workon) != 0:
          self.cros_sdk.StopWorkonPackages(self.packages_to_workon)

    return WorkonRaii(self, non_workon_packages)

  def BuildPackages(self, package_names: List[str]) -> None:
    """
    Builds given packages and preserves build artifcats.

    Raises:
      * cros_build_lib.CompletedProcess if command failed.
    """
    cmd = [
        'FEATURES="noclean"',
        'parallel_emerge',
        '--board',
        self.setup.board,
    ] + package_names

    cmd = ' '.join(cmd)
    self._Exec(cmd, with_sudo=True)

  def RunTests(self, package_names: List[str]) -> None:
    """
    Builds given testable packages, runs tests and preserves build artifcats.

    Raises:
      * cros_build_lib.CompletedProcess if command failed.
    """
    cmd = [
        'FEATURES="noclean"', 'cros_run_unit_tests', '--board',
        self.setup.board, '--filter-only-cros-workon',
        '--no-testable-packages-ok', '--packages',
        f'"{" ".join(package_names)}"', '--skip-packages',
        f'"{" ".join(list(PACKAGES_FAILING_TESTS))}"'
    ]

    cmd = ' '.join(cmd)
    self._Exec(cmd)

  def GenerateCompileCommands(self, chroot_build_dir: str) -> str:
    """
    Calls ninja and returns compile commands as a string.

    |chroot_build_dir|  is package's build dir inside chroot.

    Raises:
      * cros_build_lib.CompletedProcess if command failed.
    """
    ninja_cmd = ['ninja', '-C', chroot_build_dir, '-t', 'compdb', 'cc', 'cxx']

    return self._Exec(ninja_cmd, capture_output=True).stdout

  def GenerateGnTargets(self, chroot_root_dir: str,
                        chroot_build_dir: str) -> str:
    """
    Calls gn desc and returns gn targets as a string.

    |chroot_root_dir| is a package's dir containing upper most .gn file inside
    chroot.
    |chroot_build_dir| is a package's build dir inside chroot.

    Raises:
      * cros_build_lib.CompletedProcess if command failed.
    """
    gn_desc_cmd = [
        'gn',
        'desc',
        f"--root={chroot_root_dir}",
        chroot_build_dir,
        '*',
        '--format=json',
    ]

    return self._Exec(gn_desc_cmd, capture_output=True).stdout

  def GenerateDependencyTree(self, package_names: List[str]):
    """
    Generates dependency tree for given packages.

    Returns a dictionary with dependencies (see script/print_deps.py for
    detailed format).

    Utilizes chromite.lib.depgraph to fetch dependency tree. Depgraph has to be
    called from inside chroot so it lives in separate script file which is
    called via cros_sdk wrapper.

    Raises:
      * cros_build_lib.CompletedProcess if command failed.
    """
    use_flags = ''
    if self.setup.with_tests and not self.setup.with_build:
      # Add the test use flag if packages were built with tests. Do not add if
      # packages are not built yet.
      use_flags = 'USE="test"'
    cmd = [
        use_flags,
        PathHandler(self.setup).ToChroot(PRINT_DEPS_SCRIPT_PATH),
        self.setup.board
    ] + package_names
    cmd = ' '.join(cmd)
    return self._Exec(cmd, capture_output=True, with_sudo=True).stdout
