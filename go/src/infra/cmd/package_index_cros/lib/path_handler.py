# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import os
import re
from typing import Callable, List, NamedTuple, Tuple

from chromite.lib import path_util

from .logger import g_logger
from .package import PackagePathException
from .package import Package
from .setup import Setup


class PathHandler:
  """
  Provides with several helping methods. The main goal: fix paths by
  substituting temp path with actual one.
  """

  class PathNotFixedException(PackagePathException):
    """Indicates failure while trying to fix a path."""

  class FixedPath(NamedTuple):
    """
    Represents path outside of chroot and corresponding actual path (e.g. match
    between temporary downloaded src and actual src files).
    """
    original: str
    actual: str

  def __init__(self, setup: Setup):
    self.setup = setup

  @staticmethod
  def SanitizePath(path: str) -> str:
    # Remove any trailing slashes
    path = path.rstrip(os.path.sep)

    return path

  @staticmethod
  def MovePath(path: str, from_dir: str, to_dir: str) -> str:
    """
    Replaces path's base dir |from_dir| with |to_dir|.

    Raises:
      * ValueError if |path| is not in |from_dir|.
    """
    if not path.startswith(from_dir):
      raise ValueError(f"Path is not in dir: {path} vs {from_dir}")
    return os.path.realpath(
        os.path.join(to_dir, os.path.relpath(path, from_dir)))

  def FromChroot(self, chroot_path: str):
    # |path_util.ChrootPathResolver._ConvertPath| resolves symlinks outside
    # of chroot which don't exist inside chroot. E.g. /bin/ln becomes
    # /chroot/usr/bin/ln if /bin is linked as /usr/bin outside of chroot.
    # Use |_GetHostPath| directly.
    return path_util.ChrootPathResolver(
        chroot_path=self.setup.chroot_dir)._GetHostPath(chroot_path)

  def ToChroot(self, path: str):
    # Unlike |FromChroot|, here it is useful to resolve symlinks before
    # converting path to inside chroot.
    return path_util.ToChrootPath(path, chroot_path=self.setup.chroot_dir)

  def _NormalizePath(self,
                     chroot_path: str,
                     *,
                     chroot_base_dir: str = None,
                     base_dir: str = None) -> str:
    """
    Returns path outside of chroot for given |chroot_path| inside chroot.

    If path is relative, returns absolute path with |chroot_base_dir| as base
    dir.

    Either |chroot_base_dir| or |base_dir| shall be specified. If
    |chroot_base_dir| is given, |base_dir| is ignored. If |base_dir| is given,
    it is resolved to chroot path and used as |chroot_base_dir|.
    """
    assert chroot_base_dir or base_dir, \
      'Either chroot_base_dir or base_dir must be set'

    chroot_path = PathHandler.SanitizePath(chroot_path)

    if chroot_base_dir is None:
      chroot_base_dir = self.ToChroot(base_dir)

    if not chroot_path.startswith(os.sep):
      chroot_path = os.path.join(chroot_base_dir, chroot_path)
      chroot_path = os.path.realpath(chroot_path)

    return self.FromChroot(chroot_path)

  def FixPath(self,
              chroot_path: str,
              package: Package,
              *,
              conflicting_paths={}) -> 'PathHandler.FixedPath':
    """
    Returns an original and actual path outside of chroot for given
    |chroot_path| inside chroot.

    A path outside of |package.temp_dir| is considered as actual path and
    returned as is.

    If |chroot_path| is resolved to a path which is present in
    |conflicting_paths| dict, returns a path from corresponding entry.

    Arguments:
      * |chroot_path|: a path inside chroot to resolve.
      * |package|: a package it path belongs to.
      * |conflicting_paths|: dict of paths outside of chroot that have conflicts
        between packages.

    Raises:
      * PathNotFixedException if cannot resolve |path| to actual path.
      * PathNotFixedException if actual path does not exist.
      """
    chroot_path = PathHandler.SanitizePath(chroot_path)

    path = None
    if chroot_path.startswith('//'):
      # Special case. '//' indicates source dir.
      for match_dirs in package.src_dir_matches:
        path_attempt = os.path.join(match_dirs.temp, chroot_path[2:])
        if os.path.exists(path_attempt):
          path = path_attempt
          break
    else:
      path = self._NormalizePath(chroot_path, base_dir=package.build_dir)

    if not path or not os.path.exists(path):
      raise PathHandler.PathNotFixedException(package,
                                              'Given path does not exist', path,
                                              path)

    def Fix():

      if path in conflicting_paths:
        return conflicting_paths[path]

      if not path.startswith(package.temp_dir) or path.startswith(
          package.build_dir):
        # Don't care about paths outside of temp_dir.
        # Build dir can be subdir of temp_dir, but we don't care either.
        return path

      for matching_dirs in package.src_dir_matches:
        if not path.startswith(matching_dirs.temp):
          continue
        actual_path = os.path.realpath(
            PathHandler.MovePath(path, matching_dirs.temp,
                                 matching_dirs.actual))
        if os.path.exists(actual_path):
          return actual_path

      raise PathHandler.PathNotFixedException(
          package, 'Could not find path in any of source dirs', path)

    def Check(actual_path):
      if not os.path.exists(actual_path):
        raise PathHandler.PathNotFixedException(package,
                                                'Found path does not exist',
                                                path, actual_path)

    actual_path = os.path.realpath(Fix())
    Check(actual_path)
    return PathHandler.FixedPath(path, actual_path)

  def _FixPathFromBasedir(self,
                          chroot_path: str,
                          package: Package,
                          *,
                          conflicting_paths={},
                          ignorable_dir=None) -> 'PathHandler.FixedPath':
    """
    Fixes basedir of given |chroot_path| and appends basename of |chroot_path|
    to the fixed dir.

    Will attempt to fix base dir until |ignorable_dir| has at least one dir
    containing given chroot_path. For example:
    0. ignorable_dir == '/a/b/c' which does not exist but '/a/b' does exist
    1. chroot_path == '/a/b/c/d/e' which does not exist, parent also: go up.
    2. chroot_path == '/a/b/c/d' which does not exist, parent also: go up.
    3. chroot_path == '/a/b/c' which does not exist, but parent does - fix it.
    4. return '/fixed-a-b/' + 'c/d/e'

    Raises:
      * PathNotFixedException if cannot resolve path most possible parent dir to
        actual path.
      * PathNotFixedException if actual path's most possible parent dir does not
        exist.
    """
    chroot_path = PathHandler.SanitizePath(chroot_path)
    if ignorable_dir:
      ignorable_dir = PathHandler.SanitizePath(ignorable_dir)

    # Ignorable dir is the upper most possible parent which may not exist. If it
    # is not given, chroot_path is ignorable dir and it's parent must exist.
    chroot_ignorable_dir = self.ToChroot(
        ignorable_dir) if ignorable_dir else chroot_path

    chroot_path_base_dir = os.path.dirname(chroot_path)
    chroot_path_basename = os.path.basename(chroot_path)

    assert chroot_ignorable_dir

    # Trying to fix chroot_path's parent dir until chroot_path is inside
    # chroot_ignorable_dir.
    while chroot_path and chroot_path.startswith(chroot_ignorable_dir):
      # chroot_path is still inside chroot_ignorable_dir or is
      # chroot_ignorable_dir. Try to fix it's parent.
      try:
        path_basedir, actual_path_basedir = self.FixPath(
            chroot_path_base_dir, package, conflicting_paths=conflicting_paths)
        path = os.path.join(path_basedir, chroot_path_basename)
        actual_path = os.path.join(actual_path_basedir, chroot_path_basename)
        return PathHandler.FixedPath(path, actual_path)
      except PathHandler.PathNotFixedException:
        # Failed not fix chroot_path's parent dir. Walk up and try to fix parent
        # dir.
        chroot_path = os.path.dirname(chroot_path)
        chroot_path_basename = os.path.join(
            os.path.basename(chroot_path_base_dir), chroot_path_basename)
        chroot_path_base_dir = os.path.dirname(chroot_path_base_dir)

    raise PathHandler.PathNotFixedException(package,
                                            "Failed for fix from base dir",
                                            chroot_path, chroot_path)

  def FixPathWithIgnores(
      self,
      chroot_path: str,
      package: Package,
      *,
      conflicting_paths={},
      ignore_highly_volatile: bool = False,
      ignore_generated: bool = False,
      ignore_stable: bool = False,
      ignorable_dirs: List[str] = [],
      ignorable_extensions: List[str] = []) -> 'PathHandler.FixedPath':
    """
    Returns an actual path outside of chroot similar to |FixPath|.

    Does not fail if given or actual path does not exist according to given
    arguments.

    If |FixPath| fails but the issue can be ignored, attempts to fix
    |chroot_path| parent dir or parent's parent dir until prefix matches. If
    this attempt fails as well - report failure.

    Arguments:
    * |chroot_path|: a path inside chroot to resolve.
    * |package|: a package it path belongs to.
    * |conflicting_paths|: dict of paths outside of chroot that have conflicts
      between packages.
    * |ignore_generated|: do not fail if |chroot_path| in |package.build_dir|,
      return as is. Unlike |ignorable_dirs|, we ignore anything that happens
      inside |package.build_dir|, not just path's parent dir.
    * |ignore_stable|: do not fail if |chroot_path| belongs to a stably built
      package.
    * |ignore_highly_volatile|: do not fail if |package| is considered as
      highly volatile (may contain patches which create/delete files).
    * |ignorable_dirs|: do not fail if path is inside one of given dirs
      outside of chroot (aka has a dir as prefix).
    * |ignorable_extensions|: do not fail if path ends with one of given
      extensions.

    Raises:
      * PathNotFixedException if cannot resolve |path| to actual path.
      * PathNotFixedException if actual path does not exist.
    """
    chroot_path = PathHandler.SanitizePath(chroot_path)

    try:
      return self.FixPath(chroot_path,
                          package,
                          conflicting_paths=conflicting_paths)
    except PathHandler.PathNotFixedException as e:
      # Failed to fix as is. Check if error can be ignored and try to fix from
      # parent dir.
      path = self._NormalizePath(chroot_path, base_dir=package.build_dir)

      if ignore_generated and path.startswith(package.build_dir):
        # Path inside build dir and ignorable, return as is.
        g_logger.debug('%s: Failed to fix generated path: %s',
                       package.full_name, path)
        return PathHandler.FixedPath(path, path)

      can_ignore = False
      if ignore_highly_volatile and package.is_highly_volatile:
        g_logger.debug('%s: Failed to fix path for highly volatile package: %s',
                       package.full_name, path)
        can_ignore = True
      if ignore_stable and not package.is_built_from_actual_sources:
        g_logger.debug('%s: Failed to fix path for stable package: %s',
                       package.full_name, path)
        can_ignore = True
      elif ignorable_dirs:
        g_logger.debug('%s: Failed to fix path in ignorable dir: %s',
                       package.full_name, path)
        can_ignore = True
      elif ignorable_extensions and any(
          path.endswith(ignorable_ext)
          for ignorable_ext in ignorable_extensions):
        g_logger.debug('%s: Failed to fix path with ignorable extension: %s',
                       package.full_name, path)
        can_ignore = True

      if can_ignore:
        # Try to find matching ignorable dir containing path.
        ignorable_parent_dirs = [
            ignorable_dir for ignorable_dir in ignorable_dirs
            if path.startswith(ignorable_dir)
        ]
        assert len(ignorable_parent_dirs) <= 1, "Expecting one match at most"
        ignorable_parent_dir = ignorable_parent_dirs[
            0] if ignorable_parent_dirs else None
        return self._FixPathFromBasedir(chroot_path,
                                        package,
                                        conflicting_paths=conflicting_paths,
                                        ignorable_dir=ignorable_parent_dir)
      # Issue cannot be ignored. Report failure.
      raise e

  g_common_name_regex = '(?:\w[\w\d\-_\.]*)'
  # Matches:
  # * $ENV_VAR
  # * ${env_var}
  g_common_env_var_name_regex = (
      "(?:"
      f"\${g_common_name_regex}|(?:{{{g_common_name_regex}}})"
      ")")
  # Matches:
  # * some-name
  # * some_other.name
  # * name_number_3
  # Name can only start with letter and include letters, numbers, '.'. '-', and
  # '_'.
  g_path_simple_name_regex = g_common_name_regex
  # Matches:
  # * {{place_holder}}
  g_path_placeholder_name_regex = (
      "(?:"
      # {{ is encoded into a single {
      f"{{{{{g_path_simple_name_regex}}}}}"
      f"{g_path_simple_name_regex}?"
      ")")
  # Matches:
  # * .
  # * ..
  g_path_special_name_regex = '(?:\.\.?)'
  # Matches any path name above.
  g_path_name_regex = ("(?:"
                       f"{g_path_simple_name_regex}|"
                       f"{g_path_special_name_regex}|"
                       f"{g_path_placeholder_name_regex}|"
                       f"{g_common_env_var_name_regex}"
                       ")")
  # Matches:
  # * /
  # * //
  g_abs_path_prefix_regex = '(?:\/\/?)'
  # Matches:
  # * some/path/
  # * /some/abs/path
  # * //some/other/abs/path
  # * ./some/rel/path
  # * ../.././some/other/rel/path
  # * short_path/
  # Does not match:
  # * some_path: needs at least one slash
  g_path_regex = (
      f"(?:"
      # Abs path or nothing
      f"{g_abs_path_prefix_regex}?"
      # First name ending with /
      f"{g_path_name_regex}\/"
      # Any number of names possibly ending with /
      f"(?:{g_path_name_regex}\/?)*"
      ")")

  g_include_path_arg_prefix_regex = '(?:-I)'
  g_colon_arg_prefix_regex = '(?::)'
  # Matches:
  # --i_am_argument=
  # -another-argument=
  g_explicit_arg_prefix_regex = f"(?:--?{g_common_name_regex}=)"
  # Matches:
  # * --argument=another-argument=
  # * --argument=-L
  g_explicit_repeating_arg_prefix_regex = (
      f"(?:"
      f"{g_explicit_arg_prefix_regex}"
      f"(?:(?:{g_common_name_regex}=)|(?:-\w))"
      ")")
  # Matches:
  #
  g_explicit_proto_arg_prefix_regex = '(?:M[\w_]+\.proto=)'

  # Matches any prefix above.
  g_argument_prefix_regex = (f"(?:"
                             f"{g_include_path_arg_prefix_regex}|"
                             f"{g_colon_arg_prefix_regex}|"
                             f"{g_explicit_arg_prefix_regex}|"
                             f"{g_explicit_repeating_arg_prefix_regex}|"
                             f"{g_explicit_proto_arg_prefix_regex}"
                             ")")

  # Captures:
  # 1. Group 1: arg prefix
  # 2. Group 2: path
  g_argument_regexes = (
      "^"
      f"({g_argument_prefix_regex}?)"
      # Path may be inside quote marks.
      f'(?:\\\\")?({g_path_regex})(?:\\\\")?'
      "$")

  # Matches:
  # //some_target
  # //some_target:subtarget
  g_gn_target_regex = ("^(?:"
                       f"(?:\/\/{g_common_name_regex})"
                       f"(?:\:{g_common_name_regex})?"
                       ")$")

  @staticmethod
  def FixPathInArgument(
      arg: str, fixer_callback: Callable[[str], str]) -> Tuple[str, str]:
    """
    Parses |arg| into prefix and path. Returns tuple of prefix and actual path
    fixed with given |fixer_callback|.

    If cannot parse |arg|, returns tuple of arg and empty string.

    See |PathHandler.g_path_regex| for acceptable paths.
    See |g_argument_prefix_regex| for acceptable arguments.

    |fixer_callback| shall have chroot path as an argument and return
    corresponding actual path.

    Raises:
      * PathNotFixedException if cannot resolve |path| to actual path.
      * PathNotFixedException if actual path does not exist.
    """
    # Do not sanitize the arg as it can have trailing separators required for
    # regex match.

    # Include argument may not have a path with a separator in it which is
    # required for regex. Handle it separately.
    if arg[0:2] == '-I':
      chroot_path = arg[2:]
      return ('-I', fixer_callback(chroot_path))

    match = re.match(PathHandler.g_argument_regexes, arg)
    if not match:
      if not re.match(PathHandler.g_gn_target_regex, arg):
        assert os.sep not in arg, f"Unknown arg with possible path: {arg}"

      # Argument is a gn target. Nothing to fix.

      return (arg, '')

    assert os.sep in arg, f"Unknown arg: {arg}"
    prefix = match.group(1)
    chroot_path = match.group(2)

    if chroot_path[0] == '$':
      # Path starts with env. Do not fix.
      return (arg, '')

    return (prefix, fixer_callback(chroot_path))
