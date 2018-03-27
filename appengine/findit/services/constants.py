# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

INVALID_FAILURE_LOG = 'invalid'
FLAKY_FAILURE_LOG = 'flaky'
WRONG_FORMAT_LOG = 'not in desired test result formats'

# Swarming task states.
STATE_PENDING = 'PENDING'
STATE_RUNNING = 'RUNNING'
STATE_COMPLETED = 'COMPLETED'
STATE_NOT_STOP = (STATE_PENDING, STATE_RUNNING)

# TODO(crbug.com/785463): Use enum for error codes.

# Swarming task stopped error codes.
BOT_DIED = 30
CANCELED = 40
EXPIRED = 50
TIMED_OUT = 60

STATES_NOT_RUNNING_TO_ERROR_CODES = {
    'BOT_DIED': BOT_DIED,
    'CANCELED': CANCELED,
    'EXPIRED': EXPIRED,
    'TIMED_OUT': TIMED_OUT,
}

# Error codes when getting results of a task.
# Outputs_ref is None.
NO_TASK_OUTPUTS = 300
# Unable to retrieve output json.
NO_OUTPUT_JSON = 320
# Other/miscellaneous error codes.
UNKNOWN = 1000
# Unable to recognize the format of output json.
UNRECOGNIZABLE = 10

ERROR_CODE_TO_MESSAGE = {
    TIMED_OUT: 'Process swarming task result timed out',
    NO_TASK_OUTPUTS: 'outputs_ref is None',
    NO_OUTPUT_JSON: 'No swarming task failure log',
    UNKNOWN: 'Unknown error',
    UNRECOGNIZABLE: 'Test results format is unrecognized, cannot find a parser.'
}


def GenerateError(code):
  return {
      'code': code,
      'message': ERROR_CODE_TO_MESSAGE.get(code, ERROR_CODE_TO_MESSAGE[UNKNOWN])
  }
