#!/bin/bash
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# Change APP_ID to setup views for prod
APP_ID=chrome-test-health-staging
DATASET=test

if ! [ -z "$1" ]
  then
    APP_ID=$1
fi
if ! [  -z "$2" ]
  then
    DATASET=$2
fi

echo "creating tables for project $project_name"
bq --project_id $APP_ID mk -d "$DATASET"
sed -e s/APP_ID/$APP_ID/g -e s/DATASET/$DATASET/g \
  sql/create_raw_table.sql | \
  bq --project_id $APP_ID query --use_legacy_sql=false
sed -e s/APP_ID/$APP_ID/g -e s/DATASET/$DATASET/g \
  sql/create_daily_summary_table.sql | \
  bq --project_id $APP_ID query --use_legacy_sql=false
sed -e s/APP_ID/$APP_ID/g -e s/DATASET/$DATASET/g \
  sql/create_daily_file_summary_table.sql | \
  bq --project_id $APP_ID query --use_legacy_sql=false
sed -e s/APP_ID/$APP_ID/g -e s/DATASET/$DATASET/g \
  sql/create_weekly_file_summary_table.sql | \
  bq --project_id $APP_ID query --use_legacy_sql=false
sed -e s/APP_ID/$APP_ID/g -e s/DATASET/$DATASET/g \
  sql/create_weekly_summary_table.sql | \
  bq --project_id $APP_ID query --use_legacy_sql=false
sed -e s/APP_ID/$APP_ID/g -e s/DATASET/$DATASET/g \
  sql/create_rdb_swarming_corrections.sql | \
  bq --project_id $APP_ID query --use_legacy_sql=false
