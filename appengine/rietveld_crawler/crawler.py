#!/usr/bin/env python3

# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import base64
import os
import posixpath
import re
import sys
import urllib.parse

import bs4
import scrapy.crawler
import scrapy.http
import scrapy.linkextractors
import scrapy.spiders

from google.cloud import storage


IMAGE_PATH_RE = re.compile(r'\d+/image/')


class RietveldSpider(scrapy.spiders.Spider):
  name = 'rietveld'

  def __init__(self, bucket):
    self.bucket = bucket
    self.name = config.domain
    self.allowed_domains = [config.domain]
    super().__init__()

  def start_requests(self):
    for issue, private in datastore_utils.get_issues():
      yield scrapy.http.Request(
        url=f'https://{config.domain}/{issue}/',
        headers=auth.get_auth_headers(),
        cb_kwargs=dict(private=private))

  def _process_page(self, issue, response):
    if url.path.startswith(f'{issue}/{image}'):
      return set(), response

    response = response.decode('utf-8', 'ignore')
    response = response.replace('Please Sign in to add in-line comments.', '')
    html = bs4.BeautifulSoup(response, features='lxml')
    crawler_utils.decompose_uncrawled_elements(html)
    links = crawler_utils.extract_links(config.domain, issue, html)

    return links, html.encode('utf-8', 'ignore-issues')

  def parse(self, response, private):
    url = urllib.parse.urlsplit(response.url)
    path = crawler_utils.normalize_path(url.path)
    issue = crawler_utils.get_issue(path)
    page, links = self._process_page(issue, response.text)

    blob = bucket.blob(base64.b64encode(path.encode('utf-8')))
    blob.metadata = blob.metadata or {}
    blob.medatada['private'] = private
    blob.upload_from_string(
        page, content_type=response.headers['Content-Type'].decode('utf-8'))

    for link in links:
      yield scrapy.http.Request(
          url=link.url,
          headers=auth.get_auth_headers(),
          cb_kwargs=dict(private=private))


def main():
  config.get_config_from_args()

  client = storage.Client(project=config.storage_project)
  bucket = client.get_bucket(config.bucket_name)

  process = scrapy.crawler.CrawlerProcess(dict(
      LOG_LEVEL='INFO',
      CLOSESPIDER_ERRORCOUNT=100,
      REDIRECT_ENABLED=False,
      HTTPCACHE_ENABLED=True,
      HTTPCACHE_STORAGE='scrapy.extensions.httpcache.DbmCacheStorage',
      # CONCURRENT_REQUESTS are set to large values just to give AUTOTHROTTLE
      # some space for maneuvers, but the latter is what limits concurrency.
      CONCURRENT_REQUESTS=1000,
      CONCURRENT_REQUESTS_PER_DOMAIN=500,
      AUTOTHROTTLE_TARGET_CONCURRENCY=20,
      AUTOTHROTTLE_ENABLED=True,
      AUTOTHROTTLE_START_DELAY=1,
  ))
  process.crawl(RietveldSpider, bucket)
  process.start()


if __name__ == '__main__':
  sys.exit(main())
