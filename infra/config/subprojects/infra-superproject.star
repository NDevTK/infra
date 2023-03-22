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
        cpu = None,
        recipe = None,
        experiment_percentage = None,
        properties = None,
        in_cq = True,
        use_python3 = True):
    infra.builder(
        name = name,
        bucket = "try",
        executable = infra.recipe(recipe or "infra_repo_trybot", use_python3 = use_python3),
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
    name = "infra-superproject-try-frontend",
    os = "Ubuntu-18.04",
    recipe = "infra_frontend_tester",
)

try_builder(
    name = "infra-superproject-try-bionic-64",
    os = "Ubuntu-18.04",
    properties = {"go_version_variant": "bleeding_edge"},
)

try_builder(
    name = "infra-superproject-try-mac",
    os = "Mac-10.15",
    properties = {"go_version_variant": "legacy"},
)

# It is occasionally useful to test code on OSX 10.14, but we don't have enough
# capacity to have this trybot in CQ by default. It can be triggered manually
# though.
try_builder(
    name = "infra-superproject-try-mac-10.14",
    os = "Mac-10.14",
    properties = {"go_version_variant": "legacy"},
    in_cq = False,
)

try_builder(
    name = "infra-superproject-try-win",
    os = "Windows-10",
)
