#!/bin/bash
# Copyright 2018 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"
DEPS_PREFIX="$2"

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# Transform the patch version into a "git tag" name that will be
# embedded into the python build info.
GITTAG_NAME=$(echo -n "$_3PP_PATCH_VERSION" \
  | tr '[:upper:]' '[:lower:]' | tr -c '.[:alnum:]' '[-*]')
GITTAG="/bin/echo -n ${GITTAG_NAME}"

# The Python build system normally uses git to fill in part of the
# build id. However, when building from cached source, .git directories
# are not preserved. Always remove .git in order to get a consistent
# result.
rm -rf .git

# Make sure we don't pick up any modules from the host PYTHONPATH.
export PYTHONPATH=""

CPPFLAGS="-I$DEPS_PREFIX/include"
LDFLAGS="-L$DEPS_PREFIX/lib"

export CONFIG_ARGS="--host $CROSS_TRIPLE"

SETUP_LOCAL_SKIP=(
  # `crypt` module is deprecated in python3.11 and slated for removal.
  # See peps.python.org/pep-0594/#crypt.
  "_crypt"
  # `nis` module is deprecated in python3.11 and slated for removal.
  # See peps.python.org/pep-0594/#nis.
  "nis"
  # These modules are broken, and seem to reference non-existent symbols
  # at compile time.
  "_testcapi"
  "_testinternalcapi"
  # We don't have these libraries on Mac, and don't want to depend on
  # them on Linux either.
  "_tkinter"
)
SETUP_LOCAL_ATTACH=(
  "$DEPS_PREFIX/lib/libbz2.a"
  "$DEPS_PREFIX/lib/libreadline.a"
  "$DEPS_PREFIX/lib/libpanelw.a"
  "$DEPS_PREFIX/lib/libncursesw.a"
  "$DEPS_PREFIX/lib/libsqlite3.a"
  "$DEPS_PREFIX/lib/libz.a"
  "$DEPS_PREFIX/lib/liblzma.a"
  "$DEPS_PREFIX/lib/libssl.a"
  "$DEPS_PREFIX/lib/libcrypto.a"
  "$DEPS_PREFIX/lib/libuuid.a"

  # We always use the OSS ncurses headers; on OS X the system headers are weird
  # and python's configure file works around that by not setting the
  # XOPEN_SOURCE defines. Unfortunately that means that on OS X the configure
  # script gets it wrong.
  #
  # We set the NCURSES_WIDECHAR variable explicitly, as that's the only intended
  # side effect of setting the XOPEN_SOURCE defines. Setting XOPEN_SOURCE
  # defines ourselves leads to problems in other headers which we still use :(.
  "_curses:: -DNCURSES_WIDECHAR=1"
  "_curses_panel:: -DNCURSES_WIDECHAR=1"
)

WITH_LIBS="-lpthread"

# TODO(iannucci): Remove this once the fleet is using GLIBC 2.25 and
# macOS 10.12 or higher.
#
# See comment in 3pp/openssl/install.sh for more detail.
export ac_cv_func_getentropy=0

# Never link against libcrypt, even if it is present, as it will not
# necessarily be present on the target system.
export ac_cv_search_crypt=no
export ac_cv_search_crypt_r=no

if [[ $_3PP_PLATFORM == mac* ]]; then
  PYTHONEXE=python.exe
  USE_SYSTEM_FFI=true

  # Instruct Mac to prefer ".a" files in earlier library search paths
  # rather than search all of the paths for a ".dylib" and then, failing
  # that, do a second sweep for ".a".
  LDFLAGS="$LDFLAGS -Wl,-search_paths_first"

  # For use with cross-compiling.
  if [[ $_3PP_TOOL_PLATFORM == mac-arm64 ]]; then
      host_cpu="aarch64"
  else
      host_cpu="x86_64"
  fi
  EXTRA_CONFIGURE_ARGS="$EXTRA_CONFIGURE_ARGS --build=${host_cpu}-apple-darwin"
