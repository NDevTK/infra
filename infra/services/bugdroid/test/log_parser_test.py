# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import unittest

import infra.services.bugdroid.log_parser as log_parser


class BugLineParserTest(unittest.TestCase):
  def test_matching_bug(self):
    for bug, bug_line in [
        # Keep distinct bug numbers for easy search in case of test failures.
        (123, 'BUG=123'),
        (124, 'Bug: 124'),
        ('chromium:125', 'Bugs: chromium:125'),
    ]:
      m = log_parser.BUG_LINE_REGEX.match(bug_line)
      self.assertIsNotNone(m, '"%s" line must be matched' % bug_line)
      self.assertEqual(m.groups()[-1], str(bug),
                       '"%s" line matched to %s but %s expected.' % (
                       bug_line, m.groups()[-1], str(bug)))

  def test_not_matching_bug(self):
    for bug_line in [
        # Keep distinct bug numbers for easy search in case of test failures.
        'BUGr=123',
        'BUGS/124',
        'someBugs:',
    ]:
      m = log_parser.BUG_LINE_REGEX.match(bug_line)
      self.assertIsNone(m, '"%s" line must not be matched (got %s)' %
                           (bug_line, m.groups()) if m else None)

  def test_should_send_email(self):
    for test_case, result in [
      (None, True),
      ("Random stuff\nhereman\nBug: 12", True),
      ("Bugdroid-Send-Email: yaaaman", True),
      ("Bugdroid-Send-Email: no", False),
      ("Bugdroid-Send-Email: false", False),
    ]:
      self.assertEqual(
          result, log_parser.should_send_email(test_case))
