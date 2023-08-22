# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import mock
from datetime import datetime
import json

from google.appengine.ext import ndb

from handlers.code_coverage import utils
from waterfall.test.wf_testcase import WaterfallTestCase
from model.code_coverage import CoverageReportModifier
from model.code_coverage import DependencyRepository
from model.code_coverage import FileCoverageData
from model.code_coverage import PostsubmitReport
from model.code_coverage import SummaryCoverageData
from services.code_coverage import author_coverage


class AuthorCoverageTest(WaterfallTestCase):

  def setUp(self):
    super(AuthorCoverageTest, self).setUp()

  def tearDown(self):
    self.UpdateUnitTestConfigSettings('code_coverage_settings', {})
    super(AuthorCoverageTest, self).tearDown()

  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(author_coverage, 'GetV2Build')
  def testAuthorModiferExists_FileCoverageGetsCreated(self, mock_get_build,
                                                      mock_get_file_content):
    CoverageReportModifier(author='abc@google.com', id=123).put()
    postsubmit_report = PostsubmitReport.Create(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        bucket='ci',
        builder='android-code-coverage',
        commit_timestamp=datetime(2020, 1, 7),
        manifest=[
            DependencyRepository(
                path='//',
                server_host='chromium.googlesource.com',
                project='chromium/src.git',
                revision='latest')
        ],
        summary_metrics={},
        build_id=2000,
        visible=True)
    postsubmit_report.put()
    file_coverage_data = FileCoverageData.Create(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        path='//a/myfile.java',
        bucket='ci',
        builder='android-code-coverage',
        data={
            'lines': [{
                'count': 0,
                'first': 1,
                'last': 3
            }, {
                'count': 10,
                'first': 4,
                'last': 5
            }]
        })
    file_coverage_data.put()
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.builder.project = 'chrome'
    build.builder.bucket = 'coverage'
    build.builder.builder = 'android-code-coverage'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'postsubmit/chromium.googlesource.com/chromium/src/'
            'aaaaa/coverage/android-code-coverage/123456789/metadata',
        ]), ('mimic_builder_names', ['android-code-coverage'])
    ]
    build.input.gitiles_commit = mock.Mock(
        host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        id='123')
    mock_get_build.return_value = build
    blamelist = {'//a/myfile.java': {'abc@google.com': ['1', '2']}}
    chromium_to_googler_mapping = {'def@chromium.org': 'def@google.com'}
    mock_get_file_content.side_effect = [
        json.dumps(chromium_to_googler_mapping),
        json.dumps(blamelist)
    ]

    author_coverage.CreateAuthorCoverage('android-code-coverage')

    entity1 = FileCoverageData.Get(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        path='//a/myfile.java',
        bucket='ci',
        builder='android-code-coverage',
        modifier_id=123)
    self.assertEqual(
        entity1.data, {
            'path': '//a/myfile.java',
            'lines': [{
                'first': 1,
                'last': 2,
                'count': 0
            }],
            'summaries': [{
                'name': 'line',
                'total': 2,
                'covered': 0
            }],
            'revision': 'latest'
        })

  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(author_coverage, 'GetV2Build')
  def testAuthorModifierDoesntExist_FileCoverageGetsCreated(
      self, mock_get_build, mock_get_file_content):
    postsubmit_report = PostsubmitReport.Create(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        bucket='ci',
        builder='android-code-coverage',
        commit_timestamp=datetime(2020, 1, 7),
        manifest=[
            DependencyRepository(
                path='//',
                server_host='chromium.googlesource.com',
                project='chromium/src.git',
                revision='latest')
        ],
        summary_metrics={},
        build_id=2000,
        visible=True)
    postsubmit_report.put()

    file_coverage_data = FileCoverageData.Create(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        path='//a/myfile.java',
        bucket='ci',
        builder='android-code-coverage',
        data={
            'lines': [{
                'count': 0,
                'first': 1,
                'last': 3
            }, {
                'count': 10,
                'first': 4,
                'last': 5
            }]
        })
    file_coverage_data.put()

    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.builder.project = 'chrome'
    build.builder.bucket = 'coverage'
    build.builder.builder = 'android-code-coverage'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'postsubmit/chromium.googlesource.com/chromium/src/'
            'aaaaa/coverage/android-code-coverage/123456789/metadata',
        ]), ('mimic_builder_names', ['android-code-coverage'])
    ]
    build.input.gitiles_commit = mock.Mock(
        host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        id='123')
    mock_get_build.return_value = build

    blamelist = {'//a/myfile.java': {'abc@google.com': ['1', '2']}}
    chromium_to_googler_mapping = {'def@chromium.org': 'def@google.com'}
    mock_get_file_content.side_effect = [
        json.dumps(chromium_to_googler_mapping),
        json.dumps(blamelist)
    ]

    author_coverage.CreateAuthorCoverage('android-code-coverage')

    modifier = CoverageReportModifier.query().fetch()[0]
    entity = FileCoverageData.Get(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        path='//a/myfile.java',
        bucket='ci',
        builder='android-code-coverage',
        modifier_id=modifier.key.id())
    self.assertEqual(
        entity.data, {
            'path': '//a/myfile.java',
            'lines': [{
                'first': 1,
                'last': 2,
                'count': 0
            }],
            'summaries': [{
                'name': 'line',
                'total': 2,
                'covered': 0
            }],
            'revision': 'latest'
        })

  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(author_coverage, 'GetV2Build')
  def testAuthorHasChromiumAccount_FileCoverageGetsCreated(
      self, mock_get_build, mock_get_file_content):
    postsubmit_report = PostsubmitReport.Create(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        bucket='ci',
        builder='android-code-coverage',
        commit_timestamp=datetime(2020, 1, 7),
        manifest=[
            DependencyRepository(
                path='//',
                server_host='chromium.googlesource.com',
                project='chromium/src.git',
                revision='latest')
        ],
        summary_metrics={},
        build_id=2000,
        visible=True)
    postsubmit_report.put()
    file_coverage_data = FileCoverageData.Create(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        path='//a/myfile.java',
        bucket='ci',
        builder='android-code-coverage',
        data={
            'lines': [{
                'count': 0,
                'first': 1,
                'last': 3
            }, {
                'count': 10,
                'first': 4,
                'last': 5
            }]
        })
    file_coverage_data.put()
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.builder.project = 'chrome'
    build.builder.bucket = 'coverage'
    build.builder.builder = 'android-code-coverage'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'postsubmit/chromium.googlesource.com/chromium/src/'
            'aaaaa/coverage/android-code-coverage/123456789/metadata',
        ]), ('mimic_builder_names', ['android-code-coverage'])
    ]
    build.input.gitiles_commit = mock.Mock(
        host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        id='123')
    mock_get_build.return_value = build

    blamelist = {'//a/myfile.java': {'def@chromium.org': ['4']}}
    chromium_to_googler_mapping = {'def@chromium.org': 'def@google.com'}
    mock_get_file_content.side_effect = [
        json.dumps(chromium_to_googler_mapping),
        json.dumps(blamelist)
    ]

    author_coverage.CreateAuthorCoverage('android-code-coverage')

    modifier = CoverageReportModifier.query().fetch()[0]
    entity = FileCoverageData.Get(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        path='//a/myfile.java',
        bucket='ci',
        builder='android-code-coverage',
        modifier_id=modifier.key.id())
    self.assertEqual(
        entity.data, {
            'path': '//a/myfile.java',
            'lines': [{
                'first': 4,
                'last': 4,
                'count': 10
            }],
            'summaries': [{
                'name': 'line',
                'total': 1,
                'covered': 1
            }],
            'revision': 'latest'
        })
    self.assertEqual(modifier.author, 'def@google.com')

  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(author_coverage, 'GetV2Build')
  def testSummaryCoverageGetsCreated(self, mock_get_build,
                                     mock_get_file_content):
    CoverageReportModifier(author='abc@google.com', id=123).put()
    postsubmit_report = PostsubmitReport.Create(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        bucket='ci',
        builder='android-code-coverage',
        commit_timestamp=datetime(2020, 1, 7),
        manifest=[
            DependencyRepository(
                path='//',
                server_host='chromium.googlesource.com',
                project='chromium/src.git',
                revision='latest')
        ],
        summary_metrics={},
        build_id=2000,
        visible=True)
    postsubmit_report.put()
    file_coverage_data = FileCoverageData.Create(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        path='//a/myfile.java',
        bucket='ci',
        builder='android-code-coverage',
        data={
            'lines': [{
                'count': 0,
                'first': 1,
                'last': 3
            }, {
                'count': 10,
                'first': 4,
                'last': 5
            }]
        })
    file_coverage_data.put()
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.builder.project = 'chrome'
    build.builder.bucket = 'coverage'
    build.builder.builder = 'android-code-coverage'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'postsubmit/chromium.googlesource.com/chromium/src/'
            'aaaaa/coverage/android-code-coverage/123456789/metadata',
        ]), ('mimic_builder_names', ['android-code-coverage'])
    ]
    build.input.gitiles_commit = mock.Mock(
        host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        id='123')
    mock_get_build.return_value = build
    blamelist = {'//a/myfile.java': {'abc@google.com': ['1', '2']}}
    chromium_to_googler_mapping = {'def@chromium.org': 'def@google.com'}
    mock_get_file_content.side_effect = [
        json.dumps(chromium_to_googler_mapping),
        json.dumps(blamelist)
    ]

    author_coverage.CreateAuthorCoverage('android-code-coverage')

    entity1 = SummaryCoverageData.Get(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        data_type='dirs',
        path='//a/',
        bucket='ci',
        builder='android-code-coverage',
        modifier_id=123)
    entity2 = SummaryCoverageData.Get(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        data_type='dirs',
        path='//',
        bucket='ci',
        builder='android-code-coverage',
        modifier_id=123)
    self.assertEqual(
        entity1.data, {
            'dirs': [],
            'path':
                '//a/',
            'summaries': [{
                'covered': 0,
                'total': 2,
                'name': 'line'
            }],
            'files': [{
                'path': '//a/myfile.java',
                'name': 'myfile.java',
                'summaries': [{
                    'covered': 0,
                    'total': 2,
                    'name': 'line'
                }]
            }]
        })
    self.assertEqual(
        entity2.data, {
            'dirs': [{
                'path': '//a/',
                'name': 'a/',
                'summaries': [{
                    'covered': 0,
                    'total': 2,
                    'name': 'line'
                }]
            }],
            'path': '//',
            'summaries': [{
                'covered': 0,
                'total': 2,
                'name': 'line'
            }],
            'files': []
        })

  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(author_coverage, 'GetV2Build')
  def testPostsubmitReportGetsCreated(self, mock_get_build,
                                      mock_get_file_content):
    CoverageReportModifier(author='abc@google.com', id=123).put()
    postsubmit_report = PostsubmitReport.Create(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        bucket='ci',
        builder='android-code-coverage',
        commit_timestamp=datetime(2020, 1, 7),
        manifest=[
            DependencyRepository(
                path='//',
                server_host='chromium.googlesource.com',
                project='chromium/src.git',
                revision='latest')
        ],
        summary_metrics={},
        build_id=2000,
        visible=True)
    postsubmit_report.put()
    file_coverage_data = FileCoverageData.Create(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        path='//a/myfile.java',
        bucket='ci',
        builder='android-code-coverage',
        data={
            'lines': [{
                'count': 0,
                'first': 1,
                'last': 3
            }, {
                'count': 10,
                'first': 4,
                'last': 5
            }]
        })
    file_coverage_data.put()
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.builder.project = 'chrome'
    build.builder.bucket = 'coverage'
    build.builder.builder = 'android-code-coverage'
    build.output.properties.items.return_value = [
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'postsubmit/chromium.googlesource.com/chromium/src/'
            'aaaaa/coverage/android-code-coverage/123456789/metadata',
        ]), ('mimic_builder_names', ['android-code-coverage'])
    ]
    build.input.gitiles_commit = mock.Mock(
        host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        id='123')
    mock_get_build.return_value = build
    blamelist = {'//a/myfile.java': {'abc@google.com': ['1', '2']}}
    chromium_to_googler_mapping = {'def@chromium.org': 'def@google.com'}
    mock_get_file_content.side_effect = [
        json.dumps(chromium_to_googler_mapping),
        json.dumps(blamelist)
    ]

    author_coverage.CreateAuthorCoverage('android-code-coverage')

    entity = PostsubmitReport.Get(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        bucket='ci',
        builder='android-code-coverage',
        modifier_id=123)
    self.assertEqual(entity.summary_metrics, [{
        'covered': 0,
        'total': 2,
        'name': 'line'
    }])
