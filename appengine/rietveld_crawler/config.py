# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

domain = 'codereview.chromium.org'
rietveld_project = 'chromiumcodereview-hr'
storage_project = 'lemur-253922'
bucket_name = 'chromium-rietveld'

def get_config_from_args():
  global domain, rietveld_project, storage_project, bucket_name

  parser = argparse.ArgumentParser(description='Crawl Rietveld instances')
  parser.add_argument(
      '-d', '--domain', default=domain,
      help='The domain where the Rietveld instance to crawl is served.')
  parser.add_argument(
      '-p', '--rietveld-project', default=rietveld_project,
      help='The project where the Rietveld instance to crawl is deployed.')
  parser.add_argument(
      '-s', '--storage-project', default=storage_project,
      help='The project where to store the crawled issue pages.')
  parser.add_argument(
      '-b', '--bucket-name', default=bucket_name,
      help='The name of the bucket where to store the crawled issue pages.')
  args = parser.parse_args()

  domain = args.domain
  rietveld_project = args.rietveld_project
  storage_project = args.storage_project
  bucket_name = args.bucket_name
