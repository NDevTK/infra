#!/usr/bin/env python3

# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

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

import common


class RietveldSpider(scrapy.spiders.Spider):
  name = 'rietveld'

  def __init__(self, domain):
    self.name = domain
    self.allowed_domains = [domain]
    self.start_urls = ['https://' + domain]
    super().__init__()

  def parse(self, response):
    url = urllib.parse.urlsplit(response.url)
    path = common.make_path(url.path, url.query)

    is_diff_path = re.fullmatch(r'\d+/(diff|diff2|patch)/.*', path)
    is_image_path = re.fullmatch(r'\d+/(image)/.*', path)

    if isinstance(response, scrapy.http.HtmlResponse) and not is_image_path:
      text = response.text
      if is_diff_path:
        text = text.replace('Please Sign in to add in-line comments.', '')

      html = bs4.BeautifulSoup(text, features='lxml')

      for el in html.find_all('a'):
        if el.string in ('Sign in', 'Expand 10 before', 'Expand 10 after'):
          remove_until_pipe(el)

        elif el.get('href', '').startswith(('/search', '/rss/')):
          if '?project' in el['href']:
            el.replace_with_children()  # Remove link, keep text.
          else:
            el.decompose()

        elif el.get('href', '') == 'http://code.google.com/appengine/':
          el.decompose()

      for el in html.find_all('link', {'type': 'application/atom+xml'}):
        el.decompose()

      for el in html.find_all('div', class_='extra'):
        if 'This is Rietveld' in el.text:
          el.decompose()
      for el in html.find_all('div', class_='counter'):
        el.decompose()

      for el in html.find_all('span', class_='disabled',
                              string=re.compile(r'Can\'t')):
        remove_until_pipe(el)
      for el in html.find_all('input', {'value': 'Adjust View'}):
        hide(el.parent)  # Can't remove because it's used by JS code.

      if is_diff_path:
        for el in html.find_all('img'):
          if '/image/' in el['src']:
            filename = html.find('title').string.split(' -\n')[0].strip()
            filename = posixpath.basename(filename)
            query = urllib.parse.urlencode({'filename': filename})
            el['src'] += '?' + query

      response = response.replace(body=html.encode())

      for extractor in [
          scrapy.linkextractors.LinkExtractor(
              tags=['a', 'area', 'img', 'link', 'script'],
              attrs=['href', 'src'],
              deny_extensions=[],
              allow_domains=self.allowed_domains),
          ExpandSkippedLinkExtractor(),
          ToggleSectionLinkExtractor()]:
        for link in extractor.extract_links(response):
          yield scrapy.http.Request(url=link.url)

    # Confirm that the MIME types that server.py will be serving match what
    # Rietveld actually returns. This is a substitute for having to store them.
    actual = response.headers['Content-Type'].decode()
    deduced = common.content_type(url.path, url.query)
    assert actual.lower() == deduced, '{!r}: {!r} != {!r}'.format(
        path, deduced, actual)

    file_path = os.path.join(self.name, path)
    os.makedirs(os.path.dirname(file_path), exist_ok=True)
    with open(file_path, 'wb') as f:
      f.write(response.body)


def remove_until_pipe(el):
  """Remove the element and all its siblings, up to a textual pipe separator."""
  for siblings in el.next_siblings, el.previous_siblings:
    for sibling in list(siblings):
      sibling.replace_with('')
      if str(sibling).strip() == '|':
        break
  el.decompose()


def hide(el):
  el['hidden'] = None


class ExpandSkippedLinkExtractor(scrapy.linkextractors.LinkExtractor):

  def __init__(self):
    super().__init__(tags=['a'], attrs=['href'], deny_extensions=[],
                     process_value=self.process_value)

  def extract_links(self, response):
    m = re.search(r"\bvar skipped_lines_url = \('(.+)'\);", response.text)
    if not m:
      return []
    self.skipped_lines_url = m[1]
    return super().extract_links(response)

  def process_value(self, value):
    m = re.search(r"\bM_expandSkipped\((\d+), (\d+), 'a', (\d+)\)$", value)
    if m:
      return posixpath.join(
          self.skipped_lines_url, '{}/{}/a/80/8'.format(m[1], m[2]))


class ToggleSectionLinkExtractor(scrapy.linkextractors.LinkExtractor):

  def __init__(self):
    super().__init__(tags=['a'], attrs=['onclick'], deny_extensions=[],
                     process_value=self.process_value)

  def process_value(self, value):
    m = re.search(r"\bM_toggleSectionForPS\('(\d+)', '(\d+)'\)$", value)
    if m:
      return '/{}/patchset/{}'.format(m[1], m[2])


def main(argv):
  (domain,) = argv[1:]

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
  process.crawl(RietveldSpider, domain)
  process.start()


if __name__ == '__main__':
  sys.exit(main(sys.argv))
