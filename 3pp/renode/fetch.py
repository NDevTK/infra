#!/usr/bin/env python3
# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import json
import os
import ssl
import re
import urllib.request
from html.parser import HTMLParser
from packaging import version as pkg_version

import certifi

# Make sure up-to-date root certificates are used.
urllib.request.install_opener(
    urllib.request.build_opener(
        urllib.request.HTTPSHandler(
            context=ssl.create_default_context(cafile=certifi.where())
        )
    )
)

PLATFORM_RE = re.compile("linux-.*")

RELEASES = "https://builds.renode.io/renode-%s.linux-portable.tar.gz"


def get_download_url(version, platform):
    if not PLATFORM_RE.match(platform):
        raise ValueError("unsupported platform {}" % platform)

    manifest = {
        "url": [(RELEASES % version)],
        "ext": ".tar.gz",
    }
    print(json.dumps(manifest))


class ReleaseHTMLParser(HTMLParser):
    def __init__(self):
        super().__init__()
        self.current_version = None
        self.full_version_string = None
        self.filename_re = re.compile(
            r"renode-(\d+\.\d+\.\d+)" r"(\+\d{8}git[0-9a-f]{6,})?.linux-portable.tar.gz"
        )

    def error(self, message):
        pass

    def handle_starttag(self, tag, attrs):
        # Skip non links
        if tag != "a":
            return
        url = None

        # Search for the href attribute and save the url
        for attr, value in attrs:
            if attr != "href":
                continue
            if self.filename_re.match(value) is None:
                continue
            url = value

        (version_obj, full_string) = self._create_version_tuple(url)

        # Bail if no version string, url must be invalid
        if not version_obj:
            return

        # If this is the first version, or if it is newer, save it
        if self.current_version is None or version_obj > self.current_version:
            self.current_version = version_obj
            self.full_version_string = full_string

    def _create_version_tuple(self, url):
        # Bail if url is None or empty
        if not url:
            return (None, None)

        # Try to match the URL and bail if no match
        match = self.filename_re.match(url)
        if not match:
            return (None, None)

        # Get the version and date strings
        version_string = match.group(1)
        full_string = match.group(1)
        if match.group(2):
            # Group 2 starts with '\+\d{8}' and we want to strip out the '+'
            # We strip the tai "git[commit]"
            version_string += match.group(2)[1:9]
            full_string += match.group(2)

        # Return a tuple of (version, date)
        return (pkg_version.parse(version_string), full_string)


def do_latest(platform):
    if not PLATFORM_RE.match(platform):
        raise ValueError("unsupported platform {}" % platform)

    page = urllib.request.urlopen("https://builds.renode.io/")
    parser = ReleaseHTMLParser()
    parser.feed(str(page.read()))
    print(parser.full_version_string)


def main():
    ap = argparse.ArgumentParser()
    sub = ap.add_subparsers(dest="action", required=True)

    latest = sub.add_parser("latest")
    latest.set_defaults(func=lambda _opts: do_latest(os.environ["_3PP_PLATFORM"]))

    download = sub.add_parser("get_url")
    download.set_defaults(
        func=lambda _opts: get_download_url(
            os.environ["_3PP_VERSION"], os.environ["_3PP_PLATFORM"]
        )
    )

    opts = ap.parse_args()
    opts.func(opts)


if __name__ == "__main__":
    main()
