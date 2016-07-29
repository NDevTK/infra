# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import collections
import hashlib
import json
import re

from recipe_engine import recipe_api

RECIPE_TRYJOB_BYPASS_REASON_TAG = "Recipe-Tryjob-Bypass-Reason"

RECIPE_DEV_WORKFLOW_DOC_LINK = (
    'https://chromium.googlesource.com/external/github.com/luci/recipes-py/+/'
    'master/doc/workflow.md')

RietveldPatch = collections.namedtuple(
    'RietveldPatch', 'project server issue patchset')

def parse_patches(failing_step, patches_raw, rietveld, issue, patchset,
                  patch_project):
  """
  gives mapping of project to patch
    expect input of
    project1:https://a.b.c/1342342#ps1,project2:https://d.ce.f/1231231#ps1
  """
  result = {}

  if rietveld and issue and patchset and patch_project:
    # convert to str because recipes don't like unicode as step names
    result[str(patch_project)] = RietveldPatch(
        patch_project, rietveld, issue, patchset)

  if not patches_raw:
    return result

  for patch_raw in patches_raw.split(','):
    project, url = patch_raw.split(':', 1)
    server, issue_and_patchset = url.rsplit('/', 1)
    issue, patchset = issue_and_patchset.split('#')
    patchset = patchset[2:]

    if project in result:
      failing_step(
          "Invalid patchset list",
          "You have two patches for %r. Patches seen so far: %r" % (
              project, result)
      )

    result[project] = RietveldPatch(project, server, issue, patchset)

  return result


PROJECTS_TO_TRY = [
  'build',
  'build_limited_scripts_slave',
  'recipe_engine',
  'depot_tools',
]

PROJECT_TO_CONTINUOUS_WATERFALL = {
  'build': 'https://build.chromium.org/p/chromium.tools.build/builders/'
    'recipe-simulation_trusty64',
  'recipe_engine': 'https://build.chromium.org/p/chromium.infra/builders/'
    'recipe_engine-recipes-tests',
  'depot_tools': 'https://build.chromium.org/p/chromium.infra/builders/'
    'depot_tools-recipes-tests',
  'build_limited_scripts_slave': (
    'https://uberchromegw.corp.google.com/i/internal.infra/builders/'
    'build-limited-recipe-tests'),
}

FILE_BUG_FOR_CONTINUOUS_LINK = 'https://goo.gl/PoAPOJ'


