# -*- coding: utf-8 -*-
# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from datetime import datetime
import mock
import webapp2

from handlers.code_coverage import serve_ci_coverage
from handlers.code_coverage import utils
from model.code_coverage import CoverageReportModifier
from model.code_coverage import DependencyRepository
from model.code_coverage import FileCoverageData
from model.code_coverage import PostsubmitReport
from model.code_coverage import SummaryCoverageData
from waterfall.test.wf_testcase import WaterfallTestCase


def _CreateSampleCoverageSummaryMetric():
  """Returns a sample coverage summary metric for testing purpose.

  Note: only use this method if the exact values don't matter.
  """
  return [{
      'covered': 1,
      'total': 2,
      'name': 'region'
  }, {
      'covered': 1,
      'total': 2,
      'name': 'function'
  }, {
      'covered': 1,
      'total': 2,
      'name': 'line'
  }]


def _CreateSampleManifest():
  """Returns a sample manifest for testing purpose.

  Note: only use this method if the exact values don't matter.
  """
  return [
      DependencyRepository(
          path='//',
          server_host='chromium.googlesource.com',
          project='chromium/src.git',
          revision='ccccc')
  ]


def _CreateSamplePostsubmitReport(manifest=None,
                                  builder='linux-code-coverage',
                                  modifier_id=0):
  """Returns a sample PostsubmitReport for testing purpose.

  Note: only use this method if the exact values don't matter.
  """
  manifest = manifest or _CreateSampleManifest()
  return PostsubmitReport.Create(
      server_host='chromium.googlesource.com',
      project='chromium/src',
      ref='refs/heads/main',
      revision='aaaaa',
      bucket='coverage',
      builder=builder,
      commit_timestamp=datetime(2018, 1, 1),
      manifest=manifest,
      summary_metrics=_CreateSampleCoverageSummaryMetric(),
      build_id=123456789,
      modifier_id=modifier_id,
      visible=True)


def _CreateSampleDirectoryCoverageData(builder='linux-code-coverage',
                                       modifier_id=0):
  """Returns a sample directory SummaryCoverageData for testing purpose.

  Note: only use this method if the exact values don't matter.
  """
  return SummaryCoverageData.Create(
      server_host='chromium.googlesource.com',
      project='chromium/src',
      ref='refs/heads/main',
      revision='aaaaa',
      data_type='dirs',
      path='//dir/',
      bucket='coverage',
      builder=builder,
      modifier_id=modifier_id,
      data={
          'dirs': [],
          'path':
              '//dir/',
          'summaries':
              _CreateSampleCoverageSummaryMetric(),
          'files': [{
              'path': '//dir/myfile.cc',
              'name': 'myfile.cc',
              'summaries': _CreateSampleCoverageSummaryMetric()
          }]
      })


def _CreateSampleFileCoverageData(builder='linux-code-coverage', modifier_id=0):
  """Returns a sample FileCoverageData for testing purpose.

  Note: only use this method if the exact values don't matter.
  """
  return FileCoverageData.Create(
      server_host='chromium.googlesource.com',
      project='chromium/src',
      ref='refs/heads/main',
      revision='aaaaa',
      path='//dir/myfile.cc',
      bucket='coverage',
      builder=builder,
      modifier_id=modifier_id,
      data={
          'path': '//dir/myfile.cc',
          'revision': 'bbbbb',
          'lines': [{
              'count': 100,
              'last': 2,
              'first': 1
          }],
          'timestamp': '140000',
          'uncovered_blocks': [{
              'line': 1,
              'ranges': [{
                  'first': 1,
                  'last': 2
              }]
          }]
      })


def _CreateSampleComponentCoverageData(builder='linux-code-coverage'):
  """Returns a sample component SummaryCoverageData for testing purpose.

  Note: only use this method if the exact values don't matter.
  """
  return SummaryCoverageData.Create(
      server_host='chromium.googlesource.com',
      project='chromium/src',
      ref='refs/heads/main',
      revision='aaaaa',
      data_type='components',
      path='Component>Test',
      bucket='coverage',
      builder=builder,
      data={
          'dirs': [{
              'path': '//dir/',
              'name': 'dir/',
              'summaries': _CreateSampleCoverageSummaryMetric()
          }],
          'path': 'Component>Test',
          'summaries': _CreateSampleCoverageSummaryMetric()
      })


