# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json
import os
import socket
import time
import unittest

from six.moves import http_client
from six.moves import urllib

import infra_libs
from infra_libs.ts_mon.common import http_metrics
from infra_libs import httplib2_utils
from infra_libs import ts_mon

import httplib2
import mock


class RetriableHttplib2Test(unittest.TestCase):
  def setUp(self):
    super(RetriableHttplib2Test, self).setUp()
    self.http = infra_libs.RetriableHttp(httplib2.Http())
    self.http._http.request = mock.create_autospec(self.http._http.request,
                                                   spec_set=True)

  _MOCK_REQUEST = mock.call('http://foo/', 'GET', None)

  def test_delegate_get_attr(self):
    """RetriableHttp should delegate getting attribute except request() to
       Http"""
    self.http._http.clear_credentials = mock.create_autospec(
        self.http._http.clear_credentials, spec_set=True)
    self.http.clear_credentials()
    self.http._http.clear_credentials.assert_called_once_with()

  def test_delegate_set_attr(self):
    """RetriableHttp should delegate setting attributes to Http"""
    self.http.ignore_etag = False
    self.assertFalse(self.http.ignore_etag)
    self.assertFalse(self.http._http.ignore_etag)
    self.http.ignore_etag = True
    self.assertTrue(self.http.ignore_etag)
    self.assertTrue(self.http._http.ignore_etag)

  @mock.patch('time.sleep', autospec=True)
  def test_succeed(self, _sleep):
    self.http._http.request.return_value = (
        httplib2.Response({'status': 400}), 'content')
    response, _ = self.http.request('http://foo/')
    self.assertEqual(400, response.status)
    self.http._http.request.assert_has_calls([ self._MOCK_REQUEST ])

  @mock.patch('time.sleep', autospec=True)
  def test_retry_succeed(self, _sleep):
    self.http._http.request.side_effect = iter([
      (httplib2.Response({'status': 500}), 'content'),
      httplib2.HttpLib2Error,
      (httplib2.Response({'status': 200}), 'content')
    ])
    response, _ = self.http.request('http://foo/')
    self.assertEqual(200, response.status)
    self.http._http.request.assert_has_calls([ self._MOCK_REQUEST ] * 3)

  @mock.patch('time.sleep', autospec=True)
  def test_fail_exception(self, _sleep):
    self.http._http.request.side_effect = httplib2.HttpLib2Error()
    self.assertRaises(httplib2.HttpLib2Error, self.http.request, 'http://foo/')
    self.http._http.request.assert_has_calls([ self._MOCK_REQUEST ] * 5)

  @mock.patch('time.sleep', autospec=True)
  def test_fail_status_code(self, _sleep):
    self.http._http.request.return_value = (
        httplib2.Response({'status': 500}), 'content')
    response, _ = self.http.request('http://foo/')
    self.assertEqual(500, response.status)
    self.http._http.request.assert_has_calls([ self._MOCK_REQUEST ] * 5)


