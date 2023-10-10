# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Jobs that publish tarballs with Chromium source code."""

load("//lib/build.star", "build")
load("//lib/infra.star", "infra")

def builder(name, builderless = True, cores = 8, **kwargs):
    """Defines a infra.cron tarball builder.

    Args:
      name: name of the builder
      builderless: whether to request a builderless machine or not
      cores: CPU cores to request in the build
      **kwargs: additional dimensions to request
    """
    dimensions = {
        "pool": "luci.infra.cron",
        "os": "Ubuntu-22.04",
        "cpu": "x86-64",
    }
    if builderless:
        dimensions["builderless"] = "1"
    else:
        dimensions["builder"] = name
    if cores:
        dimensions["cores"] = str(cores)
    luci.builder(
        name = name,
        bucket = "cron",
        service_account = "chromium-tarball-builder@chops-service-accounts.iam.gserviceaccount.com",
        dimensions = dimensions,
        **kwargs
    )
    luci.list_view_entry(
        builder = name,
        list_view = "cron",
    )

builder(
    name = "publish_tarball_dispatcher",
    executable = build.recipe("publish_tarball"),
    execution_timeout = 10 * time.minute,
    schedule = "37 */3 * * *",  # every 3 hours
    triggers = ["publish_tarball"],
    experiments = {
        "luci.recipes.use_python3": 100,
    },
)

builder(
    name = "publish_tarball",
    executable = build.recipe("publish_tarball"),
    execution_timeout = 8 * time.hour,
    # Each trigger from 'publish_tarball_dispatcher' should result in a build.
    triggering_policy = scheduler.greedy_batching(
        max_batch_size = 1,
    ),
    builderless = False,
    cores = None,
    triggers = ["Build From Tarball"],
    experiments = {
        "luci.recipes.use_python3": 100,
    },
)

builder(
    name = "Build From Tarball",
    executable = infra.recipe("build_from_tarball"),
    execution_timeout = 8 * time.hour,
    # Each trigger from 'publish_tarball' should result in a build.
    triggering_policy = scheduler.greedy_batching(max_batch_size = 1),
    cores = 32,
    experiments = {
        "luci.recipes.use_python3": 100,
    },
)

luci.notifier(
    name = "Build From Tarball Notifier",
    on_failure = False,
    on_status_change = True,
    notify_emails = [
        "thestig@chromium.org",
        "thomasanderson@chromium.org",
    ],
    notified_by = [
        "Build From Tarball",
    ],
)

luci.notifier(
    name = "publish_tarball Notifier",
    on_failure = True,
    on_status_change = True,
    notify_emails = [
        "chromium-packagers@chromium.org",
        # https://crbug.com/1030114
        # "raphael.kubo.da.costa@intel.com",
        "thestig@chromium.org",
        "thomasanderson@chromium.org",
    ],
    notified_by = [
        "publish_tarball",
    ],
)
