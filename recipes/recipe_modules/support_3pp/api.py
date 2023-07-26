# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Allows uniform cross-compiliation, version tracking and archival for
third-party software packages (libs+tools) for distribution via CIPD.

The purpose of the Third Party Packages (3pp) recipe/module is to generate CIPD
packages of statically-compiled software for distribution in our continuous
integration fleets, as well as software distributed to our develepers (e.g.
via depot_tools).

Target os and architecture uses the CIPD "${os}-${arch}" (a.k.a. "${platform}")
nomenclature, which is currently defined in terms of Go's GOOS and GOARCH
runtime variables (with the unfortunate exception that CIPD uses 'mac' instead
of 'darwin'). This is somewhat arbitrary, but has worked well so far for us.

#### Package Definitions

The 3pp module loads package definitions from a folder containing subfolders.
Each subfolder defines a single software package to fetch, build and upload.
For example, you might have a folder in your repo like this:

    my_repo.git/
      3pp/  # "root folder"
        .vpython3            # common vpython file for all package scripts
        zlib/                # zlib "package folder"
          3pp.pb             # REQUIRED: the Spec.proto definition for zlib
          install.sh         # a script to build zlib from source
          extra_resource_file
        other_package/
          3pp.pb             # REQUIRED
          fetch.py           # a script to fetch `other_package` in a custom way
          install.sh
          install-win.sh     # windows-specific build script
        ...

This defines two packages (`zlib`, and `other_package`). The 3pp.pb files have
references to the fetch/build scripts, and describe what dependencies the
packages have (if any).

NOTE: Only one layer of package folders is supported currently.

Packages are named by the folder that contains their definition file (3pp.pb)
and build scripts. It's preferable to have package named after software that it
contains. However, sometimes you want multiple major versions of the software to
exist side-by-side (e.g. pcre and pcre2, python and python3, etc.). In this
case, have two separate package definition folders.

Each package folder contains a package spec (3pp.pb), as well as scripts,
patches and/or utility tools to build the software from source.

The spec is a Text Proto document specified by the [spec.proto] schema.

The spec is broken up into two main sections, "create" and "upload". The
create section allows you to specify how the package software gets created, and
allows specifying differences in how it's fetched/built/tested on a per-target
basis, and the upload section has some details on how the final result gets
uploaded to CIPD.

[spec.proto]: /recipes/recipe_modules/support_3pp/spec.proto

#### Creation Stages

The 3pp.pb spec begins with a series of `create` messages, each with details on
on how to fetch+build+test the package.  Each create message contains
a "platform_re" field which works as a regex on the ${platform} value. All
matching patterns apply in order, and non-matching patterns are skipped. Each
create message is applied with a dict.update for each member message (i.e.
['source'].update, ['build'].update, etc.) to build a singular create message
for the current target platform. For list values (e.g. 'tool', 'dep' in the
Build message), you can clear them by providing a new empty value
(e.g. `tool: ""`)

Once all the create messages are merged (see schema for all keys that can be
present), the actual creation takes place.

Note that "source" is REQUIRED in the final merged instruction set. All other
messages are optional and have defaults as documented in [spec.proto].

The creation process is broken up into 4 different stages:

  * Source
  * Build
  * Package
  * Verify

##### Envvars

All scripts described below are invoked with the following environment variables
set:
  * $_3PP_PACKAGE_NAME - the name of the package currently building
  * $_3PP_PATCH_VERSION - the `patch_version` set for the version we're building
    (if any patch version was set).
  * $_3PP_PLATFORM - the platform we're targeting
  * $_3PP_TOOL_PLATFORM - the platform that we're building on (will be different
    than _3PP_PLATFORM if we're cross-compiling)
  * $_3PP_VERSION - the version we're building, e.g. 1.2.3
  * $GOOS - The golang OS name we're targeting
  * $GOARCH - The golang architecture we're targeting
  * $MACOSX_DEPLOYMENT_TARGET - On OS X, set to 10.10, for your
    semi-up-to-date OS X building needs. This needs to be consistently
    set for all packages or it will cause linker warnings/errors when
    linking in static libs that were targeting a newer version (e.g.
    if it was left unset). Binaries built with this set to 10.10 will not
    run on 10.9 or older systems.

