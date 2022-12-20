# Copyright 2020 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

USE_PYTHON3 = True


def CheckChange(input_api, output_api):
  results = []
  results += input_api.canned_checks.CheckDoNotSubmit(input_api, output_api)
  results += input_api.canned_checks.CheckChangeHasNoTabs(input_api, output_api)
  # NPM audit presubmit disabled. See: crbug.com/monorail/10572
  # results += CheckNpmAudit(input_api, output_api)
  return results


def CheckChangeOnUpload(input_api, output_api):
  return CheckChange(input_api, output_api)


def CheckChangeOnCommit(input_api, output_api):
  return CheckChange(input_api, output_api)


def CheckNpmAudit(input_api, output_api):  # pragma: no cover
  file_filter = lambda f: f.LocalPath().endswith('.js')
  affected_js_files = input_api.AffectedFiles(
      include_deletes=False, file_filter=file_filter)
  if not affected_js_files:
    return []

  import imp
  appengine_path = input_api.os_path.dirname(input_api.PresubmitLocalPath())
  js_checker_path = input_api.os_path.join(appengine_path, 'js_checker.py')
  js_checker = imp.load_source('JSChecker', js_checker_path)

  return js_checker.JSChecker(input_api, output_api).RunAuditCheck()
