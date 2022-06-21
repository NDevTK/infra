#!/bin/bash -e

# When updating the version, you must update the SHA512 sum as well, e.g.:
# shasum -a 512 "${ARCHIVE}" > "${ARCHIVE}.sum"
VERSION=0.7.1-r51
SDK_VERSION=2022.06.19.122518
ARCHIVE="shellcheck-${VERSION}.tbz2"
URL="https://storage.googleapis.com/chromeos-prebuilt/host/amd64/amd64-host/chroot-${SDK_VERSION}/packages/dev-util/${ARCHIVE}"
SUMFILE="${ARCHIVE}.sum"

die() {
  echo "$1"
  exit 1
}

[ -f "${SUMFILE}" ] || \
  die "Missing integrity file ${SUMFILE}! (wrong directory?)"

echo "Downloading ${URL} ..."
curl "${URL}" -o "${ARCHIVE}"
echo

echo "Checking archive integrity..."
shasum -a 512 -c "${SUMFILE}" || die "Integrity check failed!"
echo

echo "Extracting shellcheck binary..."
# NOTE: Transforms tar paths into bin/shellcheck/.
tar -I 'zstd -f' -xf "${ARCHIVE}" --wildcards \
	--transform='s|.*/|bin/shellcheck/|' \
	./usr/bin/shellcheck \
	./usr/share/doc/*/LICENSE.*
chmod a+rX,a-w ./bin/shellcheck/*
