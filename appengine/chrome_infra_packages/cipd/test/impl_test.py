# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import datetime
import hashlib
import random
import StringIO
import unittest
import zipfile

from google.appengine.ext import ndb
from testing_utils import testing

from components import auth
from components import utils

from cas import impl as cas_impl

from cipd import impl
from cipd import processing


class TestValidators(unittest.TestCase):
  def test_is_valid_package_name(self):
    self.assertTrue(impl.is_valid_package_name('a'))
    self.assertTrue(impl.is_valid_package_name('a/b'))
    self.assertTrue(impl.is_valid_package_name('a/b/c/1/2/3'))
    self.assertTrue(impl.is_valid_package_name('infra/tools/cipd'))
    self.assertTrue(impl.is_valid_package_name('-/_'))
    self.assertFalse(impl.is_valid_package_name(''))
    self.assertFalse(impl.is_valid_package_name('/a'))
    self.assertFalse(impl.is_valid_package_name('a/'))
    self.assertFalse(impl.is_valid_package_name('A'))
    self.assertFalse(impl.is_valid_package_name('a/B'))
    self.assertFalse(impl.is_valid_package_name('a\\b'))

  def test_is_valid_package_path(self):
    self.assertTrue(impl.is_valid_package_path('a'))
    self.assertTrue(impl.is_valid_package_path('a/b'))
    self.assertTrue(impl.is_valid_package_path('a/b/c/1/2/3'))
    self.assertTrue(impl.is_valid_package_path('infra/tools/cipd'))
    self.assertTrue(impl.is_valid_package_path('-/_'))
    self.assertFalse(impl.is_valid_package_path(''))
    self.assertFalse(impl.is_valid_package_path('/a'))
    self.assertFalse(impl.is_valid_package_path('a/'))
    self.assertFalse(impl.is_valid_package_path('A'))
    self.assertFalse(impl.is_valid_package_path('a/B'))
    self.assertFalse(impl.is_valid_package_path('a\\b'))

  def test_is_valid_instance_id(self):
    self.assertTrue(impl.is_valid_instance_id('a'*40))
    self.assertFalse(impl.is_valid_instance_id(''))
    self.assertFalse(impl.is_valid_instance_id('A'*40))

  def test_is_valid_package_ref(self):
    self.assertTrue(impl.is_valid_package_ref('ref'))
    self.assertTrue(impl.is_valid_package_ref('abc-_0123'))
    self.assertFalse(impl.is_valid_package_ref(''))
    self.assertFalse(impl.is_valid_package_ref('no-CAPS'))
    self.assertFalse(impl.is_valid_package_ref('a'*500))
    # Tags are not refs.
    self.assertFalse(impl.is_valid_package_ref('key:value'))
    self.assertFalse(impl.is_valid_package_ref('key:'))
    # Instance IDs are not refs.
    self.assertFalse(impl.is_valid_package_ref('a'*40))

  def test_is_valid_instance_tag(self):
    self.assertTrue(impl.is_valid_instance_tag('k:v'))
    self.assertTrue(impl.is_valid_instance_tag('key:'))
    self.assertTrue(impl.is_valid_instance_tag('key-_01234:#$%@\//%$SD'))
    self.assertFalse(impl.is_valid_instance_tag(''))
    self.assertFalse(impl.is_valid_instance_tag('key'))
    self.assertFalse(impl.is_valid_instance_tag('KEY:'))
    self.assertFalse(impl.is_valid_instance_tag('key:' + 'a'*500))

  def test_is_valid_counter_name(self):
    self.assertTrue(impl.is_valid_counter_name('cipd.installed'))
    self.assertTrue(impl.is_valid_counter_name('abc123-_.'))
    self.assertTrue(impl.is_valid_counter_name('a' * 300))
    self.assertFalse(impl.is_valid_counter_name('a' * 301))
    self.assertFalse(impl.is_valid_counter_name('ABC'))
    self.assertFalse(impl.is_valid_counter_name('k:v'))
    self.assertFalse(impl.is_valid_counter_name('a/b'))


