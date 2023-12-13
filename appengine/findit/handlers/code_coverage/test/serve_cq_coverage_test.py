# -*- coding: utf-8 -*-
# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json
import mock
from flask import Flask

from handlers.code_coverage import serve_cq_coverage
from model.code_coverage import CoveragePercentage
from model.code_coverage import PresubmitCoverageData
from services.code_coverage import code_coverage_util
from waterfall.test.wf_testcase import WaterfallTestCase


class ServeCodeCoverageDataTest(WaterfallTestCase):
  app_module = Flask(__name__)
  app_module.add_url_rule(
      '/coverage/api/coverage-data',
      view_func=serve_cq_coverage.ServeCodeCoverageData().Handle,
      methods=['GET'])

  def setUp(self):
    super(ServeCodeCoverageDataTest, self).setUp()
    self.UpdateUnitTestConfigSettings(
        'code_coverage_settings', {
            'serve_presubmit_coverage_data': True,
            'allowed_gerrit_configs': {
                'chromium-review.googlesource.com': ['chromium/src']
            },
        })

  def tearDown(self):
    self.UpdateUnitTestConfigSettings('code_coverage_settings', {})
    super(ServeCodeCoverageDataTest, self).tearDown()

  def testServeCLPatchsetLinesData(self):
    host = 'chromium-review.googlesource.com'
    project = 'chromium/src'
    change = 138000
    patchset = 4
    data = [{
        'path': '//dir/myfile.cc',
        'lines': [{
            'count': 100,
            'first': 1,
            'last': 2,
        }],
    }]
    PresubmitCoverageData.Create(
        server_host=host, change=change, patchset=patchset, data=data).put()

    request_url = ('/coverage/api/coverage-data?host=%s&project=%s&change=%d'
                   '&patchset=%d&concise=1') % (host, project, change, patchset)
    response = self.test_app.get(request_url)

    expected_response_body = json.dumps({
        'data': {
            'files': [{
                'path':
                    'dir/myfile.cc',
                'lines': [{
                    'count': 100,
                    'line': 1,
                }, {
                    'count': 100,
                    'line': 2,
                }]
            }]
        },
    })
    self.assertEqual(expected_response_body, response.body)

  def testServeCLPatchsetLinesDataInvalidPatchset(self):
    host = 'chromium-review.googlesource.com'
    project = 'chromium/src'
    change = 138000
    request_url = ('/coverage/api/coverage-data?host=%s&project=%s&change=%d'
                   '&patchset=NaN&concise=1') % (host, project, change)
    with self.assertRaisesRegexp(Exception, r'.*400.*'):
      self.test_app.get(request_url)

  @mock.patch.object(code_coverage_util, 'GetEquivalentPatchsets')
  def testServeCLPatchLinesDataNoEquivalentPatchsets(self,
                                                     mock_get_equivalent_ps):
    host = 'chromium-review.googlesource.com'
    project = 'chromium/src'
    change = 138000
    patchset = 4
    mock_get_equivalent_ps.return_value = []
    request_url = ('/coverage/api/coverage-data?host=%s&project=%s&change=%d'
                   '&patchset=%d&concise=1') % (host, project, change, patchset)
    response = self.test_app.get(request_url, expect_errors=True)
    self.assertEqual(404, response.status_int)

  @mock.patch.object(code_coverage_util, 'GetEquivalentPatchsets')
  def testServeCLPatchLinesDataEquivalentPatchsetsHaveNoData(
      self, mock_get_equivalent_ps):
    host = 'chromium-review.googlesource.com'
    project = 'chromium/src'
    change = 138000
    patchset_src = 3
    patchset_dest = 4
    mock_get_equivalent_ps.return_value = [patchset_src]
    request_url = ('/coverage/api/coverage-data?host=%s&project=%s&change=%d'
                   '&patchset=%d&concise=1') % (host, project, change,
                                                patchset_dest)
    response = self.test_app.get(request_url, expect_errors=True)
    self.assertEqual(404, response.status_int)

  @mock.patch.object(code_coverage_util,
                     'RebasePresubmitCoverageDataBetweenPatchsets')
  @mock.patch.object(code_coverage_util, 'GetEquivalentPatchsets')
  def testServeCLPatchLinesDataEquivalentPatchsetsMissingData(
      self, mock_get_equivalent_ps, mock_rebase_data):
    host = 'chromium-review.googlesource.com'
    project = 'chromium/src'
    change = 138000
    patchset_src = 3
    # 4 is based on 3, used to test that 5 would choose 3 instead of 4.
    patchset_mid = 4
    patchset_dest = 5
    data = [{
        'path': '//dir/myfile.cc',
        'lines': [{
            'count': 100,
            'first': 1,
            'last': 2,
        }],
    }]
    PresubmitCoverageData.Create(
        server_host=host, change=change, patchset=patchset_src,
        data=data).put()
    mid_data = PresubmitCoverageData.Create(
        server_host=host, change=change, patchset=patchset_mid, data=data)
    mid_data.based_on = patchset_src
    mid_data.put()

    mock_get_equivalent_ps.return_value = [patchset_src, patchset_mid]
    mock_rebase_data.side_effect = (
        code_coverage_util.MissingChangeDataException(''))

    request_url = ('/coverage/api/coverage-data?host=%s&project=%s&change=%d'
                   '&patchset=%d&concise=1') % (host, project, change,
                                                patchset_dest)
    response = self.test_app.get(request_url, expect_errors=True)
    self.assertEqual(404, response.status_int)

    mock_rebase_data.side_effect = RuntimeError('Some unknown http code')
    response = self.test_app.get(request_url, expect_errors=True)
    self.assertEqual(500, response.status_int)

  @mock.patch.object(code_coverage_util,
                     'RebasePresubmitCoverageDataBetweenPatchsets')
  @mock.patch.object(code_coverage_util, 'GetEquivalentPatchsets')
  def testServeCLPatchLinesDataEquivalentPatchsets(self, mock_get_equivalent_ps,
                                                   mock_rebase_data):
    host = 'chromium-review.googlesource.com'
    project = 'chromium/src'
    change = 138000
    patchset_src = 3
    # 4 is based on 3, used to test that 5 would choose 3 instead of 4.
    patchset_mid = 4
    patchset_dest = 5
    data = [{
        'path': '//dir/myfile.cc',
        'lines': [{
            'count': 100,
            'first': 1,
            'last': 2,
        }],
    }]
    data_unit = [{
        'path': '//dir/myfile.cc',
        'lines': [{
            'count': 100,
            'first': 1,
            'last': 3,
        }],
    }]
    PresubmitCoverageData.Create(
        server_host=host,
        change=change,
        patchset=patchset_src,
        data=data,
        data_unit=data_unit).put()
    mid_data = PresubmitCoverageData.Create(
        server_host=host,
        change=change,
        patchset=patchset_mid,
        data=data,
        data_unit=data_unit)
    mid_data.based_on = patchset_src
    mid_data.put()

    rebased_coverage_data = [{
        'path': '//dir/myfile.cc',
        'lines': [{
            'count': 100,
            'first': 2,
            'last': 3,
        }],
    }]

    rebased_coverage_data_unit = [{
        'path': '//dir/myfile.cc',
        'lines': [{
            'count': 100,
            'first': 2,
            'last': 4,
        }],
    }]

    mock_get_equivalent_ps.return_value = [patchset_src, patchset_mid]
    mock_rebase_data.side_effect = [
        rebased_coverage_data, rebased_coverage_data_unit
    ]

    request_url = ('/coverage/api/coverage-data?host=%s&project=%s&change=%d'
                   '&patchset=%d&concise=1') % (host, project, change,
                                                patchset_dest)
    response = self.test_app.get(request_url)

    expected_response_body = json.dumps({
        'data': {
            'files': [{
                'path':
                    'dir/myfile.cc',
                'lines': [{
                    'count': 100,
                    'line': 2,
                }, {
                    'count': 100,
                    'line': 3,
                }]
            }]
        },
    })
    self.assertEqual(expected_response_body, response.body)
    src_entity = PresubmitCoverageData.Get(host, change, patchset_src)
    dest_entity = PresubmitCoverageData.Get(host, change, patchset_dest)
    self.assertEqual(patchset_src, dest_entity.based_on)
    self.assertEqual(src_entity.absolute_percentages,
                     dest_entity.absolute_percentages)
    self.assertEqual(src_entity.incremental_percentages,
                     dest_entity.incremental_percentages)
    self.assertEqual(src_entity.absolute_percentages_unit,
                     dest_entity.absolute_percentages_unit)
    self.assertEqual(src_entity.incremental_percentages_unit,
                     dest_entity.incremental_percentages_unit)
    self.assertEqual(rebased_coverage_data, dest_entity.data)
    self.assertEqual(rebased_coverage_data_unit, dest_entity.data_unit)

  def testServeCLPatchPercentagesData(self):
    host = 'chromium-review.googlesource.com'
    project = 'chromium/src'
    change = 138000
    patchset = 4
    data = [{
        'path': '//dir/myfile.cc',
        'lines': [{
            'count': 100,
            'first': 1,
            'last': 2,
        }],
    }]
    entity = PresubmitCoverageData.Create(
        server_host=host, change=change, patchset=patchset, data=data)
    entity.absolute_percentages = [
        CoveragePercentage(
            path='//dir/myfile.cc', total_lines=2, covered_lines=1)
    ]
    entity.incremental_percentages = [
        CoveragePercentage(
            path='//dir/myfile.cc', total_lines=1, covered_lines=1)
    ]
    entity.absolute_percentages_unit = [
        CoveragePercentage(
            path='//dir/myfile.cc', total_lines=2, covered_lines=1)
    ]
    entity.incremental_percentages_unit = [
        CoveragePercentage(
            path='//dir/myfile.cc', total_lines=1, covered_lines=1)
    ]
    entity.put()

    request_url = ('/coverage/api/coverage-data?host=%s&project=%s&change=%d'
                   '&patchset=%d&type=percentages&concise=1') % (
                       host, project, change, patchset)
    response = self.test_app.get(request_url)

    expected_response_body = json.dumps({
        'data': {
            'files': [{
                "path": "dir/myfile.cc",
                "absolute_coverage": {
                    "covered": 1,
                    "total": 2,
                },
                "incremental_coverage": {
                    "covered": 1,
                    "total": 1,
                },
                "absolute_unit_tests_coverage": {
                    "covered": 1,
                    "total": 2,
                },
                "incremental_unit_tests_coverage": {
                    "covered": 1,
                    "total": 1,
                },
            }]
        },
    })
    self.assertEqual(expected_response_body, response.body)

  @mock.patch.object(code_coverage_util, 'GetEquivalentPatchsets')
  def testServeCLPatchPercentagesDataEquivalentPatchsets(
      self, mock_get_equivalent_ps):
    host = 'chromium-review.googlesource.com'
    project = 'chromium/src'
    change = 138000
    patchset_src = 3
    patchset_dest = 4
    mock_get_equivalent_ps.return_value = [patchset_src]
    data = [{
        'path': '//dir/myfile.cc',
        'lines': [{
            'count': 100,
            'first': 1,
            'last': 2,
        }],
    }]
    entity = PresubmitCoverageData.Create(
        server_host=host, change=change, patchset=patchset_src, data=data)
    entity.absolute_percentages = [
        CoveragePercentage(
            path='//dir/myfile.cc', total_lines=2, covered_lines=1)
    ]
    entity.incremental_percentages = [
        CoveragePercentage(
            path='//dir/myfile.cc', total_lines=1, covered_lines=1)
    ]
    entity.absolute_percentages_unit = [
        CoveragePercentage(
            path='//dir/myfile.cc', total_lines=2, covered_lines=1)
    ]
    entity.incremental_percentages_unit = [
        CoveragePercentage(
            path='//dir/myfile.cc', total_lines=1, covered_lines=1)
    ]
    entity.put()

    request_url = ('/coverage/api/coverage-data?host=%s&project=%s&change=%d'
                   '&patchset=%d&type=percentages&concise=1') % (
                       host, project, change, patchset_dest)
    response = self.test_app.get(request_url)

    expected_response_body = json.dumps({
        'data': {
            'files': [{
                "path": "dir/myfile.cc",
                "absolute_coverage": {
                    "covered": 1,
                    "total": 2,
                },
                "incremental_coverage": {
                    "covered": 1,
                    "total": 1,
                },
                "absolute_unit_tests_coverage": {
                    "covered": 1,
                    "total": 2,
                },
                "incremental_unit_tests_coverage": {
                    "covered": 1,
                    "total": 1,
                },
            }]
        },
    })
    self.assertEqual(expected_response_body, response.body)

  def testServeCLPatchPercentagesDataJustUnitTestCoverage(self):
    host = 'chromium-review.googlesource.com'
    project = 'chromium/src'
    change = 138000
    patchset = 4
    data = [{
        'path': '//dir/myfile.cc',
        'lines': [{
            'count': 100,
            'first': 1,
            'last': 2,
        }],
    }]
    entity = PresubmitCoverageData.Create(
        server_host=host, change=change, patchset=patchset, data=data)
    entity.absolute_percentages_unit = [
        CoveragePercentage(
            path='//dir/myfile.cc', total_lines=2, covered_lines=1)
    ]
    entity.incremental_percentages_unit = [
        CoveragePercentage(
            path='//dir/myfile.cc', total_lines=1, covered_lines=1)
    ]
    entity.put()

    request_url = ('/coverage/api/coverage-data?host=%s&project=%s&change=%d'
                   '&patchset=%d&type=percentages&concise=1') % (
                       host, project, change, patchset)
    response = self.test_app.get(request_url)

    expected_response_body = json.dumps({
        'data': {
            'files': [{
                "path": "dir/myfile.cc",
                "absolute_coverage": None,
                "incremental_coverage": None,
                "absolute_unit_tests_coverage": {
                    "covered": 1,
                    "total": 2,
                },
                "incremental_unit_tests_coverage": {
                    "covered": 1,
                    "total": 1,
                },
            }]
        },
    })
    self.assertEqual(expected_response_body, response.body)
