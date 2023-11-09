# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Definitions of CQ for the infra/gerrit-plugins repos."""

load("//lib/infra.star", "infra")

BASE_REPO_URL = "https://chromium.googlesource.com/infra/gerrit-plugins/"
BUILDER_NAME = "Gerrit Plugins Tester"

PLUGINS = [
    "binary-size",
    "buildbucket",
    "chromium-behavior",
    "chromium-binary-size",
    "chumpdetector",
    "code-coverage",
    "git-numberer",
    "landingwidget",
    "tricium",
]

luci.cq_group(
    name = "gerrit-plugins",
    watch = [
        cq.refset(
            repo = BASE_REPO_URL + plugin,
            refs = ["refs/heads/main"],
        )
        for plugin in PLUGINS
    ],
    user_limits = [
        cq.user_limit(
            name = "chromium-infra-emergency-quota",
            groups = ["chromium-infra-emergency-quota"],
            run = cq.run_limits(max_active = None),
        ),
        cq.user_limit(
            name = "luci-cv-quota-dogfooders",
            groups = ["luci-cv-quota-dogfooders"],
            run = cq.run_limits(max_active = 3),
        ),
    ],
)

luci.builder(
    name = BUILDER_NAME,
    bucket = "try",
    executable = infra.recipe("gerrit_plugins"),
    dimensions = {
        "os": "Ubuntu-22",
        "cpu": "x86-64",
        "pool": "luci.infra.try",
    },
    service_account = infra.SERVICE_ACCOUNT_TRY,
)

luci.cq_tryjob_verifier(
    builder = BUILDER_NAME,
    cq_group = "gerrit-plugins",
)
