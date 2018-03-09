# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Handles requests to disable/enable auto-commit.

This handler is mainly for troopers to turn off auto-commit when fire happens.
So it will turn on/off auto-commit for both failure types (compile and
consistent test failure) at the same time.
Admins: use config page to turn on/off auto-commit for a single type.
"""

import copy
import json

from google.appengine.api import users

from common import acl
from gae_libs import token
from gae_libs.handlers.base_handler import BaseHandler, Permission
from model import wf_config
from waterfall import waterfall_config


class ChangeAutoRevertSetting(BaseHandler):
  PERMISSION_LEVEL = Permission.CORP_USER

  @token.AddXSRFToken(action_id='config')
  def HandleGet(self):
    action_configs = waterfall_config.GetActionSettings()

    # Either config is True means the feature is on.
    auto_commit_revert = (
        action_configs.get('auto_commit_revert_compile', False) or
        action_configs.get('auto_commit_revert_test', False))
    return {
        'template': 'change_auto_revert_setting.html',
        'data': {
            'is_admin': users.is_current_user_admin(),
            'auto_commit_revert_on': auto_commit_revert,
        }
    }

  @token.VerifyXSRFToken(action_id='config')
  def HandlePost(self):
    user = users.get_current_user()
    is_admin = users.is_current_user_admin()

    # Only admin could turn auto-commit back on again.
    flag = self.request.get('auto_commit_revert', '').lower()
    auto_commit_revert = flag == 'true'
    if auto_commit_revert and not is_admin:
      return BaseHandler.CreateError('Only admin could turn auto-commit on.',
                                     403)

    message = self.request.get('update_reason', '').strip()
    if not message:
      return BaseHandler.CreateError('Please enter the reason.', 400)

    updated = False
    action_configs = waterfall_config.GetActionSettings()
    if (auto_commit_revert != action_configs.get('auto_commit_revert_compile')
        or auto_commit_revert != action_configs.get('auto_commit_revert_test')):
      action_settings = copy.deepcopy(action_configs)
      action_settings['auto_commit_revert_compile'] = auto_commit_revert
      action_settings['auto_commit_revert_test'] = auto_commit_revert
      updated = wf_config.FinditConfig.Get().Update(
          user,
          acl.IsPrivilegedUser(user.email(), is_admin),
          message=message,
          action_settings=action_settings)

    if not updated:
      return BaseHandler.CreateError('Failed to update auto-revert setting. '
                                     'Please refresh the page and try again.',
                                     400)

    return self.CreateRedirect('/waterfall/change-auto-revert-setting')
