# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json
import logging
import os
import pwd


SSH_IDENTITY_FILE_PATH = '/b/id_rsa'
SSH_CONFIG_FILE_PATH = os.path.expanduser('~chrome-bot/.ssh/config')
SSH_CONFIG_FILE_CONTENTS = (
"""# This file is generated by the infra docker service. Changes will be
# overwritten. The identity file is unique per container.
IdentityFile %s
ConnectionAttempts 4
Protocol 2
ConnectTimeout 30
ServerAliveCountMax 3
ServerAliveInterval 10
# A device's identity changes after every update. This can cause problems when
# its identity is stored on the host, so just skip the key checking.
StrictHostKeyChecking no
UserKnownHostsFile /dev/null
# LogLevel QUIET skips the "Host key has been permanently added" messages.
LogLevel QUIET
""" % SSH_IDENTITY_FILE_PATH)


def read_device_list(path_to_list):
  try:
    with open(path_to_list) as f:
      devices = json.load(f)
      if not isinstance(devices, list):
        raise TypeError(
            'Devices json file expected list, got %s: %s' % (
                type(devices), devices))
      return [d.split('.')[0] for d in devices]
  except TypeError:
    logging.exception('Unable to read device list %s.', path_to_list)
    raise


def should_write_ssh_config():
  ssh_dir = os.path.dirname(SSH_CONFIG_FILE_PATH)
  if not os.path.exists(ssh_dir):
    os.mkdir(ssh_dir, 0o700)
    chrome_bot = pwd.getpwnam('chrome-bot')
    os.chown(ssh_dir, chrome_bot.pw_uid, chrome_bot.pw_gid)
  if not os.path.exists(SSH_CONFIG_FILE_PATH):
    logging.warning('~/.ssh/config file does not exist. Creating it.')
    return True
  # Ensure the file's contents are what they should be. The file should be
  # reasonably sized, so just compare the whole thing in one go.
  with open(SSH_CONFIG_FILE_PATH) as f:
    config_contents = f.read()
  if config_contents != SSH_CONFIG_FILE_CONTENTS:
    logging.warning("SSH config file's contents are incorrect. Overwriting.")
    return True
  return False


def write_ssh_config():
  with open(SSH_CONFIG_FILE_PATH, 'w') as f:
    f.write(SSH_CONFIG_FILE_CONTENTS)