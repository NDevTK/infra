#!/bin/sh
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

rm -Rf static
if [ ! -d "frontend/build" ]
then
  echo "Please build the frontend (frontend/build) first"
  exit 1
fi
cp -R frontend/build static
gae.py upload --switch --app-id=google.com:chrome-test-health-staging
