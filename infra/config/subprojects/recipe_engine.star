# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Definitions of recipes-py.git (aka recipe_engine) CI resources."""

load("//lib/build.star", "build")
load("//lib/infra.star", "infra")
load("//lib/recipes.star", "recipes")

REPO_URL = "https://chromium.googlesource.com/infra/luci/recipes-py"

infra.console_view(
    name = "recipes-py",
    title = "recipes-py repository console",
    repo = REPO_URL,
)

luci.cq_group(
    name = "recipes-py",
    watch = cq.refset(
        repo = REPO_URL,
        refs = [r"refs/heads/master", r"refs/heads/main"],
    ),
    retry_config = cq.RETRY_TRANSIENT_FAILURES,
)

# Presubmit trybots.
build.presubmit(
    name = "recipes-py-try-presubmit",
    cq_group = "recipes-py",
    repo_name = "recipes_py",
    timeout_s = 900,
    os = "Ubuntu-18.04",
)
build.presubmit(
    name = "recipes-py-try-presubmit-win",
    cq_group = "recipes-py",
    repo_name = "recipes_py",
    timeout_s = 900,
    os = "Windows-10",
    experiment_percentage = 100,
)

# Recipes ecosystem.
recipes.simulation_tester(
    name = "recipe_engine-recipes-tests",
    project_under_test = "recipe_engine",
    triggered_by = luci.gitiles_poller(
        name = "recipe_engine-gitiles-trigger",
        bucket = "ci",
        repo = REPO_URL,
    ),
    console_view = "recipes-py",
)

# Recipe rolls from Recipe Engine.
recipes.roll_trybots(
    upstream = "recipe_engine",
    downstream = [
        "build",
        "chromiumos",
        "depot_tools",
        "fuchsia",
        "infra",
        # These repos are stuck in the roller. (http://bugs.skia.org/10401)
        #'skia',
        #'skiabuildbot',
    ],
    cq_group = "recipes-py",
)

# External testers (defined in another projects) for recipe rolls.
luci.cq_tryjob_verifier(
    builder = "infra-internal:try/build_limited Roll Tester (recipe_engine)",
    cq_group = "recipes-py",
)
luci.cq_tryjob_verifier(
    builder = "infra-internal:try/chrome_release Roll Tester (recipe_engine)",
    cq_group = "recipes-py",
)
