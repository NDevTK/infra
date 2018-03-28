# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from libs.structured_object import StructuredObject
from services.parameters import BuildKey
from services.parameters import TestHeuristicAnalysisOutput


class RunSwarmingTasksInput(StructuredObject):
  """This class defines the input of run_test_swarming_task_pipeline."""
  build_key = BuildKey
  heuristic_result = TestHeuristicAnalysisOutput
  force = bool