class TestRepoService(testing.AppengineTestCase):
  maxDiff = None

  def setUp(self):
    super(TestRepoService, self).setUp()
    self.mocked_cas_service = MockedCASService()
    self.mock(impl.cas, 'get_cas_service', lambda: self.mocked_cas_service)
    self.service = impl.get_repo_service()

  def register_fake_instance(self, pkg_name, instance_id=None):
    _, registered = self.service.register_instance(
        package_name=pkg_name,
        instance_id=instance_id or 'a'*40,
        caller=auth.Identity.from_bytes('user:abc@example.com'),
        now=datetime.datetime(2014, 1, 1, 0, 0))
    self.assertTrue(registered)

  def test_delete_package_ok(self):
    caller = auth.Identity.from_bytes('user:abc@example.com')
    # Setup all sorts of stuff associated with a package.
    self.register_fake_instance('a/b', 'a'*40)
    self.register_fake_instance('a/b', 'b'*40)
    self.service.set_package_ref('a/b', 'ref1', 'a'*40, caller)
    self.service.set_package_ref('a/b', 'ref2', 'b'*40, caller)
    self.service.attach_tags('a/b', 'a'*40, ['tag1:tag1', 'tag2:tag2'], caller)
    self.service.attach_tags('a/b', 'b'*40, ['tag1:tag1', 'tag2:tag2'], caller)

    # Another package, to make sure it stays alive.
    self.register_fake_instance('c/d')

    # Delete a/b and all associated stuff. Ensure entire entity group is nuked.
    # The implementation of delete_package doesn't use this sort of query (and
    # use explicit list of entity classes) as a reminder for future developers
    # to be mindful about what they are deleting and how.
    q = 'SELECT __key__ WHERE ANCESTOR IS :1'
    self.assertTrue(ndb.gql(q, impl.package_key('a/b')).fetch())
    self.service.delete_package('a/b')
    self.assertFalse(ndb.gql(q, impl.package_key('a/b')).fetch())
    self.assertFalse(self.service.get_instance('a/b', 'a'*40))

    # Another package is still fine.
    self.assertTrue(self.service.get_instance('c/d', 'a'*40))

  def test_delete_package_missing(self):
    self.assertIsNone(self.service.get_package('a/b'))
    self.assertFalse(self.service.delete_package('a/b'))

  def test_list_packages_no_path(self):
    self.assertIsNone(self.service.get_package('a/b'))
    self.assertIsNone(self.service.get_package('y/z'))
    self.register_fake_instance('y/z')
    self.register_fake_instance('a/b')
    self.assertEqual(([], ['a', 'y']),
                     self.service.list_packages('', False))
    self.assertEqual((['a/b', 'y/z'], ['a', 'y']),
                     self.service.list_packages('', True))

  def test_list_packages_with_path(self):
    self.assertIsNone(self.service.get_package('a/b'))
    self.assertIsNone(self.service.get_package('y/x'))
    self.assertIsNone(self.service.get_package('y/z/z'))
    self.register_fake_instance('y/x')
    self.register_fake_instance('y/z/z')
    self.register_fake_instance('a/b')
    self.assertEqual((['y/x'], ['y/z']), self.service.list_packages('y', False))
    self.assertEqual((['y/z/z'], []),
                     self.service.list_packages('y/z/z', False))
    self.assertEqual((['y/x'], ['y/z']),
                     self.service.list_packages('y/', False))
    self.assertEqual((['y/x', 'y/z/z'], ['y/z']),
                     self.service.list_packages('y', True))

  def test_list_packages_ignore_substrings(self):
    self.assertIsNone(self.service.get_package('good/path'))
    self.register_fake_instance('good/path')
    self.assertEqual((['good/path'], []),
                     self.service.list_packages('good', False))
    self.assertEqual((['good/path'], []),
                     self.service.list_packages('good/', False))
    self.assertEqual(([], []),
                     self.service.list_packages('goo', False))

  def test_list_packages_where_a_package_is_also_a_directory(self):
    self.assertIsNone(self.service.get_package('good'))
    self.assertIsNone(self.service.get_package('good/path'))
    self.register_fake_instance('good')
    self.register_fake_instance('good/path')
    self.assertEqual((['good'], ['good']),
                     self.service.list_packages('', False))
    self.assertEqual((['good', 'good/path'], ['good']),
                     self.service.list_packages('', True))
    # To keep things simple we match packages with names matching the search
    # with the trailing slash stripped.
    self.assertEqual((['good', 'good/path'], []),
                     self.service.list_packages('good/', False))

  def test_list_packages_with_an_empty_directory(self):
    self.assertIsNone(self.service.get_package('good/sub/path'))
    self.register_fake_instance('good/sub/path')
    self.assertEqual(([], ['good/sub']),
                     self.service.list_packages('good', False))
    self.assertEqual((['good/sub/path'], ['good/sub']),
                     self.service.list_packages('good', True))
    self.assertEqual((['good/sub/path'], ['good', 'good/sub']),
                     self.service.list_packages('', True))

  def test_list_packages_with_hidden_packages(self):
    self.assertIsNone(self.service.get_package('a/b'))
    self.assertIsNone(self.service.get_package('a/c'))
    self.register_fake_instance('a/b')
    self.register_fake_instance('a/c')

    # Both are visible initially.
    self.assertEqual(
        (['a/b', 'a/c'], []),
        self.service.list_packages('a/', False, False))

    def mutation(pkg):
      pkg.hidden = True
      return True
    self.service.modify_package('a/c', mutation)

    # 'a/c' is no longer visible.
    self.assertEqual(
        (['a/b'], []),
        self.service.list_packages('a/', False, False))

    # Unless asked to show hidden packages.
    self.assertEqual(
        (['a/b', 'a/c'], []),
        self.service.list_packages('a/', False, True))

  def test_modify_package_ok(self):
    self.register_fake_instance('a/b')
    pkg = self.service.get_package('a/b')
    self.assertFalse(pkg.hidden)

    def mutation(pkg):
      pkg.hidden = True
      return True
    pkg = self.service.modify_package('a/b', mutation)
    self.assertTrue(pkg.hidden)

    pkg = self.service.get_package('a/b')
    self.assertTrue(pkg.hidden)

  def test_modify_package_unchanged(self):
    self.register_fake_instance('a/b')
    pkg = self.service.get_package('a/b')
    self.assertFalse(pkg.hidden)

    def mutation(pkg):
      pkg.hidden = True
      return False
    pkg = self.service.modify_package('a/b', mutation)
    self.assertTrue(pkg.hidden) # returns whatever 'mutation' did

    pkg = self.service.get_package('a/b')
    self.assertFalse(pkg.hidden) # the change wasn't persisted though

  def test_modify_package_missing(self):
    def mutation(_):
      return False # pragma: no cover
    pkg = self.service.modify_package('a/b', mutation)
    self.assertIsNone(pkg)

  def test_register_instance_new(self):
    self.assertIsNone(self.service.get_instance('a/b', 'a'*40))
    self.assertIsNone(self.service.get_package('a/b'))
    inst, registered = self.service.register_instance(
        package_name='a/b',
        instance_id='a'*40,
        caller=auth.Identity.from_bytes('user:abc@example.com'),
        now=datetime.datetime(2014, 1, 1, 0, 0))
    self.assertTrue(registered)
    self.assertEqual(
        ndb.Key('Package', 'a/b', 'PackageInstance', 'a'*40), inst.key)
    self.assertEqual('a/b', inst.package_name)
    self.assertEqual('a'*40, inst.instance_id)
    expected = {
      'registered_by': auth.Identity(kind='user', name='abc@example.com'),
      'registered_ts': datetime.datetime(2014, 1, 1, 0, 0),
      'processors_failure': [],
      'processors_pending': [],
      'processors_success': [],
    }
    self.assertEqual(expected, inst.to_dict())
    self.assertEqual(
        expected, self.service.get_instance('a/b', 'a'*40).to_dict())
    pkg = self.service.get_package('a/b')
    self.assertTrue(pkg)
    self.assertEqual('a/b', pkg.package_name)

  def test_register_instance_existing(self):
    # First register a package.
    inst1, registered = self.service.register_instance(
        package_name='a/b',
        instance_id='a'*40,
        caller=auth.Identity.from_bytes('user:abc@example.com'))
    self.assertTrue(registered)
    # Try to register it again.
    inst2, registered = self.service.register_instance(
          package_name='a/b',
          instance_id='a'*40,
          caller=auth.Identity.from_bytes('user:def@example.com'))
    self.assertFalse(registered)
    self.assertEqual(inst1.to_dict(), inst2.to_dict())

  def test_generate_fetch_url(self):
    inst, registered = self.service.register_instance(
        package_name='a/b',
        instance_id='a'*40,
        caller=auth.Identity.from_bytes('user:abc@example.com'),
        now=datetime.datetime(2014, 1, 1, 0, 0))
    self.assertTrue(registered)
    url = self.service.generate_fetch_url(inst)
    self.assertEqual(
        'https://signed-url/SHA1/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa', url)

  def test_is_instance_file_uploaded(self):
    self.mocked_cas_service.uploaded[('SHA1', 'a'*40)] = ''
    self.assertTrue(self.service.is_instance_file_uploaded('a/b', 'a'*40))
    self.assertFalse(self.service.is_instance_file_uploaded('a/b', 'b'*40))

  def test_create_upload_session(self):
    upload_url, upload_session_id = self.service.create_upload_session(
        'a/b', 'a'*40, auth.Identity.from_bytes('user:abc@example.com'))
    self.assertEqual('http://upload_url', upload_url)
    self.assertEqual('upload_session_id', upload_session_id)

  def test_register_instance_with_processing(self):
    self.mock(utils, 'utcnow', lambda: datetime.datetime(2014, 1, 1))

    self.service.processors.append(MockedProcessor('bad', 'Error message'))
    self.service.processors.append(MockedProcessor('good'))

    tasks = []
    def mocked_enqueue_task(**kwargs):
      tasks.append(kwargs)
      return True
    self.mock(impl.utils, 'enqueue_task', mocked_enqueue_task)

    # The processors are added to the pending list.
    inst, registered = self.service.register_instance(
        package_name='a/b',
        instance_id='a'*40,
        caller=auth.Identity.from_bytes('user:abc@example.com'),
        now=datetime.datetime(2014, 1, 1, 0, 0))
    self.assertTrue(registered)
    expected = {
      'registered_by': auth.Identity(kind='user', name='abc@example.com'),
      'registered_ts': datetime.datetime(2014, 1, 1, 0, 0),
      'processors_failure': [],
      'processors_pending': ['bad', 'good'],
      'processors_success': [],
    }
    self.assertEqual(expected, inst.to_dict())

    # The processing task is enqueued.
    self.assertEqual([{
      'payload': '{"instance_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", '
          '"package_name": "a/b", "processors": ["bad", "good"]}',
      'queue_name': 'cipd-process',
      'transactional': True,
      'url': '/internal/taskqueue/cipd-process/'
          'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
    }], tasks)

    # Now execute the task.
    self.service.process_instance(
        package_name='a/b',
        instance_id='a'*40,
        processors=['bad', 'good'])

    # Assert the final state.
    inst = self.service.get_instance('a/b', 'a'*40)
    expected = {
      'registered_by': auth.Identity(kind='user', name='abc@example.com'),
      'registered_ts': datetime.datetime(2014, 1, 1, 0, 0),
      'processors_failure': ['bad'],
      'processors_pending': [],
      'processors_success': ['good'],
    }
    self.assertEqual(expected, inst.to_dict())

    good_result = self.service.get_processing_result('a/b', 'a'*40, 'good')
    self.assertEqual({
      'created_ts': datetime.datetime(2014, 1, 1),
      'error': None,
      'result': {
        'instance_id': 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
        'package_name': 'a/b',
        'processor_name': 'good',
      },
      'success': True,
    }, good_result.to_dict())

    bad_result = self.service.get_processing_result('a/b', 'a'*40, 'bad')
    self.assertEqual({
      'created_ts': datetime.datetime(2014, 1, 1),
      'error': 'Error message',
      'result': None,
      'success': False,
    }, bad_result.to_dict())

  def test_client_binary_extraction(self):
    self.mock(utils, 'utcnow', lambda: datetime.datetime(2014, 1, 1))

    # Prepare fake cipd binary package.
    out = StringIO.StringIO()
    zf = zipfile.ZipFile(out, 'w', zipfile.ZIP_DEFLATED)
    zf.writestr('cipd', 'cipd binary data here')
    zf.close()
    zipped = out.getvalue()
    digest = hashlib.sha1(zipped).hexdigest()

    # Pretend it is uploaded.
    self.mocked_cas_service.uploaded[('SHA1', digest)] = zipped

    # Register it as a package instance.
    self.mock(impl.utils, 'enqueue_task', lambda **_args: True)
    inst, registered = self.service.register_instance(
        package_name='infra/tools/cipd/linux-amd64',
        instance_id=digest,
        caller=auth.Identity.from_bytes('user:abc@example.com'),
        now=datetime.datetime(2014, 1, 1, 0, 0))
    self.assertTrue(registered)
    expected = {
      'registered_by': auth.Identity(kind='user', name='abc@example.com'),
      'registered_ts': datetime.datetime(2014, 1, 1, 0, 0),
      'processors_failure': [],
      'processors_pending': ['cipd_client_binary:v1'],
      'processors_success': [],
    }
    self.assertEqual(expected, inst.to_dict())

    # get_client_binary_info indicated that processing is not done yet.
    instance = self.service.get_instance(
        package_name='infra/tools/cipd/linux-amd64',
        instance_id=digest)
    info, error_msg = self.service.get_client_binary_info(instance)
    self.assertIsNone(info)
    self.assertIsNone(error_msg)

    # Execute post-processing task: it would extract CIPD binary.
    self.service.process_instance(
        package_name='infra/tools/cipd/linux-amd64',
        instance_id=digest,
        processors=['cipd_client_binary:v1'])

    # Ensure succeeded.
    result = self.service.get_processing_result(
        package_name='infra/tools/cipd/linux-amd64',
        instance_id=digest,
        processor_name='cipd_client_binary:v1')
    self.assertEqual({
      'created_ts': datetime.datetime(2014, 1, 1, 0, 0),
      'success': True,
      'error': None,
      'result': {
        'client_binary': {
          'hash_algo': 'SHA1',
          'hash_digest': '5a72c1535f8d132c341585207504d94e68ef8a9d',
          'size': 21,
        },
      },
    }, result.to_dict())

    # Verify get_client_binary_info works too.
    instance = self.service.get_instance(
        package_name='infra/tools/cipd/linux-amd64',
        instance_id=digest)
    info, error_msg = self.service.get_client_binary_info(
        instance, filename='boo')
    expected = impl.ClientBinaryInfo(
        sha1='5a72c1535f8d132c341585207504d94e68ef8a9d',
        size=21,
        fetch_url=(
            'https://signed-url/SHA1/5a72c1535f8d132c341585207504d94e68ef8a9d'
            '?filename=boo'))
    self.assertIsNone(error_msg)
    self.assertEqual(expected, info)

  def test_client_binary_extract_failure(self):
    self.mock(utils, 'utcnow', lambda: datetime.datetime(2014, 1, 1))

    # Pretend some fake data is uploaded.
    self.mocked_cas_service.uploaded[('SHA1', 'a'*40)] = 'not a zip'

    # Register it as a package instance.
    self.mock(impl.utils, 'enqueue_task', lambda **_args: True)
    inst, registered = self.service.register_instance(
        package_name='infra/tools/cipd/linux-amd64',
        instance_id='a'*40,
        caller=auth.Identity.from_bytes('user:abc@example.com'),
        now=datetime.datetime(2014, 1, 1, 0, 0))
    self.assertTrue(registered)
    expected = {
      'registered_by': auth.Identity(kind='user', name='abc@example.com'),
      'registered_ts': datetime.datetime(2014, 1, 1, 0, 0),
      'processors_failure': [],
      'processors_pending': ['cipd_client_binary:v1'],
      'processors_success': [],
    }
    self.assertEqual(expected, inst.to_dict())

    # Execute post-processing task: it would fail extracting CIPD binary.
    self.service.process_instance(
        package_name='infra/tools/cipd/linux-amd64',
        instance_id='a'*40,
        processors=['cipd_client_binary:v1'])

    # Ensure error is reported.
    result = self.service.get_processing_result(
        package_name='infra/tools/cipd/linux-amd64',
        instance_id='a'*40,
        processor_name='cipd_client_binary:v1')
    self.assertEqual({
      'created_ts': datetime.datetime(2014, 1, 1, 0, 0),
      'success': False,
      'error': 'File is not a zip file',
      'result': None,
    }, result.to_dict())

    # Verify get_client_binary_info reports it too.
    instance = self.service.get_instance(
        package_name='infra/tools/cipd/linux-amd64',
        instance_id='a'*40)
    info, error_msg = self.service.get_client_binary_info(instance)
    self.assertIsNone(info)
    self.assertEqual(
        'Failed to extract the binary: File is not a zip file', error_msg)

  def test_set_package_ref(self):
    ident1 = auth.Identity.from_bytes('user:abc@example.com')
    now1 = datetime.datetime(2015, 1, 1, 0, 0)

    ident2 = auth.Identity.from_bytes('user:def@example.com')
    now2 = datetime.datetime(2016, 1, 1, 0, 0)

    self.service.register_instance(
        package_name='a/b',
        instance_id='a'*40,
        caller=ident1,
        now=datetime.datetime(2014, 1, 1, 0, 0))
    self.service.register_instance(
        package_name='a/b',
        instance_id='b'*40,
        caller=ident1,
        now=datetime.datetime(2014, 1, 1, 0, 0))

    ref = self.service.set_package_ref('a/b', 'ref', 'a'*40, ident1, now1)
    self.assertEqual({
      'instance_id': 'a'*40,
      'modified_by': ident1,
      'modified_ts': now1,
    }, ref.to_dict())
    self.assertEqual('ref', ref.ref)

    # Move to the same value -> modified_ts do not change.
    ref = self.service.set_package_ref('a/b', 'ref', 'a'*40, ident2, now2)
    self.assertEqual({
      'instance_id': 'a'*40,
      'modified_by': ident1,
      'modified_ts': now1,
    }, ref.to_dict())

    # Move to another value.
    ref = self.service.set_package_ref('a/b', 'ref', 'b'*40, ident2, now2)
    self.assertEqual({
      'instance_id': 'b'*40,
      'modified_by': ident2,
      'modified_ts': now2,
    }, ref.to_dict())

    # Code coverage for package_name.
    self.assertEqual('a/b', ref.package_name)

  def test_get_package_refs(self):
    ident = auth.Identity.from_bytes('user:abc@example.com')
    now1 = datetime.datetime(2015, 1, 1, 0, 0)
    now2 = datetime.datetime(2015, 1, 2, 0, 0)

    self.register_fake_instance('a/b', 'a'*40)
    self.service.set_package_ref('a/b', 'ref1', 'a'*40, ident, now1)

    self.register_fake_instance('a/b', 'b'*40)
    self.service.set_package_ref('a/b', 'ref2', 'b'*40, ident, now2)

    refs = self.service.get_package_refs('a/b', ['ref1', 'ref2', 'missing'])
    self.assertEqual({
      'missing': None,
      'ref1': {
        'instance_id': 'a'*40,
        'modified_by': ident,
        'modified_ts': now1,
      },
      'ref2': {
        'instance_id': 'b'*40,
        'modified_by': ident,
        'modified_ts': now2,
      },
    }, {k: v.to_dict() if v else None for k, v in refs.iteritems()})

  def test_query_package_refs(self):
    ident = auth.Identity.from_bytes('user:abc@example.com')
    now1 = datetime.datetime(2015, 1, 1, 0, 0)
    now2 = datetime.datetime(2015, 1, 2, 0, 0)

    self.register_fake_instance('a/b', 'a'*40)
    self.service.set_package_ref('a/b', 'ref1', 'a'*40, ident, now1)

    self.register_fake_instance('a/b', 'b'*40)
    self.service.set_package_ref('a/b', 'ref2', 'b'*40, ident, now2)

    refs = self.service.query_package_refs('a/b')
    self.assertEqual([
      {
        'instance_id': 'b'*40,
        'modified_by': ident,
        'modified_ts': now2,
      },
      {
        'instance_id': 'a'*40,
        'modified_by': ident,
        'modified_ts': now1,
      },
    ], [ref.to_dict() for ref in refs])

  def test_query_instance_refs(self):
    ident = auth.Identity.from_bytes('user:abc@example.com')
    now1 = datetime.datetime(2015, 1, 1, 0, 0)
    now2 = datetime.datetime(2015, 1, 2, 0, 0)

    self.register_fake_instance('a/b', 'a'*40)
    self.service.set_package_ref('a/b', 'ref1', 'a'*40, ident, now1)
    self.service.set_package_ref('a/b', 'ref2', 'a'*40, ident, now2)

    # Should not appear in results.
    self.register_fake_instance('a/b', 'b'*40)
    self.service.set_package_ref('a/b', 'ref3', 'b'*40, ident, now1)

    refs = self.service.query_instance_refs('a/b', 'a'*40)
    self.assertEqual(['ref2', 'ref1'], [ref.ref for ref in refs])

  def test_attach_detach_tags(self):
    _, registered = self.service.register_instance(
        package_name='a/b',
        instance_id='a'*40,
        caller=auth.Identity.from_bytes('user:abc@example.com'),
        now=datetime.datetime(2014, 1, 1, 0, 0))
    self.assertTrue(registered)

    # Add a tag.
    attached = self.service.attach_tags(
        package_name='a/b',
        instance_id='a'*40,
        tags=['tag1:value1'],
        caller=auth.Identity.from_bytes('user:abc@example.com'),
        now=datetime.datetime(2014, 1, 1, 0, 0))
    self.assertEqual(
      {
        'tag1:value1': {
          'registered_by': auth.Identity(kind='user', name='abc@example.com'),
          'registered_ts': datetime.datetime(2014, 1, 1, 0, 0),
          'tag': 'tag1:value1',
        },
      }, {k: e.to_dict() for k, e in attached.iteritems()})
    self.assertEqual('a/b', attached['tag1:value1'].package_name)
    self.assertEqual('a'*40, attached['tag1:value1'].instance_id)

    # Attempt to attach existing one (and one new).
    attached = self.service.attach_tags(
        package_name='a/b',
        instance_id='a'*40,
        tags=['tag1:value1', 'tag2:value2'],
        caller=auth.Identity.from_bytes('user:abc@example.com'),
        now=datetime.datetime(2015, 1, 1, 0, 0))
    self.assertEqual(
      {
        'tag1:value1': {
          'registered_by': auth.Identity(kind='user', name='abc@example.com'),
          # Didn't change to 2015.
          'registered_ts': datetime.datetime(2014, 1, 1, 0, 0),
          'tag': 'tag1:value1',
        },
        'tag2:value2': {
          'registered_by': auth.Identity(kind='user', name='abc@example.com'),
          'registered_ts': datetime.datetime(2015, 1, 1, 0, 0),
          'tag': 'tag2:value2',
        },
      }, {k: e.to_dict() for k, e in attached.iteritems()})

    # Get specific tags.
    tags = self.service.get_tags('a/b', 'a'*40, ['tag1:value1', 'missing:'])
    self.assertEqual(
      {
        'tag1:value1': {
          'registered_by': auth.Identity(kind='user', name='abc@example.com'),
          'registered_ts': datetime.datetime(2014, 1, 1, 0, 0),
          'tag': 'tag1:value1',
        },
        'missing:': None,
      }, {k: e.to_dict() if e else None for k, e in tags.iteritems()})

    # Get all tags. Newest first.
    tags = self.service.query_tags('a/b', 'a'*40)
    self.assertEqual(['tag2:value2', 'tag1:value1'], [t.tag for t in tags])

    # Search by specific tag (in a package).
    found = self.service.search_by_tag('tag1:value1', package_name='a/b')
    self.assertEqual(
        [('a/b', 'a'*40)], [(e.package_name, e.instance_id) for e in found])

    # Search by specific tag (globally). Use callback to cover this code path.
    found = self.service.search_by_tag('tag1:value1')
    self.assertEqual(
        [('a/b', 'a'*40)], [(e.package_name, e.instance_id) for e in found])

    # Cover callback usage.
    found = self.service.search_by_tag(
        'tag1:value1', callback=lambda *_a: False)
    self.assertFalse(found)

    # Remove tag, search again -> missing.
    self.service.detach_tags('a/b', 'a'*40, ['tag1:value1', 'missing:'])
    found = self.service.search_by_tag('tag1:value1')
    self.assertFalse(found)

  def add_tagged_instance(self, package_name, instance_id, tags):
    self.service.register_instance(
        package_name=package_name,
        instance_id=instance_id,
        caller=auth.Identity.from_bytes('user:abc@example.com'),
        now=datetime.datetime(2014, 1, 1, 0, 0))
    self.service.attach_tags(
        package_name=package_name,
        instance_id=instance_id,
        tags=tags,
        caller=auth.Identity.from_bytes('user:abc@example.com'),
        now=datetime.datetime(2014, 1, 1, 0, 0))

  def test_resolve_version(self):
    self.add_tagged_instance('a/b', 'a'*40, ['tag1:value1', 'tag2:value2'])
    self.add_tagged_instance('a/b', 'b'*40, ['tag1:value1'])
    self.add_tagged_instance('a/b', 'c'*40, ['tag1:value1'])
    self.service.set_package_ref(
        'a/b', 'ref', 'a'*40, auth.Identity.from_bytes('user:abc@example.com'))

    self.assertEqual([], self.service.resolve_version('a/b', 'd'*40, 2))
    self.assertEqual([], self.service.resolve_version('a/b', 'tag3:', 2))
    self.assertEqual([], self.service.resolve_version('a/b/c/d', 'a'*40, 2))
    self.assertEqual([], self.service.resolve_version('a/b', 'not-such-ref', 2))
    self.assertEqual(['a'*40], self.service.resolve_version('a/b', 'ref', 2))
    self.assertEqual(['a'*40], self.service.resolve_version('a/b', 'a'*40, 2))
    self.assertEqual(
        ['a'*40], self.service.resolve_version('a/b', 'tag2:value2', 2))

    # No order guarantees when multiple results match.
    res = self.service.resolve_version('a/b', 'tag1:value1', 2)
    self.assertEqual(2, len(res))
    self.assertTrue(set(['a'*40, 'b'*40, 'c'*40]).issuperset(res))

  def test_list_instances_success(self):
    now = datetime.datetime(2018, 1, 1, 0, 0)
    pkg = 'package/name'

    def mk(iid, ts, by='user:a@example.com', procs=0):
      inst = impl.PackageInstance(
          key=impl.package_instance_key(pkg, iid),
          registered_by=auth.Identity.from_bytes(by),
          registered_ts=now+datetime.timedelta(seconds=ts),
          processors_pending=['proc']*procs)
      inst.put()
      return inst

    # No instanced yet at all.
    res, cursor = self.service.list_instances(pkg)
    self.assertEqual([], res)
    self.assertIsNone(cursor)

    # Add a bunch of instances (all are ready).
    old = mk('a'*40, -5)
    fresh = mk('b'*40, 0)
    oldest = mk('c'*40, -10)

    def do_tests():
      # Returned in correct order.
      res, cursor = self.service.list_instances(pkg)
      self.assertEqual([fresh, old, oldest], res)
      self.assertIsNone(cursor)

      # Pagination works too.
      res, cursor = self.service.list_instances(pkg, limit=2)
      self.assertEqual([fresh, old], res)
      self.assertIsNotNone(cursor)
      res, cursor = self.service.list_instances(pkg, limit=2, cursor=cursor)
      self.assertEqual([oldest], res)
      self.assertIsNone(cursor)

    do_tests()

    # Add one more not-yet-ready instance.
    mk('d'*40, 5, procs=1)

    # The listing totally ignores it.
    do_tests()

  def test_list_instances_bad_cursor(self):
    with self.assertRaises(ValueError):
      self.service.list_instances('a/b', cursor='watsup')

  def test_read_missing_counter(self):
    counter = self.service.read_counter('a/b', 'a'*40, 'test.counter')
    self.assertEqual(0, counter.value)
    self.assertIsNone(counter.created_ts)
    self.assertIsNone(counter.updated_ts)

  def test_touch_counter(self):
    self.mock(utils, 'utcnow', lambda: datetime.datetime(2014, 1, 1))

    self.service.increment_counter('a/b', 'a'*40, 'test.counter', 0)
    counter = self.service.read_counter('a/b', 'a'*40, 'test.counter')
    self.assertEqual(0, counter.value)
    self.assertEqual(datetime.datetime(2014, 1, 1), counter.created_ts)
    self.assertEqual(datetime.datetime(2014, 1, 1), counter.updated_ts)

  def test_increment_counter_timestamps(self):
    self.mock(utils, 'utcnow', lambda: datetime.datetime(2014, 1, 1))
    self.service.increment_counter('a/b', 'a'*40, 'test.counter', 1)

    self.mock(utils, 'utcnow', lambda: datetime.datetime(2014, 1, 2))
    self.service.increment_counter('a/b', 'a'*40, 'test.counter', 1)

    self.mock(utils, 'utcnow', lambda: datetime.datetime(2014, 1, 3))
    self.service.increment_counter('a/b', 'a'*40, 'test.counter', 1)

    counter = self.service.read_counter('a/b', 'a'*40, 'test.counter')
    self.assertEqual(3, counter.value)
    self.assertEqual(datetime.datetime(2014, 1, 1), counter.created_ts)
    self.assertEqual(datetime.datetime(2014, 1, 3), counter.updated_ts)

  def test_increment_counter_same_shard(self):
    self.mock(utils, 'utcnow', lambda: datetime.datetime(2014, 1, 1))
    self.mock(random, 'randint', lambda a, b: 42)
    self.service.increment_counter('a/b', 'a'*40, 'test.counter', 1)

    self.mock(utils, 'utcnow', lambda: datetime.datetime(2014, 1, 2))
    self.service.increment_counter('a/b', 'a'*40, 'test.counter', 1)

    counter = self.service.read_counter('a/b', 'a'*40, 'test.counter')
    self.assertEqual(2, counter.value)
    self.assertEqual(datetime.datetime(2014, 1, 1), counter.created_ts)
    self.assertEqual(datetime.datetime(2014, 1, 2), counter.updated_ts)


