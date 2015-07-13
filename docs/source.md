<!--
Copyright 2015 The Chromium Authors. All rights reserved.
Use of this source code is governed by a BSD-style license that can be
found in the LICENSE file.
-->

# Source code

Most of the chrome infra source code lives in these repos:

* [infra.git](..): primary public repo.
  Contains many AppEngine apps, monitoring libraries.
* [infra\_internal.git](https://chrome-internal.googlesource.com/infra/infra_internal):
  Google-internal couterpart of infra.git. Contains CQ.
* [build](https://chromium.googlesource.com/chromium/tools/build/): legacy
  repo. Contains Buildbot, recipes, gatekeeper-ng.
* [build\_internal](https://chrome-internal.googlesource.com/chrome/tools/build):
  internal couterpart of build repo.
* [build\_limited](https://chrome-internal.googlesource.com/chrome/tools/build_limited/):
  minimal set of internal code that must be checked out on buildbot slaves
  that run internal builds.
* [luci-py](https://github.com/luci/luci-py) on GitHub: chromium-independent
  reusable continuous integration services, written in Python.
  Contains Swarming, Isolate, Auth service, Config service and AppEngine
  components, used by AppEngine apps in infra.git and infra\_internal.git
* [luci-go](https://github.com/luci/luci-go): like luci-py, but in Golang.
  Contains Isolate client.

## Checkout code

If you're reading this file, you're probably involved in the Chromium
project already. If this is not the case, you might want to read
[Chromium's Get the Code page](http://dev.chromium.org/developers/how-tos/get-the-code)
to get some background information. In particular,
[_depot_tools_ needs to be installed](http://dev.chromium.org/developers/how-tos/install-depot-tools).

The proper way to check out the non-GitHub repositories is to run:

    mkdir chrome_infra   # or whatever name you please
    cd chrome_infra
    fetch infra   # or `fetch infra_internal` if you are a Googler

If you would like to work and make changes in one of the SVN-based
dependencies of infra.git, you should make sure to set up git-svn
metadata for that repository:

    cd build
    git auto-svn

or:

    cd depot_tools
    git auto-svn

## Troubleshooting

If you're not running a supported distribution, `fetch infra` will
probably fail complaining that it cannot find some packages on Cloud
Storage. This happens with architecture-dependent packages like numpy,
which need to be compiled. The workaround is to build the packages for
yourself. Just run:

    infra/bootstrap/build_deps.py
    gclient runhooks

The first command will build the packages are store them locally. The
second command deploy them into `infra/ENV`. For more details on this
see [bootstraping](/bootstrap/README.md).

## For Googlers

If you are a Googler, see [more detailed
instructions](http://sites/chrome-infrastructure/getting-started) for working in the
other infra repositories.
