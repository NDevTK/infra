# Copyright 2016 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Set of helpers for interacting with Google Cloud Storage."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import logging
import os
import six
from six.moves import urllib
import uuid

from datetime import datetime, timedelta

from google.appengine.api import app_identity
from google.appengine.api import images
from google.appengine.api import memcache
from google.appengine.api import urlfetch
from google.cloud import storage

from framework import filecontent
from framework import framework_constants
from framework import framework_helpers


ATTACHMENT_TTL = timedelta(seconds=30)

IS_DEV_APPSERVER = (
    'development' in os.environ.get('SERVER_SOFTWARE', '').lower())

RESIZABLE_MIME_TYPES = [
    'image/png', 'image/jpg', 'image/jpeg', 'image/gif', 'image/webp',
    ]

DEFAULT_THUMB_WIDTH = 250
DEFAULT_THUMB_HEIGHT = 200
LOGO_THUMB_WIDTH = 110
LOGO_THUMB_HEIGHT = 30
MAX_ATTACH_SIZE_TO_COPY = 10 * 1024 * 1024  # 10 MB
# GCS signatures are valid for 10 minutes by default, but cache them for
# 5 minutes just to be on the safe side.
GCS_SIG_TTL = 60 * 5


def _Now():
  return datetime.utcnow()


class UnsupportedMimeType(Exception):
  pass


def _RemoveLeadingSlash(text):
  if text.startswith('/'):
    return text[1:]
  return text


def DeleteObjectFromGCS(blob_name):
  storage_client = storage.Client()
  bucket_name = app_identity.get_default_gcs_bucket_name()
  bucket = storage_client.bucket(bucket_name)
  validated_blob_name = _RemoveLeadingSlash(blob_name)
  blob = bucket.get_blob(validated_blob_name)
  blob.delete()


def StoreObjectInGCS(
    content, mime_type, project_id, thumb_width=DEFAULT_THUMB_WIDTH,
    thumb_height=DEFAULT_THUMB_HEIGHT, filename=None):
  storage_client = storage.Client()
  bucket_name = app_identity.get_default_gcs_bucket_name()
  bucket = storage_client.bucket(bucket_name)
  guid = uuid.uuid4()
  blob_name = '%s/attachments/%s' % (project_id, guid)

  blob = bucket.blob(blob_name)
  if filename:
    if not framework_constants.FILENAME_RE.match(filename):
      logging.info('bad file name: %s' % filename)
      filename = 'attachment.dat'
    content_disposition = 'inline; filename="%s"' % filename
    blob.content_disposition = content_disposition
    logging.info('Writing with content_disposition %r', content_disposition)
  blob.upload_from_string(content, content_type=mime_type)

  if mime_type in RESIZABLE_MIME_TYPES:
    # Create and save a thumbnail too.
    thumb_content = None
    try:
      thumb_content = images.resize(content, thumb_width, thumb_height)
    except images.LargeImageError:
      # Don't log the whole exception because we don't need to see
      # this on the Cloud Error Reporting page.
      logging.info('Got LargeImageError on image with %d bytes', len(content))
    except Exception as e:
      # Do not raise exception for incorrectly formed images.
      # See https://bugs.chromium.org/p/monorail/issues/detail?id=597 for more
      # detail.
      logging.exception(e)
    if thumb_content:
      thumb_blob_name = '%s-thumbnail' % blob_name
      thumb_blob = bucket.blob(thumb_blob_name)
      thumb_blob.upload_from_string(thumb_content, content_type='image/png')

  # Our database, sadly, stores these with the leading slash.
  return '/%s' % blob_name


def CheckMimeTypeResizable(mime_type):
  if mime_type not in RESIZABLE_MIME_TYPES:
    raise UnsupportedMimeType(
        'Please upload a logo with one of the following mime types:\n%s' %
            ', '.join(RESIZABLE_MIME_TYPES))


def StoreLogoInGCS(file_name, content, project_id):
  mime_type = filecontent.GuessContentTypeFromFilename(file_name)
  CheckMimeTypeResizable(mime_type)
  if '\\' in file_name:  # IE insists on giving us the whole path.
    file_name = file_name[file_name.rindex('\\') + 1:]
  return StoreObjectInGCS(
      content, mime_type, project_id, thumb_width=LOGO_THUMB_WIDTH,
      thumb_height=LOGO_THUMB_HEIGHT)


@framework_helpers.retry(3, delay=0.25, backoff=1.25)
def _FetchSignedURL(url):
  """Request that devstorage API signs a GCS content URL."""
  resp = urlfetch.fetch(url, follow_redirects=False)
  redir = resp.headers["Location"]
  return six.ensure_str(redir)


def SignUrl(bucket, object_id):
  """Get a signed URL to download a GCS object.

  Args:
    bucket: string name of the GCS bucket.
    object_id: string object ID of the file within that bucket.

  Returns:
    A signed URL, or '/mising-gcs-url' if signing failed.
  """
  try:
    cache_key = 'gcs-object-url-%s' % object_id
    cached = memcache.get(key=cache_key)
    if cached is not None:
      return cached

    if IS_DEV_APPSERVER:
      attachment_url = '/_ah/gcs/%s%s' % (bucket, object_id)
    else:
      result = ('https://www.googleapis.com/storage/v1/b/'
          '{bucket}/o/{object_id}?access_token={token}&alt=media')
      scopes = ['https://www.googleapis.com/auth/devstorage.read_only']
      if object_id[0] == '/':
        object_id = object_id[1:]
      url = result.format(
          bucket=bucket,
          object_id=urllib.parse.quote_plus(object_id),
          token=app_identity.get_access_token(scopes)[0])
      attachment_url = _FetchSignedURL(url)

    if not memcache.set(key=cache_key, value=attachment_url, time=GCS_SIG_TTL):
      logging.error('Could not cache gcs url %s for %s', attachment_url,
          object_id)

    return attachment_url

  except Exception as e:
    logging.exception(e)
    return '/missing-gcs-url'


def MaybeCreateDownload(bucket_name, blob_name, filename):
  """If the obj is not huge, and no download version exists, create it."""
  validated_blob_name = _RemoveLeadingSlash(blob_name)
  dst_blob_name = '%s-download' % validated_blob_name
  logging.info('Maybe create %r from %r', dst_blob_name, validated_blob_name)

  if IS_DEV_APPSERVER:
    logging.info('dev environment never makes download copies.')
    return False

  storage_client = storage.Client()
  bucket = storage_client.bucket(bucket_name)

  # Validate "View" object.
  src_blob = bucket.get_blob(validated_blob_name)
  if not src_blob:
    return False
  # If "Download" object already exists, it's already created.
  # `Bucket.blob` doesn't make an HTTP request.
  dst_blob = bucket.blob(dst_blob_name)
  if dst_blob.exists():
    logging.info('Download version of attachment already exists')
    return True
  # If "View" object is huge, don't create a download.
  if src_blob.size > MAX_ATTACH_SIZE_TO_COPY:
    logging.info('Download version of attachment would be too big')
    return False

  copied_dst_blob = bucket.copy_blob(src_blob, bucket, dst_blob_name)
  content_disposition = 'attachment; filename="%s"' % filename
  logging.info('Copying with content_disposition %r', content_disposition)
  copied_dst_blob.content_disposition = content_disposition
  copied_dst_blob.patch()
  logging.info('done writing')

  return True
