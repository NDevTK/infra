# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from analysis.uma_regression_data import UMA_RegressionData
from analysis.type_enums import CrashClient
from common.findit import Findit
from common.model.uma_regression_analysis import UMA_RegressionAnalysis


# TODO(cweakliam): This is currently just a skeleton. Implementation will come
# later.
class PredatorForSamplingProfiler(Findit):
  """Find culprits for regression reports from the UMA Sampling Profiler."""
  @classmethod
  def _ClientID(cls):
    return CrashClient.UMA_SAMPLING_PROFILER

  def __init__(self, get_repository, config):
    super(PredatorForSamplingProfiler, self).__init__(get_repository, config)
    # meta_weight = MetaWeight({
    #     # weights go here
    # })
    # meta_feature = WrapperMetaFeature([# features go here])

    # self._predator = Predator(ChangelistClassifier(get_repository,
    #                                                meta_feature,
    #                                                meta_weight),
    #                           self._component_classifier,
    #                           self._project_classifier)

  def _Predator(self):
    return self._predator

  def CreateAnalysis(self, regression_identifiers):
    """Creates ``UMA_RegressionAnalysis`` with regression_identifiers as key."""
    return UMA_RegressionAnalysis.Create(regression_identifiers)

  def GetAnalysis(self, regression_identifiers):
    """Gets ``UMA_RegressionAnalysis`` using regression_identifiers."""
    return UMA_RegressionAnalysis.Get(regression_identifiers)

  def GetCrashData(self, raw_regression_data):
    """Gets ``UMA_RegressionData`` from ``raw_regression_data``."""
    return UMA_RegressionData(raw_regression_data)
