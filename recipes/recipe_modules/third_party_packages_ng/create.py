# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from . import spec_pb2

from . import source

from .workdir import Workdir


def build_resolved_spec(api, spec, version, spec_lookup, cache):
  """Builds a resolved spec at a specific version, and uploads it.

  If the package is already built on the remote server, just returns a CIPDSpec
  for it.

  Args:
    * api - The ThirdPartyPackagesNGApi's `self.m` module collection.
    * spec (ResolvedSpec) - The resolved spec to build.
    * version (str) - The symver (or 'latest') version of the package to build.
    * spec_lookup ((package_name, platform) -> ResolvedSpec) - A function to
      lookup (possibly cached) ResolvedSpec's for things like dependencies and
      tools.
    * cache (dict) - A map of (package_name, version, platform) -> CIPDSpec.
      The `build_resolved_spec` function fully manages the content of this
      dictionary.

  Returns the CIPDSpec of the built package.
  """
  keys = [(spec.name, version, spec.platform)]
  if keys[0] in cache:
    return cache[keys[0]]

  cipd_spec = None
  def set_cache():
    for k in keys:
      cache[k] = cipd_spec

  with api.step.nest('building %s' % (spec.name.encode('utf-8'),)):
    env = {
      '_3PP_PLATFORM': spec.platform,
      '_3PP_PACKAGE_NAME': spec.name,
    }
    if spec.create_pb.source.patch_version:
      env['_3PP_PATCH_VERSION'] = spec.create_pb.source.patch_version

    with api.context(env=env):
      # Resolve 'latest' versions. Done inside the env because 'script' based
      # sources need the $_3PP* envvars.
      is_latest = version == 'latest'
      if is_latest:
        version = source.resolve_latest(api, spec)
        keys.append((spec.name, version, spec.platform))
        if keys[1] in cache:
          cipd_spec = cache[keys[1]]
          set_cache()
          return cipd_spec

      # See if the specific version is uploaded
      cipd_spec = spec.cipd_spec(version)
      if cipd_spec.check():
        set_cache()
        return cipd_spec

      ### OK, we need to actually build this
      workdir = Workdir(api, spec, version)
      with api.context(env={'_3PP_VERSION': version}):
        api.file.ensure_directory('mkdir -p [workdir]', workdir.base)

        with api.step.nest('prep checkout'):
          source.prep_checkout(
            api, workdir, spec, version, spec_lookup,
            lambda spec, version: build_resolved_spec(
              api, spec, version, spec_lookup, cache))

        # TODO(iannucci): build

        # TODO(iannucci): package

        # TODO(iannucci): verify

        # TODO(iannucci): upload

        set_cache()
        return cipd_spec
