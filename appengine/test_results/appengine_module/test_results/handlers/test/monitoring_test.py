# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json
import os
import tempfile
import webapp2

from appengine_module.testing_utils import testing

from appengine_module.test_results import main
from appengine_module.test_results.handlers.monitoring import EventMonUploader
from appengine_module.test_results.model.testfile import TestFile
from infra_libs import event_mon


TEST_JSON = {
    'tests': {
      'web-animations-api': {
        'animation-state-changes.html': {
          'expected': 'PASS IMAGE+TEXT FOOBAR',
          'actual': 'PASS',
          'has_stderr': True,
          'time': 0.1
        }
      }
    },
    'test_locations': {
      'web-animations-api': {
        'file': 'web/animations_api.cc',
        'line': 123,
      },
    },
    'interrupted': False,
    'version': 3,
    'seconds_since_epoch': 1457612314.123,
    'path_delimiter': '\\',
}

REQ_PAYLOAD = json.dumps({
  'master': 'master',
  'builder': 'builder',
  'build_number': '123',
  'test_type': 'ui_tests',
  'step_name': 'ui_tests 12s pending',
})

class EventMonUploaderTest(testing.AppengineTestCase):
  app_module = main.app  # for testing.AppengineTestCase

  def setUp(self):
    super(EventMonUploaderTest, self).setUp()
    tmp_handle, self.event_mon_file = tempfile.mkstemp()
    os.close(tmp_handle)
    event_mon.close()
    event_mon.setup_monitoring(
        run_type='file', hostname='test_host', output_file=self.event_mon_file)

  def tearDown(self):
    event_mon.close()
    os.remove(self.event_mon_file)
    super(EventMonUploaderTest, self).tearDown()

  def read_event_mon_file(self):
    with open(self.event_mon_file, 'rb') as f:
      log_proto = event_mon.LogRequestLite.FromString(f.read())
    return [event_mon.ChromeInfraEvent.FromString(ev.source_extension)
            for ev in log_proto.log_event]

  def test_creates_event_mon_event_correctly(self):
    TestFile.add_file(
        'master', 'builder', 'ui_tests', 123, 'full_results.json',
        json.dumps(TEST_JSON))
    response = self.test_app.post(
        '/internal/monitoring/test_res/upload', REQ_PAYLOAD)

    self.assertEqual(200, response.status_int)
    events = self.read_event_mon_file()
    self.assertEqual(1, len(events))
    self.assertEqual(events[0].test_results.master_name, 'master')
    self.assertEqual(events[0].test_results.builder_name, 'builder')
    self.assertEqual(events[0].test_results.build_number, 123)
    self.assertEqual(events[0].test_results.test_type, 'ui_tests')
    self.assertEqual(events[0].test_results.step_name, 'ui_tests 12s pending')
    self.assertEqual(events[0].test_results.interrupted, False)
    self.assertEqual(events[0].test_results.version, 3)
    self.assertEqual(events[0].test_results.usec_since_epoch, 1457612314123000)
    self.assertEqual(1, len(events[0].test_results.tests))
    self.assertEqual(
        events[0].test_results.tests[0].test_name,
        'web-animations-api\\animation-state-changes.html')
    self.assertEqual(
        events[0].test_results.tests[0].expected,
        [
          event_mon.protos.chrome_infra_log_pb2.TestResultsEvent.PASS,
          event_mon.protos.chrome_infra_log_pb2.TestResultsEvent.IMAGE_TEXT,
          event_mon.protos.chrome_infra_log_pb2.TestResultsEvent.UNKNOWN,
        ])
    self.assertEqual(
        events[0].test_results.tests[0].actual,
        [event_mon.protos.chrome_infra_log_pb2.TestResultsEvent.PASS])

  def test_returns_400_on_missing_request_payload(self):
    self.test_app.post('/internal/monitoring/test_res/upload', status=400)
    self.assertEqual(0, len(self.read_event_mon_file()))

  def test_returns_400_on_non_json_request_payload(self):
    self.test_app.post(
        '/internal/monitoring/test_res/upload', 'foobar', status=400)
    self.assertEqual(0, len(self.read_event_mon_file()))

  def test_returns_400_on_missing_request_params(self):
    self.test_app.post('/internal/monitoring/test_res/upload', '{}', status=400)
    self.assertEqual(0, len(self.read_event_mon_file()))

  def test_returns_400_on_missing_file(self):
    self.test_app.post(
        '/internal/monitoring/test_res/upload', REQ_PAYLOAD, status=400)
    self.assertEqual(0, len(self.read_event_mon_file()))

  def test_does_not_crash_on_missing_required_fields_in_json(self):
    TestFile.add_file(
        'master', 'builder', 'ui_tests', 123, 'full_results.json', '{}')
    response = self.test_app.post(
        '/internal/monitoring/test_res/upload', REQ_PAYLOAD)
    self.assertEqual(200, response.status_int)
    events = self.read_event_mon_file()
    self.assertEqual(1, len(events))
    self.assertFalse(events[0].test_results.HasField('interrupted'))
    self.assertFalse(events[0].test_results.HasField('version'))
    self.assertFalse(events[0].test_results.HasField('usec_since_epoch'))
    self.assertEqual(0, len(events[0].test_results.tests))

  def test_uploads_test_locations_correctly(self):
    TestFile.add_file(
        'master', 'builder', 'ui_tests', 123, 'full_results.json',
        json.dumps(TEST_JSON))
    response = self.test_app.post(
        '/internal/monitoring/test_loc/upload', REQ_PAYLOAD)

    self.assertEqual(200, response.status_int)
    events = self.read_event_mon_file()
    self.assertEqual(1, len(events))
    self.assertEqual(events[0].test_locations_event.bucket_name, 'master')
    self.assertEqual(events[0].test_locations_event.builder_name, 'builder')
    self.assertEqual(events[0].test_locations_event.build_number, 123)
    self.assertEqual(
        events[0].test_locations_event.step_name, 'ui_tests 12s pending')
    self.assertEqual(
        events[0].test_locations_event.usec_since_epoch, 1457612314123000)
    self.assertEqual(1, len(events[0].test_locations_event.locations))
    self.assertEqual(events[0].test_locations_event.locations[0].file,
                     'web/animations_api.cc')
    self.assertEqual(events[0].test_locations_event.locations[0].line, 123)

  def test_does_not_crash_on_missing_fields_when_uploading_test_locations(self):
    TestFile.add_file(
        'master', 'builder', 'ui_tests', 123, 'full_results.json', '{}')
    response = self.test_app.post(
        '/internal/monitoring/test_loc/upload', REQ_PAYLOAD)
    self.assertEqual(200, response.status_int)
    events = self.read_event_mon_file()
    self.assertEqual(1, len(events))
    self.assertFalse(
        events[0].test_locations_event.HasField('usec_since_epoch'))
    self.assertEqual(0, len(events[0].test_locations_event.locations))

  def test_handles_invalid_requests_for_uploading_test_locations(self):
    self.test_app.post(
        '/internal/monitoring/test_loc/upload', REQ_PAYLOAD, status=400)
    self.assertEqual(0, len(self.read_event_mon_file()))