class RecipeTryjobApi(recipe_api.RecipeApi):
  """
  This is intended as a utility module for recipe tryjobs. Currently it's just a
  refactored version of a recipe; eventually some of this, especially the
  dependency information, will probably get moved into the recipe engine.
  """
  def _checkout_projects(self, root_dir, url_mapping, deps,
                        downstream_projects, patches):
    """Checks out projects listed in projects into root_dir.

    Args:
      root_dir: Root directory to check this project out in.
      url_mapping: Project id to url of git repository.
      downstream_projects: The mapping from project to dependent projects.
      patches: Mapping of project id to patch to apply to that project.

    Returns:
      The projects we want to test, and the locations of those projects
    """
    # TODO(martiniss): be smarter about which projects we actually run tests on

    # All the projects we want to test.
    projs_to_test  = set()
    # Projects we need to look at dependencies for.
    queue = set(patches.keys())
    # luci config project name to file system path of the checkout
    locations = {}

    while queue:
      proj = queue.pop()
      if proj not in projs_to_test:
        locations[proj] = self._checkout_project(
            proj, url_mapping[proj], root_dir, patches.get(proj))
        projs_to_test.add(proj)

        for downstream in downstream_projects.get(proj, set()):
          queue.add(downstream)
        for upstream in deps.get(proj, set()):
          queue.add(upstream)

    return projs_to_test, locations

  def _checkout_project(self, proj, proj_config, root_dir, patch=None):
    """
    Args:
      proj: luci-config project name to checkout.
      proj_config: The recipes.cfg configuration for the project.
      root_dir: The temporary directory to check the project out in.
      patch: optional patch to apply to checkout.

    Returns:
      Path to repo on disk.
    """
    checkout_path = root_dir.join(proj)
    repo_path = checkout_path.join(proj)
    self.m.file.makedirs('%s directory' % proj, repo_path)

    config = self.m.gclient.make_config(
        GIT_MODE=True, CACHE_DIR=root_dir.join("__cache_dir"))
    soln = config.solutions.add()
    soln.name = proj
    soln.url = proj_config['repo_url']

    kwargs = {
        'suffix': proj,
        'gclient_config': config,
        'force': True,
        'cwd': checkout_path,
    }
    if patch:
      kwargs['rietveld'] = patch.server
      kwargs['issue'] = patch.issue
      kwargs['patchset'] = patch.patchset
    else:
      kwargs['patch'] = False

    self.m.bot_update.ensure_checkout(**kwargs)
    return repo_path

  def get_fail_build_info(self, downstream_projects, patches):
    fail_build = collections.defaultdict(lambda: True)

    for proj, patch in patches.items():
      patch_url = "%s/%s" % (patch.server, patch.issue)
      desc = self.m.git_cl.get_description(
          patch=patch_url, codereview='rietveld', suffix=proj)

      assert desc.stdout is not None, "CL %s had no description!" % patch_url

      bypass_reason = self.m.tryserver.get_footer(
          RECIPE_TRYJOB_BYPASS_REASON_TAG, patch_text=desc.stdout)

      fail_build[proj] = not bool(bypass_reason)

    # Propogate Falses down the deps tree
    queue = list(patches.keys())
    while queue:
      item = queue.pop(0)

      if not fail_build[item]:
        for downstream in downstream_projects.get(item, []):
          fail_build[downstream] = False
          queue.append(downstream)

    return fail_build

  def simulation_test(self, proj, proj_config, repo_path, deps):
    """
    Args:
      proj: The luci-config project to simulation_test.
      proj_config: The recipes.cfg configuration for the project.
      repo_path: The path to the repository on disk.
      deps: Mapping from project name to Path. Passed into the recipes.py
        invocation via the "-O" options.

    Returns the result of running the simulation tests.
    """
    recipes_path = self.m.recipe_utils.get_recipes_path(
        proj_config) + ['recipes.py']
    recipes_py_loc = repo_path.join(*recipes_path)
    args = []
    for dep_name, location in deps.items():
      args += ['-O', '%s=%s' % (dep_name, location)]
    args += ['--package', repo_path.join('infra', 'config', 'recipes.cfg')]

    args += ['simulation_test']

    return self._python('%s tests' % proj, recipes_py_loc, args)

  def _python(self, name, script, args, **kwargs):
    """Call python from infra's virtualenv.

    This is needed because of the coverage module, which is not installed by
    default, but which infra's python has installed."""
    return self.m.step(name, [
        self.m.path['checkout'].join('ENV', 'bin', 'python'),
        '-u', script] + args, **kwargs)

  def run_tryjob(self, patches_raw, rietveld, issue, patchset, patch_project):
    patches = parse_patches(
        self.m.python.failing_step, patches_raw, rietveld, issue, patchset,
        patch_project)

    root_dir = self.m.path['slave_build']

    # Needed to set up the infra checkout, for _python
    self.m.gclient.set_config('infra')
    self.m.gclient.c.solutions[0].revision = 'origin/master'
    self.m.gclient.checkout()
    self.m.gclient.runhooks()

    url_mapping = self.m.luci_config.get_projects()

    # List of all the projects we care about testing. luci-config names
    all_projects = set(p for p in url_mapping if p in PROJECTS_TO_TRY)

    deps, downstream_projects = self.m.recipe_utils.get_deps_info(
        all_projects)
    should_fail_build_mapping = self.get_fail_build_info(
        downstream_projects, patches)

    projs_to_test, locations = self._checkout_projects(
        root_dir, url_mapping, deps, downstream_projects, patches)

    bad_projects = []
    for proj in projs_to_test:
      deps_locs = {dep: locations[dep] for dep in deps[proj]}

      try:
        result = self.simulation_test(
          proj, self.m.recipe_utils.get_project_config(proj), locations[proj],
          deps_locs)
      except recipe_api.StepFailure as f:
        result = f.result
        if should_fail_build_mapping.get(proj, True):
          bad_projects.append(proj)
      finally:
        link = PROJECT_TO_CONTINUOUS_WATERFALL.get(proj)
        if link:
          result.presentation.links['reference builder'] = link
        # no cover because at the moment we don't test anything which isn't
        # included in PROJECT_TO_TRY, and all those projects have reference
        # builders.
        else: #pragma: no cover
          result.presentation.links[
              'no reference builder; file a bug to get one?'] = (
                  FILE_BUG_FOR_CONTINUOUS_LINK)


    if bad_projects:
      try:
        self.m.python.failing_step(
            'failed tests', "One or more projects failed tests: %r" % (
              ','.join(bad_projects)))
      except self.m.step.StepFailure as f:
        f.result.presentation.links['Confused? README here'] = (
            RECIPE_DEV_WORKFLOW_DOC_LINK)
        raise



