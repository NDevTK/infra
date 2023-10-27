#!/bin/bash
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e # Exit immediately if a command exits with a non-zero status.
set -x # Print commands and their arguments as they are executed.
set -o pipefail

PREFIX="$1"

. $(dirname $0)/cross_util.sh

if [[ "$_3PP_PLATFORM" == 'linux-amd64' ]]; then
  binary='ipxe.efi'
  build_path="bin-x86_64-efi"
elif [[ "$_3PP_PLATFORM" == 'linux-arm64' ]]; then
  binary='ipxe-arm64.efi'
  build_path="bin-arm64-efi"
fi

config="$(pwd)/src/config"
enable_config=(
    'NET_PROTO_LLDP'
    'DOWNLOAD_PROTO_HTTPS'
    'CONSOLE_FRAMEBUFFER'
    'NSLOOKUP_CMD'
    'TIME_CMD'
    'REBOOT_CMD'
    'POWEROFF_CMD'
    'NEIGHBOUR_CMD'
    'PING_CMD'
    'CONSOLE_CMD'
    'NTP_CMD'
)

enable_console=(
    'CONSOLE_FRAMEBUFFER'
)

for setting in "${enable_config[@]}"; do
  sed -i "s/^.*${setting}.*$/#define ${setting}/" \
      "${config}/general.h"
done

for setting in "${enable_console[@]}"; do
  sed -i "s/^.*${setting}.*$/#define ${setting}/" \
      "${config}/console.h"
done

# Set HOST_CC if cross-compiling.
MAKE_ARGS=
if [[ $_3PP_PLATFORM != $_3PP_TOOL_PLATFORM ]]; then
  MAKE_ARGS+=HOST_CC=$(3pp_toggle_host; echo $CC)
fi

# The build system tries to get the version from .git, but this won't
# exist when building from a cached source.
rm -rf .git
MAKE_ARGS+="\
  VERSION_MAJOR=1 \
  VERSION_MINOR=21 \
  VERSION_PATCH=1 \
  EXTRAVERSION=${_3PP_VERSION} \
"

cd src
make ${MAKE_ARGS} "${build_path}/ipxe.efi"
cp "${build_path}/ipxe.efi" "${PREFIX}/${binary}"
