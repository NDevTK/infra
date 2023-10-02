# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from datetime import datetime
import json
import mock
import webapp2

from components.prpc import client as prpc_client

from go.chromium.org.luci.buildbucket.proto import build_pb2
from go.chromium.org.luci.buildbucket.proto import builder_common_pb2
from go.chromium.org.luci.buildbucket.proto import builds_service_pb2
from go.chromium.org.luci.buildbucket.proto import common_pb2

from gae_libs.handlers.base_handler import BaseHandler
from handlers.code_coverage import process_coverage
from handlers.code_coverage import utils
from model.code_coverage import BlockingStatus
from model.code_coverage import CoveragePercentage
from model.code_coverage import DependencyRepository
from model.code_coverage import FileCoverageData
from model.code_coverage import LowCoverageBlocking
from model.code_coverage import PostsubmitReport
from model.code_coverage import PresubmitCoverageData
from model.code_coverage import SummaryCoverageData
from services.code_coverage import code_coverage_util
from waterfall.test.wf_testcase import WaterfallTestCase


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


def _CreateSampleFileCoverageData(builder='linux-code-coverage', modifier_id=0):
  """Returns a sample FileCoverageData for testing purpose.

  Note: only use this method if the exact values don't matter.
  """
  return FileCoverageData.Create(
      server_host='chromium.googlesource.com',
      project='chromium/src',
      ref='refs/heads/main',
      revision='aaaaa',
      path='//dir/test.cc',
      bucket='coverage',
      builder=builder,
      modifier_id=modifier_id,
      data={
          'path': '//dir/test.cc',
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


def _CreateSampleRootComponentCoverageData(builder='linux-code-coverage'):
  """Returns a sample component SummaryCoverageData for >> for testing purpose.

  Note: only use this method if the exact values don't matter.
  """
  return SummaryCoverageData.Create(
      server_host='chromium.googlesource.com',
      project='chromium/src',
      ref='refs/heads/main',
      revision='aaaaa',
      data_type='components',
      path='>>',
      bucket='coverage',
      builder=builder,
      data={
          'dirs': [{
              'path': 'Component>Test',
              'name': 'Component>Test',
              'summaries': _CreateSampleCoverageSummaryMetric()
          }],
          'path': '>>'
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
              'path': '//dir/test.cc',
              'name': 'test.cc',
              'summaries': _CreateSampleCoverageSummaryMetric()
          }]
      })


