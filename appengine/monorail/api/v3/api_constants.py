# Copyright 2020 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Some constants used by Monorail's v3 API."""

# Max comments per page in the ListComment API.
MAX_COMMENTS_PER_PAGE = 100

# Max issues per page in the SearchIssues API.
MAX_ISSUES_PER_PAGE = 100

# Max issues to fetch in the BatchGetIssues API.
MAX_BATCH_ISSUES = 1000

# Max issues to modify at once in the ModifyIssues API.
MAX_MODIFY_ISSUES = 100

# Max impacted issues allowed in a ModifyIssues API.
MAX_MODIFY_IMPACTED_ISSUES = 50

# Max approval values to modify at once in the ModifyIssueApprovalValues API.
MAX_MODIFY_APPROVAL_VALUES = 100

# Max users to fetch in the BatchGetUsers API.
MAX_BATCH_USERS = 100

# Max component defs to fetch in the ListComponentDefs API
MAX_COMPONENTS_PER_PAGE = 100
