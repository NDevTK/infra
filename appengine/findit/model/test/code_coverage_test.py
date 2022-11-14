# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import datetime

from google.appengine.api import datastore_errors
from google.appengine.ext import ndb

from model.code_coverage import CoveragePercentage
from model.code_coverage import CoverageReportModifier
from model.code_coverage import DependencyRepository
from model.code_coverage import FileCoverageData
from model.code_coverage import GitilesCommit
from model.code_coverage import PostsubmitReport
from model.code_coverage import PresubmitCoverageData
from model.code_coverage import SummaryCoverageData
from waterfall.test.wf_testcase import WaterfallTestCase


class CodeCoverageTest(WaterfallTestCase):

  def testCreateAndGetPostsubmitReport(self):
    server_host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    revision = '99999'
    bucket = 'coverage'
    builder = 'linux-code-coverage'
    commit_position = 100
    commit_timestamp = datetime.datetime(2018, 1, 1)

    manifest = [
        DependencyRepository(
            path='//src',
            server_host='chromium.googlesource.com',
            project='chromium/src.git',
            revision='88888')
    ]

    summary_metrics = [{
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

    build_id = 123456789
    visible = True
    modifier_id = 123

    report = PostsubmitReport.Create(
        server_host=server_host,
        project=project,
        ref=ref,
        revision=revision,
        bucket=bucket,
        builder=builder,
        commit_position=commit_position,
        commit_timestamp=commit_timestamp,
        manifest=manifest,
        summary_metrics=summary_metrics,
        build_id=build_id,
        visible=visible,
        modifier_id=modifier_id)
    report.put()

    # Test key.
    self.assertEqual(
        'chromium.googlesource.com$chromium/src$refs/heads/main$99999$'
        'coverage$linux-code-coverage$123', report.key.id())

    # Test Create.
    fetched_reports = PostsubmitReport.query().fetch()
    self.assertEqual(1, len(fetched_reports))
    self.assertEqual(report, fetched_reports[0])

    # Test Get.
    self.assertEqual(
        report,
        PostsubmitReport.Get(
            server_host=server_host,
            project=project,
            ref=ref,
            revision=revision,
            bucket=bucket,
            builder=builder,
            modifier_id=modifier_id))

  def testCreateAndGetPostsubmitReport_DefaultModifierId(self):
    server_host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    revision = '99999'
    bucket = 'coverage'
    builder = 'linux-code-coverage'
    commit_position = 100
    commit_timestamp = datetime.datetime(2018, 1, 1)

    manifest = [
        DependencyRepository(
            path='//src',
            server_host='chromium.googlesource.com',
            project='chromium/src.git',
            revision='88888')
    ]

    summary_metrics = [{
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

    build_id = 123456789
    visible = True

    report = PostsubmitReport.Create(
        server_host=server_host,
        project=project,
        ref=ref,
        revision=revision,
        bucket=bucket,
        builder=builder,
        commit_position=commit_position,
        commit_timestamp=commit_timestamp,
        manifest=manifest,
        summary_metrics=summary_metrics,
        build_id=build_id,
        visible=visible)
    report.put()

    # Test key.
    self.assertEqual(
        'chromium.googlesource.com$chromium/src$refs/heads/main$99999$'
        'coverage$linux-code-coverage$0', report.key.id())

    # Test Create.
    fetched_reports = PostsubmitReport.query().fetch()
    self.assertEqual(1, len(fetched_reports))
    self.assertEqual(report, fetched_reports[0])

    # Test Get.
    self.assertEqual(
        report,
        PostsubmitReport.Get(
            server_host=server_host,
            project=project,
            ref=ref,
            revision=revision,
            bucket=bucket,
            builder=builder))

  def testGetPostsubmitReport_LegacyKey(self):
    server_host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    revision = '99999'
    bucket = 'coverage'
    builder = 'linux-code-coverage'
    commit_position = 100
    commit_timestamp = datetime.datetime(2018, 1, 1)

    manifest = [
        DependencyRepository(
            path='//src',
            server_host='chromium.googlesource.com',
            project='chromium/src.git',
            revision='88888')
    ]

    summary_metrics = [{
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

    build_id = 123456789
    visible = True
    legacy_id = '%s$%s$%s$%s$%s$%s' % (server_host, project, ref, revision,
                                       bucket, builder)
    report = PostsubmitReport(
        id=legacy_id,
        gitiles_commit=GitilesCommit(
            server_host=server_host,
            project=project,
            ref=ref,
            revision=revision),
        bucket=bucket,
        builder=builder,
        commit_position=commit_position,
        commit_timestamp=commit_timestamp,
        manifest=manifest,
        summary_metrics=summary_metrics,
        build_id=build_id,
        visible=visible)
    report.put()

    # Test Get.
    self.assertEqual(
        report,
        PostsubmitReport.Get(
            server_host=server_host,
            project=project,
            ref=ref,
            revision=revision,
            bucket=bucket,
            builder=builder))

  def testCreateAndGetCoveragePercentage(self):
    path = '//base/test.cc'
    total_lines = 10
    covered_lines = 1

    percentage = CoveragePercentage(
        path=path, total_lines=total_lines, covered_lines=covered_lines)
    self.assertEqual(path, percentage.path)
    self.assertEqual(total_lines, percentage.total_lines)
    self.assertEqual(covered_lines, percentage.covered_lines)

  def testCreateCoveragePercentageRaiseError(self):
    with self.assertRaises(datastore_errors.BadValueError):
      CoveragePercentage(path='//base/test.cc', total_lines=0, covered_lines=0)

  def testCreateAndGetPresubmitCoverageData(self):
    server_host = 'chromium-review.googlesource.com'
    change = 138000
    patchset = 4
    data = [
        {
            'path': '//base1/test1.cc',
            'lines': [{
                'count': 100,
                'first': 1,
                'last': 3,
            }],
        },
        {
            'path': '//dir2/test2.cc',
            'lines': [{
                'count': 0,
                'first': 5,
                'last': 10,
            }],
        },
    ]

    data_unit = [
        {
            'path': '//base1/test1.cc',
            'lines': [{
                'count': 100,
                'first': 2,
                'last': 3,
            }],
        },
    ]

    coverage_data = PresubmitCoverageData.Create(
        server_host=server_host,
        change=change,
        patchset=patchset,
        data=data,
        data_unit=data_unit)
    coverage_data.put()

    # Test key.
    self.assertEqual('chromium-review.googlesource.com$138000$4',
                     coverage_data.key.id())

    # Test Create.
    fetched_coverage_data = PresubmitCoverageData.query().fetch()
    self.assertEqual(1, len(fetched_coverage_data))
    self.assertEqual(coverage_data, fetched_coverage_data[0])

    # Test Get.
    self.assertEqual(
        coverage_data,
        PresubmitCoverageData.Get(
            server_host=server_host, change=change, patchset=patchset))

  def testCreateAndGetFileCoverageData(self):
    server_host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    revision = '99999'
    path = '//dir/test.cc'
    bucket = 'coverage'
    builder = 'linux-code-coverage'
    modifier_id = 123
    data = {
        'path': '//dir/test.cc',
        'lines': [{
            'count': 100,
            'first': 1,
            'last': 5,
        }],
        'timestamp': 1357,
        'revision': '12345'
    }

    file_coverage_data = FileCoverageData.Create(
        server_host=server_host,
        project=project,
        ref=ref,
        revision=revision,
        path=path,
        bucket=bucket,
        builder=builder,
        modifier_id=modifier_id,
        data=data)
    file_coverage_data.put()

    # Test key.
    self.assertEqual(
        ('chromium.googlesource.com$chromium/src$refs/heads/main$99999$'
         '//dir/test.cc$coverage$linux-code-coverage$123'),
        file_coverage_data.key.id())

    # Test Create.
    fetched_file_coverage_data = FileCoverageData.query().fetch()
    self.assertEqual(1, len(fetched_file_coverage_data))
    self.assertEqual(file_coverage_data, fetched_file_coverage_data[0])

    # Test Get.
    self.assertEqual(
        file_coverage_data,
        FileCoverageData.Get(
            server_host=server_host,
            project=project,
            ref=ref,
            revision=revision,
            path=path,
            bucket=bucket,
            builder=builder,
            modifier_id=modifier_id))

  def testCreateAndGetFileCoverageData_DefaultModifierId(self):
    server_host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    revision = '99999'
    path = '//dir/test.cc'
    bucket = 'coverage'
    builder = 'linux-code-coverage'
    data = {
        'path': '//dir/test.cc',
        'lines': [{
            'count': 100,
            'first': 1,
            'last': 5,
        }],
        'timestamp': 1357,
        'revision': '12345'
    }

    file_coverage_data = FileCoverageData.Create(
        server_host=server_host,
        project=project,
        ref=ref,
        revision=revision,
        path=path,
        bucket=bucket,
        builder=builder,
        data=data)
    file_coverage_data.put()

    # Test key.
    self.assertEqual(
        ('chromium.googlesource.com$chromium/src$refs/heads/main$99999$'
         '//dir/test.cc$coverage$linux-code-coverage$0'),
        file_coverage_data.key.id())

    # Test Create.
    fetched_file_coverage_data = FileCoverageData.query().fetch()
    self.assertEqual(1, len(fetched_file_coverage_data))
    self.assertEqual(file_coverage_data, fetched_file_coverage_data[0])

    # Test Get.
    self.assertEqual(
        file_coverage_data,
        FileCoverageData.Get(
            server_host=server_host,
            project=project,
            ref=ref,
            revision=revision,
            path=path,
            bucket=bucket,
            builder=builder))

  def testGetModifiedLineCoverage(self):
    file_coverage_data = FileCoverageData.Create(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        revision='latest',
        path='//a/myfile.cc',
        bucket='ci',
        builder='linux-code-coverage',
        data={
            'path': '//a/myfile.cc',
            'lines': [{
                'first': 1,
                'last': 10,
                'count': 10
            }],
            'summaries': [{
                'name': 'line',
                'total': 10,
                'covered': 10
            }],
            'revision': 'latest'
        })

    expected = {
        'path': '//a/myfile.cc',
        'lines': [{
            'first': 4,
            'last': 5,
            'count': 10
        }, {
            'first': 7,
            'last': 7,
            'count': 10
        }],
        'summaries': [{
            'name': 'line',
            'total': 3,
            'covered': 3
        }],
        'revision': 'latest'
    }
    actual = FileCoverageData.GetModifiedLineCoverage(file_coverage_data,
                                                      [4, 5, 7])

    self.assertDictEqual(actual, expected)

  def testGetFileCoverageData_LegacyKey(self):
    server_host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    revision = '99999'
    path = '//dir/test.cc'
    bucket = 'coverage'
    builder = 'linux-code-coverage'
    data = {
        'path': '//dir/test.cc',
        'lines': [{
            'count': 100,
            'first': 1,
            'last': 5,
        }],
        'timestamp': 1357,
        'revision': '12345'
    }
    legacy_id = '%s$%s$%s$%s$%s$%s$%s' % (server_host, project, ref, revision,
                                          path, bucket, builder)
    file_coverage_data = FileCoverageData(
        id=legacy_id,
        gitiles_commit=GitilesCommit(
            server_host=server_host,
            project=project,
            ref=ref,
            revision=revision),
        path=path,
        bucket=bucket,
        builder=builder,
        data=data)
    file_coverage_data.put()

    # Test Get.
    self.assertEqual(
        file_coverage_data,
        FileCoverageData.Get(
            server_host=server_host,
            project=project,
            ref=ref,
            revision=revision,
            path=path,
            bucket=bucket,
            builder=builder))

  def testAndCreateAndGetSummaryCoverageData(self):
    server_host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    revision = '99999'
    data_type = 'dirs'
    path = '//dir/'
    bucket = 'coverage'
    builder = 'linux-code-coverage'
    data = {
        'dirs': [],
        'files': [],
        'summaries': [{
            'covered': 1,
            'total': 1,
            'name': 'region'
        }, {
            'covered': 1,
            'total': 1,
            'name': 'function'
        }, {
            'covered': 1,
            'total': 1,
            'name': 'line'
        }],
        'path':
            '//dir/',
    }
    modifier_id = 123
    dir_coverage_data = SummaryCoverageData.Create(
        server_host=server_host,
        project=project,
        ref=ref,
        revision=revision,
        data_type=data_type,
        path=path,
        bucket=bucket,
        builder=builder,
        data=data,
        modifier_id=modifier_id)
    dir_coverage_data.put()

    # Test key.
    self.assertEqual(
        'chromium.googlesource.com$chromium/src$refs/heads/main$99999$'
        'dirs$//dir/$coverage$linux-code-coverage$123',
        dir_coverage_data.key.id())

    # Test Create.
    fetched_dir_coverage_data = SummaryCoverageData.query().fetch()
    self.assertEqual(1, len(fetched_dir_coverage_data))
    self.assertEqual(dir_coverage_data, fetched_dir_coverage_data[0])

    # Test Get.
    self.assertEqual(
        dir_coverage_data,
        SummaryCoverageData.Get(
            server_host=server_host,
            project=project,
            ref=ref,
            revision=revision,
            data_type=data_type,
            path=path,
            bucket=bucket,
            builder=builder,
            modifier_id=modifier_id))

  def testAndCreateAndGetSummaryCoverageData_DefaultModifierId(self):
    server_host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    revision = '99999'
    data_type = 'dirs'
    path = '//dir/'
    bucket = 'coverage'
    builder = 'linux-code-coverage'
    data = {
        'dirs': [],
        'files': [],
        'summaries': [{
            'covered': 1,
            'total': 1,
            'name': 'region'
        }, {
            'covered': 1,
            'total': 1,
            'name': 'function'
        }, {
            'covered': 1,
            'total': 1,
            'name': 'line'
        }],
        'path': '//dir/',
    }
    dir_coverage_data = SummaryCoverageData.Create(
        server_host=server_host,
        project=project,
        ref=ref,
        revision=revision,
        data_type=data_type,
        path=path,
        bucket=bucket,
        builder=builder,
        data=data)
    dir_coverage_data.put()

    # Test key.
    self.assertEqual(
        'chromium.googlesource.com$chromium/src$refs/heads/main$99999$'
        'dirs$//dir/$coverage$linux-code-coverage$0',
        dir_coverage_data.key.id())

    # Test Create.
    fetched_dir_coverage_data = SummaryCoverageData.query().fetch()
    self.assertEqual(1, len(fetched_dir_coverage_data))
    self.assertEqual(dir_coverage_data, fetched_dir_coverage_data[0])

    # Test Get.
    self.assertEqual(
        dir_coverage_data,
        SummaryCoverageData.Get(
            server_host=server_host,
            project=project,
            ref=ref,
            revision=revision,
            data_type=data_type,
            path=path,
            bucket=bucket,
            builder=builder))

  def testGetSummaryCoverageData_LegacyKey(self):
    server_host = 'chromium.googlesource.com'
    project = 'chromium/src'
    ref = 'refs/heads/main'
    revision = '99999'
    data_type = 'dirs'
    path = '//dir/'
    bucket = 'coverage'
    builder = 'linux-code-coverage'
    data = {
        'dirs': [],
        'files': [],
        'summaries': [{
            'covered': 1,
            'total': 1,
            'name': 'region'
        }, {
            'covered': 1,
            'total': 1,
            'name': 'function'
        }, {
            'covered': 1,
            'total': 1,
            'name': 'line'
        }],
        'path': '//dir/',
    }

    legacy_id = '%s$%s$%s$%s$%s$%s$%s$%s' % (
        server_host, project, ref, revision, data_type, path, bucket, builder)
    dir_coverage_data = SummaryCoverageData(
        id=legacy_id,
        gitiles_commit=GitilesCommit(
            server_host=server_host,
            project=project,
            ref=ref,
            revision=revision),
        data_type=data_type,
        path=path,
        bucket=bucket,
        builder=builder,
        data=data)
    dir_coverage_data.put()

    # Test Get.
    self.assertEqual(
        dir_coverage_data,
        SummaryCoverageData.Get(
            server_host=server_host,
            project=project,
            ref=ref,
            revision=revision,
            data_type=data_type,
            path=path,
            bucket=bucket,
            builder=builder))

  def testGetCoverageReportModifier(self):
    CoverageReportModifier(gerrit_hashtag='my_feature', id=123).put()
    self.assertEqual(
        CoverageReportModifier.Get(123).gerrit_hashtag, 'my_feature')

  def testInsertModifierForAuthor(self):
    CoverageReportModifier.InsertModifierForAuthorIfNeeded(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        author='xyz@google.com')
    CoverageReportModifier.InsertModifierForAuthorIfNeeded(
        server_host='chromium.googlesource.com',
        project='chromium/src',
        author='xyz@google.com')
    report_modifiers = CoverageReportModifier.query().fetch()
    self.assertEqual(1, len(report_modifiers))
    self.assertTrue(report_modifiers[0].is_active)