Additionally, on cross-compile environments, the $CROSS_TRIPLE environment
variable is set to a GCC cross compile target triplet of cpu-vendor-os.

##### Source

The source is used to fetch the raw sources for assembling the package. In some
cases the sources may actually be binary artifacts (e.g. prebuilt windows
installers).

The source is unpacked to a checkout directory, possibly in some specified
subdirectory. Sources can either produce the actual source files, or they can
produce a single archive file (e.g. zip or tarball), which can be unpacked with
the 'unpack_archive' option. In addition, patches can be applied to the source
with the 'patch_dir' option (the patches should be in `git format-patch` format,
and will be applied with `git apply`).

  * `git` - This checks out a semver tag in the repo.
  * `cipd` - This fetches data from a CIPD package.
  * `url` - This is used for packages that do not provide a stable distribution
    like git for their source code. An original download url will be passed in
    this method to download the source artifact from third party distribution.
  * `script` - Used for "weird" packages which are distributed via e.g.
    an HTML download page or an API. The script must be able to return the
    'latest' version of its source, as well as to actually fetch a specified
    version. Python fetch scripts will be executed with `vpython3`, and so
    may have a .vpython3 file (or similar) in the usual manner to pull in
    dependencies like `requests`.

Additionally the Source message contains a `patch_version` field to allow symver
disambiguation of the built packages when they contain patches or other
alterations which need to be versioned. This string will be joined with a '.' to
the source version being built when uploading the result to CIPD.

##### Build

The build message allows you to specify `deps`, and `tools`, as well as a script
`install` which contains your logic to transform the source into the result
package.

Deps are libraries built for the target `${platform}` and are typically used for
linking your package.

Tools are binaries built for the host; they're things like `automake` or `sed`
that are used during the configure/make phase of your build, but aren't linked
into the built product. These tools will be available on $PATH (both '$tools'
and '$tools/bin' are added to $PATH, because many packages are set up with their
binaries at the base of the package, and some are set up with them in a /bin
folder)

Installation occurs by invoking the script indicated by the 'install' field
(with the appropriate interpreter, depending on the file extension) like:

    <interpreter> "$install[*]" "$PREFIX" "$DEPS_PREFIX"

Where:

  * The current working directory is the base of the source checkout w/o subdir.
  * `$install[*]` are all of the tokens in the 'install' field.
  * `$PREFIX` is the directory which the script should install everything to;
    this directory will be archived into CIPD verbatim.
  * `$DEPS_PREFIX` is the path to a prefix directory containing the union of all
    of your packages' transitive deps. For example, all of the headers of your
    deps are located at `$DEPS_PREFIX/include`.
  * All `tools` are in $PATH

If the 'install' script is omitted, it is assumed to be 'install.sh'.

If the ENTIRE build message is omitted, no build takes place. Instead the
result of the 'source' stage will be packaged.

During the execution of the build phase, the package itself and its dependent
packages (e.g. "dep" and "tool" in PB file) will be copied into the source
checkout in the .3pp directory, and the script will be invoked as
`/path/to/checkout/.3pp/<cipd_pkg_name>/$script_name`. If the package has
shared resources (like `.vpython3` files or helper scripts) which are outside of
the package directory, you would need to create a symbolic link for it. See
chromium.googlesource.com/infra/infra/+/main/3pp/cpython_common/ssl_suffix.py as
an example.

##### Package

Once the build stage is complete, all files in the $PREFIX folder passed to the
install script will be zipped into a CIPD package.

It is strongly recommended that if your package is a library or tool with many
files that it be packaged in the standard POSIXey PREFIX format (e.g. bin, lib,
include, etc.). If your package is a collection of one or more standalone
binaries, it's permissible to just have the binaries in the root of the output
$PREFIX.

If the build stage is skipped (i.e. the build message is omitted) then the
output of the source stage will be packaged instead (this is mostly useful when
using a 'script' source).

##### Verify