else
  PYTHONEXE=python
  USE_SYSTEM_FFI=

  EXTRA_CONFIGURE_ARGS="--with-dbmliborder=bdb:gdbm"
  # NOTE: This can break building on Mac builder, causing it to freeze
  # during execution.
  #
  # Maybe look into this if we have time later.
  # Also disable PGO when cross-compiling, since a profile can't be generated.
  if [[ $_3PP_TOOL_PLATFORM == $_3PP_PLATFORM ]]; then
    EXTRA_CONFIGURE_ARGS="$EXTRA_CONFIGURE_ARGS --enable-optimizations"
  else
    # We still want this flag, which is for some reason only used when
    # PGO is enabled.
    CFLAGS_NODIST="-fno-semantic-interposition"
    LDFLAGS_NODIST="-fno-semantic-interposition"
  fi

  # TODO(iannucci) This assumes we're building for linux under docker (which is
  # currently true).
  EXTRA_CONFIGURE_ARGS="$EXTRA_CONFIGURE_ARGS --build=x86_64-linux-gnu"

  # OpenSSL 1.1.1 depends on pthread, so it needs to come LAST. Python's
  # Makefile has BASEMODLIBS which is used last when linking the final
  # executable.
  BASEMODLIBS="-lpthread"

  # Linux requires -lrt.
  WITH_LIBS+=" -lrt"

  # sqlite3 requires -lm (log function).
  WITH_LIBS+=" -lm"

  # On Linux, we need to ensure that most symbols from our static-embedded
  # libraries (notably OpenSSL) don't get exported. If they do, they can
  # conflict with the same libraries from wheels or other dynamically
  # linked sources.
  #
  # This set of symbols was determined by trial, see:
  # - crbug.com/763792
  #
  # We use LDFLAGS_NODIST instead of LDFLAGS so that distutils doesn't use this
  # for building extensions. It would break the build, as gnu_version_script.txt
  # isn't available when we build wheels. It's not necessary there anyway.
  LDFLAGS_NODIST=" -Wl,--version-script=$SCRIPT_DIR/gnu_version_script.txt"

  if [[ $_3PP_PLATFORM != $_3PP_TOOL_PLATFORM ]]; then
    # -pthread detection does not work when cross-compiling, but we need this
    # flag in order for OpenSSL libraries to get their symbols resolved.
    export ac_cv_pthread=yes
    export ac_cv_cxx_thread=yes
    export ac_cv_pthread_system_supported=yes
  fi
fi

# Assert blindly that the target distro will have /dev/ptmx and not /dev/ptc.
# This is likely to be true, since Mac and all linuxes that we know of have this
# configuration.
export ac_cv_file__dev_ptmx=y
export ac_cv_file__dev_ptc=n

if [[ $USE_SYSTEM_FFI ]]; then
  EXTRA_CONFIGURE_ARGS+=" --with-system-ffi"
  SETUP_LOCAL_ATTACH+=("_ctypes::-lffi")
else
  EXTRA_CONFIGURE_ARGS+=" --without-system-ffi"
  SETUP_LOCAL_ATTACH+=("$DEPS_PREFIX/lib/libffi.a")
fi

if [[ $_3PP_PLATFORM != $_3PP_TOOL_PLATFORM ]]; then  # cross compiling
  BUILD_PYTHON=`which python3`
  EXTRA_CONFIGURE_ARGS+=" --with-build-python=${BUILD_PYTHON}"
fi

# Avoid querying altstack size dynamically on armv6l because the dockcross
# image we are using don't have the sys/auxv.h in glibc. The autoconf doesn't
# take sys/auxv.h into account so we need to manually disable the detection of
# linux/auxvec.h. This can be removed when we move to a newer version of the
# glibc.
# See also: https://github.com/python/cpython/pull/31789
if [[ $_3PP_PLATFORM  == "linux-armv6l" ]]; then
  export ac_cv_header_linux_auxvec_h=n
fi

# Python tends to hard-code /usr/include and /usr/local/include in its setup.py
# file which can end up picking up headers and stuff from wherever.
sed -i \
  "s+/usr/include+$DEPS_PREFIX/include+" \
  setup.py
