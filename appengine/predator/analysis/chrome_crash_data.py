# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json
import logging

from analysis import detect_regression_range
from analysis.chromecrash_parser import CracasCrashParser
from analysis.chromecrash_parser import FracasCrashParser
from analysis.crash_data import CrashData
from analysis.dependency_analyzer import DependencyAnalyzer
from decorators import cached_property


class ChromeCrashData(CrashData):
  """Parsed Chrome crash data from Cracas/Fracas.

  Properties:
    identifiers (dict): The key value pairs to uniquely identify a
      ``CrashData``.
    crashed_version (str): The version of project in which the crash occurred.
    signature (str): The signature of the crash.
    platform (str): The platform name; e.g., 'win', 'mac', 'linux', 'android',
      'ios', etc.
    stacktrace (Stacktrace): The stacktrace of the crash. N.B., this is
      an object generated by parsing the string containing the stack trace;
      we do not store the string itself.
    regression_range (pair or None): a pair of the last-good and first-bad
      versions. N.B., because this is an input, it is up to clients
      to call ``DetectRegressionRange`` (or whatever else) in order to
      provide this information. In addition, while this class does
      support storing ``None`` to indicate a missing regression range
      (because the ClusterFuzz client wants that feature), the
      CL-classifier doesn't actually support that so you won't get a
      very good Culprit. The Component- and project-classifiers do still
      return some results at least.
    dependencies (dict): A dict from dependency paths to
      ``Dependency`` objects. The keys are all those deps which are
      used by both the ``crashed_version`` of the code, and at least
      one frame in the ``stacktrace.crash_stack``.
    dependency_rolls (dict) A dict from dependency
      paths to ``DependencyRoll`` objects. The keys are all those
      dependencies which (1) occur in the regression range for the
      ``platform`` where the crash occurred, (2) neither add nor delete
      a dependency, and (3) are also keys of ``dependencies``.
  """

  def __init__(self, crash_data, dep_fetcher, top_n_frames=None):
    """
    Args:
      crash_data (dict): Dicts sent through Pub/Sub by Cracas/Fracas. Example:
      {
          'stack_trace': 'CRASHED [0x43507378...',
          # The Chrome version that produced the stack trace above.
          'chrome_version': '52.0.2743.41',
          # Client could provide customized data.
          'customized_data': {
              'trend_type': 'd',  # see supported types below
              'channel': 'beta',
              # Historical data about crash per million pageload by Chrome
              # version. (Right now last 20 versions)
              'historical_metadata': [
                  {
                      'report_number': 0,
                      'cpm': 0.0,
                      'client_number': 0,
                      'chrome_version': '51.0.2704.103'
                  },
                  ...
                  {
                      'report_number': 10,
                      'cpm': 2.1,
                      'client_number': 8,
                      'chrome_version': '53.0.2768.0'
                  },
              ]
          },
          'platform': 'mac',    # On which platform the crash occurs.
          'client_id': 'fracas',   # Identify which client this request is from.
          'signature': '[ThreadWatcher UI hang] base::RunLoopBase::Run',
          'crash_identifiers': {    # A list of key-value to identify a crash.
              'platform': 'mac',
              'version': '52.0.2743.41',
              'process_type': 'browser',
              'channel': 'beta',
              # Signature for the stack trace.
              'signature': '[ThreadWatcher UI hang] base::RunLoopBase::Run'
          }
      }
      dep_fetcher (ChromeDependencyFetcher): Dependency fetcher that can fetch
        all dependencies related to crashed version.
      top_n_frames (int): number of the frames in stacktrace we should parse.
    """
    super(ChromeCrashData, self).__init__(crash_data)
    self._crashed_version = crash_data['chrome_version']
    self._channel = crash_data['customized_data']['channel']
    self._historical_metadata = crash_data['customized_data'][
        'historical_metadata']
    self._identifiers = crash_data['crash_identifiers']

    self._top_n_frames = top_n_frames
    self._dependency_analyzer = DependencyAnalyzer(self._platform,
                                                   self._crashed_version,
                                                   self.regression_range,
                                                   dep_fetcher)

  @property
  def channel(self):
    return self._channel

  @property
  def historical_metadata(self):
    return self._historical_metadata

  @cached_property
  def stacktrace(self):
    """Parses stacktrace and returns parsed ``Stacktrace`` object."""
    stacktrace = self.StacktraceParser().Parse(
        self._raw_stacktrace,
        self._dependency_analyzer.regression_version_deps,
        top_n_frames=self._top_n_frames)
    if not stacktrace:
      logging.warning('Failed to parse the stacktrace %s',
                      self._raw_stacktrace)
    return stacktrace

  @cached_property
  def regression_range(self):
    """Detects regression range from ``historical_metadata`` and returns it."""
    regression_range = detect_regression_range.DetectRegressionRange(
        self.historical_metadata)
    if regression_range is None: # pragma: no cover
      logging.warning('Got ``None`` for the regression range.')
    else:
      regression_range = tuple(regression_range)

    return regression_range

  @cached_property
  def dependencies(self):
    """Get all dependencies that are in the crash stack of stacktrace."""
    return self._dependency_analyzer.GetDependencies(
        [self.stacktrace.crash_stack] if self.stacktrace else [])

  @cached_property
  def dependency_rolls(self):
    """Gets all dependency rolls of ``dependencies`` in regression range."""
    return self._dependency_analyzer.GetDependencyRolls(
        [self.stacktrace.crash_stack] if self.stacktrace else [])

  @property
  def identifiers(self):
    return self._identifiers

  @classmethod
  def StacktraceParser(cls):
    """The class of stacktrace parser."""
    raise NotImplementedError()


class FracasCrashData(ChromeCrashData):

  @classmethod
  def StacktraceParser(cls):
    """The class of stacktrace parser."""
    return FracasCrashParser()


class CracasCrashData(ChromeCrashData):

  @classmethod
  def StacktraceParser(cls):
    """The class of stacktrace parser."""
    return CracasCrashParser()

  @property
  def raw_stacktrace(self):
    return json.dumps(self._raw_stacktrace)

  @property
  def redo(self):
    return True