# Copyright 2016 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Unit tests for the framework_helpers module."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import mock
import unittest

from google.appengine.api import urlfetch
from google.appengine.ext import testbed
from google.cloud import storage

from framework import gcs_helpers
from testing import testing_helpers


class GcsHelpersTest(unittest.TestCase):

  def setUp(self):
    self.testbed = testbed.Testbed()
    self.testbed.activate()
    self.testbed.init_memcache_stub()
    self.testbed.init_app_identity_stub()

    self.test_storage_client = mock.MagicMock()
    mock.patch.object(
        storage, 'Client', return_value=self.test_storage_client).start()

  def tearDown(self):
    self.testbed.deactivate()
    self.test_storage_client = None
    mock.patch.stopall()

  def testDeleteObjectFromGCS(self):
    object_id = 'aaaaa'
    gcs_helpers.DeleteObjectFromGCS(object_id)
    # Verify order of client calls.
    self.test_storage_client.assert_has_calls(
        [
            mock.call.bucket().get_blob(object_id),
            mock.call.bucket().get_blob().delete()
        ])

  def testDeleteLegacyObjectFromGCS(self):
    # A previous python module expected object ids with leading '/'
    object_id = '/aaaaa'
    object_id_without_leading_slash = 'aaaaa'
    gcs_helpers.DeleteObjectFromGCS(object_id)
    # Verify order of client calls.
    self.test_storage_client.assert_has_calls(
        [
            mock.call.bucket().get_blob(object_id_without_leading_slash),
            mock.call.bucket().get_blob().delete()
        ])

  @mock.patch(
      'google.appengine.api.images.resize', return_value=mock.MagicMock())
  @mock.patch('uuid.uuid4')
  def testStoreObjectInGCS_ResizableMimeType(self, mock_uuid4, mock_resize):
    guid = 'aaaaa'
    mock_uuid4.return_value = guid
    project_id = 100
    blob_name = '%s/attachments/%s' % (project_id, guid)
    thumb_blob_name = '%s/attachments/%s-thumbnail' % (project_id, guid)
    mime_type = 'image/png'
    content = 'content'

    ret_id = gcs_helpers.StoreObjectInGCS(
        content, mime_type, project_id, gcs_helpers.DEFAULT_THUMB_WIDTH,
        gcs_helpers.DEFAULT_THUMB_HEIGHT)
    self.assertEqual('/%s' % blob_name, ret_id)
    self.test_storage_client.assert_has_calls(
        [
            mock.call.bucket().blob(blob_name),
            mock.call.bucket().blob().upload_from_string(
                content, content_type=mime_type),
            mock.call.bucket().blob(thumb_blob_name),
        ])
    mock_resize.assert_called()

  @mock.patch(
      'google.appengine.api.images.resize', return_value=mock.MagicMock())
  @mock.patch('uuid.uuid4')
  def testStoreObjectInGCS_NotResizableMimeType(self, mock_uuid4, mock_resize):
    guid = 'aaaaa'
    mock_uuid4.return_value = guid
    project_id = 100
    blob_name = '%s/attachments/%s' % (project_id, guid)
    mime_type = 'not_resizable_mime_type'
    content = 'content'

    ret_id = gcs_helpers.StoreObjectInGCS(
        content, mime_type, project_id, gcs_helpers.DEFAULT_THUMB_WIDTH,
        gcs_helpers.DEFAULT_THUMB_HEIGHT)
    self.assertEqual('/%s' % blob_name, ret_id)
    self.test_storage_client.assert_has_calls(
        [
            mock.call.bucket().blob(blob_name),
            mock.call.bucket().blob().upload_from_string(
                content, content_type=mime_type),
        ])
    mock_resize.assert_not_called()

  def testCheckMimeTypeResizable(self):
    for resizable_mime_type in gcs_helpers.RESIZABLE_MIME_TYPES:
      gcs_helpers.CheckMimeTypeResizable(resizable_mime_type)

    with self.assertRaises(gcs_helpers.UnsupportedMimeType):
      gcs_helpers.CheckMimeTypeResizable('not_resizable_mime_type')

  @mock.patch('framework.filecontent.GuessContentTypeFromFilename')
  @mock.patch('framework.gcs_helpers.StoreObjectInGCS')
  def testStoreLogoInGCS(self, mock_store_object, mock_guess_content):
    blob_name = 123
    mock_store_object.return_value = blob_name
    mime_type = 'image/png'
    mock_guess_content.return_value = mime_type
    file_name = 'test_file.png'
    content = 'test content'
    project_id = 100

    ret_id = gcs_helpers.StoreLogoInGCS(file_name, content, project_id)
    self.assertEqual(blob_name, ret_id)

  @mock.patch('google.appengine.api.urlfetch.fetch')
  def testFetchSignedURL_Success(self, mock_fetch):
    mock_fetch.return_value = testing_helpers.Blank(
        headers={'Location': 'signed url'})
    actual = gcs_helpers._FetchSignedURL('signing req url')
    mock_fetch.assert_called_with('signing req url', follow_redirects=False)
    self.assertEqual('signed url', actual)

  @mock.patch('google.appengine.api.urlfetch.fetch')
  def testFetchSignedURL_UnderpopulatedResult(self, mock_fetch):
    mock_fetch.return_value = testing_helpers.Blank(headers={})
    self.assertRaises(
        KeyError, gcs_helpers._FetchSignedURL, 'signing req url')

  @mock.patch('google.appengine.api.urlfetch.fetch')
  def testFetchSignedURL_DownloadError(self, mock_fetch):
    mock_fetch.side_effect = urlfetch.DownloadError
    self.assertRaises(
        urlfetch.DownloadError,
        gcs_helpers._FetchSignedURL, 'signing req url')

  @mock.patch('framework.gcs_helpers._FetchSignedURL')
  def testSignUrl_Success(self, mock_FetchSignedURL):
    with mock.patch(
        'google.appengine.api.app_identity.get_access_token') as gat:
      gat.return_value = ['token']
      mock_FetchSignedURL.return_value = 'signed url'
      signed_url = gcs_helpers.SignUrl('bucket', '/object')
      self.assertEqual('signed url', signed_url)

  @mock.patch('framework.gcs_helpers._FetchSignedURL')
  def testSignUrl_DownloadError(self, mock_FetchSignedURL):
    mock_FetchSignedURL.side_effect = urlfetch.DownloadError
    self.assertEqual(
        '/missing-gcs-url', gcs_helpers.SignUrl('bucket', '/object'))