class ServeCodeCoverageDataTest(WaterfallTestCase):
  app_module = webapp2.WSGIApplication(
      [('/coverage/p/.*', serve_ci_coverage.ServeCodeCoverageData)], debug=True)

  def setUp(self):
    super(ServeCodeCoverageDataTest, self).setUp()
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_gitiles_configs': {
                'chromium.googlesource.com': {
                    'chromium/src': ['refs/heads/main',]
                }
            },
            'postsubmit_platform_info_map': {
                'chromium': {
                    'linux': {
                        'bucket': 'coverage',
                        'builder': 'linux-code-coverage',
                        'coverage_tool': 'clang',
                        'ui_name': 'Linux (C/C++)',
                    },
                },
            },
            'default_postsubmit_report_config': {
                'chromium': {
                    'host': 'chromium.googlesource.com',
                    'project': 'chromium/src',
                    'ref': 'refs/heads/main',
                    'platform': 'linux',
                },
            },
        })

  def tearDown(self):
    self.UpdateUnitTestConfigSettings('code_coverage_settings', {})
    super(ServeCodeCoverageDataTest, self).tearDown()

  def testServeFullRepoProjectView(self):
    self.mock_current_user(user_email='test@google.com', is_admin=False)

    host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    platform = 'linux'

    report = _CreateSamplePostsubmitReport()
    report.put()

    request_url = ('/coverage/p/chromium?host=%s&project=%s&ref=%s&platform=%s'
                   '&list_reports=true') % (host, project, ref, platform)
    response = self.test_app.get(request_url)
    self.assertEqual(200, response.status_int)

  def testServeFullRepoProjectView_WithModifier(self):
    self.mock_current_user(user_email='test@google.com', is_admin=False)

    host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    platform = 'linux'

    report = _CreateSamplePostsubmitReport(modifier_id=123)
    report.put()

    request_url = ('/coverage/p/chromium?host=%s&project=%s&ref=%s&platform=%s'
                   '&list_reports=true&modifier_id=%d') % (host, project, ref,
                                                           platform, 123)
    response = self.test_app.get(request_url)
    self.assertEqual(200, response.status_int)

  def testServeFullRepoProjectViewDefaultReportConfig(self):
    self.mock_current_user(user_email='test@google.com', is_admin=False)
    report = _CreateSamplePostsubmitReport()
    report.put()

    response = self.test_app.get('/coverage/p/chromium?&list_reports=true')
    self.assertEqual(200, response.status_int)

  def testServeFullRepoDirectoryView(self):
    self.mock_current_user(user_email='test@google.com', is_admin=False)

    host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    revision = 'aaaaa'
    path = '//dir/'
    platform = 'linux'

    report = _CreateSamplePostsubmitReport()
    report.put()

    dir_coverage_data = _CreateSampleDirectoryCoverageData()
    dir_coverage_data.put()

    request_url = (
        '/coverage/p/chromium/dir?host=%s&project=%s&ref=%s&revision=%s'
        '&path=%s&platform=%s') % (host, project, ref, revision, path, platform)
    response = self.test_app.get(request_url)
    self.assertEqual(200, response.status_int)

  def testServeFullRepoDirectoryView_WithModifierAndRevision(self):
    self.mock_current_user(user_email='test@google.com', is_admin=False)

    host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    revision = 'aaaaa'
    path = '//dir/'
    platform = 'linux'

    report = _CreateSamplePostsubmitReport(modifier_id=123)
    report.put()

    dir_coverage_data = _CreateSampleDirectoryCoverageData(modifier_id=123)
    dir_coverage_data.put()

    request_url = (
        '/coverage/p/chromium/dir?host=%s&project=%s&ref=%s&revision=%s'
        '&path=%s&platform=%s&modifier_id=%d') % (host, project, ref, revision,
                                                  path, platform, 123)
    response = self.test_app.get(request_url)
    self.assertEqual(200, response.status_int)

  def testServeFullRepoDirectoryView_WithModifier_WithoutRevision(self):
    self.mock_current_user(user_email='test@google.com', is_admin=False)

    host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    path = '//dir/'
    platform = 'linux'

    report = _CreateSamplePostsubmitReport(modifier_id=123)
    report.put()

    dir_coverage_data = _CreateSampleDirectoryCoverageData(modifier_id=123)
    dir_coverage_data.put()

    request_url = ('/coverage/p/chromium/dir?host=%s&project=%s&ref=%s'
                   '&path=%s&platform=%s&modifier_id=%d') % (
                       host, project, ref, path, platform, 123)
    response = self.test_app.get(request_url)
    self.assertEqual(200, response.status_int)

  def testServeFullRepoComponentView(self):
    self.mock_current_user(user_email='test@google.com', is_admin=False)

    host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    revision = 'aaaaa'
    path = 'Component>Test'
    platform = 'linux'

    report = _CreateSamplePostsubmitReport()
    report.put()

    component_coverage_data = _CreateSampleComponentCoverageData()
    component_coverage_data.put()

    request_url = ('/coverage/p/chromium/component?host=%s&project=%s&ref=%s'
                   '&revision=%s&path=%s&platform=%s') % (
                       host, project, ref, revision, path, platform)
    response = self.test_app.get(request_url)
    self.assertEqual(200, response.status_int)

  def testServeFullRepo_UnitTestsOnly(self):
    self.mock_current_user(user_email='test@google.com', is_admin=False)

    host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    revision = 'aaaaa'
    path = '//dir/'
    platform = 'linux'

    report = _CreateSamplePostsubmitReport(builder='linux-code-coverage_unit')
    report.put()

    dir_coverage_data = _CreateSampleDirectoryCoverageData(
        builder='linux-code-coverage_unit')
    dir_coverage_data.put()

    request_url = (
        '/coverage/p/chromium/dir?host=%s&project=%s&ref=%s&revision=%s'
        '&path=%s&platform=%s&test_suite_type=unit') % (
            host, project, ref, revision, path, platform)
    response = self.test_app.get(request_url)
    self.assertEqual(200, response.status_int)

  @mock.patch.object(utils, 'GetFileContentFromGs')
  def testServeFullRepoFileView(self, mock_get_file_from_gs):
    self.mock_current_user(user_email='test@google.com', is_admin=False)
    mock_get_file_from_gs.return_value = 'line one/nline two'

    host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    revision = 'aaaaa'
    path = '//dir/myfile.cc'
    platform = 'linux'

    report = _CreateSamplePostsubmitReport()
    report.put()

    file_coverage_data = _CreateSampleFileCoverageData()
    file_coverage_data.put()

    request_url = ('/coverage/p/chromium/file?host=%s&project=%s&ref=%s'
                   '&revision=%s&path=%s&platform=%s') % (
                       host, project, ref, revision, path, platform)
    response = self.test_app.get(request_url)
    self.assertEqual(200, response.status_int)
    mock_get_file_from_gs.assert_called_with(
        '/source-files-for-coverage/chromium.googlesource.com/chromium/'
        'src.git/dir/myfile.cc/bbbbb')

  @mock.patch.object(utils, 'GetFileContentFromGs')
  def testServeFullRepoFileView_WithModifierAndRevision(self,
                                                        mock_get_file_from_gs):
    self.mock_current_user(user_email='test@google.com', is_admin=False)
    mock_get_file_from_gs.return_value = 'line one/nline two'

    host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    revision = 'aaaaa'
    path = '//dir/myfile.cc'
    platform = 'linux'

    report = _CreateSamplePostsubmitReport(modifier_id=123)
    report.put()

    file_coverage_data = _CreateSampleFileCoverageData(modifier_id=123)
    file_coverage_data.put()

    request_url = ('/coverage/p/chromium/file?host=%s&project=%s&ref=%s'
                   '&revision=%s&path=%s&platform=%s&modifier_id=%d') % (
                       host, project, ref, revision, path, platform, 123)
    response = self.test_app.get(request_url)
    self.assertEqual(200, response.status_int)
    mock_get_file_from_gs.assert_called_with(
        '/source-files-for-coverage/chromium.googlesource.com/chromium/'
        'src.git/dir/myfile.cc/bbbbb')

  @mock.patch.object(utils, 'GetFileContentFromGs')
  def testServeFullRepoFileView_WithModifier_WithoutRevision(
      self, mock_get_file_from_gs):
    self.mock_current_user(user_email='test@google.com', is_admin=False)
    mock_get_file_from_gs.return_value = 'line one/nline two'

    host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    path = '//dir/myfile.cc'
    platform = 'linux'

    report = _CreateSamplePostsubmitReport(modifier_id=123)
    report.put()

    file_coverage_data = _CreateSampleFileCoverageData(modifier_id=123)
    file_coverage_data.put()

    request_url = ('/coverage/p/chromium/file?host=%s&project=%s&ref=%s'
                   '&path=%s&platform=%s&modifier_id=%d') % (
                       host, project, ref, path, platform, 123)
    response = self.test_app.get(request_url)
    self.assertEqual(200, response.status_int)
    mock_get_file_from_gs.assert_called_with(
        '/source-files-for-coverage/chromium.googlesource.com/chromium/'
        'src.git/dir/myfile.cc/bbbbb')

  @mock.patch.object(utils, 'GetFileContentFromGs')
  def testServeFullRepoFileViewWithNonAsciiChars(self, mock_get_file_from_gs):
    self.mock_current_user(user_email='test@google.com', is_admin=False)
    mock_get_file_from_gs.return_value = 'line one\n═══════════╪'
    report = _CreateSamplePostsubmitReport()
    report.put()

    file_coverage_data = _CreateSampleFileCoverageData()
    file_coverage_data.put()

    request_url = ('/coverage/p/chromium/file?host=%s&project=%s&ref=%s'
                   '&revision=%s&path=%s&platform=%s') % (
                       'chromium.googlesource.com', 'chromium/src',
                       'refs/heads/main', 'aaaaa', '//dir/myfile.cc', 'linux')
    response = self.test_app.get(request_url)
    self.assertEqual(200, response.status_int)


