# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import mimetypes
import posixpath
import re


mimetypes.init()


def content_type(path, query=''):
  path = path.lstrip('/')
  result = None

  if re.fullmatch(r'static/.*', path):
    result = mimetypes.guess_type(path)[0]
  elif re.fullmatch(r'\d+/image/.*', path):
    result = mimetypes.guess_type(query)[0]
  elif re.fullmatch(r'tarball/.*', path):
    result = 'application/x-gtar'
  elif re.fullmatch(r'download/.*\.diff', path):
    result = 'text/plain; charset=utf-8'
  elif re.fullmatch(r'\d+/diff2?_skipped_lines/.*', path):
    result = 'application/json; charset=utf-8'

  if result == 'image/vnd.microsoft.icon':
    result = 'image/x-icon'
  return result or 'text/html; charset=utf-8'


def make_path(path, query):
  path = path.lstrip('/')
  if re.fullmatch(r'\d+', path):
    path += '/'
  if not posixpath.basename(path):
    path += 'index.html'
  if query and path in ('index.html', 'all'):
    path += '?' + query
  return path
