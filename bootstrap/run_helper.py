#!/usr/bin/env python3
# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Code supporting run.py implementation.

Reused across infra/run.py and infra_internal/run.py.
"""

import os
import sys


def run_py_main(args, runpy_path, package):
  import argparse
  import runpy
  import shlex
  import textwrap

  import argcomplete

  # Impersonate the argcomplete 'protocol'
  completing = os.getenv('_ARGCOMPLETE') == '1'
  if completing:
    assert not args
    line = os.getenv('COMP_LINE')
    args = shlex.split(line)[1:]
    if len(args) == 1 and not line.endswith(' '):
      args = []

  if not args or not args[0].startswith('%s.' % package):
    commands = []
    for root, _, files in os.walk(package):
      if '__main__.py' in files:
        commands.append(root.replace(os.path.sep, '.'))
    commands = sorted(commands)

    if completing:
      # Argcomplete is listening for strings on fd 8
      with os.fdopen(8, 'wb') as f:
        f.write('\n'.join(commands))
      return

    print(textwrap.dedent("""\
    usage: run.py %s.<module.path.to.tool> [args for tool]

    %s

    Available tools are:""") %
          (package, sys.modules['__main__'].__doc__.strip()))
    for command in commands:
      print('  *', command)
    return 1

  if completing:
    to_nuke = ' ' + args[0]
    os.environ['COMP_LINE'] = os.environ['COMP_LINE'].replace(to_nuke, '', 1)
    os.environ['COMP_POINT'] = str(int(os.environ['COMP_POINT']) - len(to_nuke))
    orig_parse_args = argparse.ArgumentParser.parse_args
    def new_parse_args(self, *args, **kwargs):
      argcomplete.autocomplete(self)
      return orig_parse_args(*args, **kwargs)
    argparse.ArgumentParser.parse_args = new_parse_args
  else:
    # remove the module from sys.argv
    del sys.argv[1]

  runpy.run_module(args[0], run_name='__main__', alter_sys=True)
