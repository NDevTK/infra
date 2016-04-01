# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from model import wf_analysis_status

# Additional status for swarming tasks and try jobs.
NO_SWARMING_TASK_FOUND = 110
NON_SWARMING_NO_RERUN = 120
#Additional reasons for no try job information.
SWARMING_TASK_PENDING = 130
SWARMING_TASK_RUNNING = 140
SWARMING_TASK_ERROR = 150
NO_FAILURE_RESULT_MAP = 160
FLAKY = 200

NO_TRY_JOB_REASON_MAP = {
    NO_SWARMING_TASK_FOUND: NO_SWARMING_TASK_FOUND,
    NON_SWARMING_NO_RERUN: NON_SWARMING_NO_RERUN,
    NO_FAILURE_RESULT_MAP: NO_FAILURE_RESULT_MAP,
    wf_analysis_status.PENDING: SWARMING_TASK_PENDING,
    wf_analysis_status.ANALYZING: SWARMING_TASK_RUNNING,
    wf_analysis_status.ERROR: SWARMING_TASK_ERROR,
}

STATUS_MESSAGE_MAP = {
    wf_analysis_status.PENDING: 'Try job is pending.',
    wf_analysis_status.ANALYZING: 'Try job is running.',
    wf_analysis_status.ANALYZED: 'Not Found.',
    wf_analysis_status.ERROR: 'Try job failed.',
    NO_SWARMING_TASK_FOUND: 'No swarming task found, hence no try job.',
    NON_SWARMING_NO_RERUN: ('No swarming task nor try job will be triggered'
                            ' for non-swarming steps.'),
    SWARMING_TASK_PENDING: 'Swarming task is pending, no try job yet.',
    SWARMING_TASK_RUNNING: 'Swarming task is running, no try job yet.',
    SWARMING_TASK_ERROR: (
        'Swarming task failed, try job will not be triggered.'),
    NO_FAILURE_RESULT_MAP: 'No swarming task nor try job was triggered.',
    FLAKY: 'Flaky tests.',
}
