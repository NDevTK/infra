# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import json
import requests
import tempfile
import unittest

import mock

from testing_support import auto_stub

from infra_libs.ts_mon import config
from infra_libs.ts_mon.common import interface
from infra_libs.ts_mon.common import standard_metrics
from infra_libs.ts_mon.common.test import stubs


class GlobalsTest(auto_stub.TestCase):

  def setUp(self):
    super(GlobalsTest, self).setUp()
    self.mock(interface, 'state', stubs.MockState())
    self.mock(config, 'load_machine_config', lambda x: {})

  def tearDown(self):
    # It's important to call close() before un-setting the mock state object,
    # because any FlushThread started by the test is stored in that mock state
    # and needs to be stopped before running any other tests.
    interface.close()
    super(GlobalsTest, self).tearDown()

  @mock.patch('requests.get', autospec=True)
  @mock.patch('socket.getfqdn', autospec=True)
  @mock.patch('infra_libs.ts_mon.api_monitor.ApiMonitor', autospec=True)
  @mock.patch('infra_libs.ts_mon.common.targets.DeviceTarget', autospec=True)
  def test_default_monitor_args(self, fake_target, fake_monitor, fake_fqdn,
                                fake_get):
    singleton = mock.Mock()
    fake_monitor.return_value = singleton
    fake_target.return_value = singleton
    fake_fqdn.return_value = 'slave1-a1.reg.tld'
    fake_get.return_value.side_effect = requests.exceptions.ConnectionError
    p = argparse.ArgumentParser()
    config.add_argparse_options(p)
    args = p.parse_args([
        '--ts-mon-credentials', '/path/to/creds.p8.json',
        '--ts-mon-endpoint',
        'https://www.googleapis.com/acquisitions/v1_mon_shared/storage'])
    config.process_argparse_options(args)
    fake_monitor.assert_called_once_with(
        '/path/to/creds.p8.json',
        'https://www.googleapis.com/acquisitions/v1_mon_shared/storage',
        use_instrumented_http=True)
    self.assertIs(interface.state.global_monitor, singleton)
    fake_target.assert_called_once_with('reg', 'default', '1', 'slave1-a1')
    self.assertIs(interface.state.target, singleton)
    self.assertEquals(args.ts_mon_flush, 'auto')
    self.assertIsNotNone(interface.state.flush_thread)
    self.assertTrue(standard_metrics.up.get())

  @mock.patch('requests.get', autospec=True)
  @mock.patch('socket.getfqdn', autospec=True)
  @mock.patch('infra_libs.ts_mon.api_monitor.ApiMonitor', autospec=True)
  @mock.patch('infra_libs.ts_mon.common.targets.DeviceTarget', autospec=True)
  def test_default_monitor_args_uppercase_fqdn(self, fake_target, fake_monitor,
                                               fake_fqdn, fake_get):
    singleton = mock.Mock()
    fake_monitor.return_value = singleton
    fake_target.return_value = singleton
    fake_fqdn.return_value = 'SLAVE1-A1.REG.TLD'
    fake_get.return_value.side_effect = requests.exceptions.ConnectionError
    p = argparse.ArgumentParser()
    config.add_argparse_options(p)
    args = p.parse_args([
        '--ts-mon-credentials', '/path/to/creds.p8.json',
        '--ts-mon-endpoint',
        'https://www.googleapis.com/acquisitions/v1_mon_shared/storage'])
    config.process_argparse_options(args)
    fake_target.assert_called_once_with('reg', 'default', '1', 'slave1-a1')

  @mock.patch('requests.get', autospec=True)
  @mock.patch('socket.getfqdn', autospec=True)
  @mock.patch('infra_libs.ts_mon.api_monitor.ApiMonitor', autospec=True)
  @mock.patch('infra_libs.ts_mon.common.targets.DeviceTarget', autospec=True)
  def test_fallback_monitor_args(self, fake_target, fake_monitor, fake_fqdn,
                                 fake_get):
    singleton = mock.Mock()
    fake_monitor.return_value = singleton
    fake_target.return_value = singleton
    fake_fqdn.return_value = 'foo'
    fake_get.return_value.side_effect = requests.exceptions.ConnectionError
    p = argparse.ArgumentParser()
    config.add_argparse_options(p)
    args = p.parse_args([
        '--ts-mon-credentials', '/path/to/creds.p8.json',
        '--ts-mon-endpoint',
        'https://www.googleapis.com/acquisitions/v1_mon_shared/storage'])
    config.process_argparse_options(args)
    fake_monitor.assert_called_once_with(
        '/path/to/creds.p8.json',
        'https://www.googleapis.com/acquisitions/v1_mon_shared/storage',
        use_instrumented_http=True)
    self.assertIs(interface.state.global_monitor, singleton)
    fake_target.assert_called_once_with('', 'default', '', 'foo')
    self.assertIs(interface.state.target, singleton)

  @mock.patch('socket.getfqdn', autospec=True)
  @mock.patch('infra_libs.ts_mon.api_monitor.ApiMonitor', autospec=True)
  @mock.patch('infra_libs.ts_mon.common.targets.DeviceTarget', autospec=True)
  def test_manual_flush(self, fake_target, fake_monitor, fake_fqdn):
    singleton = mock.Mock()
    fake_monitor.return_value = singleton
    fake_target.return_value = singleton
    fake_fqdn.return_value = 'foo'
    p = argparse.ArgumentParser()
    config.add_argparse_options(p)
    args = p.parse_args(['--ts-mon-flush', 'manual'])
    config.process_argparse_options(args)
    self.assertIsNone(interface.state.flush_thread)

  @mock.patch('infra_libs.ts_mon.api_monitor.ApiMonitor', autospec=True)
  def test_monitor_args(self, fake_monitor):
    singleton = mock.Mock()
    fake_monitor.return_value = singleton
    p = argparse.ArgumentParser()
    config.add_argparse_options(p)
    args = p.parse_args(['--ts-mon-credentials', '/path/to/creds.p8.json',
                         '--ts-mon-endpoint', 'https://foo.tld/api'])
    config.process_argparse_options(args)
    fake_monitor.assert_called_once_with(
        '/path/to/creds.p8.json', 'https://foo.tld/api',
        use_instrumented_http=True)
    self.assertIs(interface.state.global_monitor, singleton)

  @mock.patch('infra_libs.ts_mon.common.monitors.PubSubMonitor', autospec=True)
  def test_pubsub_args(self, fake_monitor):
    singleton = mock.Mock()
    fake_monitor.return_value = singleton
    p = argparse.ArgumentParser()
    config.add_argparse_options(p)
    args = p.parse_args(['--ts-mon-credentials', '/path/to/creds.p8.json',
                         '--ts-mon-endpoint', 'pubsub://mytopic/myproject'])
    config.process_argparse_options(args)
    fake_monitor.assert_called_once_with(
        '/path/to/creds.p8.json', 'mytopic', 'myproject',
        use_instrumented_http=True)
    self.assertIs(interface.state.global_monitor, singleton)

  @mock.patch('infra_libs.ts_mon.common.monitors.DebugMonitor', auto_spec=True)
  def test_dryrun_args(self, fake_monitor):
    singleton = mock.Mock()
    fake_monitor.return_value = singleton
    p = argparse.ArgumentParser()
    config.add_argparse_options(p)
    args = p.parse_args(['--ts-mon-endpoint', 'file://foo.txt'])
    config.process_argparse_options(args)
    fake_monitor.assert_called_once_with('foo.txt')
    self.assertIs(interface.state.global_monitor, singleton)

  @mock.patch('infra_libs.ts_mon.api_monitor.ApiMonitor', autospec=True)
  @mock.patch('infra_libs.ts_mon.common.targets.DeviceTarget', autospec=True)
  def test_device_args(self, fake_target, _fake_monitor):
    singleton = mock.Mock()
    fake_target.return_value = singleton
    p = argparse.ArgumentParser()
    config.add_argparse_options(p)
    args = p.parse_args(['--ts-mon-credentials', '/path/to/creds.p8.json',
                         '--ts-mon-target-type', 'device',
                         '--ts-mon-device-region', 'reg',
                         '--ts-mon-device-role', 'role',
                         '--ts-mon-device-network', 'net',
                         '--ts-mon-device-hostname', 'host'])
    config.process_argparse_options(args)
    fake_target.assert_called_once_with('reg', 'role', 'net', 'host')
    self.assertIs(interface.state.target, singleton)

  @mock.patch('infra_libs.ts_mon.api_monitor.ApiMonitor', autospec=True)
  @mock.patch('infra_libs.ts_mon.common.targets.TaskTarget', autospec=True)
  def test_task_args(self, fake_target, _fake_monitor):
    singleton = mock.Mock()
    fake_target.return_value = singleton
    p = argparse.ArgumentParser()
    config.add_argparse_options(p)
    args = p.parse_args(['--ts-mon-credentials', '/path/to/creds.p8.json',
                         '--ts-mon-target-type', 'task',
                         '--ts-mon-task-service-name', 'serv',
                         '--ts-mon-task-job-name', 'job',
                         '--ts-mon-task-region', 'reg',
                         '--ts-mon-task-hostname', 'host',
                         '--ts-mon-task-number', '1'])
    config.process_argparse_options(args)
    fake_target.assert_called_once_with('serv', 'job' ,'reg', 'host', 1)
    self.assertIs(interface.state.target, singleton)

  @mock.patch('infra_libs.ts_mon.common.monitors.NullMonitor', autospec=True)
  def test_no_args(self, fake_monitor):
    singleton = mock.Mock()
    fake_monitor.return_value = singleton
    p = argparse.ArgumentParser()
    config.add_argparse_options(p)
    args = p.parse_args([])
    config.process_argparse_options(args)
    self.assertEqual(1, len(fake_monitor.mock_calls))
    self.assertIs(interface.state.global_monitor, singleton)

  @mock.patch('requests.get', autospec=True)
  def test_gce_region(self, mock_get):
    r = mock_get.return_value
    r.status_code = 200
    r.text = 'projects/182615506979/zones/us-central1-f'

    self.assertEquals('us-central1-f', config._default_region('foo.golo'))

  @mock.patch('requests.get', autospec=True)
  def test_gce_region_timeout(self, mock_get):
    mock_get.side_effect = requests.exceptions.Timeout
    self.assertEquals('golo', config._default_region('foo.golo'))

  @mock.patch('requests.get', autospec=True)
  def test_gce_region_404(self, mock_get):
    r = mock_get.return_value
    r.status_code = 404

    self.assertEquals('golo', config._default_region('foo.golo'))


class ConfigTest(unittest.TestCase):

  def test_load_machine_config(self):
    with tempfile.NamedTemporaryFile() as fh:
      json.dump({'foo': 'bar'}, fh)
      fh.flush()
      self.assertEquals({'foo': 'bar'}, config.load_machine_config(fh.name))

  def test_load_machine_config_bad(self):
    with tempfile.NamedTemporaryFile() as fh:
      fh.write('not a json file')
      fh.flush()
      with self.assertRaises(ValueError):
        config.load_machine_config(fh.name)

  def test_load_machine_config_not_exists(self):
    self.assertEquals({}, config.load_machine_config('does not exist'))
