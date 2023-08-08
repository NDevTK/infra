#!/bin/bash
# Copyright 2019 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# Change APP_ID to setup views for prod
APP_ID=sheriff-o-matic-staging
# APP_ID=sheriff-o-matic

define_views() {
  local project_name=$1
  local project_condition=$2

  echo "creating data set and views for project $project_name"
  bq --project_id $APP_ID mk -d "$project_name"
  sed -e s/APP_ID/$APP_ID/g -e s/PROJECT_NAME/"$project_name"/g -e s/PROJECT_FILTER_CONDITIONS/"$project_condition"/g step_status_transitions_customized.sql | bq --project_id $APP_ID query --use_legacy_sql=false
  sed -e s/APP_ID/$APP_ID/g -e s/PROJECT_NAME/"$project_name"/g -e s/PROJECT_FILTER_CONDITIONS/"$project_condition"/g failing_steps_customized.sql | bq query --project_id $APP_ID --use_legacy_sql=false
  sed -e s/APP_ID/$APP_ID/g -e s/PROJECT_NAME/"$project_name"/g sheriffable_failures.sql | bq --project_id $APP_ID query --use_legacy_sql=false
}

define_views "chrome" "create_time > TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY) AND (project IN ('chromium', 'chrome') OR STARTS_WITH(project, 'chromium-m') OR STARTS_WITH(project, 'chrome-m'))"
define_views "angle" "create_time > TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY) AND ((project = \"angle\" AND bucket=\"ci\") OR (project = \"chromium\" AND bucket=\"ci\"))"
define_views "fuchsia" "create_time > TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY) AND project IN ('fuchsia', 'turquoise', 'cobalt-analytics')"
define_views "chromeos" "create_time > TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY) AND project IN ('chromeos')"
