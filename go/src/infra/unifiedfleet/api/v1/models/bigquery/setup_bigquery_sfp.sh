#!/bin/sh
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# Historic Reference: http://google3/third_party/luci_py/latest/appengine/swarming/setup_bigquery.sh

set -eu

cd "$(dirname $0)"

if ! (which bq) > /dev/null; then
  echo "Please install 'bq' from gcloud SDK"
  echo "  https://cloud.google.com/sdk/install"
  exit 1
fi

if ! (which bqschemaupdater) > /dev/null; then
  echo "Please install 'bqschemaupdater' from Chrome's infra.git"
  echo "  Checkout infra.git then run: eval \`./go/env.py\`"
  exit 1
fi

if [ $# != 1 ]; then
  echo "usage: setup_bigquery.sh <instanceid>"
  echo ""
  echo "Pass one argument which is the instance name"
  exit 1
fi

APPID=$1
DATASET=sfp

echo "- Make sure the BigQuery API is enabled for the project:"
# It is enabled by default for new projects, but it wasn't for older projects.
gcloud services enable --project "${APPID}" bigquery-json.googleapis.com

# Permission is grantes via overground, skipping here

echo "- Create the dataset:"
echo ""
echo "  Warning: On first 'bq' invocation, it'll try to find out default"
echo "    credentials and will ask to select a default app; just press enter to"
echo "    not select a default."

if ! (bq --location=US mk --dataset \
  --description 'unified fleet system statistics' "${APPID}":"${DATASET}"); then
  echo ""
  echo "Dataset creation failed. Assuming the dataset already exists. At worst"
  echo "the following command will fail."
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.ChromePlatformRow  \
    -table "${APPID}"."${DATASET}".chrome_platforms); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.chrome_platforms"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.ChromePlatformRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".chrome_platforms_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.chrome_platforms_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.VlanRow  \
    -table "${APPID}"."${DATASET}".vlans); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.vlans"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.VlanRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".vlans_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.vlans"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.MachineRow  \
    -table "${APPID}"."${DATASET}".machines); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.machines"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.MachineRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".machines_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.machines_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.RackRow  \
    -table "${APPID}"."${DATASET}".racks); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.racks"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.RackRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".racks_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.racks_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.MachineLSEPrototypeRow  \
    -table "${APPID}"."${DATASET}".machine_lse_prototypes); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.machine_lse_prototypes"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.MachineLSEPrototypeRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".machine_lse_prototypes_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.machine_lse_prototypes_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.MachineLSERow  \
    -table "${APPID}"."${DATASET}".machine_lses); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.machine_lses"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.MachineLSERow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".machine_lses_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.machine_lses_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.RPMRow  \
    -table "${APPID}"."${DATASET}".rpms); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.rpms"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.RPMRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".rpms_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.rpms_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.SwitchRow  \
    -table "${APPID}"."${DATASET}".switches); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.switches"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.SwitchRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".switches_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.switches_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.NicRow  \
    -table "${APPID}"."${DATASET}".nics); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.nics"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.NicRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".nics_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.nics_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.DHCPConfigRow  \
    -table "${APPID}"."${DATASET}".dhcps); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.dhcps"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.DHCPConfigRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".dhcps_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.dhcps_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.IPRow  \
    -table "${APPID}"."${DATASET}".ips); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.ips"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.IPRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".ips_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.ips_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.StateRecordRow  \
    -table "${APPID}"."${DATASET}".state_records); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.state_records"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.StateRecordRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".state_records_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.state_records_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.ChangeEventRow  \
    -table "${APPID}"."${DATASET}".change_events); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.change_events"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.VMRow  \
    -table "${APPID}"."${DATASET}".vms); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.vms"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.VMRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".vms_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.vms_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.AssetRow  \
    -table "${APPID}"."${DATASET}".assets); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.assets"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.AssetRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".assets_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.assets_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.DUTStateRecordRow  \
    -table "${APPID}"."${DATASET}".dutstates); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.dutstates"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.DUTStateRecordRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".dutstates_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.dutstates_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.CachingServiceRow  \
    -table "${APPID}"."${DATASET}".caching_services); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.caching_services"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.CachingServiceRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".caching_services_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.caching_services_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.MachineLSEDeploymentRow  \
    -table "${APPID}"."${DATASET}".machine_lse_deployments); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.machine_lse_deployments"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.MachineLSEDeploymentRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".machine_lse_deployments_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.machine_lse_deployments_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.SchedulingUnitRow  \
    -table "${APPID}"."${DATASET}".scheduling_units); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.scheduling_units"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.SchedulingUnitRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".scheduling_units_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.scheduling_units_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.HwidDataRow  \
    -table "${APPID}"."${DATASET}".hwid_data); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.hwid_data"
  echo ""
  echo "and run this script again."
  exit 1
fi

echo "- Populate the BigQuery schema:"
echo ""
echo "  Warning: On first 'bqschemaupdater' invocation, it'll request default"
echo "    credentials which is stored independently than 'bq'."
if ! (bqschemaupdater -force \
    -I ../../../../../../ \
    -message unifiedfleet.api.v1.models.bigquery.HwidDataRow  \
    -partitioning-type HOUR \
    -partitioning-expiration 3999h \
    -table "${APPID}"."${DATASET}".hwid_data_hourly); then
  echo ""
  echo ""
  echo "Oh no! You may need to restart from scratch. You can do so with:"
  echo ""
  echo "  bq rm ${APPID}:${DATASET}.hwid_data_hourly"
  echo ""
  echo "and run this script again."
  exit 1
fi
