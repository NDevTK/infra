# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# THIS MUST BE KEPT IN SYNC WITH infra/3pp/android_toolchain_canary/install.sh.

set -e
set -x
set -o pipefail

# An auto-created directory whose content will ultimately be uploaded to CIPD.
# The commands below should output the built product to this directory.
PREFIX="$1"

# Glob patterns to exclude from the final output.
GLOB_EXCLUDES=(
  # These files are excluded because they are duplicated. This is cruft in the NDK,
  # but the duplicated files (e.g. ipt_ECN.h and ipt_ecn.h) will cause problems
  # in file systems which are case-insensitive. See https://g-issues.chromium.org/issues/40273594.
  toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/include/linux/netfilter_ipv4/ipt_ECN.h
  toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/include/linux/netfilter_ipv4/ipt_TTL.h
  toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/include/linux/netfilter_ipv6/ip6t_HL.h
  toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/include/linux/netfilter/xt_CONNMARK.h
  toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/include/linux/netfilter/xt_DSCP.h
  toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/include/linux/netfilter/xt_MARK.h
  toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/include/linux/netfilter/xt_RATEEST.h
  toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/include/linux/netfilter/xt_TCPMSS.h
)

# Move all files into the output directory.
mv * "$PREFIX"

# Remove excluded files from the staging directory.
for pattern in "${GLOB_EXCLUDES[@]}"; do
  rm -rf "${PREFIX}/${pattern}"
done