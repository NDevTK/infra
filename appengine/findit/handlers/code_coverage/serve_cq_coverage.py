# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import logging
import re
import six

from common.base_handler import BaseHandler, Permission
from handlers.code_coverage import utils
from model.code_coverage import PresubmitCoverageData
from services.code_coverage import code_coverage_util
from waterfall import waterfall_config


def _IsServePresubmitCoverageDataEnabled():
  """Returns True if the feature to serve presubmit coverage data is enabled.

  Returns:
    Returns True if it is enabled, otherwise, False.
  """
  # Unless the flag is explicitly set, assuming disabled by default.
  return waterfall_config.GetCodeCoverageSettings().get(
      'serve_presubmit_coverage_data', False)


class ServeCodeCoverageData(BaseHandler):
  PERMISSION_LEVEL = Permission.ANYONE

  def _ServePerCLCoverageData(self):
    """Serves per-cl coverage data.

    There are two types of requests: 'lines' and 'percentages', and the reason
    why they're separate is that:
    1. Calculating lines takes much longer than percentages, especially when
       data needs to be shared between two equivalent patchsets, while for
       percentages, it's assumed that incremental coverage percentages would be
       the same for equivalent patchsets and no extra work is needed.
    2. Percentages are usually requested much earlier than lines by the Gerrit
       plugin because the later won't be displayed until the user actually
       expands the diff view.

    The format of the returned data conforms to:
    https://chromium.googlesource.com/infra/gerrit-plugins/code-coverage/+/213d226a5f1b78c45c91d49dbe32b09c5609e9bd/src/main/resources/static/coverage.js#93
    """

    def _ServeLines(lines_data):
      """Serves lines coverage data."""
      lines_data = lines_data or []
      formatted_data = {'files': []}
      for file_data in lines_data:
        formatted_data['files'].append({
            'path':
                file_data['path'][2:],
            'lines':
                code_coverage_util.DecompressLineRanges(file_data['lines']),
        })

      return {'data': {'data': formatted_data,}, 'allowed_origin': '*'}

    def _ServePercentages(abs_coverage, inc_coverage, abs_unit_tests_coverage,
                          inc_unit_tests_coverage):
      """Serves percentages coverage data."""

      def _GetCoverageMetricsPerFile(coverage):
        coverage_per_file = {}
        for e in coverage:
          coverage_per_file[e.path] = {
              'covered': e.covered_lines,
              'total': e.total_lines,
          }
        return coverage_per_file

      abs_coverage_per_file = _GetCoverageMetricsPerFile(abs_coverage)
      inc_coverage_per_file = _GetCoverageMetricsPerFile(inc_coverage)
      abs_unit_tests_coverage_per_file = _GetCoverageMetricsPerFile(
          abs_unit_tests_coverage)
      inc_unit_tests_coverage_per_file = _GetCoverageMetricsPerFile(
          inc_unit_tests_coverage)

      formatted_data = {'files': []}
      for p in set(
          list(abs_coverage_per_file.keys()) +
          list(abs_unit_tests_coverage_per_file.keys())):
        # Do not return results for test files
        if re.match(utils.TEST_FILE_REGEX, p):
          continue
        formatted_data['files'].append({
            'path':
                p[2:],
            'absolute_coverage':
                abs_coverage_per_file.get(p, None),
            'incremental_coverage':
                inc_coverage_per_file.get(p, None),
            'absolute_unit_tests_coverage':
                abs_unit_tests_coverage_per_file.get(p, None),
            'incremental_unit_tests_coverage':
                inc_unit_tests_coverage_per_file.get(p, None),
        })

      return {'data': {'data': formatted_data,}, 'allowed_origin': '*'}

    gerrit_host = self.request.values.get('host')
    project = self.request.values.get('project')
    try:
      change = int(self.request.values.get('change'))
      patchset = int(self.request.values.get('patchset'))
    except ValueError as ve:
      return BaseHandler.CreateError(
          error_message=(
              'Invalid value for change(%r) or patchset(%r): need int, %s' %
              (self.request.values.get('change'),
               self.request.values.get('patchset'), six.text_type(ve))),
          return_code=400,
          allowed_origin='*')

    data_type = self.request.values.get('type', 'lines')

    logging.info('Serving coverage data for CL:')
    logging.info('host=%s', gerrit_host)
    logging.info('change=%d', change)
    logging.info('patchset=%d', patchset)
    logging.info('type=%s', data_type)

    configs = utils.GetAllowedGerritConfigs()
    if project not in configs.get(gerrit_host, []):
      return BaseHandler.CreateError(
          error_message='"%s/%s" is not supported.' % (gerrit_host, project),
          return_code=400,
          allowed_origin='*',
          is_project_supported=False)

    if data_type not in ('lines', 'percentages'):
      return BaseHandler.CreateError(
          error_message=(
              'Invalid type: "%s", must be "lines" (default) or "percentages"' %
              data_type),
          return_code=400,
          allowed_origin='*')

    if not _IsServePresubmitCoverageDataEnabled():
      # TODO(crbug.com/908609): Switch to 'is_service_enabled'.
      kwargs = {'is_project_supported': False}
      return BaseHandler.CreateError(
          error_message='The functionality has been temporarity disabled.',
          return_code=400,
          allowed_origin='*',
          **kwargs)

    entity = PresubmitCoverageData.Get(
        server_host=gerrit_host, change=change, patchset=patchset)
    is_serving_percentages = (data_type == 'percentages')
    if entity:
      if is_serving_percentages:
        return _ServePercentages(entity.absolute_percentages,
                                 entity.incremental_percentages,
                                 entity.absolute_percentages_unit,
                                 entity.incremental_percentages_unit)

      return _ServeLines(entity.data)

    # If coverage data of the requested patchset is not available, we check
    # previous equivalent patchsets try to reuse their data if applicable.
    equivalent_patchsets = code_coverage_util.GetEquivalentPatchsets(
        gerrit_host, project, change, patchset)
    if not equivalent_patchsets:
      return BaseHandler.CreateError(
          'Requested coverage data is not found.', 404, allowed_origin='*')

    latest_entity = None
    for ps in sorted(equivalent_patchsets, reverse=True):
      latest_entity = PresubmitCoverageData.Get(
          server_host=gerrit_host, change=change, patchset=ps)
      if latest_entity and latest_entity.based_on is None:
        break

    if latest_entity is None:
      return BaseHandler.CreateError(
          'Requested coverage data is not found.', 404, allowed_origin='*')

    if is_serving_percentages:
      return _ServePercentages(latest_entity.absolute_percentages,
                               latest_entity.incremental_percentages,
                               latest_entity.absolute_percentages_unit,
                               latest_entity.incremental_percentages_unit)

    try:
      rebased_coverage_data = \
        code_coverage_util.RebasePresubmitCoverageDataBetweenPatchsets(
          host=gerrit_host,
          project=project,
          change=change,
          patchset_src=latest_entity.cl_patchset.patchset,
          patchset_dest=patchset,
          coverage_data_src=latest_entity.data) if latest_entity.data else None
      rebased_coverage_data_unit = \
        code_coverage_util.RebasePresubmitCoverageDataBetweenPatchsets(
          host=gerrit_host,
          project=project,
          change=change,
          patchset_src=latest_entity.cl_patchset.patchset,
          patchset_dest=patchset,
          coverage_data_src=latest_entity.data_unit
      ) if latest_entity.data_unit else None
    except code_coverage_util.MissingChangeDataException as mcde:
      return BaseHandler.CreateError(
          'Requested coverage data is not found. %s' % six.text_type(mcde),
          404,
          allowed_origin='*')

    entity = PresubmitCoverageData.Create(
        server_host=gerrit_host,
        change=change,
        patchset=patchset,
        data=rebased_coverage_data,
        data_unit=rebased_coverage_data_unit)
    entity.absolute_percentages = latest_entity.absolute_percentages
    entity.incremental_percentages = latest_entity.incremental_percentages
    entity.absolute_percentages_unit = latest_entity.absolute_percentages_unit
    entity.incremental_percentages_unit = \
      latest_entity.incremental_percentages_unit
    entity.based_on = latest_entity.cl_patchset.patchset
    entity.put()
    return _ServeLines(entity.data)

  def HandleGet(self, **kwargs):
    return self._ServePerCLCoverageData()
