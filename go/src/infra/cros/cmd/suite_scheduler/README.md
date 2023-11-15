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
The current project is capable of fetching the configuration files, ingesting
them into memory, and building CTP requests. 

The project is not able to pull in build information from the release pub/sub
stream so the CTP requests are only partially filled in. The missing data is
related to build image requirements.

## Testing
Currently the main function has some integration tests that can be validated by
reading the output. To confirm the correctness of the CTP Requests. The output
given was tossed into a LED test (with the build image information copied over
from a valid CTP run) and verified.

Example run: 

https://chromeos-swarming.appspot.com/task?id=65bf847779131c10

Original Run:

[go/bbid/8765603166160534657](https://goto2.corp.google.com/bbid/8765603166160534657)