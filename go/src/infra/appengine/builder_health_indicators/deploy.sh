#!/bin/sh

# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -euf -o pipefail

if [[ "$*" == *--prod* ]]; then
  project=cr-builder-health-indicators
else
  project=cr-builder-health-ind-staging
fi

gae.py upload --switch --app-id="$project"
