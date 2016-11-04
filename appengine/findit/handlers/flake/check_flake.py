# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from google.appengine.api import users

from common import auth_util
from common.base_handler import BaseHandler
from common.base_handler import Permission
from lib import time_util
from model import analysis_status
from model import triage_status
from model.flake.flake_analysis_request import FlakeAnalysisRequest
from model.flake.master_flake_analysis import MasterFlakeAnalysis
from waterfall.flake import flake_analysis_service
from waterfall.flake import triggering_sources


def _GetSuspectedFlakeAnalysisAndTriageResult(analysis):
  if analysis.suspected_flake_build_number is not None:
    return {
        'build_number': analysis.suspected_flake_build_number,
        'triage_result': (
            analysis.triage_history[-1].triage_result if analysis.triage_history
            else triage_status.UNTRIAGED)
    }

  return {}


class CheckFlake(BaseHandler):
  PERMISSION_LEVEL = Permission.ANYONE

  def _ValidateInput(self, master_name, builder_name, build_number, step_name,
                     test_name, bug_id):
    """Ensures the input is valid and generates an error otherwise.

    Args:
      master_name (str): The name of the master the flaky test was found on.
      builder_name (str): The name of the builder the flaky test was found on.
      build_number (str): The build number the flaky test was found on.
      step_name (str): The step the flaky test was found on.
      test_name (str): The name of the flaky test.
      bug_id (str): The bug number associated with the flaky test.

    Returns:
      None if all input fields are valid, or an error dict otherwise.
    """

    if not master_name:
      return self.CreateError('Master name must be specified', 400)

    if not builder_name:
      return self.CreateError('Builder name must be specified', 400)

    if not build_number or not build_number.isdigit():
      return self.CreateError('Build number must be specified as an int', 400)

    if not step_name:
      return self.CreateError('Step name must be specified', 400)

    if not test_name:
      return self.CreateError('Test name must be specified', 400)

    if bug_id and not bug_id.isdigit():
      return self.CreateError('Bug id (optional) must be an int', 400)

    return None

  def HandleGet(self):
    master_name = self.request.get('master_name', '').strip()
    builder_name = self.request.get('builder_name', '').strip()
    build_number = self.request.get('build_number', '').strip()
    step_name = self.request.get('step_name', '').strip()
    test_name = self.request.get('test_name', '').strip()
    bug_id = self.request.get('bug_id', '').strip()
    # TODO(lijeffrey): Add support for force flag to trigger a rerun.

    error = self._ValidateInput(
        master_name, builder_name, build_number, step_name, test_name, bug_id)

    if error:  # pragma: no cover
      return error

    build_number = int(build_number)
    bug_id = int(bug_id) if bug_id else None
    user_email = auth_util.GetUserEmail()
    is_admin = auth_util.IsCurrentUserAdmin()

    request = FlakeAnalysisRequest.Create(test_name, False, bug_id)
    request.AddBuildStep(master_name, builder_name, build_number, step_name,
                         time_util.GetUTCNow())
    scheduled = flake_analysis_service.ScheduleAnalysisForFlake(
        request, user_email, is_admin, triggering_sources.FINDIT_UI)

    analysis = MasterFlakeAnalysis.GetVersion(
        master_name, builder_name, build_number, step_name, test_name)

    if not analysis:
      if scheduled is None:
        # User does not have permission to trigger, nor was any previous
        # analysis triggered to view.
        return {
            'template': 'error.html',
            'data': {
                'error_message':
                    ('You could schedule an analysis for flaky test only after '
                     'you login with google.com account.'),
                'login_url': self.GetLoginUrl(),
            },
            'return_code': 401,
        }

      # Check if a previous request has already covered this analysis so use the
      # results from that analysis.
      request = FlakeAnalysisRequest.GetVersion(key=test_name)

      if request and request.analyses:
        analysis = request.analyses[-1].get()
      else:
        return {
            'template': 'error.html',
            'data': {
                'error_message': (
                    'Flake analysis is not supported for this request. Either '
                    'the build step may not be supported or the test is not '
                    'swarmed.'),
            },
            'return_code': 401,
        }

    suspected_flake = _GetSuspectedFlakeAnalysisAndTriageResult(analysis)

    data = {
        'pass_rates': [],
        'analysis_status': analysis.status_description,
        'master_name': master_name,
        'builder_name': builder_name,
        'build_number': build_number,
        'step_name': step_name,
        'test_name': test_name,
        'version_number': analysis.version_number,
        'suspected_flake': suspected_flake,
        'request_time': time_util.FormatDatetime(
            analysis.request_time),
        'task_number': len(analysis.data_points),
        'error': analysis.error_message,
        'iterations_to_rerun': analysis.iterations_to_rerun,
        'show_debug_info': self._ShowDebugInfo()
    }

    if (users.is_current_user_admin() and analysis.completed and
        analysis.triage_history):
      data['triage_history'] = analysis.GetTriageHistory()

    data['pending_time'] = time_util.FormatDuration(
        analysis.request_time,
        analysis.start_time or time_util.GetUTCNow())
    if analysis.status != analysis_status.PENDING:
      data['duration'] = time_util.FormatDuration(
          analysis.start_time,
          analysis.end_time or time_util.GetUTCNow())

    coordinates = []
    for data_point in analysis.data_points:
      coordinates.append([data_point.build_number, data_point.pass_rate])

    # Order by build number from earliest to latest.
    coordinates.sort(key=lambda x: x[0])

    data['pass_rates'] = coordinates
    return {
        'template': 'flake/result.html',
        'data': data
    }