After the package is built it can be optionally tested. The recipe will run your
test script in an empty directory with the path to the
packaged-but-not-yet-uploaded cipd package file and it can do whatever testing
it needs to it (exiting non-zero if something is wrong). You can use the `cipd
pkg-deploy` command to deploy it (or whatever cipd commands you like, though
I wouldn't recommend uploading it to CIPD, as the 3pp recipe will do that after
the test exits 0).

Additionally, vpython3 for the tool platform will be guaranteed to be in $PATH.

##### Upload

Once the test comes back positive, the CIPD package will be uploaded to the CIPD
server and registered with the prefix indicated in the upload message. The full
CIPD package name is constructed as:

    <prefix>/<pkg_name>/${platform}

So for example with the prefix `infra`, the `bzip2` package on linux-amd64 would
be uploaded to `infra/bzip2/linux-amd64` and tagged with the version that was
built (e.g. `version:1.2.3.patch_version1`).

You can also mark the upload as a `universal` package, which will:
  * Omit the `${platform}` suffix from the upload name
  * Only build the package on the `linux-amd64' platform. This was chosen
    to ensure that "universal" packages build consistently.

#### Versions

Every package will try to build the latest identifiable semver of its source, or
will attempt to build the semver requested as an input property to the
`3pp` recipe. This semver is also used to tag the uploaded artifacts in CIPD.

Because some of the packages here are used as dependencies for others (e.g.
curl and zlib are dependencies of git, and zlib is a dependency of curl), each
package used as a dependency for others should specify its version explicitly
(currently this is only possible to do with the 'cipd' source type). So e.g.
zlib and curl specify their source versions, but git and python float at 'head',
always building the latest tagged version fetched from git.

When building a floating package (e.g. python, git) you may explicitly
state the symver that you wish to build as part of the recipe invocation.

The symver of a package (either specified in the package definition, in the
recipe properties or discovered while fetching its source code (e.g. latest git
tag)) is also used to tag the package when it's uploaded to CIPD (plus the
patch_version in the source message).

#### Cross Compilation

Third party packages are currently compiled on linux using the
'infra.tools.dockerbuild' tool from the infra.git repo. This uses a slightly
modified version of the [dockcross] Docker cross-compile environment. Windows
and OS X targets are built using the 'osx_sdk' and 'windows_sdk' recipe modules,
each of which provides a hermetic (native) build toolchain for those platforms.

For linux, we can support all the architectures implied by dockerbuild,
including:
  * linux-arm64
  * linux-armv6l
  * linux-mips32
  * linux-mips64
  * linux-amd64

[dockcross]: https://github.com/dockcross/dockcross

#### Dry runs / experiments

If the recipe is run with `force_build` it will always build all packages
indicated. Dependencies will be built if they do not exist in CIPD. None
of the built packages will be uploaded to CIPD.

The recipe must always be run with a package_prefix (by assigning to the
.package_prefix property on the Support3ppApi). If the recipe is run in
experimental mode, 'experimental/' will be prepended to this. Additionally, you
may specify `experimental: true` in the Create message for a package, which will
have the same effect when running the recipe in production (to allow adding new
packages or package/platform combintations experimentally).

#### Examples

As an example of the package definition layout in action, take a look at the
[3pp](/3pp) folder in this infra.git repo.

#### Caches

This module uses the following named caches:
  * `3pp_cipd` - Caches all downloaded and uploaded CIPD packages. Currently
    tag lookups are performed every time against the CIPD server, but this will
    hold the actual package files.
  * `osx_sdk` - Cache for `depot_tools/osx_sdk`. Only on Mac.
  * `windows_sdk` - Cache for `depot_tools/windows_sdk`. Only on Windows.
"""

import hashlib
import itertools
import posixpath
import re

from google.protobuf import json_format as jsonpb
from google.protobuf import text_format as textpb

from recipe_engine import config_types
from recipe_engine import recipe_api

from .resolved_spec import (ResolvedSpec, parse_name_version, tool_platform,
                            get_cipd_pkg_name)
from .resolved_spec import platform_for_host
from .exceptions import BadParse, DuplicatePackage, UnsupportedPackagePlatform

from . import create
from . import cipd_spec

from PB.recipe_modules.infra.support_3pp.spec import Spec


def _flatten_spec_pb_for_platform(orig_spec, platform):
  """Transforms a copy of `orig_spec` so that it contains exactly one 'create'
  message.

  Every `create` message in the original spec will be removed if its
  `platform_re` field doesn't match `platform`. The matching messages will be
  applied, in their original order, to a single `create` message by doing
  a `dict.update` operation for each submessage.

  Args:
    * orig_spec (Spec) - The message to flatten.
    * platform (str) - The CIPD platform value we're targeting (e.g.
    'linux-amd64')

  Returns a new `Spec` with exactly one `create` message, or None if
  the original spec is not supported on the given platform.
  """
  spec = Spec()
  spec.CopyFrom(orig_spec)

  resolved_create = {}

  applied_any = False
  for create_msg in spec.create:
    plat_re = "^(%s)$" % (create_msg.platform_re or '.*')
    if not re.match(plat_re, platform):
      continue

    if create_msg.unsupported:
      return None

    # We're going to apply this message to our resolved_create.
    applied_any = True

    # To get the effect of the documented dict.update behavior, round trip
    # through JSONPB. It's a bit dirty, but it works.
    dictified = jsonpb.MessageToDict(
      create_msg, preserving_proto_field_name=True)
    for k, v in dictified.items():
      if isinstance(v, dict):
        resolved_create.setdefault(k, {}).update(v)
        to_clear = set()
        for sub_k, sub_v in resolved_create[k].items():
          if isinstance(sub_v, list) and not any(val for val in sub_v):
            to_clear.add(sub_k)
        for sub_k in to_clear:
          resolved_create[k].pop(sub_k)
      else:
        resolved_create[k] = v

  if not applied_any:
    return None

  # To make this create rule self-consistent instead of just having the last
  # platform_re to be applied.
  resolved_create.pop('platform_re', None)

  # Clear list of create messages, and parse the resolved one into the sole
  # member of the list.
  del spec.create[:]
  jsonpb.ParseDict(resolved_create, spec.create.add())

  return spec


class Support3ppApi(recipe_api.RecipeApi):
  def __init__(self, **kwargs):
    super(Support3ppApi, self).__init__(**kwargs)
    # map of <cipd package name> -> (pkg_dir, Spec)
    self._loaded_specs = {}
    # all of the base_paths that have been loaded via load_packages_from_path.
    self._package_roots = set()
    # map of <package directory> -> <cipd package name>
    self._pkg_dir_to_pkg_name = {}
    # map of (name, platform) -> ResolvedSpec
    self._resolved_packages = {}
    # map of (name, version, platform) -> CIPDSpec
    self._built_packages = {}
    # Used by CIPDSpec; must be defined here so that it doesn't persist
    # accidentally between recipe tests.
    self._cipd_spec_pool = None

    # hashlib.sha256 object storing hash of support_3pp related files/modules.
    # This hash is initialized with support_3pp's self_hash, and will be updated
    # each time when a new package is loaded by load_packages_from_path method.
    self._infra_3pp_hash = hashlib.sha256()

    # (required) The package prefix to use for built packages (not for package
    # sources).
    #
    # NOTE: If `runtime.is_experimental` or `self._experimental`, then this is
    # prepended with 'experimental/'.
    self._package_prefix = ''

    self._source_cache_prefix = 'sources'

    # If running in experimental mode
    self._experimental = False

  def initialize(self):
    self._cipd_spec_pool = cipd_spec.CIPDSpecPool(self.m)
    self._infra_3pp_hash.update(self._self_hash().encode('utf-8'))
    self._experimental = self.m.runtime.is_experimental

  def package_prefix(self, experimental=False):
    """Returns the CIPD package name prefix (str), if any is set.

    This will prepend 'experimental/' to the currently set prefix if:
      * The recipe is running in experimental mode; OR
      * You pass experimental=True
    """
    pieces = []
    if experimental or self._experimental:
      pieces.append('experimental')
    pieces.append(self._package_prefix)
    return get_cipd_pkg_name(pieces)

  def set_package_prefix(self, prefix):
    """Set the CIPD package name prefix (str).

    All CIPDSpecs for built packages (not sources) will have this string
    prepended to them.
    """
    assert isinstance(prefix, str)
    self._package_prefix = prefix.strip('/')

  def set_source_cache_prefix(self, prefix):
    """Set the CIPD namespace (str) to store the source of the packages."""
    assert isinstance(prefix, str)
    assert prefix, 'A non-empty source cache prefix is required.'
    self._source_cache_prefix = prefix.strip('/')

  def set_experimental(self, experimental):
    """Set the experimental mode (bool)."""
    self._experimental = experimental

  BadParse = BadParse
  DuplicatePackage = DuplicatePackage
  UnsupportedPackagePlatform = UnsupportedPackagePlatform

  def _resolve_for(self, cipd_pkg_name, platform):
    """Resolves the build instructions for a package for the given platform.

    This will recursively resolve any 'deps' or 'tools' that the package needs.
    The resolution process is entirely 'offline' and doesn't run any steps, just
    transforms the loaded 3pp.pb specs into usable ResolvedSpec objects.

    The results of this method are cached.

    Args:
      * cipd_pkg_name (str) - The CIPD package name to resolve, excluding the
        package_prefix and platform suffix. Must be a package registered via
        `load_packages_from_path`.
      * platform (str) - A CIPD `${platform}` value to resolve this package for.
        Always in the form of `GOOS-GOARCH`, e.g. `windows-amd64`.

    Returns a ResolvedSpec.

    Raises `UnsupportedPackagePlatform` if the package cannot be built for this
    platform.
    """
    key = (cipd_pkg_name, platform)
    ret = self._resolved_packages.get(key)
    if ret:
      return ret

    pkg_dir, orig_spec = self._loaded_specs[cipd_pkg_name]

    spec = _flatten_spec_pb_for_platform(orig_spec, platform)
    if not spec:
      raise UnsupportedPackagePlatform(
        "%s not supported for %s" % (cipd_pkg_name, platform))

    create_pb = spec.create[0]
    tool_plat = tool_platform(self.m, platform, spec)

    # See the description of this logic at the top of the file.
    if spec.upload.universal:
      if tool_plat != 'linux-amd64':
        raise UnsupportedPackagePlatform(
            'Skipping universal package %s when building for %s' %
            (cipd_pkg_name, platform))
      if not create_pb.build.no_docker_env:
        # Always use the native Docker container for universal.
        platform = 'linux-amd64'

    source_pb = create_pb.source
    method = source_pb.WhichOneof('method')

    if not method:
      raise UnsupportedPackagePlatform(
          "%s not supported for %s" % (cipd_pkg_name, platform))

    if method == 'cipd':
      if not source_pb.cipd.original_download_url:  # pragma: no cover
        raise Exception(
          'CIPD Source for `%s/%s` must have `original_download_url`.' % (
          cipd_pkg_name, platform))

    assert source_pb.patch_version == source_pb.patch_version.strip('.'), (
      'source.patch_version has starting/trailing dots')

    assert spec.upload.pkg_prefix == spec.upload.pkg_prefix.strip('/'), (
      'upload.pkg_prefix has starting/trailing slashes')

    # Recursively resolve all deps
    deps = []
    for dep_cipd_pkg_name in create_pb.build.dep:
      if dep_cipd_pkg_name not in self._loaded_specs:
        raise self.m.step.StepFailure(
            'No spec was found for dep "%s" of package "%s"' %
            (dep_cipd_pkg_name, cipd_pkg_name))
      deps.append(self._resolve_for(dep_cipd_pkg_name, platform))

    # Recursively resolve all unpinned tools. Pinned tools are always
    # installed from CIPD.
    unpinned_tools = []
    for tool in create_pb.build.tool:
      tool_cipd_pkg_name, tool_version = parse_name_version(tool)
      if tool_cipd_pkg_name not in self._loaded_specs:
        raise self.m.step.StepFailure(
            'No spec was found for tool "%s" of package "%s"' %
            (tool_cipd_pkg_name, cipd_pkg_name))
      if tool_version == 'latest':
        unpinned_tools.append(self._resolve_for(tool_cipd_pkg_name, tool_plat))

    # Parse external CIPD package specs.
    def parse_external_package(pkg, tool_plat=tool_plat):
      name, version = parse_name_version(pkg)
      name = name.replace('${platform}', tool_plat)
      return name, version

    external_tools = []
    for tool in create_pb.build.external_tool:
      name, version = parse_external_package(tool)
      external_tools.append(self._cipd_spec_pool.get(name, version))

    external_deps = []
    for dep in create_pb.build.external_dep:
      name, version = parse_external_package(dep)
      external_deps.append(self._cipd_spec_pool.get(name, version))

    ret = ResolvedSpec(self.m, self._cipd_spec_pool,
                       self.package_prefix(create_pb.experimental),
                       self._source_cache_prefix, cipd_pkg_name, platform,
                       pkg_dir, spec, deps, unpinned_tools, external_tools,
                       external_deps)
    self._resolved_packages[key] = ret
    return ret

  def _build_resolved_spec(self, spec, version, force_build_packages,
                           skip_upload):
    """Builds the resolved spec. All dependencies for this spec must already be
    built.

    Args:
      * spec (ResolvedSpec) - The spec to build.
      * version (str) - The symver of the package that we want to build.
      * force_build_packages (set[str]) - Packages to build even if they already
        exist in CIPD.
      * skip_upload (bool) - If True, never upload built packages to CIPD.

    Returns CIPDSpec for the built package.
    """
    return create.build_resolved_spec(self.m, self._resolve_for,
                                      self._built_packages,
                                      force_build_packages, skip_upload, spec,
                                      version, self._infra_3pp_hash.hexdigest())

  def _glob_paths_from_git(self,
                           name,
                           source,
                           pattern):
    """glob rules for `pattern` using `git ls-files`.

    Args:
      * name (str): The name of the step.
      * source (Path): The directory whose contents should be globbed. It must
        be part of the git repository.
      * pattern (str): The glob pattern to apply under `source`.
    """
    assert isinstance(source, config_types.Path)
    result = self.m.step(
        name,
        ['git', '-C', source, 'ls-files', pattern],
        stdout=self.m.raw_io.output_text(),
    )

    # `git ls-files` always returns posix path.
    ret = [
        source.join(*x.split(posixpath.sep))
        for x in result.stdout.splitlines()
    ]
    result.presentation.logs["glob_from_git"] = [str(x) for x in ret]
    return ret


  def load_packages_from_path(self,
                              base_path,
                              glob_pattern='**/3pp.pb',
                              check_dup=True):
    """Loads all package definitions from the given base_path and glob pattern
    inside the git repository.

    To include package definitions across multiple git repository or git
    submodules, load_packages_from_path need to be called for each of the
    repository.

    This will parse and intern all the 3pp.pb package definition files so that
    packages can be identified by their cipd package name.
    For example, if you pass:

      path/
        pkgname/
          3pp.pb
          install.sh

    And the file "path/pkgname/3pp.pb" has the following content:

      upload { pkg_prefix: "my_pkg_prefix" }

    Its cipd package name will be "my_pkg_prefix/pkgname".

    Args:
      * base_path (Path) - A path of a directory where the glob_pattern will be
        applied to.
      * glob_pattern (str) - A glob pattern to look for package definition files
        "3pp.pb" whose behavior is defined by 3pp.proto. Default to "**/3pp.pb".
      * check_dup (bool): When set, it will raise DuplicatePackage error if
        a package spec is already loaded. Default to True.

    Returns a set(str) containing the cipd package names of the packages which
    were loaded.

    Raises a DuplicatePackage exception if this function encounters a cipd
    package name which is already registered. This could occur if you call
    load_packages_from_path multiple times, and one of the later calls tries to
    load a package which was registered under one of the earlier calls.
    """

    self._package_roots.add(self._NormalizePath(base_path))
    known_package_specs = self._glob_paths_from_git('find package specs',
                                                    base_path, glob_pattern)

    discovered = set()

    with self.m.step.nest('load package specs'):
      for spec_path in known_package_specs:
        # For foo/bar/3pp.pb:
        #   * the pkg_name is "bar"
        #   * the pkg_dir is "foo/bar"
        # For foo/bar/3pp/3pp.pb:
        #   * the pkg_name is "bar"
        #   * the pkg_dir is "foo/bar/3pp"

        # Relative path to the base_path (splitted in list)
        spec_path_rel = self.m.path.relpath(spec_path, base_path)
        spec_path_rel_pieces = spec_path_rel.split(self.m.path.sep)
        pkg_dir = base_path.join(*spec_path_rel_pieces[:-1])
        try:
          pkg_name = spec_path_rel_pieces[-2]
          if pkg_name == '3pp':
            pkg_name = spec_path_rel_pieces[-3]
        except IndexError:
          raise Exception('Fail to get package name from %r. Expecting '
                          'the spec PB in a deeper folder' % spec_path)

        data = self.m.file.read_text('read %r' % spec_path_rel, spec_path)
        if not data:
          self.m.step.active_result.presentation.status = self.m.step.FAILURE
          raise self.m.step.StepFailure('Bad spec PB for %r' % pkg_name)

        try:
          pkg_spec = textpb.Merge(data, Spec())
        except Exception as ex:
          raise BadParse('While adding %r: %r' % (spec_path_rel, str(ex)))

        # CIPD package name, excluding the package_prefix and platform suffix
        cipd_pkg_name = get_cipd_pkg_name(
            [pkg_spec.upload.pkg_prefix, pkg_name])
        if cipd_pkg_name in self._loaded_specs:
          if check_dup:
            raise DuplicatePackage(cipd_pkg_name)
        else:
          # Add package hash each time recipe loads a new package
          self._infra_3pp_hash.update(
              self.m.file.compute_hash('Compute hash for %r' % cipd_pkg_name,
                                       [pkg_dir], self.m.path['start_dir'],
                                       'deadbeef').encode('utf-8'))
          self._loaded_specs[cipd_pkg_name] = (pkg_dir, pkg_spec)

          self._pkg_dir_to_pkg_name[self._NormalizePath(
              self.m.path.dirname(spec_path))] = cipd_pkg_name

        discovered.add(cipd_pkg_name)

    return discovered

  def _self_hash(self):
    """Calculates SHA256 hash for support_3pp module and recipe.

    This includes:
      * The hash of this file (3pp.py)
      * The hash of the support_3pp recipe module
      * The hash of the dockerbuild script directory

    Returns:
      A list[str] of 3 hashes computed.

    Raises a BadDirectoryPermission exception if cannot access a file in given
    directory. This in turn might stop further recipe execution.
    """
    # Base path for calculating hash.
    base_path = self.repo_resource()

    # File path of support_3pp recipe module.
    recipe_module_path = self.repo_resource('recipes', 'recipe_modules',
                                            'support_3pp')

    # File path of 3pp.py file.
    recipe_file_path = self.repo_resource('recipes', 'recipes', '3pp.py')

    # File path of infra/tools/dockerbuild.
    dockerbuild_path = self.repo_resource('infra', 'tools', 'dockerbuild')

    paths = [recipe_module_path, recipe_file_path, dockerbuild_path]

    return self.m.file.compute_hash(name='compute recipe file hash',
                                    paths=paths,
                                    base_path=base_path,
                                    test_data='deadbeef')

  # Ensures that the given path is resolved using the longest matching
  # Path root. This avoids confusion about whether the paths passed to
  # this module are rooted in 'cache' vs 'checkout'.
  def _NormalizePath(self, p):
    return self.m.path.abs_to_path(str(p))

  def _ComputeAffectedPackages(self, tryserver_affected_files):
    reverse_deps = {}  # cipd_pkg_name -> Set[cipd_pkg_name]
    for cipd_pkg_name, dir_spec in self._loaded_specs.items():
      for create_msg in dir_spec[1].create:
        for dep in itertools.chain(create_msg.build.dep, create_msg.build.tool):
          # Some packages depend on themselves for cross-compiling, ignore this.
          if dep != cipd_pkg_name:
            reverse_deps.setdefault(dep, set()).add(cipd_pkg_name)

    affected_packages = set()

    def _AddDependenciesRecurse(pkg_name):
      if pkg_name in affected_packages:
        # All reverse deps are already added.
        return
      affected_packages.add(pkg_name)
      for dep in reverse_deps.get(pkg_name, []):
        _AddDependenciesRecurse(dep)

    for f in tryserver_affected_files:
      # Map each affected file to a package. If there are files under the
      # 3pp root that cannot be mapped to a package, conservatively rebuild
      # everything.

      # Re-normalize the affected file path, as we did for the package
      # directories in load_packages_from_path().
      resolved_file = self._NormalizePath(f)
      found_pkg = False
      for pkg_dir, pkg_name in self._pkg_dir_to_pkg_name.items():
        if pkg_dir.is_parent_of(resolved_file):
          _AddDependenciesRecurse(pkg_name)
          found_pkg = True
          break
      if not found_pkg and any(
          dir.is_parent_of(resolved_file) for dir in self._package_roots):
        # The file is under a package root, but not corresponding to a
        # particular package.
        return []

    return list(sorted(affected_packages))

  def ensure_uploaded(self,
                      packages=(),
                      platform='',
                      force_build=False,
                      tryserver_affected_files=()):
    """Executes entire {fetch,build,package,verify,upload} pipeline for all the
    packages listed, targeting the given platform.

    Args:
      * packages (seq[str]) - A sequence of packages to ensure are
        uploaded. Packages must be listed as either 'pkgname' or
        'pkgname@version'. If empty, builds all loaded packages.
      * platform (str) - If specified, the CIPD ${platform} to build for.
        If unspecified, this will be the appropriate CIPD ${platform} for the
        current host machine.
      * force_build (bool) - If True, all applicable packages and their
        dependencies will be built, regardless of the presence in CIPD.
        The source and built packages will not be uploaded to CIPD.
      * tryserver_affected_files (seq[Path]) - If given, run in tryserver mode
        where the specified files (which must correspond to paths that were
        passed to load_packages_from_path) have been modified in the CL. All
        affected packages, and any that depend on them (recursively) are built.
        If any files are modified which cannot be mapped to a specific package,
        all packages are rebuilt. Overrides 'packages', and forces
        force_build=True (packages are never uploaded in this mode).

    Returns (list[(cipd_pkg, cipd_version)], set[str]) of built CIPD packages
    and their tagged versions, as well as a list of unsupported packages.
    """
    if tryserver_affected_files:
      force_build = True
      with self.m.step.nest('compute affected packages') as p:
        packages = self._ComputeAffectedPackages(tryserver_affected_files)
        p.step_text = ','.join(packages) if packages else 'all packages'

    unsupported = set()

    explicit_build_plan = []
    packages = packages or list(self._loaded_specs.keys())
    force_build_packages = set(packages) if force_build else set()
    platform = platform or platform_for_host(self.m)
    with self.m.step.nest('compute build plan'):
      for pkg in packages:
        cipd_pkg_name, version = parse_name_version(pkg)
        if cipd_pkg_name not in self._loaded_specs:
          raise self.m.step.StepFailure(
              'No spec found for requested package "%s"' % cipd_pkg_name)
        try:
          resolved = self._resolve_for(cipd_pkg_name, platform)
        except UnsupportedPackagePlatform:
          unsupported.add(pkg)
          continue
        explicit_build_plan.append((resolved, version))

      expanded_build_plan = set()
      for spec, version in explicit_build_plan:
        expanded_build_plan.add((spec, version))
        # Anything pulled in either via deps or tools dependencies should be
        # explicitly built at 'latest'.
        for sub_spec in spec.all_possible_deps_and_tools:
          expanded_build_plan.add((sub_spec, 'latest'))

    ret = []
    with self.m.step.defer_results():
      for spec, version in sorted(expanded_build_plan):
        # Never upload packages for the build platform which are incidentally
        # built when cross-compiling. These end up racing with the native
        # builders and leading to multiple CIPD instances being tagged with
        # the same version.
        skip_upload = force_build or (spec.platform != platform)
        ret.append(
            self._build_resolved_spec(spec, version, force_build_packages,
                                      skip_upload))
    return ret, unsupported
