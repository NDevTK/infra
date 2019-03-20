# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Jobs that publish tarballs with Chromium source code."""

load('//lib/infra.star', 'infra')


luci.builder(
    name = 'publish_tarball_dispatcher',
    bucket = 'cron',
    recipe = infra.recipe('publish_tarball'),
    dimensions = {
        'pool': 'luci.infra.cron',
        'builder': 'publish_tarball',  # runs on same bots as 'publish_tarball'
        'os': 'Ubuntu-14.04',
    },
    service_account = 'chromium-tarball-builder@chops-service-accounts.iam.gserviceaccount.com',
    execution_timeout = 10 * time.minute,
    schedule = '37 */3 * * *',  # every 3 hours
    triggers = ['publish_tarball'],
)

luci.builder(
    name = 'publish_tarball',
    bucket = 'cron',
    recipe = infra.recipe('publish_tarball'),
    dimensions = {
        'pool': 'luci.infra.cron',
        'builder': 'publish_tarball',
        'os': 'Ubuntu-14.04',
    },
    service_account = 'chromium-tarball-builder@chops-service-accounts.iam.gserviceaccount.com',
    execution_timeout = 5 * time.hour,
    # Each trigger from 'publish_tarball_dispatcher' should result in a build.
    triggering_policy = scheduler.greedy_batching(max_batch_size=1),
    triggers = ['Build From Tarball'],
)

luci.builder(
    name = 'Build From Tarball',
    bucket = 'cron',
    recipe = infra.recipe('build_from_tarball'),
    dimensions = {
        'pool': 'luci.infra.cron',
        'builder': 'Build From Tarball',
        'os': 'Ubuntu-14.04',
    },
    service_account = 'chromium-tarball-builder@chops-service-accounts.iam.gserviceaccount.com',
    execution_timeout = 3 * time.hour,
    # Each trigger from 'publish_tarball' should result in a build.
    triggering_policy = scheduler.greedy_batching(max_batch_size=1),
)
