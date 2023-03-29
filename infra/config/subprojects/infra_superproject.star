# Copyright 2023 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Definitions of infra_superproject.git CI resources."""

load("//lib/infra.star", "infra")

REPO_URL = "https://chromium.googlesource.com/infra/infra_superproject"

infra.console_view(
    name = "infra_superproject",
    title = "infra/infra_superproject repository console",
    repo = REPO_URL,
)

infra.cq_group(
    name = "infra_superproject",
    repo = REPO_URL,
    tree_status_host = "infra-status.appspot.com",
)

def try_builder(
        name,
        os,
        recipe,
        cpu = None,
        experiment_percentage = None,
        properties = None,
        in_cq = True,
        use_python3 = True):
    infra.builder(
        name = name,
        bucket = "try",
        executable = infra.recipe(recipe, use_python3 = use_python3),
        os = os,
        cpu = cpu,
        properties = properties,
    )
    if in_cq:
        luci.cq_tryjob_verifier(
            builder = name,
            cq_group = "infra_superproject",
            experiment_percentage = experiment_percentage,
        )

try_builder(
    name = "infra-superproject-tests",
    os = "Ubuntu-18.04",
    recipe = "infra_superproject_tester",
)