class MockedCASService(object):
  def __init__(self):
    self.uploaded = {}

  def generate_fetch_url(self, algo, digest, filename=None):
    r = 'https://signed-url/%s/%s' % (algo, digest)
    if filename:
      r += '?filename=%s' % filename
    return r

  def is_object_present(self, algo, digest):
    return (algo, digest) in self.uploaded

  def create_upload_session(self, _algo, _digest, _caller):
    class UploadSession(object):
      upload_url = 'http://upload_url'
    return UploadSession(), 'upload_session_id'

  def open(self, hash_algo, hash_digest, read_buffer_size):
    assert read_buffer_size > 0
    if not self.is_object_present(hash_algo, hash_digest):  # pragma: no cover
      raise cas_impl.NotFoundError()
    return StringIO.StringIO(self.uploaded[(hash_algo, hash_digest)])

  def start_direct_upload(self, hash_algo):
    assert hash_algo == 'SHA1'
    return cas_impl.DirectUpload(
        file_obj=StringIO.StringIO(),
        hasher=hashlib.sha1(),
        callback=lambda *_args: None)


class MockedProcessor(processing.Processor):
  def __init__(self, name, error=None):
    self.name = name
    self.error = error

  def should_process(self, instance):
    return True

  def run(self, instance, data):
    if self.error:
      raise processing.ProcessingError(self.error)
    return {
      'instance_id': instance.instance_id,
      'package_name': instance.package_name,
      'processor_name': self.name,
    }
