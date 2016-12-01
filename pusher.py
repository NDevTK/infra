import logging
import os
import subprocess
import sys
import time

FMT='%(asctime)s %(levelname)-8s %(filename)s:%(lineno)d %(message)s'
logging.basicConfig(level=logging.DEBUG, format=FMT)

log = logging.getLogger('MAIN')

def git(*a):
  log.debug('running %s', ' '.join(['git'] + list(a)))
  return subprocess.check_output(['git'] + list(a)).strip()

def refs_to_go(origin='origin', which='younger'):
  assert which in ('younger', 'older')
  log.info("finding refs which are %s", which)
  def yielder():
    remor = 'remotes/' + origin
    for r in git('branch', '-a').splitlines():
      r = r.strip()
      if r.startswith(remor):
        if r.startswith(remor + '/heads/'):
          yield r, 'refs' + r[len(remor):]
        elif r.startswith(remor + '/branch-heads/'):
          yield r, 'refs' + r[len(remor):]

  def filter():
    BA = 'refs/remotes/origin/heads/master'
    cutoff_at = int(git('show', '-s', '--format=%ct', BA).strip())
    cutoff_at -= 365 * 24 * 60 * 60  # 1 year
    for r, d in yielder():
      committed_at = int(git('show', '-s', '--format=%ct', r).strip())
      is_older = committed_at < cutoff_at
      if which == 'younger' and is_older:
        log.warn('skipping too old commit %s (%s vs cutoff %s)', r,
            str(time.ctime(committed_at)),
            str(time.ctime(cutoff_at)))
        continue
      if which == 'older' and not is_older:
        log.warn('skipping too young commit %s (%s vs cutoff %s)', r,
            str(time.ctime(committed_at)),
            str(time.ctime(cutoff_at)))
        continue
      yield r, d

  refs = dict((d, (r,d)) for r, d in filter())
  log.info('found %s %d refs', which, len(refs))
  if 'refs/heads/master' in refs:
    yield refs.pop('refs/heads/master')
  for d in sorted(refs):
    yield refs[d]

def do_push(origin, r, d):
  # log.debug('pushing %s ref %s => %s', origin, r, d)
  git('push', origin, '%s:%s' % (r, d))
  # git('push', origin, '--force', '--delete', d)

def attention_required(action):
  raw_input('\n' * 10 + ' ' * 10 + '!!!!!!  ATTENTION REQUIRED  !!!!!!\n\n' +
            ' ' * 16 + action + '\n' * 2 + 'press enter when done: ')
  print

def main(origin_url, test_url, repopath):
  if not os.path.exists(repopath):
    os.mkdir(repopath)
    os.chdir(repopath)
    git('init')
    with open('.git/config', 'w') as f:
      f.write("""
					[remote "origin"]
						url = %s
						pushurl = https://bad.url/
						fetch = +refs/heads/*:refs/remotes/origin/heads/*
						fetch = +refs/branch-heads/*:refs/remotes/origin/branch-heads/*
					[remote "gnumbd"]
						url = %s
						fetch = +refs/heads/*:refs/remotes/gnumbd/heads/*
						fetch = +refs/branch-heads/*:refs/remotes/gnumbd/branch-heads/*
      """ % (origin_url, test_url))
  os.chdir(repopath)
  git('fetch', 'origin')
  git('fetch', 'gnumbd')

  log.info('disable plugin now...')
  attention_required('disable plugin now')
  for r, d in refs_to_go('origin', 'older'):
    do_push('gnumbd', r, d)

  log.info('enable plugin now...')
  attention_required('enable plugin now')
  for r, d in refs_to_go('origin', 'younger'):
    do_push('gnumbd', r, d)

if __name__ == '__main__':
  name = sys.argv[1]
  urls = {
      'v8': ('https://chromium.googlesource.com/v8/v8/',
             'https://chromium.googlesource.com/playground/gnumbd-v8/'),
  }
  origin_url, test_url = urls[name]

  file_log = logging.FileHandler('%s.push.log' % name)
  file_log.setFormatter(logging.Formatter(FMT))
  logging.getLogger().addHandler(file_log)
  log.debug("\n" * 10)
  log.info('starting on %s', name)

  try:
    main(origin_url, test_url, name)
  except:
    log.exception('aborting because...')
    sys.exit(1)


# Hints for V8:
# $ export BA=c6e74e707913f176c1bca1e8b7155e20aa4b3c7d   # Dec 31 2014.
# $ git log $BA..remotes/origin/heads/master --pretty=oneline | wc -l
#   15475 commits
# Push slowly:
# $ git push gnumbd remotes/origin/heads/master~10000:refs/heads/master
# $ git push gnumbd remotes/origin/heads/master~5000:refs/heads/master
# $ git push gnumbd remotes/origin/heads/master~1:refs/heads/master
