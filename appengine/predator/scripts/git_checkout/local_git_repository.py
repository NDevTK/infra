# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import logging
import os
from urlparse import urlparse
import subprocess
import threading

from scripts.git_checkout import local_git_parsers
from libs.gitiles.git_repository import GitRepository
from local_libs import script_util

_CHANGELOG_FORMAT_STRING = ('commit %H%n'
                            'author %an%n'
                            'author-mail %ae%n'
                            'author-time %ad%n%n'
                            'committer %cn%n'
                            'committer-mail %ce%n'
                            'committer-time %cd%n%n'
                            '--Message start--%n%B%n--Message end--%n')
_CHANGELOGS_FORMAT_STRING = ('**Changelog start**%%n%s' %
                             _CHANGELOG_FORMAT_STRING)
CHECKOUT_ROOT_DIR = os.path.join(os.path.expanduser('~'), '.local_checkouts')


def ConvertRemoteCommitToLocal(revision):
  """Converts remote commit from gitile to local git checkout revision."""
  return 'HEAD' if revision == 'master' else revision


class LocalGitRepository(GitRepository):
  """Represents local checkout of git repository on chromium host.

  Note, to automatically check out internal repos which you have access to,
  follow the instructions in 'go/internal-repo-checkout-setup'.
  """
  lock = threading.Lock()
  # Keep track all the updated repos, so every repo only get updated once.
  _updated_repos = set()

  def __init__(self, repo_url=None):
    self._host = None
    self._repo_path = None
    self._repo_url = repo_url
    if repo_url is not None:
      parsed_url = urlparse(repo_url)
      self._host = parsed_url.netloc
      # Remove the / in the front of path.
      self._repo_path = parsed_url.path[1:]
      self._CloneOrUpdateRepoIfNeeded()

    self.changelog_parser = local_git_parsers.GitChangeLogParser()
    self.changelogs_parser = local_git_parsers.GitChangeLogsParser()
    self.blame_parser = local_git_parsers.GitBlameParser()
    self.diff_parser = local_git_parsers.GitDiffParser()

  @classmethod
  def Factory(cls): # pragma: no cover
    """Construct a factory for creating ``LocalGitRepository`` instances.

    Returns:
      A function from repo urls to ``LocalGitRepository`` instances. All
      instances produced by the returned function are novel (i.e., newly
      allocated).
    """
    return lambda repo_url: cls(repo_url)  # pylint: disable=W0108

  @property
  def repo_path(self):
    return self._repo_path

  @property
  def real_repo_path(self):
    """Absolute path of the local repository."""
    return os.path.join(CHECKOUT_ROOT_DIR, self._host, self.repo_path)

  @property
  def repo_url(self):
    """Url of remote repository which the local repo checks out from."""
    return self._repo_url

  def _CloneOrUpdateRepoIfNeeded(self):
    """Clones repo, or update it if it didn't got updated before."""
    with LocalGitRepository.lock:
      if self.repo_url in LocalGitRepository._updated_repos:
        return

      # Clone the repo if needed.
      if not os.path.exists(self.real_repo_path):
        try:
          subprocess.check_call(['git', 'clone',
                                 self.repo_url, self.real_repo_path])
        except subprocess.CalledProcessError as e:  # pragma: no cover.
          raise Exception(
              'Exception while cloning %s: %s' % (self.repo_url, e))
      # Update repo if it's already cloned.
      else:
        try:
          # Disable verbose of cd and git pull.
          with open(os.devnull, 'w') as null_handle:
            subprocess.check_call(
                'cd %s && git pull' % self.real_repo_path,
                stdout=null_handle, stderr=null_handle, shell=True)
        except subprocess.CalledProcessError as e:  # pragma: no cover.
          raise Exception(
              'Exception while updating %s: %s' % (self.repo_path, e))

      LocalGitRepository._updated_repos.add(self.repo_url)

  def _GetFinalCommand(self, command, utc=False):
    # Change local time to utc time.
    if utc:
      command = 'TZ=UTC %s --date=format-local:"%s"' % (
                command, local_git_parsers.DATETIME_FORMAT)
    return 'cd %s && %s' % (self.real_repo_path, command)

  def GetChangeLog(self, revision):
    """Returns the change log of the given revision."""
    command = ('git log --pretty=format:"%s" --max-count=1 --raw '
               '--no-abbrev %s' % (_CHANGELOG_FORMAT_STRING,
                                   ConvertRemoteCommitToLocal(revision)))
    output = script_util.GetCommandOutput(self._GetFinalCommand(command, True))
    return self.changelog_parser(output, self.repo_url)

  def GetChangeLogs(self, start_revision, end_revision):  # pylint: disable=W
    """Returns change log list in (start_revision, end_revision]."""
    command = ('git log --pretty=format:"%s" --raw --no-abbrev %s' % (
                   _CHANGELOGS_FORMAT_STRING,
                   '%s..%s' % (ConvertRemoteCommitToLocal(start_revision),
                               ConvertRemoteCommitToLocal(end_revision))))
    output = script_util.GetCommandOutput(self._GetFinalCommand(command, True))
    return self.changelogs_parser(output, self.repo_url)

  def GetChangeDiff(self, revision, path=None):  # pylint: disable=W
    """Returns the diff of the given revision."""
    command = ('git log --format="" --max-count=1 %s' %
               ConvertRemoteCommitToLocal(revision))
    if path:
      command += ' -p %s' % path
    output = script_util.GetCommandOutput(self._GetFinalCommand(command))
    return self.diff_parser(output)

  def GetBlame(self, path, revision):
    """Returns blame of the file at ``path`` of the given revision."""
    command = 'git blame --incremental %s -- %s' % (
        ConvertRemoteCommitToLocal(revision), path)
    output = script_util.GetCommandOutput(self._GetFinalCommand(command))
    return self.blame_parser(output, path, revision)

  def GetSource(self, path, revision):
    """Returns source code of the file at ``path`` of the given revision."""
    # Check whether the requested file exist or not.
    command = 'git show %s:%s' % (ConvertRemoteCommitToLocal(revision), path)
    output = script_util.GetCommandOutput(self._GetFinalCommand(command))
    return output
