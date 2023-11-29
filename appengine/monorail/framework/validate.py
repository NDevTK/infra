# Copyright 2016 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""A set of Python input field validators."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import re

# RFC 5322-compliant email address regex
# https://stackoverflow.com/a/201378
_RFC_2821_EMAIL_REGEX = r"""
  (?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|
  "(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|
  \\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@
  (?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|
  \[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|
  [0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:
  (?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|
  \\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])
  """

# object used with <re>.search() or <re>.sub() to find email addresses
# within a string (or with <re>.match() to find email addresses at the
# beginning of a string that may be followed by trailing characters,
# since <re>.match() implicitly anchors at the beginning of the string)
RE_EMAIL_SEARCH = re.compile(_RFC_2821_EMAIL_REGEX, re.X)

# object used with <re>.match to find strings that contain *only* a single
# email address (by adding the end-of-string anchor $)
RE_EMAIL_ONLY = re.compile('^%s$' % _RFC_2821_EMAIL_REGEX, re.X)

_SCHEME_PATTERN = r'(?:https?|ftp)://'
_SHORT_HOST_PATTERN = (
    r'(?=[a-zA-Z])[-a-zA-Z0-9]*[a-zA-Z0-9](:[0-9]+)?'
    r'/'  # Slash is manditory for short host names.
    r'[^\s]*'
    )
_DOTTED_HOST_PATTERN = (
    r'[-a-zA-Z0-9.]+\.[a-zA-Z]{2,9}(:[0-9]+)?'
    r'(/[^\s]*)?'
    )
_URL_REGEX = r'%s(%s|%s)' % (
    _SCHEME_PATTERN, _SHORT_HOST_PATTERN, _DOTTED_HOST_PATTERN)

# A more complete URL regular expression based on a combination of the
# existing _URL_REGEX and the pattern found for URI regular expressions
# found in the URL RFC document. It's detailed here:
# http://www.ietf.org/rfc/rfc2396.txt
RE_COMPLEX_URL = re.compile(r'^%s(\?([^# ]*))?(#(.*))?$' % _URL_REGEX)


def IsValidEmail(s):
  """Return true iff the string is a properly formatted email address."""
  return RE_EMAIL_ONLY.match(s)


def IsValidMailTo(s):
  """Return true iff the string is a properly formatted mailto:."""
  return s.startswith('mailto:') and RE_EMAIL_ONLY.match(s[7:])


def IsValidURL(s):
  """Return true iff the string is a properly formatted web or ftp URL."""
  return RE_COMPLEX_URL.match(s)
