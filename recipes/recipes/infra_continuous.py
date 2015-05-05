# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

DEPS = [
  'bot_update',
  'gclient',
  'json',
  'path',
  'properties',
  'python',
  'step',
]


# Path to a service account credentials to use to talk to CIPD backend.
# Deployed by Puppet.
CIPD_BUILDER_CREDS = '/creds/service_accounts/service-account-cipd-builder.json'


def build_cipd_packages(api):
  try:
    return api.python(
        'build cipd packages',
        api.path['checkout'].join('build', 'build.py'),
        [
          '--upload',
          '--service-account-json', CIPD_BUILDER_CREDS,
          '--json-output', api.json.output(),
        ])
  finally:
    step_result = api.step.active_result
    output = step_result.json.output or {}
    p = step_result.presentation
    for pkg in output.get('succeeded', []):
      info = pkg['info']
      title = '%s %s' % (info['package'], info['instance_id'])
      p.links[title] = info.get('url', 'http://example.com/not-implemented-yet')


def GenSteps(api):
  builder_name = api.properties.get('buildername')
  if builder_name.startswith('infra-internal-continuous'):
    project_name = 'infra_internal'
  elif builder_name.startswith('infra-continuous'):
    project_name = 'infra'
  else:  #pragma: no cover
    raise ValueError('This recipe is not intended for builder %s. '
                     % builder_name)

  api.gclient.set_config(project_name)
  api.bot_update.ensure_checkout(force=True)
  api.gclient.runhooks()

  with api.step.defer_results():
    api.python('infra python tests',
               'test.py', ['test'], cwd=api.path['checkout'])
    # Note: env.py knows how to expand 'python' into sys.executable.
    api.python('infra go tests', api.path['checkout'].join('go', 'env.py'),
               ['python', api.path['checkout'].join('go', 'test.py')])

  build_cipd_packages(api)


def GenTests(api):
  cipd_json_output = {
    'succeeded': [
      {
        'info': {
          'instance_id': 'abcdefabcdef63ad814cd1dfffe2fcfc9f81299c',
          'package': 'infra/tools/some_tool/linux-bitness',
        },
        'pkg_def_name': 'some_tool',
      },
    ],
    'failed': [],
  }

  yield (
    api.test('infra') +
    api.properties.tryserver(
        buildername='infra-continuous') +
    api.override_step_data(
        'build cipd packages', api.json.output(cipd_json_output))
  )
  yield (
    api.test('infra_internal') +
    api.properties.tryserver(
        buildername='infra-internal-continuous') +
    api.override_step_data(
        'build cipd packages', api.json.output(cipd_json_output))
  )
