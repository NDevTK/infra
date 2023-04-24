#!/bin/bash
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# this is a temporary hack for testing purpose
# this environmental variable should be set differently
# once the folder containing the service_account.json is loaded
export GOOGLE_APPLICATION_CREDENTIALS="service_account.json"
./satlabrpcserver