class InstrumentedHttplib2Test(unittest.TestCase):
  def setUp(self):
    super(InstrumentedHttplib2Test, self).setUp()
    self.mock_time = mock.create_autospec(time.time, spec_set=True)
    self.mock_time.return_value = 42
    self.http = infra_libs.InstrumentedHttp('test', time_fn=self.mock_time)
    self.http._request = mock.Mock()
    ts_mon.reset_for_unittest()

  def test_success_status(self):
    self.http._request.return_value = (
        httplib2.Response({'status': 200}),
        'content')

    response, _ = self.http.request('http://foo/')
    self.assertEqual(200, response.status)
    self.assertEqual(1, http_metrics.response_status.get(
        {'name': 'test', 'client': 'httplib2', 'status': 200}))
    self.assertIsNone(http_metrics.response_status.get(
        {'name': 'test', 'client': 'httplib2', 'status': 404}))

  def test_error_status(self):
    self.http._request.return_value = (
        httplib2.Response({'status': 404}),
        'content')

    response, _ = self.http.request('http://foo/')
    self.assertEqual(404, response.status)
    self.assertIsNone(http_metrics.response_status.get(
        {'name': 'test', 'client': 'httplib2', 'status': 200}))
    self.assertEqual(1, http_metrics.response_status.get(
        {'name': 'test', 'client': 'httplib2', 'status': 404}))

  def test_timeout(self):
    self.http._request.side_effect = socket.timeout

    with self.assertRaises(socket.timeout):
      self.http.request('http://foo/')
    self.assertIsNone(http_metrics.response_status.get(
        {'name': 'test', 'client': 'httplib2', 'status': 200}))
    self.assertEqual(1, http_metrics.response_status.get(
        {'name': 'test', 'client': 'httplib2',
         'status': http_metrics.STATUS_TIMEOUT}))

  def test_connection_error(self):
    self.http._request.side_effect = socket.error

    with self.assertRaises(socket.error):
      self.http.request('http://foo/')
    self.assertIsNone(http_metrics.response_status.get(
        {'name': 'test', 'client': 'httplib2', 'status': 200}))
    self.assertEqual(1, http_metrics.response_status.get(
        {'name': 'test', 'client': 'httplib2',
         'status': http_metrics.STATUS_ERROR}))

  def test_exception(self):
    self.http._request.side_effect = httplib2.HttpLib2Error

    with self.assertRaises(httplib2.HttpLib2Error):
      self.http.request('http://foo/')
    self.assertIsNone(http_metrics.response_status.get(
        {'name': 'test', 'client': 'httplib2', 'status': 200}))
    self.assertEqual(1, http_metrics.response_status.get(
        {'name': 'test', 'client': 'httplib2',
         'status': http_metrics.STATUS_EXCEPTION}))

  def test_httplib_exception(self):
    self.http._request.side_effect = http_client.HTTPException

    with self.assertRaises(http_client.HTTPException):
      self.http.request('http://foo/')
    self.assertIsNone(http_metrics.response_status.get(
        {'name': 'test', 'client': 'httplib2', 'status': 200}))
    self.assertEqual(1, http_metrics.response_status.get(
        {'name': 'test', 'client': 'httplib2',
         'status': http_metrics.STATUS_EXCEPTION}))

  def test_gae_httplib_timeout_exception(self):
    self.http._request.side_effect = http_client.HTTPException(
        'Deadline exceeded while waiting for HTTP response from URL: '
        'http://foo/')

    with self.assertRaises(http_client.HTTPException):
      self.http.request('http://foo/')
    self.assertIsNone(http_metrics.response_status.get(
        {'name': 'test', 'client': 'httplib2', 'status': 200}))
    self.assertEqual(1, http_metrics.response_status.get(
        {'name': 'test', 'client': 'httplib2',
         'status': http_metrics.STATUS_TIMEOUT}))

  def test_response_bytes(self):
    self.http._request.return_value = (
        httplib2.Response({'status': 200}),
        'content')

    _, content = self.http.request('http://foo/')
    self.assertEqual('content', content)
    self.assertEqual(1, http_metrics.response_bytes.get(
        {'name': 'test', 'client': 'httplib2'}).count)
    self.assertEqual(7, http_metrics.response_bytes.get(
        {'name': 'test', 'client': 'httplib2'}).sum)

  def test_request_bytes(self):
    self.http._request.return_value = (
        httplib2.Response({'status': 200}),
        'content')

    _, content = self.http.request('http://foo/', body='wibblewibble')
    self.assertEqual('content', content)
    self.assertEqual(1, http_metrics.request_bytes.get(
        {'name': 'test', 'client': 'httplib2'}).count)
    self.assertEqual(12, http_metrics.request_bytes.get(
        {'name': 'test', 'client': 'httplib2'}).sum)

  def test_duration(self):
    current_time = [4.2]

    def time_side_effect():
      ret = current_time[0]
      current_time[0] += 0.3
      return ret
    self.mock_time.side_effect = time_side_effect

    self.http._request.return_value = (
        httplib2.Response({'status': 200}),
        'content')

    _, _ = self.http.request('http://foo/')
    self.assertEqual(1, http_metrics.durations.get(
        {'name': 'test', 'client': 'httplib2'}).count)
    self.assertAlmostEqual(300, http_metrics.durations.get(
        {'name': 'test', 'client': 'httplib2'}).sum)


