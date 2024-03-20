# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Functions and constants related to recipes ecosystem support."""

load("//lib/build.star", "build")
load("//lib/infra.star", "infra")

# Recipes repo ID => (repo URL, name to use in builders).
_KNOWN_REPOS = {
    "build": ("https://chromium.googlesource.com/chromium/tools/build", "Build"),
    "chromiumos": ("https://chromium.googlesource.com/chromiumos/infra/recipes", "ChromiumOS"),
    "depot_tools": ("https://chromium.googlesource.com/chromium/tools/depot_tools", "Depot Tools"),
    "fuchsia": ("https://fuchsia.googlesource.com/infra/recipes", "Fuchsia"),
    "infra": ("https://chromium.googlesource.com/infra/infra", "Infra"),
    "pigweed": ("https://pigweed.googlesource.com/infra/recipes", "Pigweed"),
    "recipe_engine": ("https://chromium.googlesource.com/infra/luci/recipes-py", "Recipe Engine"),
    "skia": ("https://skia.googlesource.com/skia", "Skia"),
    "skiabuildbot": ("https://skia.googlesource.com/buildbot", "Skia Buildbot"),
}

def _repo_url(proj):
    return _KNOWN_REPOS[proj][0]

def _friendly(proj):
    return _KNOWN_REPOS[proj][1]

def simulation_tester(
        name,
        project_under_test,
        triggered_by,
        console_view = None,
        console_category = None,
        os = "Ubuntu-22.04"):
    """Defines a CI builder that runs recipe simulation tests.

    Args:
      name: name of the builder.
      project_under_test: the recipient project of the recipe roll.
      triggered_by: what builders trigger this one.
      console_view: a console to add it to.
      console_category: a category to use in the console.
      os: the target OS dimension.
    """

    # Normally, this builder will be triggered on specific commit in this
    # git_repo, and hence additional git_repo property is redundant. However, if
    # one uses LUCI scheduler "Trigger Now" feature, there will be no associated
    # commit and hence we need git_repo property.
    properties = {"git_repo": _repo_url(project_under_test)}
    luci.builder(
        name = name,
        bucket = "ci",
        executable = infra.recipe("recipe_simulation", use_python3 = True),
        properties = properties,
        dimensions = {
            "os": os,
            "cpu": "x86-64",
            "pool": "luci.infra.ci",
        },
        service_account = infra.SERVICE_ACCOUNT_CI,
        build_numbers = True,
        triggered_by = [triggered_by],
        notifies = infra.tree_closing_notifiers(),
    )
    if console_view:
        luci.console_view_entry(
            builder = name,
            console_view = console_view,
            category = console_category,
        )

def roll_trybots(upstream, downstream, cq_group, os = "Ubuntu"):
    """Defines a bunch of recipe roller trybots, one per downstream project.

    Args:
      upstream: an upstream project to roll from.
      downstream: a list of downstream projects to roll into.
      cq_group: a CQ group to add the builders to as verifiers.
      os: the target OS dimension.
    """
    for proj in downstream:
        name = "%s downstream Recipe Roll tester from %s" % (_friendly(proj), _friendly(upstream))
        pool = "luci.infra.try"
        if os.startswith("Mac"):
            pool = "luci.flex.try"
        luci.builder(
            name = name,
            bucket = "try",
            executable = infra.recipe("recipe_roll_tryjob", use_python3 = True),
            properties = {
                "upstream_id": upstream,
                "upstream_url": _repo_url(upstream),
                "downstream_id": proj,
                "downstream_url": _repo_url(proj),
            },
            dimensions = {
                "os": os,
                "cpu": "x86-64",
                "pool": pool,
            },
            service_account = infra.SERVICE_ACCOUNT_TRY,
        )
        luci.cq_tryjob_verifier(
            builder = name,
            cq_group = cq_group,
        )

def led_recipes_tester(name, cq_group, repo_name, os = "Ubuntu-22.04"):
    """Defines a builder that uses LED to test recipe changes."""
    luci.builder(
        name = name,
        bucket = "try",
        executable = build.recipe("led_recipes_tester", use_python3 = True),
        properties = {"repo_name": repo_name},
        dimensions = {
            "os": os,
            "cpu": "x86-64",
            "pool": "luci.infra.try",
        },
        service_account = "infra-try-recipes-tester@chops-service-accounts.iam.gserviceaccount.com",
        execution_timeout = 3 * time.hour,
    )
    luci.cq_tryjob_verifier(
        builder = name,
        cq_group = cq_group,
    )

recipes = struct(
    simulation_tester = simulation_tester,
    roll_trybots = roll_trybots,
    led_recipes_tester = led_recipes_tester,
)
