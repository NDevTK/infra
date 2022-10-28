# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import webapp2

import gae_ts_mon

from gae_libs import appengine_util
from handlers.flake.detection import detect_flakes
from handlers.flake.detection import process_flakes
from handlers.flake.detection import update_flake_counts
from handlers.flake.reporting import generate_report


# "flake-detection-backend" module.
flake_detection_backend_web_pages_handler_mappings = [
    ('/flake/detection/cron/detect-hidden-flakes',
     detect_flakes.DetectHiddenFlakesCronJob),
    ('/flake/detection/cron/detect-non-hidden-flakes',
     detect_flakes.DetectNonHiddenFlakesCronJob),
    ('/flake/detection/cron/generate-flakiness-report',
     generate_report.PrepareFlakinessReport),
    ('/flake/detection/cron/process-flakes',
     process_flakes.ProcessFlakesCronJob),
    ('/flake/detection/cron/update-flake-counts',
     update_flake_counts.UpdateFlakeCountsCron),
    ('/flake/detection/task/detect-flakes', detect_flakes.FlakeDetection),
    ('/flake/detection/task/detect-flakes-from-build',
     detect_flakes.DetectFlakesFromFlakyCQBuild),
    ('/flake/detection/task/process-flakes', process_flakes.FlakeAutoAction),
    ('/flake/detection/task/update-flake-counts',
     update_flake_counts.UpdateFlakeCountsTask),
]
flake_detection_backend_web_application = webapp2.WSGIApplication(
    flake_detection_backend_web_pages_handler_mappings, debug=False)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(flake_detection_backend_web_application)

