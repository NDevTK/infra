#!/bin/bash
# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e # Exit immediately if a command exits with a non-zero status.
set -x # Print commands and their arguments as they are executed.
set -o pipefail

PREFIX="$1"
DEPS="$2"

# Include the headers for dependencies.
EXTRA_CFLAGS="-I${DEPS}/include -I${DEPS}/include/pixman-1"

# Linker Flags to link pixman statically
EXTRA_LDFLAGS="-L${DEPS}/lib -Wl,-Bstatic,-lpixman-1,-Bdynamic -pthread"

# Include the pixman pkg_config path. Configure will complain otherwise
PKG_CONFIG_PATH="${PKG_CONFIG_PATH}:${DEPS}/lib/pkgconfig"

# Run configure with the following changes
# * Install in the PREFIX dir
# * Disable all default features, Will enable required ones
# * Enable KVM for faster amd64 and X86 execution
# * Enable VNC to help in debugging
# * Enable tools to build qemu-img
# * Disable pie as pixman doesn't link statically otherwise
# * Targets are X86_64, arm and aarch64
# * Include compiler and linker flags
./configure \
  --prefix="${PREFIX}" \
  --without-default-features \
  --enable-kvm \
  --enable-vnc \
  --enable-tools \
  --enable-cap-ng \
  --enable-attr \
  --enable-virtfs \
  --disable-pie \
  --target-list=x86_64-softmmu,arm-softmmu,aarch64-softmmu \
  --extra-ldflags="${EXTRA_LDFLAGS}" \
  --extra-cflags="${EXTRA_CFLAGS}"

# Dump config log and meson log to stdout for debugging
cat ./build/config.log
cat ./build/meson-logs/meson-log.txt

# Build qemu
make install -j $(nproc) VERBOSE=1

