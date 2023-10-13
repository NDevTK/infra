<!--
Copyright 2023 The Chromium Authors
Use of this source code is governed by a BSD-style license that can be
found in the LICENSE file.
-->

# Suite Scheduler **__v1.5__**
SuiteScheduler v1.5 is the partial rewrite of the suite scheduler cron scheduler
before we fully re-imagine the service. This service will offer just the core
services of SuiteScheduler and not implement pipelines such as android build,
firmware builds, nor multi-dut builds.

More information can be found at: go/suitescheduler-v15


## Current State: WIP
The current project is capable of fetching the configuration files and ingesting
them memory for application usage.


Until b/305286743 is resolved, the .cfg files will need to be stored locally.
Once that issue is resolved then we can delete the local CLs so that the program
only functions using the internet fetched file just like the current production
Suite Scheduler does.
