#!/usr/bin/env vpython
# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""This script rebuilds Python & Go universes of infra.git multiverse and
invokes CIPD client to package and upload chunks of it to the CIPD repository as
individual packages.

See build/packages/*.yaml for definition of packages and README.md for more
details.
"""

import argparse
import collections
import contextlib
import copy
import errno
import glob
import hashlib
import json
import os
import platform
import re
import shutil
import socket
import subprocess
import sys
import tempfile

import yaml


# Root of infra.git repository.
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# Where to upload packages to by default.
PACKAGE_REPO_SERVICE = 'https://chrome-infra-packages.appspot.com'

# Hash algorithm to use for calculating instance IDs.
HASH_ALGO = 'sha256'

# True if running on Windows.
IS_WINDOWS = sys.platform == 'win32'

# .exe on Windows.
EXE_SUFFIX = '.exe' if IS_WINDOWS else ''

# All GOARCHs we are willing to cross-compile for.
KNOWN_GOARCHS = frozenset([
    '386',
    'amd64',
    'arm',
    'arm64',
    'mips',
    'mips64',
    'mips64le',
    'mipsle',
    'ppc64',
    'ppc64le',
    'riscv64',
    's390x',
])

# All platforms support 'go build -race'.
RACE_SUPPORTED_PLATFORMS = frozenset([
    'linux-amd64',
    'freebsd-amd64',
    'darwin-amd64',
    'windows-amd64',
    'linux-ppc64le',
    'linux-arm64',
])

# A package prefix => cwd to use when building this package.
INFRA_MODULE_MAP = {
    # The luci-go module is checked out separately, use its go.mod.
    'go.chromium.org/luci/':
        os.path.join(ROOT, 'go', 'src', 'go.chromium.org', 'luci'),
    # All infra packages should use go.mod in infra.git.
    'infra/':
        os.path.join(ROOT, 'go', 'src', 'infra'),
    # Use infra's go.mod when building goldctl.
    'go.skia.org/infra/gold-client/cmd/goldctl':
        os.path.join(ROOT, 'go', 'src', 'infra')
}


class PackageDefException(Exception):
  """Raised if a package definition is invalid."""
  def __init__(self, path, msg):
    super(PackageDefException, self).__init__('%s: %s' % (path, msg))


class BuildException(Exception):
  """Raised on errors during package build step."""


class SearchException(Exception):
  """Raised on errors during package search step."""


class TagException(Exception):
  """Raised on errors during package tag step."""


class UploadException(Exception):
  """Raised on errors during package upload step."""


class PackageDef(collections.namedtuple(
    '_PackageDef', ('path', 'pkg_def'))):
  """Represents parsed package *.yaml file."""

  @property
  def name(self):
    """Returns name of YAML file (without the directory path and extension)."""
    return os.path.splitext(os.path.basename(self.path))[0]

  @property
  def disabled(self):
    """Returns True if the package should be excluded from the build."""
    return self.pkg_def.get('disabled', False)

  @property
  def uses_python_env(self):
    """Returns True if 'uses_python_env' in the YAML file is set."""
    return bool(self.pkg_def.get('uses_python_env'))

  @property
  def update_latest_ref(self):
    """Returns True if 'update_latest_ref' in the YAML file is set.

    Defaults to True."""
    return bool(self.pkg_def.get('update_latest_ref', True))

  @property
  def go_packages(self):
    """Returns a list of Go packages that must be installed for this package."""
    return self.pkg_def.get('go_packages') or []

  def cgo_enabled(self, target_goos):
    """Either True, False or None (meaning "let go decide itself").

    Ignored when cross-compiling.
    """
    val = self.pkg_def.get('go_build_environ', {}).get('CGO_ENABLED')
    if isinstance(val, dict):
      val = val.get(target_goos)
    return None if val is None else bool(val)

  @property
  def pkg_root(self):
    """Absolute path to a package root directory."""
    root = self.pkg_def['root'].replace('/', os.sep)
    if os.path.isabs(root):
      return root
    return os.path.abspath(os.path.join(os.path.dirname(self.path), root))

  def with_race(self, target_goos, target_goarch):
    """Returns True if should build with `-race` flag.

    To build with race:
      - race should be enabled on the target_goos or in general;
      - cgo should be enabled on target_goos;
      - target_goos should be one of the supported platforms.
    """
    val = self.pkg_def.get('go_build_environ', {}).get('race')
    if isinstance(val, dict):
      val = val.get(target_goos)

    if not val:
      return False

    cgo_enabled = self.cgo_enabled(target_goos)
    if not cgo_enabled:
      print(
          'go build -race cannot be enabled because CGO is not enabled on %s' %
          target_goos)
      return False

    platform = '%s-%s' % (target_goos, target_goarch)
    if platform in RACE_SUPPORTED_PLATFORMS:
      return True
    print('go build -race is not supported on %s' % platform)
    return False

  def validate(self):
    """Raises PackageDefException if the package definition looks invalid."""
    for var_name in self.pkg_def.get('go_build_environ', {}):
      if var_name not in ['CGO_ENABLED', 'race']:
        raise PackageDefException(
            self.path,
            'Only "CGO_ENABLED" and "race" is supported in "go_build_environ" currently'
        )

  def should_visit(self):
    """Returns True if package targets the current platform."""
    # If the package doesn't have 'platforms' set, assume it doesn't want to be
    # cross-compiled, and supports only native host platform or it's platform
    # independent. Otherwise build it only if the target of the compilation is
    # declared as supported. Note that these are CIPD-flavored platform strings
    # (e.g. "mac-amd64"), exactly like they appear in CIPD package names.
    platforms = self.pkg_def.get('platforms')
    if not platforms:
      return not is_cross_compiling()
    return get_package_vars()['platform'] in platforms

  def preprocess(self, build_root, pkg_vars, cipd_exe, sign_id=None):
    """Parses the definition and filters/extends it before passing to CIPD.

    This process may generate additional files that are put into the package.

    Args:
      build_root: root directory for building cipd package.
      pkg_vars: dict with variables passed to cipd as -pkg-var.
      cipd_exe: path to cipd executable.
      sign_id: identity used for Mac codesign.

    Returns:
      Path to filtered package definition YAML.

    Raises:
      BuildException on error.
    """
    pkg_def = copy.deepcopy(self.pkg_def)

    pkg_def['root'] = build_root

    bat_files = [
        d['file'] for d in pkg_def['data'] if d.get('generate_bat_shim')
    ]

    def process_cipd_export(ensure_contents,
                            dest,
                            pkg_vars=pkg_vars,
                            cipd_exe=cipd_exe):
      # Render target_platform in the ensure file.
      ensure_contents = ensure_contents.replace('${target_platform}',
                                                pkg_vars['platform'])
      cipd_export(ensure_contents, dest, cipd_exe)

    if 'mac_bundle' in pkg_def:
      bundle_def = pkg_def['mac_bundle']
      bundle = create_mac_bundle(build_root, bundle_def)
      pkg_def['data'].append({
          'dir':
              os.path.relpath(bundle['root'], build_root).replace(os.sep, '/')
      })

      for d in bundle_def['data']:
        if 'file' in d:
          file_path = render_path(d['file'], pkg_vars)
          src = os.path.join(self.pkg_root, file_path)
          dst = os.path.join(bundle['files_root'], d['path'],
                             os.path.basename(file_path))
          shutil.copy(src, dst)
        elif 'cipd_export' in d:
          process_cipd_export(d['cipd_export'], bundle['root'])

      if 'codesign' in bundle_def:
        cmd = ['/usr/bin/codesign', '--deep', '--force']
        if sign_id:
          for k, v in bundle_def['codesign'].items():
            cmd.extend(['--' + k, v])
          cmd.extend(['--sign', sign_id])
        else:
          # Ignoring all codesign args and use ad-hoc signing for testing.
          cmd.extend(['--sign', '-'])
        cmd.append(bundle['root'])

        print('Running %s' % ' '.join(cmd))
        subprocess.check_call(cmd)

    for cp in pkg_def.get('copies', ()):
      plat = cp.get('platforms')
      if plat and pkg_vars['platform'] not in plat:
        continue
      dst = os.path.join(build_root, render_path(cp['dst'], pkg_vars))
      shutil.copy(os.path.join(self.pkg_root, render_path(cp['src'], pkg_vars)),
                  dst)
      pkg_def['data'].append(
          {'file': os.path.relpath(dst, build_root).replace(os.sep, '/')})
      if cp.get('generate_bat_shim'):
        bat_files.append(cp['dst'])

    if 'cipd_export' in pkg_def:
      process_cipd_export(pkg_def['cipd_export'], build_root)

    # Copy all included files into build root if not existed. This must be after
    # steps generating files and before any steps referring a symbolic link.
    for d in self.pkg_def['data']:
      path = d.get('file') or d.get('dir')
      if path:
        copy_if_not_exist(self.pkg_root, build_root, path, pkg_vars)

    if not is_targeting_windows(pkg_vars):
      for sym in pkg_def.get('posix_symlinks', ()):
        dst = os.path.join(build_root, render_path(sym['dst'], pkg_vars))
        try:
          os.remove(dst)
        except OSError:
          pass
        os.symlink(
            os.path.join(build_root, render_path(sym['src'], pkg_vars)), dst)
        pkg_def['data'].append(
            {'file': os.path.relpath(dst, build_root).replace(os.sep, '/')})

    # Generate *.bat shims when targeting Windows.
    if is_targeting_windows(pkg_vars):
      for f in bat_files:
        # Generate actual *.bat.
        bat_abs = generate_bat_shim(build_root, render_path(f, pkg_vars))
        # Make it part of the package definition (use slash paths there).
        pkg_def['data'].append(
            {'file': os.path.relpath(bat_abs, build_root).replace(os.sep, '/')})
        # Stage it for cleanup.

    # Keep generated yaml in the same directory to avoid rewriting paths.
    out_path = os.path.join(build_root, self.name + '.processed_yaml')
    with open(out_path, 'w') as f:
      json.dump(pkg_def, f)
    return out_path

  def on_change_info(self, pkg_vars):
    """Returns tags and path to check package changed."""
    on_change_tags = [
        get_on_change_tag(self.pkg_root, d, pkg_vars)
        for d in self.pkg_def.get('upload_on_change', [])
    ]
    pkg_path = render_path(
        self.pkg_def.get('package'), pkg_vars, replace_sep=False)
    return on_change_tags, pkg_path

# Carries modifications for go-related env vars and cwd.
#
# If a field has value None, it will be popped from the environment in
# 'apply_to_environ'.
class GoEnviron(
    collections.namedtuple(
        'GoEnviron', ['GOOS', 'GOARCH', 'CGO_ENABLED', 'cwd', 'with_race'])):

  @staticmethod
  def host_native():
    """Returns GoEnviron that instructs Go to not cross-compile."""
    return GoEnviron(
        GOOS=None,
        GOARCH=None,
        CGO_ENABLED=None,
        cwd=os.getcwd(),
        with_race=False,
    )

  @staticmethod
  def from_environ():
    """Reads GoEnviron from the current os.environ.

    If CGO_ENABLED is not given, picks the default based on whether we are
    cross-compiling or not. cgo is disabled by default when cross-compiling.
    """
    cgo = os.environ.get('CGO_ENABLED')
    if cgo is None:
      cgo = not os.environ.get('GOOS')
    else:
      cgo = cgo == '1'
    return GoEnviron(
        GOOS=os.environ.get('GOOS'),
        GOARCH=os.environ.get('GOARCH'),
        CGO_ENABLED=cgo,
        cwd=os.getcwd(),
        with_race=False,
    )

  def apply_to_environ(self):
    """Applies GoEnviron to the current os.environ and cwd."""
    if self.GOOS is not None:
      os.environ['GOOS'] = self.GOOS
    else:
      os.environ.pop('GOOS', None)
    if self.GOARCH is not None:
      os.environ['GOARCH'] = self.GOARCH
    else:
      os.environ.pop('GOARCH', None)
    if self.CGO_ENABLED is not None:
      os.environ['CGO_ENABLED'] = '1' if self.CGO_ENABLED else '0'
    else:
      os.environ.pop('CGO_ENABLED', None)
    if self.cwd is not None:
      os.chdir(self.cwd)


def render_path(p, pkg_vars, replace_sep=True):
  """Renders ${...} substitutions in paths, converts them to native slash."""
  for k, v in pkg_vars.iteritems():
    assert '${' not in v  # just in case, to avoid recursive expansion
    p = p.replace('${%s}' % k, v)
  if replace_sep:
    return os.path.normpath(p.replace('/', os.sep))
  return p


def copy_if_not_exist(src_root, dst_root, path, pkg_vars):
  """Copies a file from src_root to dst_root if it doesn't exist there."""
  file_path = render_path(path, pkg_vars)
  src = os.path.join(src_root, file_path)
  dst = os.path.join(dst_root, file_path)
  if os.path.exists(dst):
    return

  try:
    os.makedirs(os.path.dirname(dst))
  except OSError as e:
    if e.errno != errno.EEXIST:
      raise

  copy_tree(src_root, src, dst)


def cipd_export(ensure_contents, dst_root, cipd_exe):
  """Installs cipd_pkg with the given version tag to dst_root."""
  args = [cipd_exe, 'export', '-ensure-file', '-', '-root', dst_root]
  cmd = subprocess.Popen(
      args,
      stdin=subprocess.PIPE,
      stderr=subprocess.STDOUT,
      executable=cipd_exe)
  out, _ = cmd.communicate(ensure_contents)
  if cmd.returncode:
    raise subprocess.CalledProcessError(cmd.returncode, args, output=out)


def copy_tree(src_root, src, dst):
  """Copies a directory from src to dst. If it's a symlink, convert it pointing
  to relative path."""
  if os.path.islink(src):
    linkto = os.readlink(src)
    if os.path.commonprefix([src_root, linkto]) == src_root:
      linkto = os.path.relpath(linkto, os.path.dirname(src))
    os.symlink(linkto, dst)
  elif os.path.isdir(src):
    os.mkdir(dst)
    for name in os.listdir(src):
      copy_tree(src_root, os.path.join(src, name), os.path.join(dst, name))
  else:
    shutil.copy(src, dst)


def generate_bat_shim(pkg_root, target_rel):
  """Writes a shim file side-by-side with target and returns abs path to it."""
  target_name = os.path.basename(target_rel)
  bat_name = os.path.splitext(target_name)[0] + '.bat'
  base_dir = os.path.dirname(os.path.join(pkg_root, target_rel))
  bat_path = os.path.join(base_dir, bat_name)
  with open(bat_path, 'w') as fd:
    fd.write('\n'.join([  # python turns \n into CRLF
        '@set CIPD_EXE_SHIM="%%~dp0%s"' % (target_name,),
        '@shift',
        '@%CIPD_EXE_SHIM% %*',
        '',
    ]))
  return bat_path


def create_mac_bundle(pkg_root, bundle_def):
  """
  Generate the Mac Bundle structure.

  something.app
    Contents
      Info.plist
      MacOS
        something
      _CodeSignature # Generated by codesign
        CodeResources
  """
  bundle_root = os.path.join(pkg_root, bundle_def['name'])
  shutil.rmtree(bundle_root, ignore_errors=True)

  contents_path = os.path.join(bundle_root, 'Contents')
  os.makedirs(contents_path)

  with open(os.path.join(contents_path, 'Info.plist'), 'w') as info_plist:
    info_plist.write(bundle_def['info'])

  files_root = os.path.join(contents_path, 'MacOS')
  os.mkdir(files_root)
  return {
      'root': bundle_root,
      'files_root': files_root,
  }


def is_cross_compiling():
  """Returns True if using GOOS or GOARCH env vars.

  We also check at the start of the script that if one of them is used, then
  the other is specified as well.
  """
  return bool(os.environ.get('GOOS')) or bool(os.environ.get('GOARCH'))


def get_env_dot_py():
  if os.environ.get('GOOS') == 'android':
    return 'mobile_env.py'
  else:
    return 'env.py'


def run_python(script, args):
  """Invokes a python script via the root python interpreter.

  Escapes virtualenv if finds itself running with VIRTUAL_ENV env var set.

  Raises:
    BuildException if couldn't find a proper python binary.
    subprocess.CalledProcessError on non zero exit code.
  """
  environ = os.environ.copy()
  python_exe = sys.executable

  venv = environ.pop('VIRTUAL_ENV')
  if venv:
    path = environ['PATH'].split(os.pathsep)
    path = [p for p in path if not p.startswith(venv+os.sep)]
    environ['PATH'] = os.pathsep.join(path)
    # Popen doesn't use new env['PATH'] to search for binaries. Do it ourselves.
    for p in path:
      candidate = os.path.join(p, 'python' + EXE_SUFFIX)
      if os.path.exists(candidate):
        python_exe = candidate
        break
    else:
      raise BuildException(
          'Could\'n find python%s in %s' % (EXE_SUFFIX, environ['PATH']))

  print 'Running %s %s' % (script, ' '.join(args))
  print '  via %s' % python_exe
  print '  in  %s' % os.getcwd()
  subprocess.check_call(
      args=['python', '-u', script] + list(args),
      executable=python_exe,
      env=environ)


def find_cipd():
  """Finds a CIPD client in PATH."""
  exts = ('.exe', '.bat') if sys.platform == 'win32' else ('',)
  for p in os.environ.get('PATH', '').split(os.pathsep):
    base = os.path.join(p, 'cipd')
    for ext in exts:
      candidate = base + ext
      if os.path.isfile(candidate):
        return candidate
  return 'cipd' + EXE_SUFFIX


def run_cipd(cipd_exe, cmd, args):
  """Invokes CIPD, parsing -json-output result.

  Args:
    cipd_exe: path to cipd client binary to run.
    cmd: cipd subcommand to run.
    args: list of command line arguments to pass to the subcommand.

  Returns:
    (Process exit code, parsed JSON output or None).
  """
  temp_file = None
  try:
    fd, temp_file = tempfile.mkstemp(suffix='.json', prefix='cipd_%s' % cmd)
    os.close(fd)

    cmd_line = [cipd_exe, cmd, '-json-output', temp_file] + list(args)

    print 'Running %s' % ' '.join(cmd_line)
    exit_code = subprocess.call(args=cmd_line, executable=cmd_line[0])
    try:
      with open(temp_file, 'r') as f:
        json_output = json.load(f)
    except (IOError, ValueError):
      json_output = None

    return exit_code, json_output
  finally:
    try:
      if temp_file:
        os.remove(temp_file)
    except OSError:
      pass


def print_title(title):
  """Pretty prints a banner to stdout."""
  sys.stdout.flush()
  sys.stderr.flush()
  print
  print '-' * 80
  print title
  print '-' * 80


def print_go_step_title(title):
  """Same as 'print_title', but also appends values of GOOS, GOARCH, etc."""
  go_mod = None
  if os.environ.get('GO111MODULE') != 'off' and os.path.exists('go.mod'):
    go_mod = os.path.abspath('go.mod')
  go_vars = [(k, os.environ[k])
             for k in ('GOOS', 'GOARCH', 'GOARM', 'CGO_ENABLED')
             if k in os.environ]
  if go_vars or go_mod:
    title += '\n' + '-' * 80
  if go_mod:
    title += '\n  go.mod: %s' % go_mod
  for k, v in go_vars:
    title += '\n  %s=%s' % (k, v)
  print_title(title)


@contextlib.contextmanager
def workspace_env(go_environ):
  """Puts Go env vars from go_environ into os.environ and changes cwd.

  Args:
    go_environ: instance of GoEnviron object with go related env vars.
  """
  orig_cwd = os.getcwd()
  orig_environ = os.environ.copy()

  # Change os.environ and cwd.
  go_environ.apply_to_environ()

  # Make sure we build ARMv6 code even if the host is ARMv7. See the comment in
  # get_host_package_vars for reasons why. Also explicitly set GOARM to 6 when
  # cross-compiling (it should be '6' in this case by default anyway).
  plat = platform.machine().lower()
  if plat.startswith('arm') or os.environ.get('GOARCH') == 'arm':
    os.environ['GOARM'] = '6'
  else:
    os.environ.pop('GOARM', None)

  try:
    yield
  finally:
    os.chdir(orig_cwd)
    # Apparently 'os.environ = orig_environ' doesn't actually modify process
    # environment, only modifications of os.environ object itself do.
    for k, v in orig_environ.iteritems():
      os.environ[k] = v
    for k in os.environ.keys():
      if k not in orig_environ:
        os.environ.pop(k)


def bootstrap_go_toolset(go_workspace):
  """Makes sure go is installed and returns its 'go env' and version.

  Used to verify that our platform detection in get_host_package_vars() matches
  the Go toolset being used.
  """
  with workspace_env(GoEnviron.host_native()):
    print_go_step_title('Making sure Go toolset is installed')
    # env.py does the actual job of bootstrapping if the toolset is missing.
    output = subprocess.check_output(
        args=[
            'python', '-u',
            os.path.join(go_workspace, get_env_dot_py()), 'go', 'env'
        ],
        executable=sys.executable)
    # See https://github.com/golang/go/blob/master/src/cmd/go/env.go for format
    # of the output.
    print 'Go environ:'
    print output.strip()
    env = {}
    for line in output.splitlines():
      k, _, v = line.lstrip('set ').partition('=')
      if v.startswith('"') and v.endswith('"'):
        v = v.strip('"')
      env[k] = v

    # This would be something like "go version go1.15.8 darwin/amd64".
    print_go_step_title('Go version')
    output = subprocess.check_output(
        args=[
            'python', '-u',
            os.path.join(go_workspace, get_env_dot_py()), 'go', 'version'
        ],
        executable=sys.executable)
    print output.strip()

    # We want only "go1.15.8" part.
    version = re.match(r'go version (go[\d\.]+)', output).group(1)

    return env, version


def run_go_clean(go_workspace, go_environ, packages):
  """Removes object files and executables left from building given packages.

  Transitively cleans all dependencies (including stdlib!) and removes
  executables from GOBIN. In Go modules mode this also appears to be downloading
  modules.

  Args:
    go_workspace: path to 'infra/go' or 'infra_internal/go'.
    go_environ: instance of GoEnviron object with go related env vars.
    packages: list of go packages to clean (can include '...' patterns).
  """
  with workspace_env(go_environ):
    print_go_step_title('Preparing:\n  %s' % '\n  '.join(packages))
    subprocess.check_call(
        args=[
            'python', '-u',
            os.path.join(go_workspace, get_env_dot_py()), 'go', 'clean', '-i',
            '-r'
        ] + list(packages),
        executable=sys.executable,
        stderr=subprocess.STDOUT)
    # Above command is either silent (without '-x') or too verbose (with '-x').
    # Prefer silent version, but add a note that it's alright.
    print 'Done.'


def run_go_install(go_workspace, go_environ, packages):
  """Builds (and installs) Go packages into GOBIN via 'go install ...'.

  Compiles and installs packages into default GOBIN, which is <go_workspace>/bin
  (it is setup by go/env.py).

  Args:
    go_workspace: path to 'infra/go' or 'infra_internal/go'.
    go_environ: instance of GoEnviron object with go related env vars.
    packages: list of go packages to build (can include '...' patterns).
    rebuild: if True, will forcefully rebuild all dependences.
  """
  args = [
      'python', '-u',
      os.path.join(go_workspace, get_env_dot_py()), 'go', 'install',
      '-trimpath', '-ldflags=-buildid=', '-v'
  ]
  if go_environ.with_race:
    args.append('-race')

  args += list(packages)
  with workspace_env(go_environ):
    print_go_step_title('Building:\n  %s' % '\n  '.join(packages))
    subprocess.check_call(
        args=args, executable=sys.executable, stderr=subprocess.STDOUT)


def run_go_build(go_workspace, go_environ, package, output):
  """Builds single Go package.

  Args:
    go_workspace: path to 'infra/go' or 'infra_internal/go'.
    go_environ: instance of GoEnviron object with go related env vars.
    package: go package to build.
    output: where to put the resulting binary.
  """
  args = [
      'python', '-u',
      os.path.join(go_workspace, get_env_dot_py()), 'go', 'build', '-trimpath',
      '-ldflags=-buildid=', '-v', '-o', output
  ]
  if go_environ.with_race:
    args.append('-race')

  args.append(package)
  with workspace_env(go_environ):
    print_go_step_title('Building %s' % (package,))
    subprocess.check_call(
        args=args, executable=sys.executable, stderr=subprocess.STDOUT)


def find_main_module(module_map, pkg):
  """Returns a path to the main module to use when building `pkg`.

  Args:
    module_map: a dict "go package prefix => directory with main module".
    pkg: a Go package name to look up.
  """
  matches = set()
  for pfx, main_dir in module_map.items():
    if pkg.startswith(pfx):
      matches.add(main_dir)
  if len(matches) == 0:
    raise BuildException(
        'Package %r is not in the module map %s' %
        (pkg, module_map))
  if len(matches) > 1:
    raise BuildException(
        'Package %r matches multiple modules in the module map %s' %
        (pkg, module_map))
  return list(matches)[0]


def build_go_code(go_workspace, module_map, pkg_defs):
  """Builds and installs all Go packages used by the given PackageDefs.

  Understands GOOS and GOARCH and uses slightly different build strategy when
  cross-compiling. In the end <go_workspace>/bin will have all built binaries,
  and only them (regardless of whether we are cross-compiling or not).

  Args:
    go_workspace: path to 'infra/go' or 'infra_internal/go'.
    module_map: a dict "go package prefix => directory with main module".
    pkg_defs: list of PackageDef objects that define what to build.
  """
  # Exclude all disabled packages.
  pkg_defs = [p for p in pkg_defs if not p.disabled]

  # Whatever GOOS, GOARCH, etc were passed from outside. They are set when
  # cross-compiling.
  default_environ = GoEnviron.from_environ()

  # The OS we compiling for (defaulting to the host OS).
  target_goos = default_environ.GOOS or get_host_goos()
  host_vars = get_host_package_vars()
  target_goarch = default_environ.GOARCH or host_vars['platform'].split('-')[1]

  # Grab a set of all go packages we need to build and install into GOBIN,
  # figuring out a go environment (and cwd) they want.
  go_packages = {}  # go package name => [GoEnviron]

  for pkg_def in pkg_defs:
    pkg_env = default_environ
    pkg_env = pkg_env._replace(
        with_race=pkg_def.with_race(target_goos, target_goarch))
    if not is_cross_compiling():
      cgo_enabled = pkg_def.cgo_enabled(target_goos)
      if cgo_enabled is not None:
        pkg_env = pkg_env._replace(CGO_ENABLED=cgo_enabled)
    for name in pkg_def.go_packages:
      pkg_env = pkg_env._replace(cwd=find_main_module(module_map, name))
      if name not in go_packages:
        go_packages[name] = []
      go_packages[name].append(pkg_env)

  # Group packages by the environment they want.
  packages_per_env = {}  # GoEnviron => [str]
  for name, pkg_envs in go_packages.iteritems():
    for pkg_env in pkg_envs:
      packages_per_env.setdefault(pkg_env, []).append(name)

  # Execute build command for each individual environment.
  for pkg_env, to_install in sorted(packages_per_env.iteritems()):
    to_install = sorted(to_install)
    if not to_install:
      continue

    # Make sure there are no stale files in the workspace.
    run_go_clean(go_workspace, pkg_env, to_install)

    if not is_cross_compiling():
      # If not cross-compiling, build all Go code in a single "go install" step,
      # it's faster that way. We can't do that when cross-compiling, since
      # 'go install' isn't supposed to be used for cross-compilation and the
      # toolset actively complains with "go install: cannot install
      # cross-compiled binaries when GOBIN is set".
      run_go_install(go_workspace, pkg_env, to_install)
    else:
      # Prebuild stdlib once. 'go build' calls below are discarding build
      # results, so it's better to install as much shared stuff as possible
      # beforehand.
      run_go_install(go_workspace, pkg_env, ['std'])

      # Build packages one by one and put the resulting binaries into GOBIN, as
      # if they were installed there. It's where the rest of the build.py code
      # expects them to be (see also 'root' property in package definition
      # YAMLs).
      go_bin = os.path.join(go_workspace, 'bin')
      exe_suffix = get_package_vars()['exe_suffix']
      for pkg in to_install:
        bin_name = pkg[pkg.rfind('/')+1:] + exe_suffix
        run_go_build(go_workspace, pkg_env, pkg, os.path.join(go_bin, bin_name))


def enumerate_packages(package_def_dir, package_def_files):
  """Returns a list PackageDef instances for files in build/packages/*.yaml.

  Args:
    package_def_dir: path to build/packages dir to search for *.yaml.
    package_def_files: optional list of filenames to limit results to.

  Returns:
    List of PackageDef instances parsed from *.yaml files under packages_dir.
  """
  paths = []
  if not package_def_files:
    # All existing packages by default.
    paths = glob.glob(os.path.join(package_def_dir, '*.yaml'))
  else:
    # Otherwise pick only the ones in 'package_def_files' list.
    for name in package_def_files:
      abs_path = os.path.abspath(os.path.join(package_def_dir, name))
      if not os.path.isfile(abs_path):
        raise PackageDefException(name, 'No such package definition file')
      paths.append(abs_path)
  # Load and validate YAMLs.
  pkgs = []
  for p in sorted(paths):
    pkg = PackageDef(p, read_yaml(p))
    pkg.validate()
    pkgs.append(pkg)
  return pkgs


def read_yaml(path):
  """Returns content of YAML file as python dict."""
  with open(path, 'rb') as f:
    return yaml.safe_load(f)


def get_package_vars():
  """Returns a dict with variables that describe the package target environment.

  Variables can be referenced in the package definition YAML as
  ${variable_name}. It allows to reuse exact same definition file for similar
  packages (e.g. packages with same cross platform binary, but for different
  platforms).

  If running in cross-compilation mode, uses GOOS and GOARCH to figure out the
  target platform instead of examining the host environment.
  """
  if is_cross_compiling():
    return get_target_package_vars()
  return get_host_package_vars()


def get_target_package_vars():
  """Returns a dict with variables that describe cross-compilation target env.

  Examines os.environ for GOOS, GOARCH and GOARM.

  The returned dict contains only 'platform' and 'exe_suffix' entries.
  """
  assert is_cross_compiling()
  goos = os.environ['GOOS']
  goarch = os.environ['GOARCH']

  if goarch not in KNOWN_GOARCHS:
    raise BuildException('Unsupported GOARCH %s' % goarch)

  # There are many ARMs, pick the concrete instruction set. 'v6' is the default,
  # don't try to support other variants for now. Note that 'GOARM' doesn't apply
  # to 'arm64' arch.
  #
  # See:
  #   https://golang.org/doc/install/source#environment
  #   https://github.com/golang/go/wiki/GoArm
  if goarch == 'arm':
    goarm = os.environ.get('GOARM', '6')
    if goarm != '6':
      raise BuildException('Unsupported GOARM value %s' % goarm)
    arch = 'armv6l'
  else:
    arch = goarch

  # We use 'mac' instead of 'darwin'.
  if goos == 'darwin':
    goos = 'mac'

  return {
      'exe_suffix': '.exe' if goos == 'windows' else '',
      'platform': '%s-%s' % (goos, arch),
  }


def get_linux_host_arch():
  """Returns: The Linux host architecture, or None if it could not be resolved.
  """
  try:
    # Query "dpkg" to identify the userspace architecture.
    return subprocess.check_output(['dpkg', '--print-architecture']).strip()
  except OSError:
    # This Linux distribution doesn't use "dpkg".
    return None


def get_host_package_vars():
  """Returns a dict with variables that describe the current host environment.

  The returned platform may not match the machine environment exactly, but it is
  compatible with it.

  For example, on ARMv7 machines we claim that we are in fact running ARMv6
  (which is subset of ARMv7), since we don't really care about v7 over v6
  difference and want to reduce the variability in supported architectures
  instead.

  Similarly, if running on 64-bit Linux with 32-bit user space (based on python
  interpreter bitness), we claim that machine is 32-bit, since most 32-bit Linux
  Chrome Infra bots are in fact running 64-bit kernels with 32-bit userlands.
  """
  # linux, mac or windows.
  platform_variant = {
      'darwin': 'mac',
      'linux2': 'linux',
      'win32': 'windows',
  }.get(sys.platform)
  if not platform_variant:
    raise ValueError('Unknown OS: %s' % sys.platform)

  sys_arch = None
  if sys.platform == 'linux2':
    sys_arch = get_linux_host_arch()

  # If we didn't override our system architecture, identify it using "platform".
  sys_arch = sys_arch or platform.machine()

  # amd64, 386, etc.
  platform_arch = {
      'amd64': 'amd64',
      'i386': '386',
      'i686': '386',
      'x86': '386',
      'x86_64': 'amd64',
      'arm64': 'arm64',
      'armv6l': 'armv6l',
      'armv7l': 'armv6l',  # we prefer to use older instruction set for builds
  }.get(sys_arch.lower())
  if not platform_arch:
    raise ValueError('Unknown machine arch: %s' % sys_arch)

  # Most 32-bit Linux Chrome Infra bots are in fact running 64-bit kernel with
  # 32-bit userland. Detect this case (based on bitness of the python
  # interpreter) and report the bot as '386'.
  if (platform_variant == 'linux' and
      platform_arch == 'amd64' and
      sys.maxsize == (2 ** 31) - 1):
    platform_arch = '386'

  return {
      # e.g. '.exe' or ''.
      'exe_suffix': EXE_SUFFIX,
      # e.g. 'linux-amd64'
      'platform': '%s-%s' % (platform_variant, platform_arch),
  }


def get_host_goos():
  """Returns GOOS value matching the host that builds the package."""
  goos = {
      'darwin': 'darwin',
      'linux2': 'linux',
      'win32': 'windows',
  }.get(sys.platform)
  if not goos:
    raise ValueError('Unknown OS: %s' % sys.platform)
  return goos


def is_targeting_windows(pkg_vars):
  """Returns true if 'platform' in pkg_vars indicates Windows."""
  return pkg_vars['platform'].startswith('windows-')


def get_on_change_tag(root, pkg_data, pkg_vars):
  """Get the tag for detecting package on change"""
  h = hashlib.sha256()
  data_file = render_path(pkg_data['file'], pkg_vars)
  with open(os.path.join(root, data_file), 'rb') as f:
    for chunk in iter(lambda: f.read(h.block_size * 256), b""):
      h.update(chunk)
  return ':'.join(['on_change', data_file, h.name, h.hexdigest()])


def build_pkg(cipd_exe, pkg_def, out_file, package_vars, sign_id=None):
  """Invokes CIPD client to build a package.

  Args:
    cipd_exe: path to cipd client binary to use.
    pkg_def: instance of PackageDef representing this package.
    out_file: where to store the built package.
    package_vars: dict with variables to pass as -pkg-var to cipd.
    sign_id: identity used for Mac codesign.

  Returns:
    {'package': <name>, 'instance_id': <hash>}

  Raises:
    BuildException on error.
  """
  print_title('Building: %s' % os.path.basename(out_file))

  # Make sure not stale output remains.
  if os.path.isfile(out_file):
    os.remove(out_file)

  try:
    build_root = tempfile.mkdtemp(prefix="build_py")

    # Parse the definition and filter/extend it before passing to CIPD. This
    # process may generate additional files that are put into the package.
    processed_yaml = pkg_def.preprocess(
        build_root, package_vars, cipd_exe, sign_id=sign_id)

    # Build the package.
    args = ['-pkg-def', processed_yaml]
    for k, v in sorted(package_vars.items()):
      args.extend(['-pkg-var', '%s:%s' % (k, v)])
    args.extend(['-out', out_file])
    args.extend(['-hash-algo', HASH_ALGO])
    exit_code, json_output = run_cipd(cipd_exe, 'pkg-build', args)
    if exit_code:
      print
      print >> sys.stderr, 'FAILED! ' * 10
      raise BuildException('Failed to build the CIPD package, see logs')

    # Expected result is {'package': 'name', 'instance_id': 'hash'}
    info = json_output['result']
    print '%s %s' % (info['package'], info['instance_id'])
    return info
  finally:
    shutil.rmtree(build_root, ignore_errors=True)


def upload_pkg(cipd_exe, pkg_file, service_url, tags, update_latest_ref,
               service_account):
  """Uploads existing *.cipd file to the storage and tags it.

  Args:
    cipd_exe: path to cipd client binary to use.
    pkg_file: path to *.cipd file to upload.
    service_url: URL of a package repository service.
    tags: a list of tags to attach to uploaded package instance.
    update_latest_ref: a bool of whether or not to update the 'latest' CIPD ref
    service_account: path to *.json file with service account to use.

  Returns:
    {'package': <name>, 'instance_id': <hash>}

  Raises:
    UploadException on error.
  """
  print_title('Uploading: %s' % os.path.basename(pkg_file))

  args = ['-service-url', service_url]
  for tag in sorted(tags):
    args.extend(['-tag', tag])
  if update_latest_ref:
    args.extend(['-ref', 'latest'])
  if service_account:
    args.extend(['-service-account-json', service_account])
  args.extend(['-hash-algo', HASH_ALGO])
  args.append(pkg_file)
  exit_code, json_output = run_cipd(cipd_exe, 'pkg-register', args)
  if exit_code:
    print
    print >> sys.stderr, 'FAILED! ' * 10
    raise UploadException('Failed to upload the CIPD package, see logs')
  info = json_output['result']
  info['url'] = '%s/p/%s/+/%s' % (
      service_url, info['package'], info['instance_id'])
  print '%s %s' % (info['package'], info['instance_id'])
  return info


def search_pkg(cipd_exe, pkg_name, service_url, tags, service_account):
  """Search existing cipd packages with given tags.

  Args:
    cipd_exe: path to cipd client binary to use.
    pkg_name: name of the cipd package.
    service_url: URL of a package repository service.
    tags: tags to search in the package repository.
    service_account: path to *.json file with service account to use.

  Returns:
    {'package': <name>, 'instance_id': <hash>}

  Raises:
    SearchException on error.
  """
  print_title('Searching: %s by on_change tags %s' % (pkg_name, tags))

  args = ['-service-url', service_url]
  for tag in tags:
    args.extend(['-tag', tag])
  if service_account:
    args.extend(['-service-account-json', service_account])
  args.append(pkg_name)
  exit_code, json_output = run_cipd(cipd_exe, 'search', args)
  if exit_code:
    print
    print >> sys.stderr, 'FAILED! ' * 10
    raise SearchException('Failed to search the CIPD package, see logs')
  result = json_output['result']
  if result and len(result) > 1:
    print
    print >> sys.stderr, 'FAILED! ' * 10
    raise SearchException('Multiple CIPD package matched, %s', result)
  return result[0] if result else None


def tag_pkg(cipd_exe, pkg_name, pkg_version, service_url, tags,
            service_account):
  """Tag existing cipd package with given tags.

  Args:
    cipd_exe: path to cipd client binary to use.
    pkg_name: name of the cipd package.
    pkg_version: version of the cipd package.
    service_url: URL of a package repository service.
    tags: tags to set to the cipd package
    service_account: path to *.json file with service account to use.

  Raises:
    TagException on error.
  """
  print_title('Tagging: %s, %s' % (pkg_name, tags))

  args = ['-service-url', service_url]
  for tag in tags:
    args.extend(['-tag', tag])
  if service_account:
    args.extend(['-service-account-json', service_account])
  args.extend(['-version', pkg_version])
  args.append(pkg_name)
  exit_code, _ = run_cipd(cipd_exe, 'set-tag', args)
  if exit_code:
    print
    print >> sys.stderr, 'FAILED! ' * 10
    raise TagException('Failed to tag the CIPD package, see logs')


def get_build_out_file(package_out_dir, pkg_def):
  """Returns a path where to put built *.cipd package file.

  Args:
    package_out_dir: root directory where to put *.cipd files.
    pkg_def: instance of PackageDef being built.
  """
  # When cross-compiling, append a suffix to package file name to indicate that
  # it's for foreign platform.
  sfx = ''
  if is_cross_compiling():
    sfx = '+' + get_target_package_vars()['platform']
  return os.path.join(package_out_dir, pkg_def.name + sfx + '.cipd')


def run(
    go_workspace,
    build_callback,
    builder,
    package_def_dir,
    package_out_dir,
    package_def_files,
    build,
    upload,
    sign_id,
    service_url,
    tags,
    service_account_json,
    json_output,
    want_refresh_python,
):
  """Rebuilds python and Go universes and CIPD packages.

  Args:
    go_workspace: path to 'infra/go' or 'infra_internal/go'.
    build_callback: called to build binaries, virtual environment, etc.
    builder: name of CI buildbot builder that invoked the script.
    package_def_dir: path to build/packages dir to search for *.yaml.
    package_out_dir: where to put built packages.
    package_def_files: names of *.yaml files in package_def_dir or [] for all.
    build: False to skip building packages (valid only when upload==True).
    upload: True to also upload built packages, False just to build them.
    sign_id: identity used for Mac codesign.
    service_url: URL of a package repository service.
    tags: a list of tags to attach to uploaded package instances.
    service_account_json: path to *.json service account credential.
    json_output: path to *.json file to write info about built packages to.
    want_refresh_python: want a refreshed python ENV.

  Returns:
    0 on success, 1 or error.
  """
  assert build or upload, 'Both build and upload are False, nothing to do'

  # We need both GOOS and GOARCH or none.
  if is_cross_compiling():
    if not os.environ.get('GOOS') or not os.environ.get('GOARCH'):
      print >> sys.stderr, (
          'When cross-compiling both GOOS and GOARCH environment variables '
          'must be set.')
      return 1
    if os.environ.get('GOARM', '6') != '6':
      print >> sys.stderr, 'Only GOARM=6 is supported for now.'
      return 1

  # Load all package definitions and pick ones we want to build (based on
  # whether we are cross-compiling or not).
  try:
    defs = enumerate_packages(package_def_dir, package_def_files)
  except PackageDefException as exc:
    print >> sys.stderr, exc
    return 1
  packages_to_visit = [p for p in defs if p.should_visit()]

  # Make sure we have a Go toolset and it matches the host platform we detected
  # in get_host_package_vars(). Otherwise we may end up uploading wrong binaries
  # under host platform CIPD package suffix. It's important on Linux with 64-bit
  # kernel and 32-bit userland (we must use 32-bit Go in that case, even if
  # 64-bit Go works too).
  go_env, go_ver = bootstrap_go_toolset(go_workspace)
  host_vars = get_host_package_vars()
  expected_arch = host_vars['platform'].split('-')[1]
  if go_env['GOHOSTARCH'] != expected_arch:
    print >> sys.stderr, (
        'Go toolset GOHOSTARCH (%s) doesn\'t match expected architecture (%s)' %
        (go_env['GOHOSTARCH'], expected_arch))
    return 1

  # Append tags related to the build host. They are especially important when
  # cross-compiling: cross-compiled packages can be identified by comparing the
  # platform in the package name with value of 'build_host_platform' tag.
  tags = list(tags)
  tags.append('build_host_hostname:' + socket.gethostname().split('.')[0])
  tags.append('build_host_platform:' + host_vars['platform'])
  tags.append('go_version:' + go_ver)

  print_title('Overview')
  if upload:
    print 'Service URL: %s' % service_url
    print
  if builder:
    print 'Package definition files to process on %s:' % builder
  else:
    print 'Package definition files to process:'
  for pkg_def in packages_to_visit:
    print '  %s' % pkg_def.name
  if not packages_to_visit:
    print '  <none>'
  print
  print 'Variables to pass to CIPD:'
  package_vars = get_package_vars()
  for k, v in sorted(package_vars.items()):
    print '  %s = %s' % (k, v)
  if upload and tags:
    print
    print 'Tags to attach to uploaded packages:'
    for tag in sorted(tags):
      print '  %s' % tag
  if not packages_to_visit:
    print
    print 'Nothing to do.'
    return 0

  # Find a CIPD client in PATH to use for building and uploading packages.
  print_title('CIPD client')
  cipd_exe = find_cipd()
  print 'Binary: %s' % cipd_exe
  subprocess.check_call(['cipd', 'version'], executable=cipd_exe)

  # Remove old build artifacts to avoid stale files in case the script crashes
  # for some reason.
  if build:
    print_title('Cleaning %s' % package_out_dir)
    if not os.path.exists(package_out_dir):
      os.makedirs(package_out_dir)
    cleaned = False
    for pkg_def in packages_to_visit:
      out_file = get_build_out_file(package_out_dir, pkg_def)
      if os.path.exists(out_file):
        print 'Removing stale %s' % os.path.basename(out_file)
        os.remove(out_file)
        cleaned = True
    if not cleaned:
      print 'Nothing to clean'

  # Build the world.
  if build:
    should_refresh_python = (
        want_refresh_python and
        any(p.uses_python_env for p in packages_to_visit) and
        not is_cross_compiling())

    build_callback(packages_to_visit, should_refresh_python)

  # Package it.
  failed = []
  succeeded = []
  for pkg_def in packages_to_visit:
    if pkg_def.disabled:
      print_title('Skipping building disabled %s' % pkg_def.name)
      continue
    out_file = get_build_out_file(package_out_dir, pkg_def)
    try:
      info = None
      if build:
        info = build_pkg(
            cipd_exe, pkg_def, out_file, package_vars, sign_id=sign_id)
      if upload:
        if pkg_def.uses_python_env and not builder:
          print(
              'Not uploading %s, since it uses a system Python enviornment '
              'and that enviornment is only valid on builders.' %
              (pkg_def.name,))
          continue

        on_change_tags, pkg_path = pkg_def.on_change_info(package_vars)
        if on_change_tags:
          existed_pkg = search_pkg(cipd_exe, pkg_path, service_url,
                                   on_change_tags, service_account_json)
          if existed_pkg:
            print('Not uploading %s, since all change tags are present.'
                  ' result: %s' % (pkg_def.name, existed_pkg))
            tag_pkg(
                cipd_exe,
                existed_pkg['package'],
                existed_pkg['instance_id'],
                service_url,
                tags,
                service_account_json,
            )
            succeeded.append({
                'pkg_def_name': pkg_def.name,
                'info': existed_pkg
            })
            continue
          tags.extend(on_change_tags)

        info = upload_pkg(
            cipd_exe,
            out_file,
            service_url,
            tags,
            pkg_def.update_latest_ref,
            service_account_json,
        )
      assert info is not None
      succeeded.append({'pkg_def_name': pkg_def.name, 'info': info})
    except (BuildException, UploadException) as e:
      failed.append({'pkg_def_name': pkg_def.name, 'error': str(e)})

  print_title('Summary')
  for d in failed:
    print 'FAILED %s, see log above' % d['pkg_def_name']
  for d in succeeded:
    print '%s %s' % (d['info']['package'], d['info']['instance_id'])

  if json_output:
    with open(json_output, 'w') as f:
      summary = {
          'failed': failed,
          'succeeded': succeeded,
          'tags': sorted(tags),
          'vars': package_vars,
      }
      json.dump(summary, f, sort_keys=True, indent=2, separators=(',', ': '))

  return 1 if failed else 0


def build_infra(pkg_defs, should_refresh_python):
  """Builds infra.git multiverse.

  Args:
    pkg_defs: list of PackageDef instances for packages being built.
  """
  if should_refresh_python:
    print_title('Making sure python virtual environment is fresh')
    run_python(
        script=os.path.join(ROOT, 'bootstrap', 'bootstrap.py'),
        args=[
            '--deps_file',
            os.path.join(ROOT, 'bootstrap', 'deps.pyl'),
            os.path.join(ROOT, 'ENV'),
        ])
  # Build all necessary go binaries.
  build_go_code(os.path.join(ROOT, 'go'), INFRA_MODULE_MAP, pkg_defs)


def main(
    args,
    build_callback=build_infra,
    go_workspace=os.path.join(ROOT, 'go'),
    package_def_dir=os.path.join(ROOT, 'build', 'packages'),
    package_out_dir=os.path.join(ROOT, 'build', 'out')):
  parser = argparse.ArgumentParser(description='Builds infra CIPD packages')
  parser.add_argument(
      'yamls', metavar='YAML', type=str, nargs='*',
      help='name of a file in build/packages/* with the package definition')
  parser.add_argument(
      '--signing-identity',
      metavar='IDENTITY',
      dest='sign_id',
      default=None,
      help='Signing identity used for mac codesign. '
      'Use adhoc sign if not provided.')
  parser.add_argument(
      '--upload',  action='store_true', dest='upload', default=False,
      help='upload packages into the repository')
  parser.add_argument(
      '--no-rebuild',  action='store_false', dest='build', default=True,
      help='when used with --upload means upload existing *.cipd files')
  parser.add_argument(
      '--builder', metavar='NAME', type=str,
      help='Name of the CI buildbot builder that invokes this script.')
  parser.add_argument(
      '--service-url', metavar='URL', dest='service_url',
      default=PACKAGE_REPO_SERVICE,
      help='URL of the package repository service to use')
  parser.add_argument(
      '--service-account-json', metavar='PATH', dest='service_account_json',
      help='path to credentials for service account to use')
  parser.add_argument(
      '--json-output', metavar='PATH', dest='json_output',
      help='where to dump info about built package instances')
  parser.add_argument(
      '--tags', metavar='KEY:VALUE', type=str, dest='tags', nargs='*',
      help='tags to attach to uploaded package instances')
  parser.add_argument(
      '--no-freshen-python-env',
      action='store_false',
      dest='refresh_python',
      default=True,
      help=('skip freshening the python env. ' +
            'Only use if you know the env is clean.'))
  args = parser.parse_args(args)
  if not args.build and not args.upload:
    parser.error('--no-rebuild doesn\'t make sense without --upload')
  return run(
      go_workspace,
      build_callback,
      args.builder,
      package_def_dir,
      package_out_dir,
      [n + '.yaml' if not n.endswith('.yaml') else n for n in args.yamls],
      args.build,
      args.upload,
      args.sign_id,
      args.service_url,
      args.tags or [],
      args.service_account_json,
      args.json_output,
      args.refresh_python,
  )


if __name__ == '__main__':
  sys.exit(main(sys.argv[1:]))
