# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

PYTHON_VERSION_COMPATIBILITY = "PY2+3"

DEPS = [
    'recipe_engine/buildbucket',
]


def RunSteps(api):
  """Runs infra{_internal} builds for infra_superproject changes."""

  # TODO(crbug.com/1415507): schedule infra and infra_internal builds
  return


def GenTests(api):
  yield (api.test('basic', status='INFRA_FAILURE') +
         api.buildbucket.try_build(project='infra/infra_superproject'))
