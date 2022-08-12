# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
from PB.recipes.infra.windows_image_builder import actions


def get_src_from_action(action):
  """ get_src_from_action returns src ref in action if any
      Args:
        action: proto Action object that might contain a src
  """
  if action.WhichOneof('action') == 'add_file':
    return action.add_file.src
  if action.WhichOneof('action') == 'add_windows_package':
    return action.add_windows_package.src
  if action.WhichOneof('action') == 'add_windows_driver':
    return action.add_windows_driver.src


def pin_src_from_action(action, sources):
  if action.WhichOneof('action') == 'add_file':
    action.add_file.src.CopyFrom(sources.pin(action.add_file.src))
  if action.WhichOneof('action') == 'add_windows_package':
    action.add_windows_package.src.CopyFrom(
        sources.pin(action.add_windows_package.src))
  if action.WhichOneof('action') == 'add_windows_driver':
    action.add_windows_driver.src.CopyFrom(
        sources.pin(action.add_windows_driver.src))


def get_build_offline_customization(offline_customization):
  """ get_build_offline_customization returns actions.OfflineAction object
      same as oc, but with all name strings reset
      Args:
        offline_customization: actions.OfflineAction proto object representing
        a sub-customization to be performed.
      Example:
        Given a config
          OfflineAction{
            name: "add diskparts"
            actions: [...]
          }
        returns config
          OfflineAction{
            name: ""
            actions: [...]
          }
  """
  acts = [get_build_actions(act) for act in offline_customization.actions]
  return actions.OfflineAction(actions=acts)


def ensure_dirs(m_file, dirs):
  """ ensure_dirs ensures that the given dirs are created on the bot
      Args:
        m_file: ref to recipe_engine/file module object
        dirs: list of paths to dirs that need to be ensured
  """
  for d in dirs:
    m_file.ensure_directory('Ensure {}'.format(d), d)


def get_build_actions(action):
  """ get_build_actions returns a actions.Action object same as given action
      but with name string reset
      Args:
        action: proto actions.Action object representing an action to be
        performed
      Example:
        Given a config
          Action{
            xyz_action: XYZAction{
              name: "do this"
              ...
            }
          }
        returns config
          Action{
            xyz_action: XYZAction{
              name: ""
              ...
            }
          }
  """
  if action.WhichOneof('action') == 'add_file':
    return actions.Action(
        add_file=actions.AddFile(
            src=action.add_file.src,
            dst=action.add_file.dst,
        ))
  if action.WhichOneof('action') == 'add_windows_package':
    return actions.Action(
        add_windows_package=actions.AddWindowsPackage(
            src=action.add_windows_package.src,
            args=action.add_windows_package.args,
        ))
  if action.WhichOneof('action') == 'add_windows_driver':
    return actions.Action(
        add_windows_driver=actions.AddWindowsDriver(
            src=action.add_windows_driver.src,
            args=action.add_windows_driver.args,
        ))
  if action.WhichOneof('action') == 'edit_offline_registry':
    eor = actions.EditOfflineRegistry()
    eor.CopyFrom(action.edit_offline_registry)
    eor.name = ''
    return actions.Action(edit_offline_registry=eor)


def conv_to_win_path(path):
  return '\\'.join(path.split('/'))
