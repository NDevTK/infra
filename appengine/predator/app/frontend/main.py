# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import import_utils

import_utils.FixImports()

import google.cloud.logging

client = google.cloud.logging.Client()
client.setup_logging()

from flask import Flask
from google.appengine.api import wrap_wsgi_app

from frontend.handlers import clusterfuzz_dashboard
from frontend.handlers import clusterfuzz_result_feedback
from frontend.handlers import clusterfuzz_public_dashboard
from frontend.handlers import cracas_dashboard
from frontend.handlers import cracas_result_feedback
from frontend.handlers import crash_config
from frontend.handlers import crash_handler
from frontend.handlers import fracas_dashboard
from frontend.handlers import fracas_result_feedback
from frontend.handlers import triage_analysis
from frontend.handlers import uma_sampling_profiler_dashboard
from frontend.handlers import uma_sampling_profiler_result_feedback


frontend_web_pages_handler_mappings = [
    ('/clusterfuzz/dashboard', 'clusterfuzz_dashboard',
     clusterfuzz_dashboard.ClusterfuzzDashBoard().Handle, ['GET']),
    ('/clusterfuzz/public-dashboard', 'clusterfuzz_public_dashboard',
     clusterfuzz_public_dashboard.ClusterfuzzPublicDashBoard().Handle, ['GET']),
    ('/uma-sampling-profiler/dashboard', 'uma_sampling_profiler_dashboard',
     uma_sampling_profiler_dashboard.UMASamplingProfilerDashboard().Handle,
     ['GET']),
    ('/cracas/dashboard', 'cracas_dashboard',
     cracas_dashboard.CracasDashBoard().Handle, ['GET']),
    ('/clusterfuzz/result-feedback', 'clusterfuzz_result_feedback',
     clusterfuzz_result_feedback.ClusterfuzzResultFeedback().Handle, ['GET']),
    ('/uma-sampling-profiler/result-feedback',
     'uma_sampling_profiler_result_feedback',
     uma_sampling_profiler_result_feedback.UMASamplingProfilerResultFeedback(
     ).Handle, ['GET']),
    ('/cracas/result-feedback', 'cracas_result_feedback',
     cracas_result_feedback.CracasResultFeedback().Handle, ['GET']),
    ('/config', 'config', crash_config.CrashConfig().Handle, ['GET', 'POST'])
]

triage_analysis_view = triage_analysis.TriageAnalysis().Handle
frontend_web_pages_handler_mappings += [
    ('/clusterfuzz/triage-analysis', 'clusterfuzz_triage_analysis',
     triage_analysis_view, ['POST']),
    ('/cracas/triage-analysis', 'cracas_triage_analysis', triage_analysis_view,
     ['POST']),
    ('/uma-sampling-profiler/triage-analysis',
     'uma_sampling_profiler_triage_analysis', triage_analysis_view, ['POST']),
]

crash_handler_view = crash_handler.CrashHandler().Handle
frontend_web_pages_handler_mappings += [
    ('/_ah/push-handlers/crash/cracas', 'cracas_crash_handler',
     crash_handler_view, ['POST']),
    ('/_ah/push-handlers/crash/clusterfuzz', 'clusterfuzz_crash_handler',
     crash_handler_view, ['POST']),
    ('/_ah/push-handlers/regression/uma-sampling-profiler',
     'uma_sampling_profiler_handler', crash_handler_view, ['POST']),
]

frontend_app = Flask(__name__)
frontend_app.wsgi_app = wrap_wsgi_app(frontend_app.wsgi_app)

for url, endpoint, view_func, methods in frontend_web_pages_handler_mappings:
  frontend_app.add_url_rule(
      url, endpoint=endpoint, view_func=view_func, methods=methods)
