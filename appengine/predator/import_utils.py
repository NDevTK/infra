# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Adds third_party packages to their respective package namespaces."""

import os
import six
import sys


def FixImports():
  """Adds third_party packages to their respective package namespaces."""
  _AddThirdPartyToPath()
  _ImportProtocolBuffer()


def _AddThirdPartyToPath():
  """Adds third_party/ to sys.path.

  This lets us find endpoints."""
  sys.path.append(_ThirdPartyDir())


def _ImportProtocolBuffer():
  """Adds google.net.proto.ProtocolBuffer to the importable packages.

  The appengine-python-standard package doesn't include
  google.net.proto.ProtocolBuffer. So, we include a local copy in
  third_party/, and modify the package __path__ to use our local copy.
  """
  # Add third_party/google/ to the google namespace.
  # This makes Python look in this additional location for google.net.proto.
  import google
  package_path = os.path.join(_ThirdPartyDir(), 'google')
  google.__path__.append(package_path)


def _ThirdPartyDir():
  return os.path.join(os.path.dirname(__file__), 'third_party')

