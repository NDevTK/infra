#!/bin/bash
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

# Make sure we target a base model CPU since we have a wide range of
# CPUs in the fleet.
case $_3PP_PLATFORM in
  mac-amd64)
    TARGET=CORE2
    ;;
  mac-arm64)
    TARGET=ARMV8
    ;;
  *)
    echo Please configure a CPU target for ${_3PP_PLATFORM}
    exit 1
    ;;
esac

MAKE_FLAGS="NO_LAPACKE=1 NO_SHARED=1 TARGET=${TARGET}"
make ${MAKE_FLAGS}
make ${MAKE_FLAGS} install PREFIX="${PREFIX}"
