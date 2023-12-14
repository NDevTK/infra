# Copyright (c) 2023 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import sys
import json
import argparse
import collections.abc

from pathlib import Path

THIS_FILE = Path(__file__).absolute()

sys.path.append(str(THIS_FILE.parent.parent / 'depot_tools'))

import gclient_eval

parser = argparse.ArgumentParser()
parser.add_argument('--DEPS-filename')
parser.add_argument('--host-os')
parser.add_argument('--host-cpu')
parser.add_argument('--target-os')
parser.add_argument('--target-cpu')
parser.add_argument('--str-var', action='append', default=[])
parser.add_argument('--bool-var', action='append', default=[])

args = parser.parse_args()

builtin_vars = {
    'checkout_android': 'android' in args.target_os,
    'checkout_chromeos': 'chromeos' in args.target_os,
    'checkout_fuchsia': 'fuchsia' in args.target_os,
    'checkout_ios': 'ios' in args.target_os,
    'checkout_linux': 'unix' in args.target_os,
    'checkout_mac': 'mac' in args.target_os,
    'checkout_win': 'win' in args.target_os,
    'host_os': args.host_os,
    'checkout_arm': 'arm' in args.target_cpu,
    'checkout_arm64': 'arm64' in args.target_cpu,
    'checkout_x86': 'x86' in args.target_cpu,
    'checkout_mips': 'mips' in args.target_cpu,
    'checkout_mips64': 'mips64' in args.target_cpu,
    'checkout_ppc': 'ppc' in args.target_cpu,
    'checkout_s390': 's390' in args.target_cpu,
    'checkout_x64': 'x64' in args.target_cpu,
    'host_cpu': args.host_cpu,
}

custom_vars = {}
for val in args.str_var:
  key, value = val.split('=', 1)
  custom_vars[key] = value
for val in args.bool_var:
  key, value = val.split('=', 1)
  custom_vars[key] = value == 'true'

deps = gclient_eval.Parse(
    sys.stdin.read(),
    filename=args.DEPS_filename,
    vars_override=custom_vars,
    builtin_vars=builtin_vars)

# deps['vars'] are the variables we just parsed out of DEPS, and are
# unevaluated.
#
# We layer builtin_vars on top, and finally custom_vars (which includes both the
# gclient definition of custom_vars (supplied on the CLI), AND ALSO the parent
# DEPS's variables).
completeVars = dict(deps['vars'])
completeVars.update(builtin_vars)
completeVars.update(custom_vars)


def _eval_var(value):
  if isinstance(value, gclient_eval.ConstantString):
    return value.value
  if isinstance(value, bool):
    return value

  try:
    return gclient_eval.EvaluateCondition(value, completeVars)
  except:
    print(f"FAILED FOR {value!r}", file=sys.stderr)
    raise


ret = {
    'use_relative_paths': deps.get('use_relative_paths', False),
    'git_dependencies': deps.get('git_dependencies', 'DEPS'),
    'vars': {
        'bool_vars': {
            k: v
            for k, v in completeVars.items()
            if isinstance(v, bool) and k not in builtin_vars
        },
        'str_vars': {
            k: v
            for k, v in completeVars.items()
            if isinstance(v, (str, gclient_eval.ConstantString)) and
            k not in builtin_vars
        },
        # TODO: These are simply repeated back from the CLI for convenience, but
        # these should just be merged on the Go side.
        'target_os': args.target_os,
        'target_cpu': args.target_cpu,
        'host_os': args.host_os,
        'host_cpu': args.host_cpu,
    },
    'hooks': [],
    'git_deps': {},
    'cipd_deps': {},
    'recursedeps': deps.get('recursedeps', []),
    'gclient_gn_args_file': deps.get('gclient_gn_args_file', ''),
    'gclient_gn_args': {
        arg: _eval_var(completeVars[arg])
        for arg in deps.get('gclient_gn_args', [])
    },
}

for deppath, dep in deps['deps'].items():
  if 'condition' in dep:
    if not gclient_eval.EvaluateCondition(dep['condition'], completeVars):
      continue

  if dep['dep_type'] == 'git':
    repo, rev = dep['url'].split('@')
    ret['git_deps'][deppath] = {'repo': repo, 'rev': rev}
  elif dep['dep_type'] == 'cipd':
    ret['cipd_deps'][deppath] = {
        'packages': dep['packages'],
    }
  else:
    sys.exit(f'UNKNOWN dep_type in DEPS: {dep!r}')

for hook in deps.get('hooks', ()):
  if 'condition' in hook:
    if not gclient_eval.EvaluateCondition(hook['condition'], completeVars):
      continue

  ret['hooks'].append(hook)


def _fix_nodes(obj):
  if isinstance(obj, collections.abc.MutableMapping):
    return dict(obj)
  if isinstance(obj, gclient_eval.ConstantString):
    return obj.value
  return obj


json.dump(ret, sys.stdout, sort_keys=True, indent=2, default=_fix_nodes)
