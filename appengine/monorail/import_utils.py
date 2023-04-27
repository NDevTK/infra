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
  _FixProtorpcPackage()
  _FixDefaultApiStub()
  _ImportProtocolBuffer()


def _AddThirdPartyToPath():
  """Adds third_party/ to sys.path.

  This lets us find endpoints."""
  sys.path.append(_ThirdPartyDir())


def _FixProtorpcPackage():
  """Adds third_party/protorpc/ to protorpc.__path__.

  protorpc generally supports Python 3, except for a few minor issues. protorpc
  has been unmaintained and archived for years, and will not take pull requests.
  So, we have a local copy of a few of the files with Python 3 modifications,
  and update the package __path__ to use our local copy.
  """
  import protorpc
  package_path = os.path.join(_ThirdPartyDir(), 'protorpc')
  protorpc.__path__.insert(0, package_path)


def _FixDefaultApiStub():
  """Fixes "Attempted RPC call without active security ticket" error.

  In appengine-python-standard==1.0.0, default_api_stub throws an error when
  trying to access NDB outside of a Flask request. This was fixed in commit
  cc19a2e on Juy 21, 2022, but wasn't included in the 1.0.1rc1 release on
  Sep 6, 2022. It's been months since that release, so instead of waiting on
  another release, we'll just monkeypatch the file here.
  """
  if not six.PY3:
    return
  sys.path.append(os.path.join(_ThirdPartyDir(), 'appengine-python-standard'))
  import default_api_stub as fixed_default_api_stub
  from google.appengine.runtime import default_api_stub
  default_api_stub.DefaultApiRPC = fixed_default_api_stub.DefaultApiRPC


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
