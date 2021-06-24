# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import os
import subprocess

from recipe_engine import recipe_api
from recipe_engine.recipe_api import Property

WIN_ADK_PACKAGE = 'infra/3pp/tools/win_adk/windows-amd64'
WIN_ADK_WINPE_PACKAGE = 'infra/3pp/tools/win_adk_winpe/windows-amd64'


class WindowsADKApi(recipe_api.RecipeApi):
  """API for using Windows ADK distributed via CIPD."""

  def __init__(self, win_adk_refs, win_adk_winpe_refs, *args, **kwargs):
    super(WindowsADKApi, self).__init__(*args, **kwargs)

    self._win_adk_refs = win_adk_refs
    self._win_adk_winpe_refs = win_adk_winpe_refs

  def ensure(self, install=True):
    """Ensure the presence of the Windows ADK."""
    with self.m.step.nest('ensure windows adk present'):
      with self.m.context(infra_steps=True):
        if install:
          self.ensure_win_adk(refs=self._win_adk_refs)
    with self.m.step.nest('ensure win-pe add-on present'):
      with self.m.context(infra_steps=True):
        if install:
          self.ensure_win_adk_winpe(refs=self._win_adk_winpe_refs)

  # TODO(actodd): reconfigure 3pp builder to preserve the software name
  def ensure_win_adk(self, refs):
    """Downloads & Installs the Windows ADK."""
    ensure_file = self.m.cipd.EnsureFile()
    ensure_file.add_package(WIN_ADK_PACKAGE, refs)
    cipd_dir = self.m.path['start_dir'].join('cipd', '3pp', 'adk')
    # Download the installer using cipd
    self.m.cipd.ensure(cipd_dir, ensure_file)
    logs_dir = self.m.path['cleanup'].join('logs', 'adk')
    # Run the installer and install ADK. If ADK is already installed this
    # does nothing. q: quiet, l: logs, features +: all features
    self.m.step('Install ADK', [
        cipd_dir.join('raw_source_0.exe'), '/q', '/l',
        logs_dir.join('adk.log'), '/features', '+'
    ])
    with self.m.step.nest('Read install logs') as r:
      ins_logs = self.m.file.listdir(
          'Listing all logs',
          logs_dir.join('adk'),
          test_data=[logs_dir.join('adk.log')])
      for l in ins_logs:
        c = self.m.file.read_raw('Reading ' + str(l), l)
        r.presentation.logs[str(l)] = c

  # TODO(actodd): reconfigure 3pp builder to preserve the software name
  def ensure_win_adk_winpe(self, refs):
    """Ensures that the WinPE add-on is available."""
    ensure_file = self.m.cipd.EnsureFile()
    ensure_file.add_package(WIN_ADK_WINPE_PACKAGE, refs)
    cipd_dir = self.m.path['start_dir'].join('cipd', '3pp', 'winpe')
    # Download the installer using cipd
    self.m.cipd.ensure(cipd_dir, ensure_file)
    logs_dir = self.m.path['cleanup'].join('logs', 'winpe')
    # Run the installer and install WinPE. If WinPE is already installed this
    # does nothing. q: quiet, l: logs, features +: all features
    self.m.step('Install WinPE', [
        cipd_dir.join('raw_source_0.exe'), '/q', '/l',
        logs_dir.join('winpe.log'), '/features', '+'
    ])
    with self.m.step.nest('Read install logs') as r:
      ins_logs = self.m.file.listdir(
          'Listing all logs', logs_dir, test_data=[logs_dir.join('winpe.log')])
      for l in ins_logs:
        # Read all the installation logs for display on milo
        c = self.m.file.read_raw('Reading ' + str(l), l)
        r.presentation.logs[str(l)] = c

  def cleanup_win_adk(self):
    """Cleanup the Windows ADK."""
    cipd_dir = self.m.path['start_dir'].join('cipd', '3pp', 'adk')
    logs_dir = self.m.path['cleanup'].join('uninstall-logs', 'adk')
    log_file = logs_dir.join('adk.log')
    # Run the installer and uninstall ADK. q: quiet, l: logs,
    # uninstall: remove all
    self.m.step(
        'Uninstall ADK',
        [cipd_dir.join('raw_source_0.exe'), '/q', '/uninstall', '/l', log_file])
    with self.m.step.nest('Read uninstall logs') as r:
      ins_logs = self.m.file.listdir(
          'Listing all logs', logs_dir, test_data=[log_file])
      for l in ins_logs:
        c = self.m.file.read_raw('Reading ' + str(l), l)
        r.presentation.logs[str(l)] = c

  def cleanup_winpe(self):
    """Cleanup WinPE."""
    cipd_dir = self.m.path['start_dir'].join('cipd', '3pp', 'winpe')
    logs_dir = self.m.path['cleanup'].join('uninstall-logs', 'winpe')
    log_file = logs_dir.join('winpe.log')
    # Run the installer and uninstall WinPE. q: quiet, l: logs, uninstall:
    # remove all
    self.m.step(
        'Uninstall ADK',
        [cipd_dir.join('raw_source_0.exe'), '/q', '/uninstall', '/l', log_file])
    with self.m.step.nest('Read uninstall logs') as r:
      ins_logs = self.m.file.listdir(
          'Listing all logs', logs_dir.join('winpe'), test_data=[log_file])
      for l in ins_logs:
        c = self.m.file.read_raw('Reading ' + str(l), l)
        r.presentation.logs[str(l)] = c

  def cleanup(self):
    """Remove the ADK and WinPE."""
    self.cleanup_win_adk()
    self.cleanup_winpe()
