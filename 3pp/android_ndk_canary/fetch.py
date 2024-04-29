# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

#!/usr/bin/env python3
import argparse
import json

# Update this when upgrading NDK.
_VERSION = "11746098"
_URL = f"https://ci.android.com/builds/submitted/{_VERSION}/linux/latest/raw/android-ndk-{_VERSION}-linux-x86_64.zip"


def do_latest():
  return _VERSION


def do_get_url():
  partial_manifest = {
      'url': [_URL],
      'ext': '.zip',
  }
  print(json.dumps(partial_manifest))


def main():
  ap = argparse.ArgumentParser()
  sub = ap.add_subparsers(required=True)

  latest = sub.add_parser("latest")
  latest.set_defaults(func=do_latest)

  download = sub.add_parser("get_url")
  download.set_defaults(func=do_get_url)

  opts = ap.parse_args()
  opts.func()


if __name__ == '__main__':
  main()
