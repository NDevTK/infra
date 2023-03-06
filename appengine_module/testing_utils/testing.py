# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from contextlib import contextmanager
import collections
import logging
import mock
import six
import time
import unittest

if six.PY2:
  # appengine sdk is supposed to be on the path.
  import dev_appserver
  dev_appserver.fix_sys_path()

  import endpoints

from google.appengine.api import oauth
from google.appengine.api import urlfetch
from google.appengine.ext import ndb
from google.appengine.ext import testbed

import webtest


class MockPatchMixin(object):  # pragma: no cover
  """Adds patch method that can uses mock.patch and stops it in tearDown."""

  _saved_patchers = None

  def add_patcher(self, patcher):
    """Remembers |patcher| to stop it in tearDown."""
    self._saved_patchers = self._saved_patchers or []
    self._saved_patchers.append(patcher)

  def patch(self, *mock_patch_args, **mock_patch_kwargs):
    """Calls mock.patch, starts the returned patcher and stops it in tearDown.

    Returns:
      The mock returned by patch.start().

    Example of usage:
      class MyTest(unittest.TestCase, MockPatchMixin):
        def setUp(self):
          foo = self.patch('module.foo')
          foo.return_value = 'bar'
    """
    patcher = mock.patch(*mock_patch_args, **mock_patch_kwargs)
    mocked = patcher.start()
    self.add_patcher(patcher)
    return mocked

  def tearDown(self):
    """Stop patchers."""
    if self._saved_patchers:
      for p in self._saved_patchers:
        p.stop()
      self._saved_patchers = None


class AppengineTestCase(unittest.TestCase, MockPatchMixin):  # pragma: no cover
  """Base class for Appengine test cases.

  Must set app_module to use self.test_app.
  """

  # To be set in tests that wants to use test_app
  app_module = None

  # To be set in tests that want to test with custom task queues.
  taskqueue_stub_root_path = None

  # To be set in tests that want to change test datastore consistency policy.
  datastore_stub_consistency_policy = None

  def setUp(self):
    super(AppengineTestCase, self).setUp()
    self.testbed = testbed.Testbed()
    # needed because endpoints expects a . in this value
    self.testbed.setup_env(current_version_id='testbed.version')
    self.testbed.activate()
    # Can't use init_all_stubs() because PIL isn't in wheel.
    self.testbed.init_app_identity_stub()
    self.testbed.init_blobstore_stub()
    self.testbed.init_capability_stub()
    self.testbed.init_datastore_v3_stub(
        consistency_policy=self.datastore_stub_consistency_policy)
    self.testbed.init_mail_stub()
    self.testbed.init_memcache_stub()
    self.testbed.init_modules_stub()
    self.testbed.init_taskqueue_stub(root_path=self.taskqueue_stub_root_path)
    self.testbed.init_urlfetch_stub()
    self.testbed.init_user_stub()

    if six.PY2:
      # These stubs aren't in the Py3 appengine-python-standard package.
      self.testbed.init_files_stub()
      self.testbed.init_logservice_stub()
      self.testbed.init_search_stub()
      self.testbed.init_xmpp_stub()

    # Test app is lazily initialized on a first use from app_module.
    self._test_app = None

    self.taskqueue_stub = self.testbed.get_stub(testbed.TASKQUEUE_SERVICE_NAME)

  def tearDown(self):
    try:
      self.testbed.deactivate()
    finally:
      MockPatchMixin.tearDown(self)
      mock.patch.stopall()

  def mock(self, *args, **kwargs):
    # For backwards compatibility with auto_stub.AutoStubMixIn.
    patcher = mock.patch.object(*args, **kwargs)
    patcher.start()
    self.addCleanup(patcher.stop)

  @property
  def test_app(self):
    """Returns instance of webtest.TestApp that wraps app_module."""
    if self._test_app is None:
      # app_module may be a property, so access it only once.
      app = self.app_module
      if app is None:
        self.fail('self.app_module is not provided by the test class')
      self._test_app = webtest.TestApp(
          app, extra_environ={'REMOTE_ADDR': '127.0.0.1'})
    return self._test_app

  def mock_now(self, now):
    """Mocks time in ndb properties that use auto_now and auto_now_add.

    Args:
      now: instance of datetime.datetime.
    """
    self.mock(ndb.DateTimeProperty, '_now', lambda _: now)
    self.mock(ndb.DateProperty, '_now', lambda _: now.date())

  def mock_current_user(self, user_id='', user_email='', is_admin=False):
    # dev_appserver hack.
    self.testbed.setup_env(
      USER_ID=user_id,
      USER_EMAIL=user_email,
      USER_IS_ADMIN=str(int(is_admin)),
      overwrite=True)

  def mock_endpoints_user(self, user_id='', is_admin=False):
    self.mock(endpoints, 'get_current_user', lambda: user_id)
    self.mock(oauth, 'is_current_user_admin', lambda _: is_admin)


  @contextmanager
  def mock_urlfetch(self):
    class UrlHandlers:
      def __init__(self):
        self.response_class = collections.namedtuple(
            'response', ['content', 'status_code', 'headers'])

        self.urls = collections.defaultdict(lambda: self.response_class(
            content=None, status_code=404, headers={}))

      def register_handler(
          self, url, content, status_code=200, headers=None, data=None):
        self.urls[(url, data)] = self.response_class(
            content=content, status_code=status_code, headers=headers or {})

      def handle_url(self, url, payload=None, **_kwargs):
        return self.urls[(url, payload)]


    url_handlers = UrlHandlers()
    yield url_handlers
    self.mock(urlfetch, 'fetch', url_handlers.handle_url)

  def mock_sleep(self):
    self.mock(time, 'sleep', lambda _: None)

  def execute_queued_tasks(self):
    responses = []
    while True:
      # Some tasks spawn more tasks or delete existing tasks, we execute the
      # tasks one by one ordered by (ETA, queue-name, task-name) until empty.
      all_tasks = []
      for queue in self.taskqueue_stub.GetQueues():
        tasks = self.taskqueue_stub.get_filtered_tasks(
            queue_names=queue['name'])
        # Sadly, get_filtered_tasks won't set the queue name in the tasks.
        all_tasks.extend((task, queue['name']) for task in tasks)

      if not all_tasks:
        break

      all_tasks.sort(key=lambda t: (t[0].eta, t[1], t[0].name))
      task, queue_name = all_tasks[0]

      params = task.extract_params()
      extra_environ = {
          'HTTP_X_APPENGINE_TASKNAME': str(task.name),
          'HTTP_X_APPENGINE_QUEUENAME': str(queue_name or 'default'),
      }

      method = {
           'GET': self.test_app.get,
           'POST': self.test_app.post,
      }[task.method]

      responses.append(method(task.url, params, extra_environ=extra_environ))

      self.taskqueue_stub.DeleteTask(queue_name, task.name)

    return responses


