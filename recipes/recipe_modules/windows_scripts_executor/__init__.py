# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
from PB.recipes.infra.windows_image_builder import input as input_pb

DEPS = [
    'depot_tools/gclient', 'depot_tools/bot_update', 'depot_tools/git',
    'depot_tools/gitiles', 'depot_tools/gsutil', 'recipe_engine/context',
    'recipe_engine/cipd', 'recipe_engine/step', 'recipe_engine/path',
    'recipe_engine/platform', 'recipe_engine/json', 'recipe_engine/file',
    'recipe_engine/archive', 'recipe_engine/raw_io',
    'recipe_engine/buildbucket', 'powershell', 'qemu'
]

PYTHON_VERSION_COMPATIBILITY = 'PY3'

ENV_PROPERTIES = input_pb.EnvProperties
