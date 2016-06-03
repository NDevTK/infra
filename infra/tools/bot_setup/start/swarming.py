# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Sets up and starts a Swarm slave."""

import os
import re
import requests
import sys


def is_staging(slave_name):
  return (
      slave_name.startswith('swarm-staging-') or
      re.match(r'^swarm[0-9]-c4$', slave_name))


def drop_signal_file():
  """Try to drop a signal file so that swarming doesn't."""
  fn = ("C:\\Users\\chrome-bot\\AppData\\Roaming\\Microsoft\\Windows\\"
        "Start Menu\\Programs\\Startup\\run_swarm_bot.bat")
  with open(fn, 'w') as f:
    f.write("exit 0\n")


def start(slave_name, root_dir):
  try:
    os.mkdir(os.path.join(root_dir, 'swarming'))
  except OSError:
    pass

  url = 'https://chromium-swarm.appspot.com'
  if is_staging(slave_name):
    url = 'https://chromium-swarm-dev.appspot.com'

  if sys.platform.startswith('win'):
    drop_signal_file()

  exec requests.get('%s/bootstrap' % url).text

  return 0
