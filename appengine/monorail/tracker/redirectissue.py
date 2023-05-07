# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.


from __future__ import print_function
from __future__ import division
from __future__ import absolute_import


from google.appengine.ext import ndb


class RedirectIssue(ndb.Model):
 """Represents a issue redirect information."""
 ProjectName = ndb.StringProperty()
 MonorailID = ndb.IntegerProperty()
 TrackerID = ndb.IntegerProperty()


 @classmethod
 def GetRedirectIssue(cls, project, issue_local_id):
   key = project + '_' + str(issue_local_id)
   redirect_issue_entity = ndb.Key('RedirectIssue', key).get()
   if not redirect_issue_entity:
     return None
   return_entity = redirect_issue_entity.to_dict()
   return str(return_entity['TrackerID'])