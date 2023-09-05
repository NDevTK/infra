# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import gae_ts_mon

outgoing_http_errors = gae_ts_mon.CounterMetric(
    'findit/outgoinghttperrors', 'Failed http requests to various servers',
    [gae_ts_mon.StringField('host'),
     gae_ts_mon.StringField('exception')])

outgoing_http_statuses = gae_ts_mon.CounterMetric(
    'findit/outgoinghttpstatuses', 'Http requests to external services',
    [gae_ts_mon.StringField('host'),
     gae_ts_mon.StringField('status_code')])

code_coverage_cq_errors = gae_ts_mon.CounterMetric(
    'code-coverage/cq-bot-errors',
    'Number of cq builds with coverage data step failures', [
        gae_ts_mon.StringField('project'),
        gae_ts_mon.StringField('bucket'),
        gae_ts_mon.StringField('builder'),
    ])

code_coverage_full_reports = gae_ts_mon.CounterMetric(
    'code-coverage/full-reports',
    'Number of whole-codebase coverage reports',
    [
        gae_ts_mon.StringField('host'),
        gae_ts_mon.StringField('project'),  # Gerrit project.
        gae_ts_mon.StringField('ref'),
        gae_ts_mon.StringField('builder'),  # <luci_project>/<bucket>/<builder>
    ])

code_coverage_report_timestamp = gae_ts_mon.GaugeMetric(
    'code-coverage/report_timestamp',
    'Timestamp of the completion of the last report',
    [
        gae_ts_mon.StringField('host'),
        gae_ts_mon.StringField('project'),  # Gerrit project.
        gae_ts_mon.StringField('ref'),
        gae_ts_mon.StringField('builder'),  # <luci_project>/<bucket>/<builder>
        gae_ts_mon.BooleanField('is_success'),
    ])
