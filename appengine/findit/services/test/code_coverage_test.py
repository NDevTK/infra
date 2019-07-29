# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
import json
import mock

from services import code_coverage
from waterfall.test.wf_testcase import WaterfallTestCase


class CodeCoverageTest(WaterfallTestCase):

  @mock.patch.object(code_coverage.FinditHttpClient, 'Get')
  def testGetEquivalentPatchsets(self, mock_http_client):
    mock_http_client.return_value = (
        200,
        json.dumps({
            'revisions': {
                'aaaaaaaaaaa': {
                    '_number': 8,
                    'kind': 'TRIVIAL_REBASE',
                },
                'bbbbbbbbbbb': {
                    '_number': 7,
                    'kind': 'MERGE_FIRST_PARENT_UPDATE',
                },
                'ccccccccccc': {
                    '_number': 6,
                    'kind': 'NO_CODE_CHANGE',
                },
                'ddddddddddd': {
                    '_number': 5,
                    'kind': 'NO_CHANGE',
                },
                'eeeeeeeeeee': {
                    '_number': 4,
                    'kind': 'REWORK',
                },
                'fffffffffff': {
                    '_number': 3,
                    'kind': 'TRIVIAL_REBASE',
                },
            },
        }), None)

    result = code_coverage.GetEquivalentPatchsets(
        'chromium-review.googlesource.com', 'chromium/src', '12345', '8')
    self.assertListEqual([8, 7, 6, 5, 4], result)

  @mock.patch.object(code_coverage.FinditHttpClient, 'Get')
  def testNoEquivalentPatchsets(self, mock_http_client):
    mock_http_client.return_value = (200,
                                     json.dumps({
                                         'revisions': {
                                             'aaaaaaaaaaa': {
                                                 '_number': 8,
                                                 'kind': 'REWORK',
                                             },
                                             'bbbbbbbbbbb': {
                                                 '_number': 7,
                                                 'kind': 'NO_CODE_CHANGE',
                                             },
                                         },
                                     }), None)

    result = code_coverage.GetEquivalentPatchsets(
        'chromium-review.googlesource.com', 'chromium/src', '12345', '8')
    self.assertListEqual([8], result)

  @mock.patch.object(code_coverage.FinditHttpClient, 'Get')
  def testEquivalentPatchsetsIsCached(self, mock_http_client):
    mock_http_client.return_value = (200,
                                     json.dumps({
                                         'revisions': {
                                             'aaaaaaaaaaa': {
                                                 '_number': 8,
                                                 'kind': 'REWORK',
                                             },
                                         },
                                     }), None)

    self.assertEqual(0, mock_http_client.call_count)
    code_coverage.GetEquivalentPatchsets('chromium-review.googlesource.com',
                                         'chromium/src', '12345', '8')
    self.assertEqual(1, mock_http_client.call_count)
    code_coverage.GetEquivalentPatchsets('chromium-review.googlesource.com',
                                         'chromium/src', '12345', '8')
    self.assertEqual(1, mock_http_client.call_count)
