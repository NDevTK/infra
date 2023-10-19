#!/bin/bash
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# this is a temporary hack for testing purpose
# this environmental variable should be set differently
# once the folder containing the service_account.json is loaded
export SERVICE_ACCOUNT_KEY_PATH="service_account.json"
export GCS_IMAGE_BUCKET="chromeos-distributed-fleet-s4p"
./satlabrpcserver