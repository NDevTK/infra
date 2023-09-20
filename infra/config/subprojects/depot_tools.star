# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Definitions of depot_tools.git CI resources."""

load("//lib/build.star", "build")
load("//lib/infra.star", "infra")
load("//lib/recipes.star", "recipes")

REPO_URL = "https://chromium.googlesource.com/chromium/tools/depot_tools"

luci.gitiles_poller(
    name = "depot_tools-gitiles-trigger",
    bucket = "ci",
    repo = REPO_URL,
    refs = ["refs/heads/main"],
)

infra.console_view(
    name = "depot_tools",
    title = "depot_tools repository console",
    repo = REPO_URL,
)

luci.cq_group(
    name = "depot_tools",
    watch = cq.refset(
        repo = REPO_URL,
        refs = ["refs/heads/main"],
    ),
    retry_config = cq.RETRY_TRANSIENT_FAILURES,
    user_limits = [
        cq.user_limit(
            name = "cv-admins-limit",
            groups = ["service-luci-change-verifier-admins"],
            run = cq.run_limits(max_active = 5),
        ),
    ],
)

# Presubmit trybots.
build.presubmit(
    name = "Depot Tools Presubmit",
    cq_group = "depot_tools",
    repo_name = "depot_tools",
    run_hooks = False,
    os = "Ubuntu-18.04",
)

build.presubmit(
    name = "Depot Tools Presubmit (win)",
    cq_group = "depot_tools",
    repo_name = "depot_tools",
    run_hooks = False,
    os = "Windows-10",
)

# Recipes ecosystem.
recipes.simulation_tester(
    name = "depot_tools-recipes-tests",
    project_under_test = "depot_tools",
    triggered_by = "depot_tools-gitiles-trigger",
    console_view = "depot_tools",
    os = "Ubuntu-18.04",
)

# Recipe rolls from Depot Tools.
recipes.roll_trybots(
    upstream = "depot_tools",
    downstream = [
        "build",
        "chromiumos",
        "infra",
        # These repos are stuck in the roller. (http://bugs.skia.org/10401)
        # 'skia',
        # 'skiabuildbot',
    ],
    cq_group = "depot_tools",
)

# External testers (defined in another project) for recipe rolls.
luci.cq_tryjob_verifier(
    builder = "infra-internal:try/build_limited Roll Tester (depot_tools)",
    cq_group = "depot_tools",
)
luci.cq_tryjob_verifier(
    builder = "infra-internal:try/chrome_release Roll Tester (depot_tools)",
    cq_group = "depot_tools",
)
