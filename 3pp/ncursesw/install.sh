#!/bin/bash
# Copyright 2018 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

if ! which realpath; then
  realpath() {
    (cd "$@" && pwd)
  }
fi > /dev/null

. $(dirname $0)/cross_util.sh

if [[ $_3PP_TOOL_PLATFORM == mac-amd64 ]]; then
  # Undo global CCC_OVERRIDE_OPTIONS when building for native.
  export BUILD_CC="cc -arch x86_64"
fi

export CFLAGS="${CFLAGS} -fPIC"
export CXXFLAGS+=" -std=c++14"

# The "ncurses" package, by default, uses a fixed-path location for terminal
# information. This is not relocatable, so we need to disable it. Instead, we
# will compile ncurses with a set of hand-picked custom terminal information
# data baked in.
#
# To do this, we need to build in multiple stages:
# 1) Generic configure / make so that the "tic" (terminfo compiler) and
#    "toe" (table of entries) commands are built.
# 2) Use "toe" tool to dump the set of available profiles and groom it.
# 3) Build library with no database support using "tic" from (1), and
#    configure it to statically embed all of the profiles from (2).
tic_build="$(realpath ..)/tic_build"
tic_prefix="$(realpath ..)/tic_prefix"

# Make tic for host
(
  3pp_toggle_host

  src=$(realpath .)
  mkdir -p $tic_build
  cd $tic_build

  $src/configure --enable-widec --prefix $tic_prefix
  make install -j $(nproc)
)

# Run toe to strip out fallbacks with bugs.
#
# This currently leaves 1591 profiles behind, which will be statically
# compiled into the library.
#
# Some profiles do not generate valid C, either because:
# - They begin with a number, which is not valid in C.
# - They are flattened to a duplicate symbol as another profile. This
#   usually happens when there are "+" and "-" variants; we choose
#   "-".
# - They include quotes in the description name.
#
# None of these identified terminals are really important, so we will
# just avoid processing them.
fallback_exclusions=(
  9term
  guru\\+
  hp\\+
  tvi912b\\+
  tvi912b-vb
  tvi920b-vb
  att4415\\+
  nsterm\\+
  xnuppc\\+
  xterm\\+
  wyse-vp
)
joined=$(IFS='|'; echo "${fallback_exclusions[*]}")
fallbacks_array=($($tic_prefix/bin/toe | awk '{print $1}' | grep -Ev "^(${joined})"))
fallbacks=$(IFS=','; echo "${fallbacks_array[*]}")

# Run the remainder of our build with our generated "tic" on PATH.
#
# Note that we only run "install.libs". Standard "install" expects the
# full database to exist, and this will not be the case since we are
# explicitly disabling it.
PATH=$tic_prefix/bin:$PATH ./configure \
  --prefix=$PREFIX \
  --host=$CROSS_TRIPLE \
  --disable-database \
  --disable-db-install \
  --enable-widec \
  --with-fallbacks="$fallbacks"
make clean

# Build everything to get the timestamps warmed up. This will then fail to
# generate comp_captab.c (or init_keytry.h, depending on the race).
make install.libs -j $(nproc) || (
  # Then copy the good toolchain programs from $tic_build that we built earlier.
  cp $tic_build/ncurses/make_* ./ncurses
  # Huzzah, cross compiling C is terrible.
  make install.libs -j $(nproc)
)

# Some programs (like python) expect to be able to `#include <ncurses.h>`, so
# create that symlink. Ncurses also installs the actual header as `curses.h`
# (and creates a symlink for ncurses.h), so we link to the original file here.
(cd $PREFIX/include && ln -s ./ncursesw/curses.h ncurses.h)
(cd $PREFIX/include && ln -s ./ncursesw/panel.h panel.h)
(cd $PREFIX/include && ln -s ./ncursesw/term.h term.h)
