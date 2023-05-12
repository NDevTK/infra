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
  _FixMox3()


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


def _FixMox3():
  """Fixes a Python 3 warning with the mox3 library.

  mox3 uses `inspect.getargspec()`, which is deprecated since Python 3.0.
  This throws a warning when running unit tests. Update the method to use
  `inspect.getfullargspec()` instead.
  """
  from mox3 import mox
  mox.MethodSignatureChecker.__init__ = _MethodSignatureChecker


def _ThirdPartyDir():
  return os.path.join(os.path.dirname(__file__), 'third_party')


def _MethodSignatureChecker(self, method, class_to_bind=None):
  """Creates a checker.

  Args:
      # method: A method to check.
      # class_to_bind: optionally, a class used to type check first
      #                method parameter, only used with unbound methods
      method: function
      class_to_bind: type or None

  Raises:
      ValueError: method could not be inspected, so checks aren't
                  possible. Some methods and functions like built-ins
                  can't be inspected.
  """
  import inspect
  try:
    self._args, varargs, varkw, defaults, _, _, _ = inspect.getfullargspec(
        method)
  except TypeError:
    raise ValueError('Could not get argument specification for %r' % (method,))
  if (inspect.ismethod(method) or class_to_bind or
      (hasattr(self, '_args') and len(self._args) > 0 and
       self._args[0] == 'self')):
    self._args = self._args[1:]  # Skip 'self'.
  self._method = method
  self._instance = None  # May contain the instance this is bound to.
  self._instance = getattr(method, "__self__", None)

  # _bounded_to determines whether the method is bound or not
  if self._instance:
    self._bounded_to = self._instance.__class__
  else:
    self._bounded_to = class_to_bind or getattr(method, "im_class", None)

  self._has_varargs = varargs is not None
  self._has_varkw = varkw is not None
  if defaults is None:
    self._required_args = self._args
    self._default_args = []
  else:
    self._required_args = self._args[:-len(defaults)]
    self._default_args = self._args[-len(defaults):]
