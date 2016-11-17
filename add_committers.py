# Run as:
#   ENV/bin/python find_double_commits.py 100000 100
#
import re
from infra.libs.git2 import Repo, Ref, INVALID
import cPickle
import sys
import collections
import time
import threading
import Queue

url = "https://chromium.googlesource.com/chromium/src"
tmpl = url + '/+/%s'
REGEX = re.compile('Review URL: (.*/\d+)( .)?')
assert (REGEX.match("Review URL: https://codereview.chromium.org/1182113008").group(
    1) == 'https://codereview.chromium.org/1182113008')
assert (REGEX.match("Review URL: https://codereview.chromium.org/1613843002 .").group(
  1) == 'https://codereview.chromium.org/1613843002')

r = Repo(url)
r._repo_path = '/s/tmp/chromium_gclient/src'

def get_author(url):
  hsh = url.split('/+/')[-1]
  assert len(hsh) == 40, hsh
  c = r.get_commit(hsh)
  return c.data.author.email

def run(fin, fout):
  first = True
  for l in fin:
    if first:
      fout.write(l.strip() + ',author\n')
      first = False
      continue
    parts = l.strip().split(',')
    a1, a2 = map(get_author, (parts[1], parts[3]))
    assert a1 == a2
    parts.append(a2)
    fout.write(','.join(parts) + '\n')

if __name__ == "__main__":
  run(sys.stdin, sys.stdout)
