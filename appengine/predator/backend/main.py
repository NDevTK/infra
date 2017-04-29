# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import webapp2

from backend.handlers import analyze_crash
from backend.handlers import rerun_analysis
from backend.handlers import triage_analysis
from backend.handlers import update_component_config


backend_handler_mappings = [
    ('/process/update-component-config',
     update_component_config.UpdateComponentConfig),
    ('/process/triage-analysis', triage_analysis.TriageAnalysis),
    ('/process/rerun-analysis', rerun_analysis.RerunAnalysis)
]
backend_app = webapp2.WSGIApplication(backend_handler_mappings, debug=False)
