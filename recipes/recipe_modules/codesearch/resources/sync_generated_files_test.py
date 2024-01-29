#!/usr/bin/env vpython3
# coding=utf-8
# Copyright 2018 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Tests for sync_generated_files_codesearch."""

import os
from pathlib import Path
import shutil
import tempfile
from typing import Iterable, List, Optional, Set
import unittest
import unittest.mock

from recipe_modules.codesearch.resources import sync_generated_files as sync


class SyncGeneratedFilesCodesearchTest(unittest.TestCase):

  def setUp(self) -> None:
    """Set up the test case."""
    super().setUp()
    self.src_root = Path(
        tempfile.mkdtemp(suffix=f'_{self._testMethodName}_src'))
    self.dest_root = Path(
        tempfile.mkdtemp(suffix=f'_{self._testMethodName}_dest'))

  def tearDown(self) -> None:
    """Tear down the test case."""
    super().tearDown()
    shutil.rmtree(self.src_root)
    shutil.rmtree(self.dest_root)

  @unittest.mock.patch('subprocess.check_call')
  @unittest.mock.patch('subprocess.check_output')
  def test_main(self, mock_check_call, mock_check_output) -> None:
    """Basic end-to-end test."""
    # Set up as follows:
    #
    # src_root/
    # | foo.cc
    # dest_root/
    #
    # The contents of src_root should be copied to dest_root.

    basename = 'foo.cc'
    contents = 'foo contents'
    src_file = self.src_root / basename
    src_file.write_text(contents)

    sync.main([
        '--copy',
        f'{self.src_root};{self.dest_root}',
        '--message',
        'my cool commit message',
        'path/to/dest/repo',
    ])

    dest_file = self.dest_root / basename
    self.assertEqual(dest_file.read_text(), contents)

  def test_copy_files_basic(self) -> None:
    """Test a basic copy."""
    # Set up as follows:
    #
    # src_root/
    # | foo.cc
    # dest_root/
    #
    # The contents of src_root should be copied to dest_root.

    basename = 'foo.cc'
    contents = 'foo contents'
    src_file = self.src_root / basename
    src_file.write_text(contents)

    sync.copy_generated_files(str(self.src_root), str(self.dest_root))

    dest_file = self.dest_root / basename
    self.assertEqual(dest_file.read_text(), contents)

  def test_copy_files_nested(self) -> None:
    """Test that we can copy files in nested dirs."""
    # Set up as follows:
    #
    # src_root/
    # | foo.cc
    # | dir1/
    # | | bar.css
    # | | baz.js
    # | dir2/
    # | | dir21/
    # | | | quux.json
    # | | | zip.txt
    # dest_root/
    #
    # The contents of src_root should be copied to dest_root.

    relative_dir_paths: Tuple(Path) = (
        Path('dir1'),
        Path('dir2'),
        Path('dir2', 'dir21'),
    )
    relative_paths_to_contents: Dict[Path, str] = {
        Path('foo.cc'): 'foo contents',
        Path('dir1', 'bar.css'): 'bar contents',
        Path('dir2', 'baz.js'): 'baz contents',
        Path('dir2', 'dir21', 'quux.json'): 'quux contents',
        Path('dir2', 'dir21', 'zip.txt'): 'zip contents',
    }

    for relative_dir_path in relative_dir_paths:
      src_dir = self.src_root / relative_dir_path
      src_dir.mkdir()
    for relative_path, contents in relative_paths_to_contents.items():
      src_file = self.src_root / relative_path
      src_file.write_text(contents)

    sync.copy_generated_files(str(self.src_root), str(self.dest_root))

    for relative_dir_path in relative_dir_paths:
      dest_dir = self.dest_root / relative_dir_path
      self.assertTrue(dest_dir.exists())
    for relative_path, contents in relative_paths_to_contents.items():
      dest_file = self.dest_root / relative_path
      self.assertEqual(dest_file.read_text(), contents)

  def test_copy_files_not_allowlisted(self) -> None:
    """Test that we don't copy any files with non-allowlisted extensions."""
    # Set up as follows:
    #
    # src_root/
    # | foo.crazy
    # dest_root/
    #
    # foo.crazy should not be copied to dest_root, since it doesn't have
    # an allowlisted file extension.

    basename = 'foo.crazy'
    contents = 'foo contents'
    src_file = self.src_root / basename
    src_file.write_text(contents)

    sync.copy_generated_files(str(self.src_root), str(self.dest_root))

    dest_file = self.dest_root / basename
    self.assertFalse(dest_file.exists())

  def test_copy_files_contents_changed(self) -> None:
    """Test that if a dest file already exists, it will be overwritten."""
    # Set up as follows:
    #
    # src_root/
    # | foo.cc
    # dest_root/
    # | foo.cc (but with different contents than in src)
    #
    # The contents of dest_root/foo.cc should be overwritten.

    basename = 'foo.cc'
    src_contents = 'new foo contents'
    src_file = self.src_root / basename
    src_file.write_text(src_contents)

    original_dest_contents = 'old foo contents'
    dest_file = self.dest_root / basename
    dest_file.write_text(original_dest_contents)

    sync.copy_generated_files(str(self.src_root), str(self.dest_root))

    self.assertEqual(dest_file.read_text(), src_contents)

  def test_copy_files_delete_no_longer_existing_files(self) -> None:
    """Test that any existng dest files that don't exist in src get deleted."""
    # Set up as follows:
    #
    # src_root/
    # dest_root/
    # | the_dir/
    # | | the_file.cc
    #
    # The contents of dest_root don't exist in src_root, so they should get
    # deleted.

    dest_dir = self.dest_root / 'the_dir'
    dest_dir.mkdir()

    dest_file = dest_dir / 'the_file.cc'
    dest_file.write_text('the data')

    sync.copy_generated_files(str(self.src_root), str(self.dest_root))

    self.assertFalse(dest_dir.exists())
    self.assertFalse(dest_file.exists())

  def test_copy_files_delete_nested_empty_dirs(self) -> None:
    """Test that we recursively delete empty dest dirs after syncing."""
    # Set up as follows:
    #
    # src_root/
    # dest_root/
    # | outer_dir/
    # | | inner_dir/
    # | | | the_file.cc
    #
    # First, the_file.cc should be deleted from dest_root, because it doesn't
    # exist in src_root.
    # Then, inner_dir should be deleted, since it's now empty.
    # Finally, outer_dir should be deleted, since it's now empty.
    dest_outer_dir = self.dest_root / 'outer_dir'
    dest_outer_dir.mkdir()

    dest_inner_dir = dest_outer_dir / 'inner_dir'
    dest_inner_dir.mkdir()

    dest_file = dest_inner_dir / 'the_file.cc'
    dest_file.write_text('the data')

    sync.copy_generated_files(str(self.src_root), str(self.dest_root))

    self.assertFalse(dest_outer_dir.exists())
    self.assertFalse(dest_inner_dir.exists())
    self.assertFalse(dest_file.exists())

  def test_copy_files_delete_excluded_files(self) -> None:
    """Test that non-allowlisted files get deleted from dest."""
    # Set up as follows:
    #
    # src_root/
    # | the_dir/
    # | | the_file.woah
    # dest_root/
    # | the_dir/
    # | | the_file.woah
    #
    # Even though the_file.woah exists in src_root, it shouldn't be copied,
    # since it doesn't have an allowlisted extension.
    # Even though the_file.woah exists in dest_root, it should be deleted,
    # since no file was synced to it.
    # Finally, dest_root/the_dir/ should be deleted, since it's now empty.
    dir_basename = 'the_dir'
    file_basename = 'the_file.woah'
    file_contents = 'the_data'

    src_dir = self.src_root / dir_basename
    src_dir.mkdir()
    src_file = src_dir / file_basename
    src_file.write_text(file_contents)

    dest_dir = self.dest_root / dir_basename
    dest_dir.mkdir()
    dest_file = dest_dir / file_basename
    dest_file.write_text(file_contents)

    sync.copy_generated_files(str(self.src_root), str(self.dest_root))

    self.assertFalse(dest_dir.exists())
    self.assertFalse(dest_file.exists())

  def test_copy_files_with_secrets(self) -> None:
    """Test that we avoid copying any files with secrets in them."""
    # Set up as follows:
    #
    # src_root/
    # | foo.json
    # | creds1.json (contains a secret)
    # | creds2.json (contains a secret)
    # | creds3.json (contains invalid utf8, and a secret)
    # dest_root/
    # | creds2.json (contains a secret)
    # | creds3.json (contains invalid utf8, and a secret)
    #
    # foo.json should be copied to dest_root.
    # None of the creds files should be copied, since they all contain secrets.
    # The creds files already present in dest-root should be deleted.

    safe_file_basename = 'foo.json'
    safe_file_contents = 'foo contents'
    src_safe_file = self.src_root / safe_file_basename
    src_safe_file.write_text(safe_file_contents)
    dest_safe_file = self.dest_root / safe_file_basename

    creds1_basename = 'creds1.json'
    creds1_contents = '"accessToken": "ya29.c.dontuploadme"'
    src_creds1_file = self.src_root / creds1_basename
    dest_creds1_file = self.dest_root / creds1_basename
    src_creds1_file.write_text(creds1_contents)

    creds2_basename = 'creds2.json'
    creds2_contents = '"code": "4/topsecret"'
    src_creds2_file = self.src_root / creds2_basename
    dest_creds2_file = self.dest_root / creds2_basename
    src_creds2_file.write_text(creds2_contents)
    dest_creds2_file.write_text(creds2_contents)

    creds3_basename = 'creds3.json'
    creds3_contents = b'\n'.join((bytes([0xfa]), b'"code": "4/topsecret"'))
    src_creds3_file = self.src_root / creds3_basename
    dest_creds3_file = self.dest_root / creds3_basename
    src_creds3_file.write_bytes(creds3_contents)
    dest_creds3_file.write_bytes(creds3_contents)

    sync.copy_generated_files(str(self.src_root), str(self.dest_root))

    self.assertEqual(dest_safe_file.read_text(), safe_file_contents)

    for dest_file, failure_message in (
        (dest_creds1_file, 'creds1.json should not be synced'),
        (dest_creds2_file, 'creds2.json should have been deleted'),
        (dest_creds3_file, 'creds3.json should have been deleted'),
    ):
      self.assertFalse(dest_file.exists(), msg=failure_message)

  def test_copy_files_kzip_suffix_set(self) -> None:
    """Test that we don't sync, and do delete, files not in the suffix set."""
    # Set up as follows:
    #
    # src_root/
    # | foo.cc
    # | bar.cc
    # dest_root/
    # | bar.cc
    # | baz.cc
    #
    # kzip_suffixes will specify that we should only sync foo.cc.
    # Thus, bar.cc shouldn't be copied, and both files already in dest_root
    # should be deleted.

    for (basename, contents) in (
        # foo.cc is mentioned in kzip_suffixes, so it'll be copied.
        ('foo.cc', 'foo contents'),
        # bar.cc isn't in kzip_suffixes, so it won't be copied.
        ('bar.cc', 'bar contents'),
    ):
      src_path = self.src_root / basename
      src_path.write_text(contents)

    for (basename, contents) in (
        # bar.cc isn't in kzip_suffixes, so it'll be deleted, even though it's
        # in the source root.
        ('bar.cc', 'bar contents'),
        # baz.cc isn't in kzip_suffixes, so it'll be deleted.
        ('baz.cc', 'baz.contents'),
    ):
      dest_path = self.dest_root / basename
      dest_path.write_text(contents)

    kzip_suffixes: Set[str] = {'foo.cc'}

    sync.copy_generated_files(
        str(self.src_root),
        str(self.dest_root),
        kzip_input_suffixes=kzip_suffixes)

    self.assertTrue((self.dest_root / 'foo.cc').exists())
    self.assertFalse((self.dest_root / 'bar.cc').exists())
    self.assertFalse((self.dest_root / 'baz.cc').exists())

  def test_copy_files_ignore(self) -> None:
    """Test that we don't copy files in the ignore set."""
    # Set up as follows:
    #
    # src_root/
    # | relevant.cc
    # | ignorefile.cc
    # | ignoredir/
    # | | inner.cc
    # dest_root/
    # | ignorefile.cc
    # | ignoredir/
    # | | inner.cc
    # | | extra.cc
    #
    # We'll sync with ignore={ignorefile.cc, ignoredir}.
    # relevant.cc should be synced like normal, since it's not ignored.
    # ignorefile.cc and ignoredir/, and the contents of ignoredir/, should not
    # be synced, and should be deleted from dest_root.

    # relevant_file exists in src, but not in dest.
    # It should be synced like normal.
    relevant_file_basename = 'relevant.cc'
    relevant_file_contents = 'relevant contents'
    src_relevant_file = self.src_root / relevant_file_basename
    src_relevant_file.write_text(relevant_file_contents)
    dest_relevant_file = self.dest_root / relevant_file_basename

    # ignorefile will be named in the ignore set.
    # Since it already exists in dest, it should get deleted.
    ignorefile_basename = 'ignorefile.cc'
    ignorefile_contents = 'ignorefile contents'
    src_ignorefile = self.src_root / ignorefile_basename
    src_ignorefile.write_text(ignorefile_contents)
    dest_ignorefile = self.dest_root / ignorefile_basename
    dest_ignorefile.write_text(ignorefile_contents)

    # ignoredir will be named in the ignore set.
    # Since it already exists in dest, it and its children should get deleted.
    ignoredir_basename = 'ignoredir'
    src_ignoredir = self.src_root / ignoredir_basename
    src_ignoredir.mkdir()
    dest_ignoredir = self.dest_root / ignoredir_basename
    dest_ignoredir.mkdir()

    # inner_ignorefile is a file within ignoredir.
    # It already exists in both src and dest.
    inner_ignorefile_basename = 'inner.cc'
    inner_ignorefile_contents = 'inner ignorefile contents'
    src_inner_ignorefile = (
        self.src_root / ignoredir_basename / inner_ignorefile_basename)
    src_inner_ignorefile.write_text(inner_ignorefile_contents)
    dest_inner_ignorefile = (
        self.dest_root / ignoredir_basename / inner_ignorefile_basename)
    dest_inner_ignorefile.write_text(inner_ignorefile_contents)

    # extra_inner_ignorefile is also a file within ignoredir.
    # It already exists in dest, but not in src.
    extra_inner_ignorefile_basename = 'extra.cc'
    extra_inner_ignorefile_contents = 'extra inner ignorefile contents'
    dest_extra_inner_ignorefile = (
        self.dest_root / ignoredir_basename / extra_inner_ignorefile_basename)
    dest_extra_inner_ignorefile.write_text(extra_inner_ignorefile_contents)

    ignore: Set[str] = {str(src_ignoredir), str(src_ignorefile)}
    sync.copy_generated_files(
        str(self.src_root), str(self.dest_root), ignore=ignore)

    self.assertTrue(dest_relevant_file.exists())
    self.assertFalse(dest_ignorefile.exists())
    self.assertFalse(dest_ignoredir.exists())
    self.assertFalse(dest_inner_ignorefile.exists())
    self.assertFalse(dest_extra_inner_ignorefile.exists())

  def test_dont_copy_tmp_files(self) -> None:
    """Make sure files in src_root/**/tmp/** don't get copied."""
    # Set up as follows:
    #
    # src_root/
    # | tmp/
    # | | foo.cc
    # | tmp.cc
    # | not_tmp/
    # | | foo.cc
    # dest_root/
    #
    # tmp/foo.cc should not be synced, since its relative path contains `tmp`.
    # However, tmp.cc and not/foo.cc should be synced, since `tmp` isn't a whole
    # dir name in either of them.

    # root/tmp/
    # Should be ignored.
    tmp_dir_basename = 'tmp'
    src_tmp_dir = self.src_root / 'tmp'
    src_tmp_dir.mkdir()

    # root/tmp/foo.cc
    # Should be igored, because it's in tmp/.
    tmp_foo_cc_basename = 'foo.cc'
    tmp_foo_cc_contents = 'tmp/foo.cc contents'
    src_tmp_foo_cc = src_tmp_dir / tmp_foo_cc_basename
    src_tmp_foo_cc.write_text(tmp_foo_cc_contents)

    # root/tmp.cc
    # Filename contains the substring "tmp", but we should only ignore exact
    # matches. Thus, should be copied.
    tmp_cc_basename = 'tmp.cc'
    tmp_cc_contents = 'tmp.cc contents'
    src_tmp_cc = self.src_root / tmp_cc_basename
    src_tmp_cc.write_text(tmp_cc_contents)

    # root/not_tmp/
    # Dir contains the substring "tmp", but we should only ignore exact matches.
    # Thus, should be copied.
    not_tmp_basename = 'not_tmp'
    src_not_tmp = self.src_root / 'not_tmp'
    src_not_tmp.mkdir()

    # root/not_tmp/foo.cc
    # Should be copied because not_tmp/ isn't ignored.
    not_tmp_foo_cc_basename = 'foo.cc'
    not_tmp_foo_cc_contents = 'not_tmp/foo.cc contents'
    src_not_tmp_foo_cc = src_not_tmp / not_tmp_foo_cc_basename
    src_not_tmp_foo_cc.write_text(not_tmp_foo_cc_contents)

    sync.copy_generated_files(str(self.src_root), str(self.dest_root))

    # /tmp/ and /tmp/foo.cc should not be synced.
    dest_tmp_dir = self.dest_root / tmp_dir_basename
    dest_tmp_foo_cc = dest_tmp_dir / tmp_foo_cc_basename
    self.assertFalse(dest_tmp_dir.exists())
    self.assertFalse(dest_tmp_foo_cc.exists())

    # /tmp.cc should be synced.
    dest_tmp_cc = self.dest_root / tmp_cc_basename
    self.assertTrue(dest_tmp_cc.exists())
    self.assertEqual(dest_tmp_cc.read_text(), tmp_cc_contents)

    # /not_tmp/ and /not_tmp/foo.cc should be synced.
    dest_not_tmp = self.dest_root / not_tmp_basename
    dest_not_tmp_foo_cc = dest_not_tmp / not_tmp_foo_cc_basename
    self.assertTrue(dest_not_tmp.exists())
    self.assertTrue(dest_not_tmp_foo_cc.exists())
    self.assertEqual(dest_not_tmp_foo_cc.read_text(), not_tmp_foo_cc_contents)


if __name__ == '__main__':
  unittest.main()
