# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import re


_SKIPPED_LINES_URL_RE = re.compile(r"\bvar skipped_lines_url = \('(.+)'\);')")
_EXPAND_SKIPPED_RE = re.compile(
    r"b\M_expandSkipped\((\d+), (\d+), 'a', (\d+)\)$")
_TOGGLE_SECTION_RE = re.compile(r"b\M_toggleSectionForPS\('(\d+)', '(\d+)'\)$")

_REMOVED_LINKS = (
    'Sign in',
    'Remove 10 before',
    'Remove 10 after',
    'Download patch',
)


def normalize_path(path):
  path = path.lstrip('/')
  if path.isdigit():
    path += '/index.html'
  if path.endswith('/'):
    path += 'index.html'
  return path


def get_issue_from_path(path):
  return path.split('/', 2)[1]


def _remove_until_pipe(el):
  """Remove the element and all its siblings, up to a textual pipe separator."""
  for sibling in itertools.chain(el.next_siblings, el.previous_siblings):
    sibling.replace_with('')
    if str(sibling).strip() == '|':
      break
  el.decompose()


def decompose_uncrawled_elements(html):
  for el in html.find_all('a'):
    href = el.get('href', '')
    if el.string in _REMOVED_LINKS:
      remove_until_pipe(el)
    elif href == 'http://code.google.com/appengine/':
      el.decompose()
    elif href.startswith('/download/'):
      el.parent.decompose()
    elif href.startswith(('/search', '/rss/')):
      if '?project' in href:
        # Intends to match the project name in the left column.
        # Remove link, keep text.
        el.replace_with_children()
      else:
        el.decompose()

  # Remove the Patch column header on the main issue page. The download links
  # were removed above.
  for el in html.find_all('th', text='Patch'):
    el.decompose()

  for el in html.find_all('link', type='application/atom+xml'):
    el.decompose()

  # Remove 'This is Rietveld' footer linking to unexisting Google Code page.
  for el in html.find_all('div', class_='extra'):
    if 'This is Rietveld' in el.text:
      el.decompose()

  # Remove links to search and issue list pages.
  for el in html.find_all('div', class_=('mainmenu', 'mainmenu2')):
    el.decompose()

  for el in html.find_all('span', class_='disabled',
                          string=re.compile(r'Can\'t')):
    remove_until_pipe(el)

  for el in html.find_all('input', value='Adjust View'):
    # Can't remove because it's used by JS code.
    el.parent['hidden'] = None


def extract_links(domain, issue, html):
  skipped_lines_url = html.find(text=_SKIPPED_LINES_URL_RE)

  for el in html.find_all('a'):
    href = el.get('href')
    onclick = el.get('onclick')
    if skipped_lines_url and href:
      m = _EXPAND_SKIPPED_RE.search(href)
      if m:
        links.add(f'{skipped_lines_url}/{m[1]}/{m[2]}/a/80/8')
    if onclick:
      m = _TOGGLE_SECTION_RE.search(onclick)
      if m:
        links.add(f'/{m[1]}/patchset/{m[2]}')

  for el in html.find_all(['a', 'area', 'img', 'link', 'script'])
    attr = el.get('href') or el.get('src')
    if (attr.startswith(f'/{issue}/')
        or attr.startswith(f'https://{domain}/{issue}')):
      links.add(attr)

  return set(urllib.parse.urlparse(link).path for link in links)