class EndpointsTestCase(AppengineTestCase):  # pragma: no cover
  """Base class for a test case that tests Cloud Endpoint Service.

  Usage:
    class MyTestCase(testing.EndpointsTestCase):
      api_service_cls = MyEndpointsService

      def test_stuff(self):
        response = self.call_api('my_method')
        self.assertEqual(...)

      def test_expected_fail(self):
        with self.call_should_fail(403):
          self.call_api('protected_method')
  """

  # Should be set in subclasses to a subclass of remote.Service.
  api_service_cls = None

  # See call_should_fail.
  expected_fail_status = None

  @property
  def app_module(self):
    """WSGI module that wraps the API class, used by AppengineTestCase."""
    # Import endpoints_flask here because this method is not called by
    # every user of the library and they may not have this dependency.
    from components import endpoints_flask
    return endpoints_flask.api_server([self.api_service_cls])

  def call_api(self, method, body=None, status=None):
    """Calls endpoints API method identified by its name."""
    # Because body is a dict and not a ResourceContainer, there's no way to tell
    # which parameters belong in the URL and which belong in the body when the
    # HTTP method supports both. However there's no harm in supplying parameters
    # in both the URL and the body since ResourceContainers don't allow the same
    # parameter name to be used in both places. Supplying parameters in both
    # places produces no ambiguity and extraneous parameters are safely ignored.
    assert hasattr(self.api_service_cls, method), method
    info = getattr(self.api_service_cls, method).method_info
    path = info.get_path(self.api_service_cls.api_info)

    # Identify which arguments are path parameters and which are query strings.
    body = body or {}
    query_strings = []
    for key, value in sorted(body.items()):
      if '{%s}' % str(key) in path:
        path = path.replace('{%s}' % str(key), str(value))
      else:
        # We cannot tell if the parameter is a repeated field from a dict.
        # Allow all query strings to be multi-valued.
        if not isinstance(value, list):
          value = [value]
        for val in value:
          query_strings.append('%s=%s' % (str(key), str(val)))
    if query_strings:
      path = '%s?%s' % (path, '&'.join(query_strings))

    api_info = self.api_service_cls.api_info
    path_version = (
        api_info.path_version
        if hasattr(api_info, 'path_version') else api_info.version)
    path = '/_ah/api/%s/%s/%s' % (api_info.name, path_version, path)
    status = status or self.expected_fail_status
    try:
      if info.http_method == 'DELETE':
        return self.test_app.delete_json(path, body, status=status)
      if info.http_method == 'PATCH':
        return self.test_app.patch_json(path, body, status=status)
      if info.http_method == 'POST':
        return self.test_app.post_json(path, body, status=status)
      if info.http_method == 'PUT':
        return self.test_app.put_json(path, body, status=status)
      return self.test_app.get(path, status=status)
    except Exception as e:
      # Useful for diagnosing issues in test cases.
      logging.info('%s failed: %s', path, e)
      raise

  @contextmanager
  def call_should_fail(self, status):
    """Asserts that Endpoints call inside the guarded region of code fails."""
    # TODO(vadimsh): Get rid of this function and just use
    # call_api(..., status=...). It existed as a workaround for bug that has
    # been fixed:
    # https://code.google.com/p/googleappengine/issues/detail?id=10544
    assert self.expected_fail_status is None, 'nested call_should_fail'
    assert status is not None
    self.expected_fail_status = status
    try:
      yield
    except AssertionError:
      # Assertion can happen if tests are running on GAE < 1.9.31, where
      # endpoints bug still exists (and causes webapp guts to raise assertion).
      # It should be rare (since we are switching to GAE >= 1.9.31), so don't
      # bother to check that assertion was indeed raised. Just skip it if it
      # did.
      pass
    finally:
      self.expected_fail_status = None