class ProcessCodeCoverageDataTest(WaterfallTestCase):
  app_module = webapp2.WSGIApplication([
      ('/coverage/task/process-data/.*',
       process_coverage.ProcessCodeCoverageData),
  ],
                                       debug=True)

  def setUp(self):
    super(ProcessCodeCoverageDataTest, self).setUp()
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': [
                'chromium/try/linux-rel',
                'chrome/coverage/linux-code-coverage',
            ]
        })

  def tearDown(self):
    self.UpdateUnitTestConfigSettings('code_coverage_settings', {})
    super(ProcessCodeCoverageDataTest, self).tearDown()

  def testPermissionInProcessCodeCoverageData(self):
    self.mock_current_user(user_email='test@google.com', is_admin=False)
    response = self.test_app.post(
        '/coverage/task/process-data/123?format=json', status=401)
    self.assertEqual(('Either not log in yet or no permission. '
                      'Please log in with your @google.com account.'),
                     response.json_body.get('error_message'))

  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  def testProcessCLPatchData(self, mocked_is_request_from_appself,
                             mocked_get_build, mocked_get_validated_data,
                             mocked_inc_percentages, *_):
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'linux-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'linux-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['linux-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build

    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/test.cc',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 1,
            }, {
                'count': 0,
                'first': 2,
                'last': 2,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/test.cc', total_lines=1, covered_lines=1)
    ]
    mocked_inc_percentages.return_value = inc_percentages

    request_url = '/coverage/task/process-data/build/123456789'
    response = self.test_app.post(request_url)
    self.assertEqual(200, response.status_int)
    mocked_is_request_from_appself.assert_called()

    mocked_get_validated_data.assert_called_with(
        '/code-coverage-data/presubmit/chromium-review.googlesource.com/138000/'
        '4/try/linux-rel/123456789/metadata/all.json.gz')

    expected_entity = PresubmitCoverageData.Create(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4,
        data=coverage_data['files'])
    expected_entity.absolute_percentages = [
        CoveragePercentage(
            path='//dir/test.cc', total_lines=2, covered_lines=1)
    ]
    expected_entity.incremental_percentages = inc_percentages
    expected_entity.insert_timestamp = datetime.now()
    expected_entity.update_timestamp = datetime.now()
    fetched_entities = PresubmitCoverageData.query().fetch()

    self.assertEqual(1, len(fetched_entities))
    self.assertEqual(expected_entity.cl_patchset,
                     fetched_entities[0].cl_patchset)
    self.assertEqual(expected_entity.data, fetched_entities[0].data)
    self.assertEqual(expected_entity.absolute_percentages,
                     fetched_entities[0].absolute_percentages)
    self.assertEqual(expected_entity.incremental_percentages,
                     fetched_entities[0].incremental_percentages)
    self.assertEqual(expected_entity.based_on, fetched_entities[0].based_on)

  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  def testProcessCLPatchDataUnitTestBuilder(self,
                                            mocked_is_request_from_appself,
                                            mocked_get_build,
                                            mocked_get_validated_data,
                                            mocked_inc_percentages, *_):
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'linux-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'linux-rel_unit/123456789/metadata'
        ]), ('mimic_builder_names', ['linux-rel_unit'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build

    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/test.cc',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 1,
            }, {
                'count': 0,
                'first': 2,
                'last': 2,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/test.cc', total_lines=1, covered_lines=1)
    ]
    mocked_inc_percentages.return_value = inc_percentages

    request_url = '/coverage/task/process-data/build/123456789'
    response = self.test_app.post(request_url)
    self.assertEqual(200, response.status_int)
    mocked_is_request_from_appself.assert_called()

    mocked_get_validated_data.assert_called_with(
        '/code-coverage-data/presubmit/chromium-review.googlesource.com/138000/'
        '4/try/linux-rel_unit/123456789/metadata/all.json.gz')

    expected_entity = PresubmitCoverageData.Create(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4,
        data_unit=coverage_data['files'])
    expected_entity.absolute_percentages_unit = [
        CoveragePercentage(
            path='//dir/test.cc', total_lines=2, covered_lines=1)
    ]
    expected_entity.incremental_percentages_unit = inc_percentages
    expected_entity.insert_timestamp = datetime.now()
    expected_entity.update_timestamp = datetime.now()
    fetched_entities = PresubmitCoverageData.query().fetch()

    self.assertEqual(1, len(fetched_entities))
    self.assertEqual(expected_entity.cl_patchset,
                     fetched_entities[0].cl_patchset)
    self.assertEqual(expected_entity.data_unit, fetched_entities[0].data_unit)
    self.assertEqual(expected_entity.absolute_percentages_unit,
                     fetched_entities[0].absolute_percentages_unit)
    self.assertEqual(expected_entity.incremental_percentages_unit,
                     fetched_entities[0].incremental_percentages_unit)
    self.assertEqual(expected_entity.based_on, fetched_entities[0].based_on)

  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  def testProcessCLPatchDataMergingData(self, mocked_get_build,
                                        mocked_get_validated_data,
                                        mocked_inc_percentages, *_):
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'linux-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'linux-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['linux-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build

    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path': '//dir/test.cc',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 1,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    mocked_inc_percentages.return_value = []

    existing_entity = PresubmitCoverageData.Create(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4,
        data=[{
            'path': '//dir/test.cc',
            'lines': [{
                'count': 100,
                'first': 2,
                'last': 2,
            }],
        }])
    existing_entity.put()

    request_url = '/coverage/task/process-data/build/123456789'
    response = self.test_app.post(request_url)
    self.assertEqual(200, response.status_int)

    expected_entity = PresubmitCoverageData.Create(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4,
        data=[{
            'path': '//dir/test.cc',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 2,
            }],
        }])
    expected_entity.absolute_percentages = [
        CoveragePercentage(
            covered_lines=2, path=u'//dir/test.cc', total_lines=2)
    ]
    expected_entity.incremental_percentages = []
    fetched_entities = PresubmitCoverageData.query().fetch()

    self.assertEqual(1, len(fetched_entities))
    self.assertEqual(expected_entity.cl_patchset,
                     fetched_entities[0].cl_patchset)
    self.assertEqual(expected_entity.data, fetched_entities[0].data)
    self.assertEqual(expected_entity.absolute_percentages,
                     fetched_entities[0].absolute_percentages)
    self.assertEqual(expected_entity.incremental_percentages,
                     fetched_entities[0].incremental_percentages)
    self.assertEqual(expected_entity.based_on, fetched_entities[0].based_on)

  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  def testProcessCLPatchDataRTSBuilder_NewEntity(self,
                                                 mocked_is_request_from_appself,
                                                 mocked_get_build,
                                                 mocked_get_validated_data,
                                                 mocked_inc_percentages, *_):
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'linux-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'linux-rel_unit/123456789/metadata'
        ]), ('mimic_builder_names', ['linux-rel_unit']), ('rts_was_used', True)
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build

    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/test.cc',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 1,
            }, {
                'count': 0,
                'first': 2,
                'last': 2,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/test.cc', total_lines=1, covered_lines=1)
    ]
    mocked_inc_percentages.return_value = inc_percentages

    request_url = '/coverage/task/process-data/build/123456789'
    response = self.test_app.post(request_url)
    self.assertEqual(200, response.status_int)
    mocked_is_request_from_appself.assert_called()

    mocked_get_validated_data.assert_called_with(
        '/code-coverage-data/presubmit/chromium-review.googlesource.com/138000/'
        '4/try/linux-rel_unit/123456789/metadata/all.json.gz')

    expected_entity = PresubmitCoverageData.Create(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4,
        data_unit=coverage_data['files'],
        data_unit_rts=coverage_data['files'])
    expected_entity.absolute_percentages_unit_rts = [
        CoveragePercentage(
            path='//dir/test.cc', total_lines=2, covered_lines=1)
    ]
    expected_entity.insert_timestamp = datetime.now()
    expected_entity.update_timestamp = datetime.now()
    fetched_entities = PresubmitCoverageData.query().fetch()

    self.assertEqual(1, len(fetched_entities))
    self.assertEqual(expected_entity.cl_patchset,
                     fetched_entities[0].cl_patchset)
    self.assertEqual(expected_entity.data_unit_rts,
                     fetched_entities[0].data_unit_rts)
    self.assertEqual(expected_entity.absolute_percentages_unit_rts,
                     fetched_entities[0].absolute_percentages_unit_rts)
    self.assertEqual(expected_entity.based_on, fetched_entities[0].based_on)

  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  def testProcessCLPatchDataRTSBuilder_MergeDataIntoExistingEntity(
      self, mocked_is_request_from_appself, mocked_get_build,
      mocked_get_validated_data, mocked_inc_percentages, *_):
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'linux-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'linux-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['linux-rel']), ('rts_was_used', True)
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build

    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/test.cc',
            'lines': [{
                'count': 0,
                'first': 1,
                'last': 1,
            }, {
                'count': 100,
                'first': 2,
                'last': 2,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/test.cc', total_lines=1, covered_lines=1)
    ]
    mocked_inc_percentages.return_value = inc_percentages

    existing_file_coverage = [{
        'path': '//dir/test.cc',
        'lines': [{
            'count': 100,
            'first': 3,
            'last': 3,
        }],
    }]
    existing_entity = PresubmitCoverageData.Create(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4,
        data=existing_file_coverage,
        data_rts=existing_file_coverage)
    existing_entity.absolute_percentages_rts = [
        CoveragePercentage(
            path='//dir/test.cc', total_lines=1, covered_lines=1)
    ]
    existing_entity.put()

    request_url = '/coverage/task/process-data/build/123456789'
    response = self.test_app.post(request_url)
    self.assertEqual(200, response.status_int)
    mocked_is_request_from_appself.assert_called()

    mocked_get_validated_data.assert_called_with(
        '/code-coverage-data/presubmit/chromium-review.googlesource.com/138000/'
        '4/try/linux-rel/123456789/metadata/all.json.gz')

    expected_file_coverage = [{
        'path':
            '//dir/test.cc',
        'lines': [{
            'count': 0,
            'first': 1,
            'last': 1,
        }, {
            'count': 100,
            'first': 2,
            'last': 3,
        }],
    }]
    expected_entity = PresubmitCoverageData.Create(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4,
        data=expected_file_coverage,
        data_rts=expected_file_coverage)
    expected_entity.absolute_percentages_rts = [
        CoveragePercentage(
            path='//dir/test.cc', total_lines=3, covered_lines=2)
    ]
    expected_entity.insert_timestamp = datetime.now()
    expected_entity.update_timestamp = datetime.now()
    fetched_entities = PresubmitCoverageData.query().fetch()

    self.assertEqual(1, len(fetched_entities))
    self.assertEqual(expected_entity.cl_patchset,
                     fetched_entities[0].cl_patchset)
    self.assertEqual(expected_entity.data_rts, fetched_entities[0].data_rts)
    self.assertEqual(expected_entity.absolute_percentages_rts,
                     fetched_entities[0].absolute_percentages_rts)
    self.assertEqual(expected_entity.based_on, fetched_entities[0].based_on)

  # This test case tests the scenario where multiple coverage builders
  # were triggered for a CL, and the first coverage build is being processed.
  # In this case, the said coverage build produced coverage data.
  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(process_coverage.ProcessCodeCoverageData,
                     '_MayBeBlockCLForLowCoverage')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  # pylint: disable=line-too-long
  def testLowCoverageBlocking_firstCoverageBuildWithData_dontCheckForLowCoverage(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mock_blocking_logic, mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': [
                'chromium/try/android-nougat-x86-rel',
                'chromium/try/android-pie-x86-rel',
            ],
            'block_low_coverage_changes_projects': ['chromium/src'],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir'],
                }
            },
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/myfile.java',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 0,
                'first': 11,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfile.java', total_lines=90, covered_lines=9)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # Two coverage builds were triggered for the CL
    # and both completed without error
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS),
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-pie-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    blocking_entity = LowCoverageBlocking.Get(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4)
    self.assertEqual(blocking_entity.blocking_status, BlockingStatus.DEFAULT)
    self.assertEqual(
        set(blocking_entity.expected_builders),
        set(['android-nougat-x86-rel', 'android-pie-x86-rel']))
    self.assertEqual(
        set(blocking_entity.successful_builders),
        set(['android-nougat-x86-rel', 'android-pie-x86-rel']))
    self.assertEqual(
        set(blocking_entity.processed_builders),
        set(['android-nougat-x86-rel']))
    # Assert that CL was not checked for low coverage
    self.assertEqual(len(mock_blocking_logic.call_args_list), 0)

  # This test case tests the scenario where multiple coverage builders
  # were triggered for a CL, and the first coverage build is being processed.
  # In this case, the said coverage build produced no data.
  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(process_coverage.ProcessCodeCoverageData,
                     '_MayBeBlockCLForLowCoverage')
  @mock.patch.object(process_coverage, 'GetV2Build')
  # pylint: disable=line-too-long
  def testLowCoverageBlocking_firstCoverageBuildWithNoData_dontCheckForLowCoverage(
      self, mocked_get_build, mock_blocking_logic, mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': [
                'chromium/try/android-nougat-x86-rel',
                'chromium/try/android-pie-x86-rel',
            ],
            'block_low_coverage_changes_projects': ['chromium/src'],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir'],
                }
            },
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        # coverage_gs_bucket, coverage_metadata_gs_paths and
        # mimic_builder_names properties are missing indicating
        # this build did not produce any coverage data
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build

    # Two coverage builds were triggered for the CL
    # and both completed without error
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS),
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-pie-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    blocking_entity = LowCoverageBlocking.Get(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4)
    self.assertEqual(blocking_entity.blocking_status, BlockingStatus.DEFAULT)
    self.assertEqual(
        set(blocking_entity.expected_builders),
        set(['android-nougat-x86-rel', 'android-pie-x86-rel']))
    self.assertEqual(
        set(blocking_entity.successful_builders),
        set(['android-nougat-x86-rel', 'android-pie-x86-rel']))
    self.assertEqual(
        set(blocking_entity.processed_builders),
        set(['android-nougat-x86-rel']))
    # Assert that CL was not checked for low coverage
    self.assertEqual(len(mock_blocking_logic.call_args_list), 0)

  # This test case tests the scenario where multiple coverage builders
  # were triggered for a CL, and one of them failed to complete.
  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(process_coverage.ProcessCodeCoverageData,
                     '_MayBeBlockCLForLowCoverage')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  # pylint: disable=line-too-long
  def testLowCoverageBlocking_buildFailure_dontCheckForLowCoverage(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mock_blocking_logic, mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': [
                'chromium/try/android-nougat-x86-rel',
                'chromium/try/android-pie-x86-rel',
            ],
            'block_low_coverage_changes_projects': ['chromium/src'],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir'],
                }
            },
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/myfile.java',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 0,
                'first': 11,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfile.java', total_lines=90, covered_lines=9)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # Two coverage builds were triggered for the CL
    # and both completed without error
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS),
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-pie-x86-rel'),
                status=common_pb2.Status.FAILURE)
        ]))

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    blocking_entity = LowCoverageBlocking.Get(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4)
    self.assertEqual(blocking_entity.blocking_status,
                     BlockingStatus.DONT_BLOCK_BUILDER_FAILURE)
    self.assertEqual(
        set(blocking_entity.expected_builders),
        set(['android-nougat-x86-rel', 'android-pie-x86-rel']))
    self.assertEqual(
        set(blocking_entity.successful_builders),
        set(['android-nougat-x86-rel']))
    self.assertEqual(
        set(blocking_entity.processed_builders),
        set(['android-nougat-x86-rel']))
    # Assert that CL was not checked for low coverage
    self.assertEqual(len(mock_blocking_logic.call_args_list), 0)

  # This test tests for the scenario where all coverage builders have completed
  # successfully, only the last one is pending processing and the last builder
  # has produced coverage data
  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(process_coverage.ProcessCodeCoverageData,
                     '_MayBeBlockCLForLowCoverage')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  # pylint: disable=line-too-long
  def testLowCoverageBlocking_FinalBuildHasData_checkForLowCoverage(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mock_blocking_logic, mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': [
                'chromium/try/android-nougat-x86-rel',
                'chromium/try/android-pie-x86-rel',
            ],
            'block_low_coverage_changes_projects': ['chromium/src'],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir'],
                }
            },
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/myfile.java',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 0,
                'first': 11,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfile.java', total_lines=90, covered_lines=9)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # Two coverage builds were triggered for the CL
    # and both completed without error. The third build is in the list
    # just for completeness
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS),
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-pie-x86-rel'),
                status=common_pb2.Status.SUCCESS),
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-performance-rel'),  # not a coverage build
                status=common_pb2.Status.SCHEDULED)
        ]))
    # Blocking entity created during the processing of a coverage build earlier
    # i.e. chromium/try/android-pie-x86-rel
    LowCoverageBlocking.Create(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4,
        expected_builders=['android-nougat-x86-rel', 'android-pie-x86-rel'],
        successful_builders=['android-pie-x86-rel'],
        processed_builders=['android-pie-x86-rel']).put()

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    blocking_entity = LowCoverageBlocking.Get(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4)
    self.assertEqual(blocking_entity.blocking_status,
                     BlockingStatus.READY_FOR_VERDICT)
    self.assertEqual(
        set(blocking_entity.expected_builders),
        set(['android-nougat-x86-rel', 'android-pie-x86-rel']))
    self.assertEqual(
        set(blocking_entity.successful_builders),
        set(['android-nougat-x86-rel', 'android-pie-x86-rel']))
    self.assertEqual(
        set(blocking_entity.processed_builders),
        set(['android-nougat-x86-rel', 'android-pie-x86-rel']))
    # Assert that CL was checked for low coverage
    self.assertEqual(len(mock_blocking_logic.call_args_list), 1)

  # This test tests for the scenario where all coverage builders have completed
  # successfully, only the last one is pending processing and the last builder
  # has produced coverage data
  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(process_coverage.ProcessCodeCoverageData,
                     '_MayBeBlockCLForLowCoverage')
  @mock.patch.object(process_coverage, 'GetV2Build')
  # pylint: disable=line-too-long
  def testLowCoverageBlocking_FinalBuildHasNoData_checkForLowCoverage(
      self, mocked_get_build, mock_blocking_logic, mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': [
                'chromium/try/android-nougat-x86-rel',
                'chromium/try/android-pie-x86-rel',
            ],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir']
                }
            },
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        # coverage_gs_bucket, coverage_metadata_gs_paths and
        # mimic_builder_names properties are missing indicating
        # this build did not produce any coverage data
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Two coverage builds were triggered for the CL
    # and both completed without error. The third build is in the list
    # just for completeness
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS),
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-pie-x86-rel'),
                status=common_pb2.Status.SUCCESS),
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-performance-rel'),  # not a coverage build
                status=common_pb2.Status.SCHEDULED)
        ]))
    # Blocking entity created during the processing of a coverage build earlier
    # i.e. chromium/try/android-pie-x86-rel
    LowCoverageBlocking.Create(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4,
        expected_builders=['android-nougat-x86-rel', 'android-pie-x86-rel'],
        successful_builders=['android-pie-x86-rel'],
        processed_builders=['android-pie-x86-rel']).put()

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    blocking_entity = LowCoverageBlocking.Get(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4)
    self.assertEqual(blocking_entity.blocking_status,
                     BlockingStatus.READY_FOR_VERDICT)
    self.assertEqual(
        set(blocking_entity.expected_builders),
        set(['android-nougat-x86-rel', 'android-pie-x86-rel']))
    self.assertEqual(
        set(blocking_entity.successful_builders),
        set(['android-nougat-x86-rel', 'android-pie-x86-rel']))
    self.assertEqual(
        set(blocking_entity.processed_builders),
        set(['android-nougat-x86-rel', 'android-pie-x86-rel']))
    # Assert that CL was checked for low coverage
    self.assertEqual(len(mock_blocking_logic.call_args_list), 1)

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util.FinditHttpClient, 'Post')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(code_coverage_util, 'FetchChangeDetails')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  # pylint: disable=line-too-long
  def testLowCoverageBlocking_allCoverageBuildsProcessedSuccessfully_block(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mocked_fetch_change_details, mocked_get_file_content, mock_http_client,
      mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': [
                'chromium/try/android-nougat-x86-rel',
                'chromium/try/android-pie-x86-rel',
            ],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir']
                }
            },
            'block_low_coverage_changes_projects': ['chromium/src']
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/myfile.java',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 0,
                'first': 11,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfile.java', total_lines=90, covered_lines=9)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # One coverage build, and it completed successfully
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))
    mocked_fetch_change_details.return_value = {
        'owner': {
            'email': 'john@chromium.org'
        }
    }
    mocked_get_file_content.return_value = json.dumps(
        {'john@chromium.org': 'john@google.com'})

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    blocking_entity = LowCoverageBlocking.Get(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4)
    self.assertEqual(blocking_entity.blocking_status,
                     BlockingStatus.VERDICT_BLOCK)
    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='postreview-request-queue')
    self.assertEqual(1, len(tasks))
    payload = json.loads(tasks[0].payload)
    self.assertDictEqual({'Code-Coverage': -1}, payload['data']['labels'])
    self.assertTrue('70%' in payload['data']['message'])
    self.assertTrue('clank' in payload['cohorts_matched'])
    self.assertTrue('clank' in payload['cohorts_violated'])

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util.FinditHttpClient, 'Post')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(code_coverage_util, 'FetchChangeDetails')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  def testLowCoverageBlocking_revertCL_noop(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mocked_fetch_change_details, mocked_get_file_content, mock_http_client,
      mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': ['chromium/try/android-nougat-x86-rel',],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir']
                }
            },
            'block_low_coverage_changes_projects': ['chromium/src']
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/myfile.java',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 0,
                'first': 11,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfile.java', total_lines=90, covered_lines=9)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # One coverage build was triggered and it succeeded
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))
    mocked_fetch_change_details.return_value = {
        'owner': {
            'email': 'john@google.com'
        },
        'revert_of': 123456
    }
    mocked_get_file_content.return_value = json.dumps(
        {'john@chromium.org': 'john@google.com'})

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='postreview-request-queue')
    self.assertEqual(0, len(tasks))

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util.FinditHttpClient, 'Post')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(code_coverage_util, 'FetchChangeDetails')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  def testLowCoverageBlocking_testMainExampleFile_allow(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mocked_fetch_change_details, mocked_get_file_content, mock_http_client,
      mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': ['chromium/try/android-nougat-x86-rel',],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir']
                }
            },
            'block_low_coverage_changes_projects': ['chromium/src']
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [
            {
                'path':
                    '//dir/myfileMain.java',
                'lines': [{
                    'count': 100,
                    'first': 1,
                    'last': 10,
                }, {
                    'count': 0,
                    'first': 11,
                    'last': 100,
                }],
            },
            {
                'path':
                    '//dir/myfileTests.java',
                'lines': [{
                    'count': 100,
                    'first': 1,
                    'last': 10,
                }, {
                    'count': 0,
                    'first': 11,
                    'last': 100,
                }],
            },
            {
                'path':
                    '//dir/examples/xx.java',
                'lines': [{
                    'count': 100,
                    'first': 1,
                    'last': 10,
                }, {
                    'count': 0,
                    'first': 11,
                    'last': 100,
                }],
            },
        ],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfileMain.java', total_lines=90, covered_lines=9),
        CoveragePercentage(
            path='//dir/myfileTests.java', total_lines=90, covered_lines=9)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # One coverage build was triggere and it succeeded
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))
    mocked_fetch_change_details.return_value = {
        'owner': {
            'email': 'john@chromium.org'
        }
    }
    mocked_get_file_content.return_value = json.dumps(
        {'john@chromium.org': 'john@google.com'})

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    blocking_entity = LowCoverageBlocking.Get(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4)
    self.assertEqual(blocking_entity.blocking_status,
                     BlockingStatus.VERDICT_NOT_BLOCK)
    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='postreview-request-queue')
    self.assertEqual(1, len(tasks))
    payload = json.loads(tasks[0].payload)
    self.assertDictEqual({'Code-Coverage': +1}, payload['data']['labels'])
    self.assertTrue('clank' in payload['cohorts_matched'])
    self.assertFalse('clank' in payload['cohorts_violated'])

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util.FinditHttpClient, 'Post')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(code_coverage_util, 'FetchChangeDetails')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  def testLowCoverageBlocking_highAbsoluteCoverage_allow(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mocked_fetch_change_details, mocked_get_file_content, mock_http_client,
      mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': ['chromium/try/android-nougat-x86-rel',],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir']
                }
            },
            'block_low_coverage_changes_projects': ['chromium/src']
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/myfile.java',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 10,
                'first': 11,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfile.java', total_lines=90, covered_lines=9)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # One coverage build was triggered and it succeeded
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))
    mocked_fetch_change_details.return_value = {
        'owner': {
            'email': 'john@google.com'
        }
    }
    mocked_get_file_content.return_value = json.dumps(
        {'john@chromium.org': 'john@google.com'})

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    blocking_entity = LowCoverageBlocking.Get(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4)
    self.assertEqual(blocking_entity.blocking_status,
                     BlockingStatus.VERDICT_NOT_BLOCK)
    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='postreview-request-queue')
    self.assertEqual(1, len(tasks))
    payload = json.loads(tasks[0].payload)
    self.assertDictEqual({'Code-Coverage': +1}, payload['data']['labels'])
    self.assertTrue('clank' in payload['cohorts_matched'])
    self.assertFalse('clank' in payload['cohorts_violated'])

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util.FinditHttpClient, 'Post')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(code_coverage_util, 'FetchChangeDetails')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  def testLowCoverageBlocking_authorOptInNotDefined_block(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mocked_fetch_change_details, mocked_get_file_content, mock_http_client,
      mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': ['chromium/try/android-nougat-x86-rel',],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'included_directories': ['//dir']
                }
            },
            'block_low_coverage_changes_projects': ['chromium/src']
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/myfile.java',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 0,
                'first': 11,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfile.java', total_lines=90, covered_lines=9)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # One coverage build was triggered and it succeeded
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))
    mocked_fetch_change_details.return_value = {
        'owner': {
            'email': 'john@google.com'
        }
    }
    mocked_get_file_content.return_value = json.dumps(
        {'john@chromium.org': 'john@google.com'})

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='postreview-request-queue')
    self.assertEqual(1, len(tasks))
    payload = json.loads(tasks[0].payload)
    self.assertDictEqual({'Code-Coverage': -1}, payload['data']['labels'])
    self.assertTrue('70%' in payload['data']['message'])
    self.assertTrue('clank' in payload['cohorts_matched'])
    self.assertTrue('clank' in payload['cohorts_violated'])

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util.FinditHttpClient, 'Post')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(code_coverage_util, 'FetchChangeDetails')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  def testLowCoverageBlocking_authorNotOptIn_noop(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mocked_fetch_change_details, mocked_get_file_content, mock_http_client,
      mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': ['chromium/try/android-nougat-x86-rel',],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['notjohn'],
                    'included_directories': ['//dir']
                }
            },
            'block_low_coverage_changes_projects': ['chromium/src']
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/myfile.java',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 0,
                'first': 11,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfile.java', total_lines=90, covered_lines=9)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # One coverage build was triggered and it succeeded
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))
    mocked_fetch_change_details.return_value = {
        'owner': {
            'email': 'john@google.com'
        }
    }
    mocked_get_file_content.return_value = json.dumps(
        {'john@chromium.org': 'john@google.com'})

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='postreview-request-queue')
    self.assertEqual(0, len(tasks))

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util.FinditHttpClient, 'Post')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(code_coverage_util, 'FetchChangeDetails')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  def testLowCoverageBlocking_externalAuthor_noop(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mocked_fetch_change_details, mocked_get_file_content, mock_http_client,
      mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': ['chromium/try/android-nougat-x86-rel',],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir']
                }
            },
            'block_low_coverage_changes_projects': ['chromium/src']
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/myfile.java',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 0,
                'first': 11,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfile.java', total_lines=90, covered_lines=9)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # One coverage build was triggered and it succeeded
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))
    mocked_fetch_change_details.return_value = {
        'owner': {
            # some other john from outside google created the change.
            'email': 'john@external.com'
        }
    }
    mocked_get_file_content.return_value = json.dumps(
        {'john@chromium.org': 'john@google.com'})

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='postreview-request-queue')
    self.assertEqual(0, len(tasks))

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util.FinditHttpClient, 'Post')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(code_coverage_util, 'FetchChangeDetails')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  def testLowCoverageBlocking_DirNotOptIn_allow(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mocked_fetch_change_details, mocked_get_file_content, mock_http_client,
      mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': ['chromium/try/android-nougat-x86-rel',],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': []
                }
            },
            'block_low_coverage_changes_projects': ['chromium/src']
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/myfile.java',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 0,
                'first': 11,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfile.java', total_lines=90, covered_lines=9)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # One coverage build was triggered and it succeeded
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))
    mocked_fetch_change_details.return_value = {
        'owner': {
            'email': 'john@google.com'
        }
    }
    mocked_get_file_content.return_value = json.dumps(
        {'john@chromium.org': 'john@google.com'})

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    blocking_entity = LowCoverageBlocking.Get(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4)
    self.assertEqual(blocking_entity.blocking_status,
                     BlockingStatus.READY_FOR_VERDICT)
    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='postreview-request-queue')
    self.assertEqual(0, len(tasks))

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util.FinditHttpClient, 'Post')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(code_coverage_util, 'FetchChangeDetails')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  def testLowCoverageBlocking_DirInOptOut_allow(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mocked_fetch_change_details, mocked_get_file_content, mock_http_client,
      mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': ['chromium/try/android-nougat-x86-rel',],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//'],
                    'excluded_directories': ['//dir/']
                }
            },
            'block_low_coverage_changes_projects': ['chromium/src']
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/myfile.java',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 0,
                'first': 11,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfile.java', total_lines=90, covered_lines=9)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # One coverage build was triggered and it succeeded
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))
    mocked_fetch_change_details.return_value = {
        'owner': {
            'email': 'john@google.com'
        }
    }
    mocked_get_file_content.return_value = json.dumps(
        {'john@chromium.org': 'john@google.com'})

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    blocking_entity = LowCoverageBlocking.Get(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4)
    self.assertEqual(blocking_entity.blocking_status,
                     BlockingStatus.READY_FOR_VERDICT)
    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='postreview-request-queue')
    self.assertEqual(0, len(tasks))

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util.FinditHttpClient, 'Post')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(code_coverage_util, 'FetchChangeDetails')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  def testLowCoverageBlocking_notEnoughLines_allow(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mocked_fetch_change_details, mocked_get_file_content, mock_http_client,
      mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': ['chromium/try/android-nougat-x86-rel',],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir']
                }
            },
            'block_low_coverage_changes_projects': ['chromium/src']
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build

    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/myfile.java',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 1,
            }, {
                'count': 0,
                'first': 2,
                'last': 4,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfile.java', total_lines=4, covered_lines=1)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # One coverage build was triggered and it succeeded
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))
    mocked_fetch_change_details.return_value = {
        'owner': {
            'email': 'john@google.com'
        }
    }
    mocked_get_file_content.return_value = json.dumps(
        {'john@chromium.org': 'john@google.com'})

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    blocking_entity = LowCoverageBlocking.Get(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4)
    self.assertEqual(blocking_entity.blocking_status,
                     BlockingStatus.VERDICT_NOT_BLOCK)
    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='postreview-request-queue')
    self.assertEqual(1, len(tasks))
    payload = json.loads(tasks[0].payload)
    self.assertDictEqual({'Code-Coverage': +1}, payload['data']['labels'])
    self.assertTrue('clank' in payload['cohorts_matched'])
    self.assertFalse('clank' in payload['cohorts_violated'])

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util.FinditHttpClient, 'Post')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(code_coverage_util, 'FetchChangeDetails')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  def testLowCoverageBlocking_FileNotOfBlockingFileType_allow(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mocked_fetch_change_details, mocked_get_file_content, mock_http_client,
      mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings',
        {
            'allowed_builders': ['chromium/try/android-nougat-x86-rel',],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir'],
                    # Blocking file type is different from the file in CL
                    'monitored_file_types': ['.java']
                }
            },
            'block_low_coverage_changes_projects': ['chromium/src']
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/myfile.cc',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 0,
                'first': 11,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfile.cc', total_lines=90, covered_lines=9)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # One coverage build was triggered and it succeeded
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))
    mocked_fetch_change_details.return_value = {
        'owner': {
            'email': 'john@google.com'
        }
    }
    mocked_get_file_content.return_value = json.dumps(
        {'john@chromium.org': 'john@google.com'})

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    blocking_entity = LowCoverageBlocking.Get(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4)
    self.assertEqual(blocking_entity.blocking_status,
                     BlockingStatus.READY_FOR_VERDICT)
    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='postreview-request-queue')
    self.assertEqual(0, len(tasks))

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util.FinditHttpClient, 'Post')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(code_coverage_util, 'FetchChangeDetails')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  def testLowCoverageBlocking_FileOfBlockingFileType_block(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mocked_fetch_change_details, mocked_get_file_content, mock_http_client,
      mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': ['chromium/try/android-nougat-x86-rel',],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir'],
                    'monitored_file_types': ['.cc']
                }
            },
            'block_low_coverage_changes_projects': ['chromium/src']
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/myfile.cc',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 0,
                'first': 11,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfile.cc', total_lines=90, covered_lines=9)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # One coverage build was triggered and it succeeded
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))
    mocked_fetch_change_details.return_value = {
        'owner': {
            'email': 'john@google.com'
        }
    }
    mocked_get_file_content.return_value = json.dumps(
        {'john@chromium.org': 'john@google.com'})

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    blocking_entity = LowCoverageBlocking.Get(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4)
    self.assertEqual(blocking_entity.blocking_status,
                     BlockingStatus.VERDICT_BLOCK)
    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='postreview-request-queue')
    self.assertEqual(1, len(tasks))
    payload = json.loads(tasks[0].payload)
    self.assertDictEqual({'Code-Coverage': -1}, payload['data']['labels'])
    self.assertTrue('clank' in payload['cohorts_matched'])
    self.assertTrue('clank' in payload['cohorts_violated'])

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(code_coverage_util, 'FetchChangeDetails')
  @mock.patch.object(code_coverage_util.FinditHttpClient, 'Post')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  def testLowCoverageBlocking_ProjectNotAllowed_noop(self, mocked_get_build,
                                                     mocked_get_validated_data,
                                                     mocked_inc_percentages,
                                                     *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': ['chromium/try/android-nougat-x86-rel',],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir'],
                    'monitored_file_types': ['.cc']
                }
            },
            'block_low_coverage_changes_projects': []
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir/myfile.java',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 0,
                'first': 11,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir/myfile.java', total_lines=90, covered_lines=9)
    ]
    mocked_inc_percentages.return_value = inc_percentages

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='postreview-request-queue')
    self.assertEqual(0, len(tasks))

  # This test tests for scenario where a CL makes changes which fall under
  # multiple cohorts, out of which only for some coverage requirement may
  # be getting violated.
  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util.FinditHttpClient, 'Post')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(code_coverage_util, 'FetchChangeDetails')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  # pylint: disable=line-too-long
  def testLowCoverageBlocking_multipleCohorts_taskPayloadCreatedCorrectly(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mocked_fetch_change_details, mocked_get_file_content, mock_http_client,
      mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': [
                'chromium/try/android-nougat-x86-rel',
                'chromium/try/android-pie-x86-rel',
            ],
            'block_low_coverage_changes': {
                'clank': {
                    "is_operational": True,
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir1']
                },
                'ios': {
                    "is_operational": True,
                    'included_directories': ['//dir2']
                }
            },
            'block_low_coverage_changes_projects': ['chromium/src']
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    # coverage is low only for clank changes, and not for ios
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir1/myclankfile.cc',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 0,
                'first': 11,
                'last': 100,
            }],
        }, {
            'path': '//dir2/myiosfile.cc',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir1/myclankfile.cc', total_lines=90, covered_lines=9),
        CoveragePercentage(
            path='//dir2/myiosfile.cc', total_lines=100, covered_lines=100)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # One coverage build, and it completed successfully
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))
    mocked_fetch_change_details.return_value = {
        'owner': {
            'email': 'john@chromium.org'
        }
    }
    mocked_get_file_content.return_value = json.dumps(
        {'john@chromium.org': 'john@google.com'})

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    blocking_entity = LowCoverageBlocking.Get(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4)
    self.assertEqual(blocking_entity.blocking_status,
                     BlockingStatus.VERDICT_BLOCK)
    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='postreview-request-queue')
    self.assertEqual(1, len(tasks))
    payload = json.loads(tasks[0].payload)
    self.assertDictEqual({'Code-Coverage': -1}, payload['data']['labels'])
    self.assertTrue('70%' in payload['data']['message'])
    self.assertTrue('ios' in payload['cohorts_matched'])
    self.assertTrue('clank' in payload['cohorts_matched'])
    self.assertFalse('ios' in payload['cohorts_violated'])
    self.assertTrue('clank' in payload['cohorts_violated'])

  # This test tests for scenario where a CL makes changes which fall under
  # multiple cohorts, out of which only for some coverage requirement may
  # be getting violated.
  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(prpc_client, 'service_account_credentials')
  @mock.patch.object(prpc_client, 'Client')
  @mock.patch.object(code_coverage_util.FinditHttpClient, 'Post')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(code_coverage_util, 'FetchChangeDetails')
  @mock.patch.object(code_coverage_util, 'CalculateIncrementalPercentages')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  # pylint: disable=line-too-long
  def testLowCoverageBlocking_nonOperationalCohort_noop(
      self, mocked_get_build, mocked_get_validated_data, mocked_inc_percentages,
      mocked_fetch_change_details, mocked_get_file_content, mock_http_client,
      mock_buildbucket_client, *_):
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'allowed_builders': [
                'chromium/try/android-nougat-x86-rel',
                'chromium/try/android-pie-x86-rel',
            ],
            'block_low_coverage_changes': {
                'clank': {
                    'monitored_authors': ['john'],
                    'included_directories': ['//dir1']
                }
            },
            'block_low_coverage_changes_projects': ['chromium/src']
        })
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chromium'
    build.builder.bucket = 'try'
    build.builder.builder = 'android-nougat-x86-rel'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'presubmit/chromium-review.googlesource.com/138000/4/try/'
            'android-nougat-x86-rel/123456789/metadata'
        ]), ('mimic_builder_names', ['android-nougat-x86-rel'])
    ]
    build.input.gerrit_changes = [
        mock.Mock(
            host='chromium-review.googlesource.com',
            project='chromium/src',
            change=138000,
            patchset=4)
    ]
    mocked_get_build.return_value = build
    # Mock get validated data from cloud storage.
    # coverage is low only for clank changes, and not for ios
    coverage_data = {
        'dirs': None,
        'files': [{
            'path':
                '//dir1/myclankfile.cc',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 10,
            }, {
                'count': 0,
                'first': 11,
                'last': 100,
            }],
        }, {
            'path': '//dir2/myiosfile.cc',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 100,
            }],
        }],
        'summaries': None,
        'components': None,
    }
    mocked_get_validated_data.return_value = coverage_data
    inc_percentages = [
        CoveragePercentage(
            path='//dir1/myclankfile.cc', total_lines=90, covered_lines=9),
        CoveragePercentage(
            path='//dir2/myiosfile.cc', total_lines=100, covered_lines=100)
    ]
    mocked_inc_percentages.return_value = inc_percentages
    # One coverage build, and it completed successfully
    mock_buildbucket_client.return_value.SearchBuilds.return_value = (
        builds_service_pb2.SearchBuildsResponse(builds=[
            build_pb2.Build(
                builder=builder_common_pb2.BuilderID(
                    builder='android-nougat-x86-rel'),
                status=common_pb2.Status.SUCCESS)
        ]))
    mocked_fetch_change_details.return_value = {
        'owner': {
            'email': 'john@chromium.org'
        }
    }
    mocked_get_file_content.return_value = json.dumps(
        {'john@chromium.org': 'john@google.com'})

    request_url = '/coverage/task/process-data/build/123456789'
    self.test_app.post(request_url)

    blocking_entity = LowCoverageBlocking.Get(
        server_host='chromium-review.googlesource.com',
        change=138000,
        patchset=4)
    self.assertEqual(blocking_entity.blocking_status,
                     BlockingStatus.READY_FOR_VERDICT)
    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='postreview-request-queue')
    self.assertEqual(0, len(tasks))

  @mock.patch.object(process_coverage.ProcessCodeCoverageData,
                     '_FetchAndSaveFileIfNecessary')
  @mock.patch.object(process_coverage, '_RetrieveChromeManifest')
  @mock.patch.object(process_coverage.CachedGitilesRepository, 'GetChangeLog')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  def testProcessFullRepoData(self, mocked_is_request_from_appself,
                              mocked_get_build, mocked_get_validated_data,
                              mocked_get_change_log, mocked_retrieve_manifest,
                              mocked_fetch_file, *_):
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chrome'
    build.builder.bucket = 'coverage'
    build.builder.builder = 'linux-code-coverage'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'postsubmit/chromium.googlesource.com/chromium/src/'
            'aaaaa/coverage/linux-code-coverage/123456789/metadata',
            'postsubmit/chromium.googlesource.com/chromium/src/'
            'aaaaa/coverage/linux-code-coverage_unit/123456789/metadata'
        ]),
        ('mimic_builder_names',
         ['linux-code-coverage', 'linux-code-coverage_unit'])
    ]
    build.input.gitiles_commit = mock.Mock(
        host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        id='aaaaa')
    mocked_get_build.return_value = build

    # Mock Gitiles API to get change log.
    change_log = mock.Mock()
    change_log.committer.time = datetime(2018, 1, 1)
    mocked_get_change_log.return_value = change_log

    # Mock retrieve manifest.
    manifest = _CreateSampleManifest()
    mocked_retrieve_manifest.return_value = manifest

    # Mock get validated data from cloud storage for both all.json and file
    # shard json.
    all_coverage_data = {
        'dirs': [{
            'path': '//dir/',
            'dirs': [],
            'files': [{
                'path': '//dir/test.cc',
                'name': 'test.cc',
                'summaries': _CreateSampleCoverageSummaryMetric()
            }],
            'summaries': _CreateSampleCoverageSummaryMetric()
        }],
        'file_shards': ['file_coverage/files1.json.gz'],
        'summaries':
            _CreateSampleCoverageSummaryMetric(),
        'components': [{
            'path': 'Component>Test',
            'dirs': [{
                'path': '//dir/',
                'name': 'dir/',
                'summaries': _CreateSampleCoverageSummaryMetric()
            }],
            'summaries': _CreateSampleCoverageSummaryMetric()
        }],
    }

    file_shard_coverage_data = {
        'files': [{
            'path':
                '//dir/test.cc',
            'revision':
                'bbbbb',
            'lines': [{
                'count': 100,
                'last': 2,
                'first': 1
            }],
            'timestamp':
                '140000',
            'uncovered_blocks': [{
                'line': 1,
                'ranges': [{
                    'first': 1,
                    'last': 2
                }]
            }]
        }]
    }

    mocked_get_validated_data.side_effect = [
        all_coverage_data, file_shard_coverage_data, all_coverage_data,
        file_shard_coverage_data
    ]

    request_url = '/coverage/task/process-data/build/123456789'
    response = self.test_app.post(request_url)
    self.assertEqual(200, response.status_int)
    mocked_is_request_from_appself.assert_called()

    fetched_reports = PostsubmitReport.query().fetch()
    self.assertEqual(2, len(fetched_reports))
    self.assertEqual(_CreateSamplePostsubmitReport(), fetched_reports[0])
    self.assertEqual(
        _CreateSamplePostsubmitReport(builder='linux-code-coverage_unit'),
        fetched_reports[1])
    mocked_fetch_file.assert_called_with(
        _CreateSamplePostsubmitReport(builder='linux-code-coverage_unit'),
        '//dir/test.cc', 'bbbbb')

    fetched_file_coverage_data = FileCoverageData.query().fetch()
    self.assertEqual(2, len(fetched_file_coverage_data))
    self.assertEqual(_CreateSampleFileCoverageData(),
                     fetched_file_coverage_data[0])
    self.assertEqual(
        _CreateSampleFileCoverageData(builder='linux-code-coverage_unit'),
        fetched_file_coverage_data[1])

    fetched_summary_coverage_data = SummaryCoverageData.query().fetch()
    self.assertListEqual([
        _CreateSampleRootComponentCoverageData(),
        _CreateSampleRootComponentCoverageData(
            builder='linux-code-coverage_unit'),
        _CreateSampleComponentCoverageData(),
        _CreateSampleComponentCoverageData(builder='linux-code-coverage_unit'),
        _CreateSampleDirectoryCoverageData(),
        _CreateSampleDirectoryCoverageData(builder='linux-code-coverage_unit')
    ], fetched_summary_coverage_data)

  @mock.patch.object(process_coverage.ProcessCodeCoverageData,
                     '_FetchAndSaveFileIfNecessary')
  @mock.patch.object(process_coverage, '_RetrieveChromeManifest')
  @mock.patch.object(process_coverage.CachedGitilesRepository, 'GetChangeLog')
  @mock.patch.object(process_coverage, '_GetValidatedData')
  @mock.patch.object(process_coverage, 'GetV2Build')
  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  def testProcessFullRepoData_outputGitilesCommit(
      self, mocked_is_request_from_appself, mocked_get_build,
      mocked_get_validated_data, mocked_get_change_log,
      mocked_retrieve_manifest, mocked_fetch_file, *_):
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.status = common_pb2.Status.SUCCESS
    build.builder.project = 'chrome'
    build.builder.bucket = 'coverage'
    build.builder.builder = 'linux-code-coverage'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'postsubmit/chromium.googlesource.com/chromium/src/'
            'aaaaa/coverage/linux-code-coverage/123456789/metadata',
            'postsubmit/chromium.googlesource.com/chromium/src/'
            'aaaaa/coverage/linux-code-coverage_unit/123456789/metadata'
        ]),
        ('mimic_builder_names',
         ['linux-code-coverage', 'linux-code-coverage_unit'])
    ]
    build.input.gitiles_commit = mock.Mock(
        host=None, project=None, ref=None, id=None)
    build.output.gitiles_commit = mock.Mock(
        host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        id='aaaaa')
    mocked_get_build.return_value = build

    # Mock Gitiles API to get change log.
    change_log = mock.Mock()
    change_log.committer.time = datetime(2018, 1, 1)
    mocked_get_change_log.return_value = change_log

    # Mock retrieve manifest.
    manifest = _CreateSampleManifest()
    mocked_retrieve_manifest.return_value = manifest

    # Mock get validated data from cloud storage for both all.json and file
    # shard json.
    all_coverage_data = {
        'dirs': [{
            'path': '//dir/',
            'dirs': [],
            'files': [{
                'path': '//dir/test.cc',
                'name': 'test.cc',
                'summaries': _CreateSampleCoverageSummaryMetric()
            }],
            'summaries': _CreateSampleCoverageSummaryMetric()
        }],
        'file_shards': ['file_coverage/files1.json.gz'],
        'summaries':
            _CreateSampleCoverageSummaryMetric(),
        'components': [{
            'path': 'Component>Test',
            'dirs': [{
                'path': '//dir/',
                'name': 'dir/',
                'summaries': _CreateSampleCoverageSummaryMetric()
            }],
            'summaries': _CreateSampleCoverageSummaryMetric()
        }],
    }

    file_shard_coverage_data = {
        'files': [{
            'path':
                '//dir/test.cc',
            'revision':
                'bbbbb',
            'lines': [{
                'count': 100,
                'last': 2,
                'first': 1
            }],
            'timestamp':
                '140000',
            'uncovered_blocks': [{
                'line': 1,
                'ranges': [{
                    'first': 1,
                    'last': 2
                }]
            }]
        }]
    }

    mocked_get_validated_data.side_effect = [
        all_coverage_data, file_shard_coverage_data, all_coverage_data,
        file_shard_coverage_data
    ]

    request_url = '/coverage/task/process-data/build/123456789'
    response = self.test_app.post(request_url)
    self.assertEqual(200, response.status_int)
    mocked_is_request_from_appself.assert_called()

    fetched_reports = PostsubmitReport.query().fetch()
    self.assertEqual(2, len(fetched_reports))
    self.assertEqual(_CreateSamplePostsubmitReport(), fetched_reports[0])
    self.assertEqual(
        _CreateSamplePostsubmitReport(builder='linux-code-coverage_unit'),
        fetched_reports[1])
    mocked_fetch_file.assert_called_with(
        _CreateSamplePostsubmitReport(builder='linux-code-coverage_unit'),
        '//dir/test.cc', 'bbbbb')

    fetched_file_coverage_data = FileCoverageData.query().fetch()
    self.assertEqual(2, len(fetched_file_coverage_data))
    self.assertEqual(_CreateSampleFileCoverageData(),
                     fetched_file_coverage_data[0])
    self.assertEqual(
        _CreateSampleFileCoverageData(builder='linux-code-coverage_unit'),
        fetched_file_coverage_data[1])

    fetched_summary_coverage_data = SummaryCoverageData.query().fetch()
    self.assertListEqual([
        _CreateSampleRootComponentCoverageData(),
        _CreateSampleRootComponentCoverageData(
            builder='linux-code-coverage_unit'),
        _CreateSampleComponentCoverageData(),
        _CreateSampleComponentCoverageData(builder='linux-code-coverage_unit'),
        _CreateSampleDirectoryCoverageData(),
        _CreateSampleDirectoryCoverageData(builder='linux-code-coverage_unit')
    ], fetched_summary_coverage_data)
