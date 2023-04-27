# Copyright (c) 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""appengine-specific presubmit for running pylint.
"""

USE_PYTHON3 = True

# LGTM FROM iannucci@ REQUIRED TO EDIT THIS LIST:
DISABLED_PYLINT_WARNINGS = [
    'W0231',  # __init__ method from base class is not called
    'W0232',  # Class has no __init__ method
    'unused-argument',
    'attribute-defined-outside-init',

    # TODO(crbug/1347377): Re-enable these checks.
    'no-member',
    'super-with-arguments',
    'useless-object-inheritance',
    'no-else-return',
    'syntax-error',
    'inconsistent-return-statements',
    'unnecessary-pass',
    'raise-missing-from',
    'import-outside-toplevel',
    'unnecessary-comprehension',
    'arguments-differ',
    'invalid-overridden-method',
    'trailing-newlines',
    'no-else-raise',
    'consider-using-in',
    'raising-format-tuple',
    'consider-using-set-comprehension',
    'assignment-from-none',
    'use-a-generator',
    'useless-super-delegation',
    'subprocess-run-check',
    'redefined-argument-from-local',
    'no-else-continue',
    'len-as-condition',
    'keyword-arg-before-vararg',
    'bad-option-value',
    'no-value-for-parameter',
    'implicit-str-concat',
    'useless-return',
    'useless-import-alias',
    'simplifiable-if-expression',
    'no-else-break',
    'literal-comparison',
    'consider-iterating-dictionary',
    'arguments-out-of-order',
]

DISABLED_PYLINT_FILES = [
    r'.*[/\\]__pycache__',
    r'.+_pb2\.py',
    r'.*[/\\]six\.py',
    r'.*[/\\]gae\.py',
]


# Forked from depot_tools/presubmit_canned_checks._FetchAllFiles
def FetchAllFiles(input_api, files_to_check, files_to_skip):
  import datetime
  start_time = datetime.datetime.now()
  def Find(filepath, filters):
    return any(input_api.re.match(item, filepath) for item in filters)

  repo_path = input_api.PresubmitLocalPath()
  def MakeRootRelative(dirpath, item):
    path = input_api.os_path.join(dirpath, item)
    # Poor man's relpath:
    if path.startswith(repo_path):  # pragma: no cover
      return path[len(repo_path) + 1:]
    return path  # pragma: no cover

  dirs_walked = []

  files = []
  for dirpath, dirnames, filenames in input_api.os_walk(repo_path):
    dirs_walked.append(dirpath)
    for item in dirnames[:]:
      filepath = MakeRootRelative(dirpath, item)
      if Find(filepath, files_to_skip):
        dirnames.remove(item)
    for item in filenames:
      filepath = MakeRootRelative(dirpath, item)
      if Find(filepath, files_to_check) and not Find(filepath, files_to_skip):
        files.append(filepath)
  duration = datetime.datetime.now() - start_time
  input_api.logging.info('FetchAllFiles found %s files, searching '
      '%s directories in %ss' % (len(files), len(dirs_walked),
      duration.total_seconds()))
  return files


def IgnoredPaths(input_api): # pragma: no cover
  # This computes the list if repository-root-relative paths which are
  # ignored by .gitignore files. There is probably a faster way to do this.
  status_output = input_api.subprocess.check_output(
      ['git', 'status', '--porcelain', '--ignored'])
  statuses = [(line[:2], line[3:]) for line in status_output.splitlines()]
  return [
    input_api.re.escape(path) for (mode, path) in statuses
    if mode in ('!!', '??') and not path.endswith('.pyc')
  ]


def PythonRootForPath(input_api, path):
  # For each path, walk up directories until find no more __init__.py
  # The directory above the last __init__.py is considered our root.
  root = input_api.os_path.dirname(path)
  while True:
    init_file = input_api.os_path.join(root, '__init__.py')
    if not input_api.os_path.isfile(init_file):
      break
    root = input_api.os_path.dirname(root)

  return root


def GroupPythonFilesByRoot(input_api, paths):
  sorted_paths = sorted(paths)
  import collections
  grouped_paths = collections.defaultdict(list)
  for path in sorted_paths:
    # FIXME: This doesn't actually need to touch the filesystem if we can
    # trust that 'paths' contains all __init__.py paths we care about.
    root = PythonRootForPath(input_api, path)
    grouped_paths[root].append(path)
  # Convert back to a normal dict before returning.
  return dict(grouped_paths)


def DirtyRootsFromAffectedFiles(changed_py_files, root_to_paths):
  # Compute root_groups for all python files
  path_to_root = {}
  for root, paths in root_to_paths.items():
    for path in paths:
      path_to_root[path] = root

  # Using the above mapping, compute the actual roots we need to run
  dirty_roots = set()
  for path in changed_py_files:
    dirty_roots.add(path_to_root[path])
  return dirty_roots


def GetAppEngineLibraryPaths(input_api, appengine_env_path):  # pragma: no cover
  # AppEngine has a wrapper_util module which knows where the various
  # appengine libraries are stored inside the SDK. All AppEngine scripts
  # 'import wrapper_util' and then call its various methods to get those
  # paths to fix their sys.path. Since AppEngine isn't in our sys.path yet
  # we use imp.load_source to load wrapper_util from an absolute path
  # and then call its methods to get all the paths to the AppEngine-provided
  # libraries to add to sys.path when calling pylint.
  import imp
  wrapper_util_path = input_api.os_path.join(appengine_env_path,
      'wrapper_util.py')
  wrapper_util = imp.load_source('wrapper_util', wrapper_util_path)
  wrapper_util_paths = wrapper_util.Paths(appengine_env_path)
  appengine_lib_paths = wrapper_util_paths.script_paths('dev_appserver.py')
  # Unclear if v2_extra_paths is correct here, it contains endpoints
  # and protorpc which several apps seem to depend on.
  return appengine_lib_paths + wrapper_util_paths.v2_extra_paths


# Forked with prejudice from depot_tools/presubmit_canned_checks.py
# pragma: no cover
def PylintFiles(input_api, output_api, files, pylint_root, disabled_warnings):
  input_api.logging.debug('Running pylint on: %s', files)

  # FIXME: depot_tools should be right next to infra, however DEPS
  # recursion into build/DEPS does not seem to be working: crbug.com/410070
  canned_checks_path = input_api.canned_checks.__file__
  canned_checks_path = input_api.os_path.abspath(canned_checks_path)
  depot_tools_path = input_api.os_path.dirname(canned_checks_path)

  pylint_args = ['-d', ','.join(disabled_warnings)]

  pylint_path = input_api.os_path.join(depot_tools_path, 'pylint-2.7')

  # Make paths relative to pylint_root
  for filename in files:
    assert(filename.startswith(pylint_root + '/') if pylint_root else True)
  files = [filename[len(pylint_root)+1:] if pylint_root else filename
           for filename in files]

  kwargs = {'stdin': '\n'.join(pylint_args + files).encode()}
  if pylint_root:
    kwargs['cwd'] = pylint_root

  # Pass args via stdin, because windows (command line limit).
  return input_api.Command(
      name=('Pylint (%s files%s)' %
            (len(files), ' under %s' % pylint_root if pylint_root else '')),
      cmd=['vpython3', pylint_path, '--args-on-stdin'],
      kwargs=kwargs,
      message=output_api.PresubmitError)


def PylintChecks(input_api, output_api, only_changed):  # pragma: no cover
  files_to_check = [r'.*\.py$']
  files_to_skip = list(input_api.DEFAULT_FILES_TO_SKIP)
  files_to_skip += DISABLED_PYLINT_FILES
  files_to_skip += IgnoredPaths(input_api)

  source_filter = lambda path: input_api.FilterSourceFile(
      path, files_to_check=files_to_check, files_to_skip=files_to_skip)

  # Compute paths of changed files relative to the present file. Ignore files
  # outside of PresubmitLocalPath().
  changed_py_files = []
  for f in input_api.AffectedSourceFiles(source_filter):
    abs_path = f.AbsoluteLocalPath()
    rel = input_api.os_path.relpath(abs_path, input_api.PresubmitLocalPath())
    if not rel.startswith('..'):
      changed_py_files.append(abs_path[len(input_api.PresubmitLocalPath())+1:])

  if only_changed:
    if changed_py_files:
      input_api.logging.info('Running pylint on %d files',
                             len(changed_py_files))
      return [
          PylintFiles(input_api, output_api, changed_py_files, None,
                      DISABLED_PYLINT_WARNINGS)
      ]
    return []

  all_python_files = FetchAllFiles(input_api, files_to_check, files_to_skip)
  root_to_paths = GroupPythonFilesByRoot(input_api, all_python_files)
  dirty_roots = DirtyRootsFromAffectedFiles(changed_py_files, root_to_paths)

  tests = []
  for root_path in sorted(dirty_roots):
    python_files = root_to_paths[root_path]
    if python_files:
      input_api.logging.info('Running appengine pylint on %d files under %s',
                             len(python_files), root_path or 'appengine')
      tests.append(
          PylintFiles(input_api, output_api, python_files, root_path,
                      DISABLED_PYLINT_WARNINGS))
  return tests


def JsChecks(input_api, output_api):  # pragma: no cover
  try:
    infra_root = input_api.os_path.dirname(input_api.PresubmitLocalPath())
    import sys
    import imp
    old_sys_path = sys.path[:]
    wrapper_util_path = input_api.os_path.join(
        infra_root, 'appengine', 'js_checker.py')
    js_checker = imp.load_source('JSChecker', wrapper_util_path)
    # TODO(prasadv): Determine which files to exclude.
    files_to_skip = []
    file_filter = lambda f: f.LocalPath() not in files_to_skip
    results = js_checker.JSChecker(
        input_api, output_api, file_filter).RunChecks()
  finally:
    sys.path = old_sys_path

  return results


def CommonChecks(_input_api, _output_api):  # pragma: no cover
  return []


def CheckChangeOnUpload(input_api, output_api):  # pragma: no cover
  output = CommonChecks(input_api, output_api)
  output.extend(input_api.RunTests(PylintChecks(
      input_api, output_api, only_changed=True)))
  output.extend(input_api.RunTests(JsChecks(input_api, output_api)))

  return output


def CheckChangeOnCommit(input_api, output_api):  # pragma: no cover
  output = CommonChecks(input_api, output_api)
  output.extend(input_api.RunTests(PylintChecks(
      input_api, output_api, only_changed=False)))
  return output