class HttpMockTest(unittest.TestCase):
  def test_empty(self):
    http = infra_libs.HttpMock([])
    with self.assertRaises(AssertionError):
      http.request('https://www.google.com', 'GET')

  def test_invalid_parameter(self):
    with self.assertRaises(TypeError):
      infra_libs.HttpMock(None)

  def test_uris_wrong_length(self):
    with self.assertRaises(ValueError):
      infra_libs.HttpMock([(1, 2)])

  def test_uris_wrong_type(self):
    with self.assertRaises(ValueError):
      infra_libs.HttpMock([(None,)])

  def test_invalid_uri(self):
    with self.assertRaises(TypeError):
      infra_libs.HttpMock([(1, {'status': '100'}, None)])

  def test_invalid_headers(self):
    with self.assertRaises(TypeError):
      infra_libs.HttpMock([('https://www.google.com', None, None)])

  def test_headers_without_status(self):
    with self.assertRaises(ValueError):
      infra_libs.HttpMock([('https://www.google.com', {'foo': 'bar'}, None)])

  def test_invalid_body(self):
    with self.assertRaises(TypeError):
      infra_libs.HttpMock([('https://www.google.com', {'status': '200'}, 42)])

  def test_one_uri(self):
    http = infra_libs.HttpMock([('https://www.google.com',
                                 {'status': '403'},
                                 'bar')])
    response, body = http.request('https://www.google.com', 'GET')
    self.assertIsInstance(response, httplib2.Response)
    self.assertEqual(response.status, 403)
    self.assertEqual(body, 'bar')

  def test_two_uris(self):
    http = infra_libs.HttpMock([('https://www.google.com',
                                 {'status': 200}, 'foo'),
                                ('.*', {'status': 404}, '')])
    response, body = http.request('https://mywebserver.woo.hoo', 'GET')
    self.assertIsInstance(response, httplib2.Response)
    self.assertEqual(response.status, 404)
    self.assertEqual(body, '')

    self.assertEqual(http.requests_made[0].uri, 'https://mywebserver.woo.hoo')
    self.assertEqual(http.requests_made[0].method, 'GET')
    self.assertEqual(http.requests_made[0].body, None)
    self.assertEqual(http.requests_made[0].headers, None)


class DelegateServiceAccountCredentialsTest(unittest.TestCase):

  def setUp(self):
    self.email = 'test@example.com'
    self.project = '-'
    self.scopes = ['scope']
    self.mock_http = mock.Mock()
    self.creds = httplib2_utils.DelegateServiceAccountCredentials(
        self.mock_http, self.email, self.scopes)

  def tearDown(self):
    mock.patch.stopall()

  def test_generate_assertion(self):
    self.mock_http.request.return_value = (
      httplib2.Response({'status': 200}),
      '{"accessToken": "testtoken",'
      '"expireTime":"2017-08-17T04:21:32.722952943Z"}')

    self.creds._refresh(None)
    self.assertEqual('testtoken', self.creds.access_token)

    self.mock_http.request.assert_called_once_with(
        uri='https://iamcredentials.googleapis.com/v1/projects/%s/'
        'serviceAccounts/%s:generateAccessToken' % (urllib.parse.quote_plus(
            self.project), urllib.parse.quote_plus(self.email)),
        method='POST',
        body=json.dumps({'scope': self.scopes}),
        headers={'Content-Type': 'application/json'})

    self.assertEqual(['scope'], self.creds._canonicalize_scopes('scope'))

  def test_generate_assertion_httpfail(self):
    self.mock_http.request.return_value = (
      httplib2.Response({'status': 403}), '')

    with self.assertRaises(httplib2_utils.AuthError):
      self.creds._refresh(None)

  def test_authorize_http(self):
    self.mock_http.request.return_value = (
      httplib2.Response({'status': 200}),
      '{"accessToken": "testtoken",'
      '"expireTime":"2017-08-17T04:21:32.722952943Z"}')

    self.creds._refresh(None)
    headers = {}
    self.creds.apply(headers)
    self.assertEqual('Bearer %s' % 'testtoken',
                     headers['Authorization'])
