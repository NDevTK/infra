# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Adds third_party packages to their respective package namespaces."""

import os
import sys


def FixImports():
  """Adds third_party packages to their respective package namespaces."""
  _AddThirdPartyToPath()
  _FixProtorpcPackage()
  _ImportAppEngineSearchApi()


def _AddThirdPartyToPath():
  """Adds third_party/ to sys.path.

  This lets us find antlr3, which is a dependency of search."""
  sys.path.append(_ThirdPartyDir())


def _FixProtorpcPackage():
  """Adds third_party/protorpc/ to protorpc.__path__.

  protorpc generally supports Python 3, except that protorpc.remote uses
  `async` as a field name, which is a reserved keyword starting in Python 3.5,
  with the introduction of coroutines. protorpc has been unmaintained and
  archived for years, and will not take pull requests. So, we have a local
  copy of that file with `async` renamed to `async_`, and modify the package
  __path__ to use our local copy.
  """
  import protorpc
  package_path = os.path.join(_ThirdPartyDir(), 'protorpc')
  protorpc.__path__.insert(0, package_path)


def _ImportAppEngineSearchApi():
  """Adds google.appengine.api.search to the importable packages.

  The appengine-python-standard package doesn't include search or
  google.net.proto.ProtocolBuffer. So, we include local copies in
  third_party/, and modify the package __path__ to use our local copies.
  """
  base_dir = _ThirdPartyDir()

  # Add third_party/google/ to the google namespace.
  # This makes Python look in this additional location for google.net.proto.
  import google
  package_path = os.path.join(base_dir, 'google')
  google.__path__.append(package_path)

  # Add third_party/google/ to the google.appengine.api namespace. This makes
  # Python look in this additional location for google.appengine.api.search.
  import google.appengine.api
  package_path = os.path.join(base_dir, 'google', 'appengine', 'api')
  google.appengine.api.__path__.append(package_path)

  # Add third_party/google/ to the google.appengine.datastore namespace.
  # This makes Python look in this additional location for document_pb.py.
  import google.appengine.datastore
  package_path = os.path.join(base_dir, 'google', 'appengine', 'datastore')
  google.appengine.datastore.__path__.append(package_path)


def _ThirdPartyDir():
  return os.path.join(os.path.dirname(__file__), 'third_party')
