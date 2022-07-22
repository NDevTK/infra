# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""LKGR finder cron."""

load("//lib/build.star", "build")

luci.builder(
    name = "chromium-lkgr-finder",
    bucket = "cron",
    executable = build.recipe("lkgr_finder", use_python3 = True),
    service_account = "chromium-lkgr-finder-builder@chops-service-accounts.iam.gserviceaccount.com",
    dimensions = {
        "builderless": "1",
        "os": "Ubuntu",
        "cpu": "x86-64",
        "cores": "8",
        "pool": "luci.infra.cron",
    },
    execution_timeout = 2 * time.hour,
    schedule = "with 3000s interval",
)

luci.list_view_entry(
    list_view = "cron",
    builder = "chromium-lkgr-finder",
)
