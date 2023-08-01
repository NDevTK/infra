# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import webapp2

import gae_ts_mon

from backend.handlers import rerun_analyses
from backend.handlers import rerun_analysis
from backend.handlers import update_component_config
from backend.handlers import update_inverted_index
from backend.handlers import update_repo_to_dep_path
from gae_libs.pipeline_wrapper import pipeline_handlers


# For appengine pipeline running on backend modules.
pipeline_backend_application = pipeline_handlers._APP
# TODO(crbug.com/1322775) Migrate away from the shared prodx-mon-chrome-infra
# service account and change to gae_ts_mon.initialize_prod()
gae_ts_mon.initialize_adhoc(pipeline_backend_application)


backend_handler_mappings = [
    ('/process/update-component-config',
     update_component_config.UpdateComponentConfig),
    ('/process/rerun-analyses', rerun_analyses.RerunAnalyses),
    ('/process/rerun-analysis', rerun_analysis.RerunAnalysis),
    ('/process/update-inverted-index',
     update_inverted_index.UpdateInvertedIndex),
    ('/process/update-repo-to-dep-path',
     update_repo_to_dep_path.UpdateRepoToDepPath),
]
backend_app = webapp2.WSGIApplication(backend_handler_mappings, debug=False)
# TODO(crbug.com/1322775) Migrate away from the shared prodx-mon-chrome-infra
# service account and change to gae_ts_mon.initialize_prod()
gae_ts_mon.initialize_adhoc(backend_app)