class SplitLineIntoRegionsTest(WaterfallTestCase):

  def testRejoinSplitRegions(self):
    line = 'the quick brown fox jumped over the lazy dog'
    blocks = [{
        'first': 4,
        'last': 10,
    }, {
        'first': 20,
        'last': 23,
    }, {
        'first': 42,
        'last': 43,
    }]
    regions = serve_ci_coverage._SplitLineIntoRegions(line, blocks)
    reconstructed_line = ''.join(region['text'] for region in regions)
    self.assertEqual(line, reconstructed_line)

  def testRegionsCorrectlySplit(self):
    line = 'onetwothreefourfivesixseven'
    blocks = [{
        'first': 4,
        'last': 6,
    }, {
        'first': 12,
        'last': 15,
    }, {
        'first': 20,
        'last': 22,
    }]
    regions = serve_ci_coverage._SplitLineIntoRegions(line, blocks)

    self.assertEqual('one', regions[0]['text'])
    self.assertEqual('two', regions[1]['text'])
    self.assertEqual('three', regions[2]['text'])
    self.assertEqual('four', regions[3]['text'])
    self.assertEqual('five', regions[4]['text'])
    self.assertEqual('six', regions[5]['text'])
    self.assertEqual('seven', regions[6]['text'])

    # Regions should alternate between covered and uncovered.
    self.assertTrue(regions[0]['is_covered'])
    self.assertTrue(regions[2]['is_covered'])
    self.assertTrue(regions[4]['is_covered'])
    self.assertTrue(regions[6]['is_covered'])
    self.assertFalse(regions[1]['is_covered'])
    self.assertFalse(regions[3]['is_covered'])
    self.assertFalse(regions[5]['is_covered'])

  def testPrefixUncovered(self):
    line = 'NOCOVcov'
    blocks = [{'first': 1, 'last': 5}]
    regions = serve_ci_coverage._SplitLineIntoRegions(line, blocks)
    self.assertEqual('NOCOV', regions[0]['text'])
    self.assertEqual('cov', regions[1]['text'])
    self.assertFalse(regions[0]['is_covered'])
    self.assertTrue(regions[1]['is_covered'])

  def testSuffixUncovered(self):
    line = 'covNOCOV'
    blocks = [{'first': 4, 'last': 8}]
    regions = serve_ci_coverage._SplitLineIntoRegions(line, blocks)
    self.assertEqual('cov', regions[0]['text'])
    self.assertEqual('NOCOV', regions[1]['text'])
    self.assertTrue(regions[0]['is_covered'])
    self.assertFalse(regions[1]['is_covered'])
