# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from gae_libs.pipeline_wrapper import BasePipeline
from model.wf_analysis import WfAnalysis
from model.wf_try_job_data import WfTryJobData
from pipelines.pipeline_inputs_and_outputs import BuildKey
from pipelines.pipeline_inputs_and_outputs import (
    RevertAndNotifyCulpritPipelineInput)
from pipelines.test_failure.revert_and_notify_test_culprit_pipeline import (
    RevertAndNotifyTestCulpritPipeline)
from services import build_failure_analysis
from services import git
from services.test_failure import test_try_job


class IdentifyTestTryJobCulpritPipeline(BasePipeline):
  """A pipeline to identify culprit CL info based on try job results."""

  # Arguments number differs from overridden method - pylint: disable=W0221
  def run(self, master_name, builder_name, build_number, try_job_id, result):
    """Identifies the information for failed revisions.

    Please refer to try_job_result_format.md for format check.
    """
    culprits = None
    flaky_failures = {}
    if try_job_id and result and result.get('report'):
      culprit_map, failed_revisions = (
          test_try_job.FindCulpritForEachTestFailure(result))
      culprits = git.GetCLInfo(failed_revisions)

      if not culprits:
        flaky_failures = result.get('report', {}).get('flakes')
      if culprits:
        try_job_data = WfTryJobData.Get(try_job_id)
        test_try_job.UpdateCulpritMapWithCulpritInfo(culprit_map, culprits)
        try_job_data.culprits = test_try_job.GetCulpritDataForTest(culprit_map)
        try_job_data.put()
        result['culprit'] = culprit_map

    # Store try-job results.
    test_try_job.UpdateTryJobResult(master_name, builder_name, build_number,
                                    result, try_job_id, culprits)

    # Saves cls found by heuristic approach for later use.
    # This part must be before UpdateWfAnalysisWithTryJobResult().
    analysis = WfAnalysis.Get(master_name, builder_name, build_number)
    heuristic_cls = build_failure_analysis.GetHeuristicSuspectedCLs(analysis)

    # Add try-job results to WfAnalysis.
    test_try_job.UpdateWfAnalysisWithTryJobResult(master_name, builder_name,
                                                  build_number, result,
                                                  culprits, flaky_failures)

    # TODO (chanli): Update suspected_cl for builds in the same group with
    # current build.
    # Updates suspected_cl.
    test_try_job.UpdateSuspectedCLs(master_name, builder_name, build_number,
                                    culprits, result)
    if not culprits:
      return

    yield RevertAndNotifyTestCulpritPipeline(
        RevertAndNotifyCulpritPipelineInput(
            build_key=BuildKey(
                master_name=master_name,
                builder_name=builder_name,
                build_number=build_number),
            culprits=git.GetCLKeysFromCLInfo(culprits),
            heuristic_cls=heuristic_cls))
