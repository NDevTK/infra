# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import mock
import unittest

from google.appengine.api import oauth
from google.appengine.api import users

from gae_libs.http import auth_util


class AuthUtilTest(unittest.TestCase):

  @mock.patch.object(
      auth_util.app_identity, 'get_access_token', return_value=('abc', 123))
  def testGetAuthToken(self, mocked_func):
    self.assertEqual('abc', auth_util.GetAuthToken('scope'))
    mocked_func.assert_called_once_with('scope')

  @mock.patch.object(
      auth_util.app_identity, 'get_access_token', return_value=('abc', 123))
  def testAuthenticator(self, _):
    authenticator = auth_util.Authenticator()
    self.assertEqual({}, authenticator.GetHttpHeadersFor('http://test.com'))
    self.assertEqual({}, authenticator.GetHttpHeadersFor('https://unknown.com'))
    self.assertEqual(
        {},
        authenticator.GetHttpHeadersFor('https://test.googlesource.com.fake'))
    self.assertEqual(
        {},
        authenticator.GetHttpHeadersFor('https://codereview.chromium.org.fake'))
    self.assertEqual({
        'Authorization': 'Bearer abc'
    }, authenticator.GetHttpHeadersFor('https://test.googlesource.com/cr'))
    self.assertEqual({
        'Authorization': 'Bearer abc'
    }, authenticator.GetHttpHeadersFor('https://codereview.chromium.org/api'))

  @mock.patch.object(
      auth_util.oauth,
      'get_current_user',
      return_value=users.User('email2', 'domain', 'id2'))
  @mock.patch.object(
      auth_util.users,
      'get_current_user',
      return_value=users.User('email1', 'domain', 'id1'))
  def testGetUserEmailFromCookie(self, mocked_users, mocked_oauth):
    self.assertEqual('email1', auth_util.GetUserEmail('scope'))
    mocked_users.assert_called_once_with()
    mocked_oauth.assert_not_called()

  @mock.patch.object(
      auth_util.oauth,
      'get_current_user',
      return_value=users.User('email2', 'domain', 'id2'))
  @mock.patch.object(auth_util.users, 'get_current_user', return_value=None)
  def testGetUserEmailFromOauth(self, mocked_users, mocked_oauth):
    self.assertEqual('email2', auth_util.GetUserEmail('scope'))
    mocked_users.assert_called_once_with()
    mocked_oauth.assert_called_once_with('scope')

  @mock.patch.object(
      auth_util.oauth,
      'get_current_user',
      side_effect=oauth.OAuthRequestError())
  @mock.patch.object(auth_util.users, 'get_current_user', return_value=None)
  def testGetUserEmailFromOauthException(self, mocked_users, mocked_oauth):
    self.assertIsNone(auth_util.GetUserEmail('scope'))
    mocked_users.assert_called_once_with()
    mocked_oauth.assert_called_once_with('scope')

  @mock.patch.object(
      auth_util.oauth, 'is_current_user_admin', return_value=False)
  @mock.patch.object(
      auth_util.users, 'is_current_user_admin', return_value=True)
  def testAdminFromCookie(self, mocked_users, mocked_oauth):
    self.assertTrue(auth_util.IsCurrentUserAdmin('scope'))
    mocked_users.assert_called_once_with()
    mocked_oauth.assert_not_called()

  @mock.patch.object(
      auth_util.oauth, 'is_current_user_admin', return_value=True)
  @mock.patch.object(
      auth_util.users, 'is_current_user_admin', return_value=False)
  def testAdminFromOauth(self, mocked_users, mocked_oauth):
    self.assertTrue(auth_util.IsCurrentUserAdmin('scope'))
    mocked_users.assert_called_once_with()
    mocked_oauth.assert_called_once_with('scope')

  @mock.patch.object(
      auth_util.oauth,
      'is_current_user_admin',
      side_effect=oauth.OAuthRequestError())
  @mock.patch.object(
      auth_util.users, 'is_current_user_admin', return_value=False)
  def testAdminFromOauthException(self, mocked_users, mocked_oauth):
    self.assertFalse(auth_util.IsCurrentUserAdmin('scope'))
    mocked_users.assert_called_once_with()
    mocked_oauth.assert_called_once_with('scope')

  @mock.patch.object(
      auth_util.users, 'create_login_url', return_value='http://login')
  @mock.patch.object(
      auth_util.users, 'create_logout_url', return_value='http://logout')
  @mock.patch.object(auth_util, 'GetUserEmail', return_value='email')
  @mock.patch.object(auth_util, 'IsCurrentUserAdmin', return_value=False)
  def testUserInfoAfterLogin(self, *_):
    expected_info = {
        'email': 'email',
        'is_admin': False,
        'logout_url': 'http://logout',
    }
    self.assertEqual(expected_info, auth_util.GetUserInfo())

  @mock.patch.object(
      auth_util.users, 'create_login_url', return_value='http://login')
  @mock.patch.object(
      auth_util.users, 'create_logout_url', return_value='http://logout')
  @mock.patch.object(auth_util, 'GetUserEmail', return_value=None)
  @mock.patch.object(auth_util, 'IsCurrentUserAdmin', return_value=False)
  def testUserInfoBeforeLogin(self, *_):
    expected_info = {
        'email': None,
        'is_admin': False,
        'login_url': 'http://login',
    }
    self.assertEqual(expected_info, auth_util.GetUserInfo())
