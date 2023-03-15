#!/bin/bash
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e # Exit immediately if a command exits with a non-zero status.
set -x # Print commands and their arguments as they are executed.
set -o pipefail

PREFIX="$1"
CONFIG="$(pwd)/src/config"

enable_config=(
  NET_PROTO_LLDP
  DOWNLOAD_PROTO_HTTPS
  CONSOLE_FRAMEBUFFER
  NSLOOKUP_CMD
  TIME_CMD
  REBOOT_CMD
  POWEROFF_CMD
  NEIGHBOUR_CMD
  PING_CMD
  CONSOLE_CMD
  NTP_CMD
)

enable_console=(
    CONSOLE_FRAMEBUFFER
)

for setting in "${enable_config[@]}"; do
  sed -i "s/^.*${setting}.*$/#define ${setting}/" \
      "${CONFIG}/general.h"
done

for setting in "${enable_console[@]}"; do
  sed -i "s/^.*${setting}.*$/#define ${setting}/" \
      "${CONFIG}/console.h"
done

cd src
make bin-x86_64-efi/ipxe.efi
cp bin-x86_64-efi/ipxe.efi "$PREFIX"
