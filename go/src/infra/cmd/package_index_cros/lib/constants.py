# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import os

PACKAGE_ROOT_DIR = os.path.dirname(os.path.dirname(__file__))
PACKAGE_LIB_DIR = os.path.dirname(__file__)
PACKAGE_SCRIPTS_DIR = os.path.join(PACKAGE_ROOT_DIR, 'scripts')

PRINT_DEPS_SCRIPT_PATH = os.path.join(PACKAGE_SCRIPTS_DIR, 'print_deps.py')

# Set of packages that should be fine to work with but are not handled properly
# yet.
TEMPORARY_UNSUPPORTED_PACKAGES = {
    # Reason: build dir does not contain out/Debug
    # Is built with Makefile but lists .gn in CROS_WORKON_SUBTREE.
    'chromeos-base/avtest_label_detect',

    # Reason: Fails build because it cannot find src/aosp/external/perfetto.
    # It's a go package that pretends to be an actual package. Should be
    # properly ignored.
    'dev-go/perfetto-protos',
}

# Set of packages that are not currently supported when building.
TEMPORARY_UNSUPPORTED_PACKAGES_WITH_BUILD = {}

# Set of packages that are not currently supported when building with tests.
TEMPORARY_UNSUPPORTED_PACKAGES_WITH_TESTS = {}

# Set of packages failing test run. To be skipped for test run.
PACKAGES_FAILING_TESTS = {
    "chromeos-base/vboot_reference",
    "chromeos-base/chromeos-installer",
    "chromeos-base/chromeos-init",
    "chromeos-base/chromeos-trim",
}
