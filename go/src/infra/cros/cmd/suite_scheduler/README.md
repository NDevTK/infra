<!--
Copyright 2023 The Chromium Authors
Use of this source code is governed by a BSD-style license that can be
found in the LICENSE file.
-->

# Suite Scheduler ****v1.5****

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

To interact with the current project use the CLI explained in the below section.

## Installation

### CIPD

If you have CIPD set up you can fetch the package from there using:

```bash
cipd install chromiumos/infra/suite_scheduler/linux-amd64 latest
```

If not you will need to set up CIPD:

```bash
# Create a directory for the CIPD root
mkdir ~/cipd
cd ~/cipd/

# Initialize the CIPD root for package installation
cipd init -force

# Add the export command to your .bashrc
echo "export PATH=\$PATH:$(pwd)" >> ~/.bashrc && source ~/.bashrc

# Install the package
cipd install chromiumos/infra/suite_scheduler/linux-amd64 latest
```

Once installed use the command by calling:

```bash
suite_scheduler -help
```

### Local

The provided makefile has build instructions for the program. To build the files
locally just run:

```bash
make build
```

This will install the package at the project root and you can run the program
using:

```bash
./suite_scheduler -help
```

## CLI

SuiteScheduler v1.5 is made as a CLI application. To run the program use one of
the below commands to access the project.

### Commands

#### Configs

```bash
suite_scheduler configs <flags>
```

The `configs` command is used to search through the SuiteScheduler configs. The
command will take in the user input and will output the configs which match the
criteria. To see all flags, and information about their usage, enter:

```bash
suite_scheduler help configs
```

### Filters

When ingesting and searching for configs the application defines two types of
filters, top and bottom level filters.

`Top-level` filters define the set of filters that largely define the config
trigger mechanism, e.g. `NEW_BUILD` or `DAILY`. These filters do not care about
the contents of the configs but rather are reducing the domain of configs that
will be sent to the `bottom-level` filters.

`Bottom-level` filters work on the inner contents of the configs, E.g. `name` or
`board`. The bottom level filters will receive its working set from the
top-level filters. This reduces the amount of expensive filtering that is
performed making the CLI run faster when working through large amounts of
configurations.

## Testing

Currently the main function has some integration tests that can be validated by
reading the output. To confirm the correctness of the CTP Requests. The output
given was tossed into a LED test (with the build image information copied over
from a valid CTP run) and verified.

Example run:

<https://chromeos-swarming.appspot.com/task?id=65bf847779131c10>

Original Run:

[go/bbid/8765603166160534657](https://goto2.corp.google.com/bbid/8765603166160534657)