sed -i \
  "s+/usr/include+$DEPS_PREFIX/include+" \
  configure.ac
sed -i \
  "s+/usr/local/include+$DEPS_PREFIX/include+" \
  setup.py
sed -i \
  "s+/usr/lib+$DEPS_PREFIX/lib+" \
  setup.py

# Generate our configure script.
autoconf

export LDFLAGS
export LDFLAGS_NODIST
export CPPFLAGS
export CFLAGS_NODIST
# Configure our production Python build with our static configuration
# environment and generate our basic platform.
#
# We're going to use our bootstrap python interpreter to generate our static
# module list.
if ! ./configure --prefix "$PREFIX" --host="$CROSS_TRIPLE" \
  --enable-shared --enable-ipv6 \
  --with-openssl="$DEPS_PREFIX" --with-libs="$WITH_LIBS" \
  --without-ensurepip \
  $EXTRA_CONFIGURE_ARGS; then
    # Show log when failed to run configure.
    cat config.log
    exit 1
fi


# These flags have been picked up by configure; unset them so they aren't
# appended again.
export LDFLAGS=
export LDFLAGS_NODIST=
export CPPFLAGS=
export CFLAGS_NODIST=

if [ ! $USE_SYSTEM_FFI ]; then
  # Tweak Makefile to change LIBFFI_INCLUDEDIR=<TAB>path
  sed -i \
    $'s+^LIBFFI_INCLUDEDIR=\t.*+LIBFFI_INCLUDEDIR=\t'"$DEPS_PREFIX/include+" \
    Makefile
fi

# Build production Python. BASEMODLIBS override allows -lpthread to be
# at the end of the linker command for old gcc's (like 4.9, still used on e.g.
# arm64 as of Nov 2019). This can likely go away when the dockcross base images
# update to gcc-6 or later.
make -j $(nproc) GITTAG="${GITTAG}" platform
make install BASEMODLIBS=$BASEMODLIBS GITTAG="${GITTAG}"

# Augment the Python installation.

# Read / augment / write the "ssl.py" module to implement custom SSL
# certificate loading logic.
#
# We do this here instead of "usercustomize.py" because the latter
# isn't propagated when a VirtualEnv is cut.
cat "$SCRIPT_DIR/ssl_suffix.py" >> $PREFIX/lib/python*/ssl.py

# Replace paths to the install location in _sysconfigdata with a
# placeholder string. This can then be reliably replaced with the
# real install location when we need to build wheels.
for f in "${PREFIX}/lib/python*/_sysconfigdata*.py"; do
  sed -e "s?${PREFIX}?[INSTALL_PREFIX]?g" -i $f
done

# TODO: maybe strip python executable?

INTERP=python3
if [[ $_3PP_PLATFORM == $_3PP_TOOL_PLATFORM ]]; then  # not cross compiling
  INTERP=./$PYTHONEXE
fi


export LD_LIBRARY_PATH=$(pwd):$LD_LIBRARY_PATH
$INTERP "$(which pip_bootstrap.py)" "$PREFIX"

PYTHON_MAJOR=$(cd $PREFIX/lib && echo python*)

# Cleanup!
find $PREFIX -name '*.a' -delete -print
# For some reason, docker freezes when doing `rm -rf .../test`. Specifically,
# `rm` hangs at 100% CPU forever on my mac. However, deleting it in smaller
# chunks works just fine. IDFK.
find $PREFIX/lib/$PYTHON_MAJOR/test -type f -exec rm -vf '{}' ';'
rm -vrf $PREFIX/lib/$PYTHON_MAJOR/test
rm -vrf $PREFIX/lib/$PYTHON_MAJOR/config
rm -vrf $PREFIX/lib/pkgconfig
rm -vrf $PREFIX/share

# Don't distribute __pycache__. Because the file modification times are not
# preserved in the CIPD package, Python will try to regenerate the compiled
# code, but will not overwrite an existing read-only file, effectively
# disabling the compiled code cache.
find "$PREFIX" -name __pycache__ -exec rm -rf {} +
