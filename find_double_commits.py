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



def gen(n):
  r = Repo(url)
  # r.reify()
  r._repo_path = '/s/tmp/chromium_gclient/src'
  # r.fetch()
  master = r['refs/heads/master'].commit
  # master = r.get_commit('0f54920173bda1d13e9c89c51a1995a8542c5648')
  while master != INVALID and n > 0:
    assert master.data
    yield master
    master = master.parent
    n -=1

def review_url(data):
  if data.committer.email != 'commit-bot@chromium.org':
    return None
  # Look for "Review-Url: https://codereview.chromium.org/2491073002"
  url = data.footers.get('Review-Url')
  if url:
    return url[0]
  for l in data.message_lines:
    m = REGEX.match(l)
    if m:
      return m.group(1)
  assert False
  return None


def find_pairs(commit_gen, window_size):
  window = collections.deque()
  for c in commit_gen:
    url = review_url(c.data)
    # print c, url
    for o, o_url in window:
      if url and o_url == url:
        if (o.data.footers.get('Cr-Original-Commit-Position') or
            o.data.footers.get('Committed')):
          # probably re-land
          continue
        if o.data.message_lines[0] != c.data.message_lines[0]:
          # probably a re-land.
          continue
        yield (c, o, url)
        break
    else:
      sys.stdout.write('.')
      sys.stdout.flush()
    if len(window) >= window_size:
      window.popleft()
    window.append((c, url))

def tformat(c):
  return time.ctime(c.data.committer.timestamp.secs)

def gen_thread(n):
  q = Queue.Queue()
  done = object()
  def work():
    for i in gen(n):
      q.put(i)
    q.put(done)
  t = threading.Thread(target=work)
  t.name = 'gen'
  t.daemon = True
  t.start()
  i = q.get()
  while i != done:
    yield i
    i = q.get()
  t.join()


def run(N, W):
  stored = []
  interrupt = False
  try:
    for older, newer, url in find_pairs(gen_thread(n=N), window_size=W):
      print
      print '=' * 80
      print url
      print tmpl % newer.hsh, tformat(newer)
      print tmpl % older.hsh, tformat(older)
      stored.append(','.join([url, tmpl % newer.hsh, tformat(newer), tmpl %
        older.hsh, tformat(older)]))
      # print '%s => %s: %s' %(older.hsh[:8], newer.hsh[:8], url)
      #assert older.data.message_lines == newer.data.message_lines
      #print
      #print older.data.committer
      #print '\t' + '\n\t'.join(older.data.message_lines)
      #print
      print '=' * 80
  except KeyboardInterrupt:
    interrupt = True

  print
  print 'total%s: %s' % (' so far bcz interrupted' if interrupt else '', len(stored))
  with open('stored.%s.%s.csv' % (N, W), 'w') as f:
    f.write('url,new_commit,new_commit_time,old_commit,old_commit_time\n')
    f.write('\n'.join(stored))
    f.write('\n')


if __name__ == "__main__":
  run(*map(int, sys.argv[1:]))
