# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Deprecated. Use gae_libs/base_pipeline.py instead."""

import pipeline as pipeline
from pipeline import handlers as pipeline_handlers
from pipeline import status_ui as pipeline_status_ui


# TODO(stgao): remove BasePipeline after http://crrev.com/810193002 is landed.
class BasePipeline(pipeline.Pipeline):  # pragma: no cover

  def send_result_email(self):
    """We override this so it doesn't email on completion."""
    pass

  def run_test(self, *args, **kwargs):
    pass

  def finalized_test(self, *args, **kwargs):
    pass

  def callback(self, **kwargs):
    pass

  def pipeline_status_path(self):
    """Returns an absolute path to look up the status of the pipeline."""
    return '/_ah/pipeline/status?root=%s&auto=false' % self.root_pipeline_id

  def run(self, *args, **kwargs):
    raise NotImplementedError()
