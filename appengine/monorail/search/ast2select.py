# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is govered by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""Convert a user's issue search AST into SQL clauses.

The main query is done on the Issues table.
 + Some simple conditions are implemented as WHERE conditions on the Issue
   table rows.  These are generated by the _Compare() function.
 + More complex conditions are implemented via a "LEFT JOIN ... ON ..." clause
   plus a check in the WHERE clause to select only rows where the join's ON
   condition was satisfied.  These are generated by appending a clause to
   the left_joins list plus calling _CompareAlreadyJoined().  Each such left
   join defines a unique alias to keep it separate from other conditions.

The functions that generate SQL snippets need to insert table names, column
names, alias names, and value placeholders into the generated string.  These
functions use the string format() method and the "{varname}" syntax to avoid
confusion with the "%s" syntax used for SQL value placeholders.
"""

import logging

from framework import sql
from proto import ast_pb2
from proto import tracker_pb2
from search import query2ast
from services import tracker_fulltext


NATIVE_SEARCHABLE_FIELDS = {
    'id': 'local_id',
    'is_spam': 'is_spam',
    'stars': 'star_count',
    'attachments': 'attachment_count',
    'opened': 'opened',
    'closed': 'closed',
    'modified': 'modified',
    'ownermodified': 'owner_modified',
    'statusmodified': 'status_modified',
    'componentmodified': 'component_modified',
    }


def BuildSQLQuery(query_ast, snapshot_mode=False):
  """Translate the user's query into an SQL query.

  Args:
    query_ast: user query abstract syntax tree parsed by query2ast.py.

  Returns:
    A pair of lists (left_joins, where) to use when building the SQL SELECT
    statement.  Each of them is a list of (str, [val, ...]) pairs.
  """
  left_joins = []
  where = []
  unsupported_conds = []
  # OR-queries are broken down into multiple simpler queries before they
  # are sent to the backends, so we should never see an "OR"..
  assert len(query_ast.conjunctions) == 1, 'OR-query should have been split'
  conj = query_ast.conjunctions[0]

  for cond_num, cond in enumerate(conj.conds):
    cond_left_joins, cond_where, unsupported = _ProcessCond(cond_num, cond,
        snapshot_mode)
    left_joins.extend(cond_left_joins)
    where.extend(cond_where)
    unsupported_conds.extend(unsupported)

  return left_joins, where, unsupported_conds


def _ProcessBlockedOnIDCond(cond, alias, _spare_alias, snapshot_mode):
  """Convert a blockedon_id=issue_id cond to SQL."""
  return _ProcessRelatedIDCond(cond, alias, 'blockedon', snapshot_mode)


def _ProcessBlockingIDCond(cond, alias, _spare_alias, snapshot_mode):
  """Convert a blocking_id:1,2 cond to SQL."""
  return _ProcessRelatedIDCond(cond, alias, 'blockedon', reverse_relation=True,
      snapshot_mode=snapshot_mode)


def _ProcessMergedIntoIDCond(cond, alias, _spare_alias, snapshot_mode):
  """Convert a mergedinto:1,2 cond to SQL."""
  return _ProcessRelatedIDCond(cond, alias, 'mergedinto',
      snapshot_mode=snapshot_mode)


def _ProcessRelatedIDCond(cond, alias, kind, reverse_relation=False,
                          snapshot_mode=False):
  """Convert either blocking_id, blockedon_id, or mergedinto_id cond to SQL.

  Normally, we query for issue_id values where the dst_issue_id matches the
  IDs specified in the cond.  However, when reverse_relation is True, we
  query for dst_issue_id values where issue_id matches.  This is done for
  blockedon_id.
  """
  if snapshot_mode:
    return [], [], [cond]

  matching_issue_col = 'issue_id' if reverse_relation else 'dst_issue_id'
  ret_issue_col = 'dst_issue_id' if reverse_relation else 'issue_id'

  kind_cond_str, kind_cond_args = _Compare(
      alias, ast_pb2.QueryOp.EQ, tracker_pb2.FieldTypes.STR_TYPE, 'kind',
      [kind])
  left_join_str = (
      'IssueRelation AS {alias} ON Issue.id = {alias}.{ret_issue_col} AND '
       '{kind_cond}').format(
           alias=alias, ret_issue_col=ret_issue_col, kind_cond=kind_cond_str)
  left_join_args = kind_cond_args

  field_type, field_values = _GetFieldTypeAndValues(cond)
  if field_values:
    related_cond_str, related_cond_args = _Compare(
        alias, ast_pb2.QueryOp.EQ, field_type, matching_issue_col, field_values)
    left_join_str += ' AND {related_cond}'.format(related_cond=related_cond_str)
    left_join_args += related_cond_args

  where = [_CompareAlreadyJoined(alias, cond.op, ret_issue_col)]

  return [(left_join_str, left_join_args)], where, []


def _GetFieldTypeAndValues(cond):
  """Returns the field type and values to use from the condition.

  This function should be used when we do not know what values are present on
  the condition. Eg: cond.int_values could be set if ast2ast.py preprocessing is
  first done. If that preprocessing is not done then str_values could be set
  instead.
  If both int values and str values exist on the condition then the int values
  are returned.
  """
  if cond.int_values:
    return tracker_pb2.FieldTypes.INT_TYPE, cond.int_values
  else:
    return tracker_pb2.FieldTypes.STR_TYPE, cond.str_values


def _ProcessOwnerCond(cond, alias, _spare_alias, snapshot_mode):
  """Convert an owner:substring cond to SQL."""
  if snapshot_mode:
    left_joins = [(
        'User AS {alias} ON '
        'IssueSnapshot.owner_id = {alias}.user_id'.format(alias=alias),
        [])]
  else:
    left_joins = [(
        'User AS {alias} ON (Issue.owner_id = {alias}.user_id '
        'OR Issue.derived_owner_id = {alias}.user_id)'.format(alias=alias),
        [])]
  where = [_Compare(alias, cond.op, tracker_pb2.FieldTypes.STR_TYPE, 'email',
                    cond.str_values)]

  return left_joins, where, []


def _ProcessOwnerIDCond(cond, _alias, _spare_alias, snapshot_mode):
  """Convert an owner_id=user_id cond to SQL."""
  if snapshot_mode:
    field_type, field_values = _GetFieldTypeAndValues(cond)
    explicit_str, explicit_args = _Compare(
        'IssueSnapshot', cond.op, field_type, 'owner_id', field_values)
    where = [(explicit_str, explicit_args)]
  else:
    field_type, field_values = _GetFieldTypeAndValues(cond)
    explicit_str, explicit_args = _Compare(
        'Issue', cond.op, field_type, 'owner_id', field_values)
    derived_str, derived_args = _Compare(
        'Issue', cond.op, field_type, 'derived_owner_id', field_values)
    if cond.op in (ast_pb2.QueryOp.NE, ast_pb2.QueryOp.NOT_TEXT_HAS):
      where = [(explicit_str, explicit_args), (derived_str, derived_args)]
    else:
      if cond.op == ast_pb2.QueryOp.IS_NOT_DEFINED:
        op = ' AND '
      else:
        op = ' OR '
      where = [
          ('(' + explicit_str + op + derived_str + ')',
           explicit_args + derived_args)]

  return [], where, []


def _ProcessOwnerLastVisitCond(cond, alias, _spare_alias, snapshot_mode):
  """Convert an ownerlastvisit<timestamp cond to SQL."""
  # TODO(jeffcarp): It is possible to support this on snapshots.
  if snapshot_mode:
    return [], [], [cond]

  left_joins = [(
      'User AS {alias} '
      'ON (Issue.owner_id = {alias}.user_id OR '
      'Issue.derived_owner_id = {alias}.user_id)'.format(alias=alias),
      [])]
  where = [_Compare(alias, cond.op, tracker_pb2.FieldTypes.INT_TYPE,
                    'last_visit_timestamp', cond.int_values)]
  return left_joins, where, []


def _ProcessIsOwnerBouncing(cond, alias, _spare_alias, snapshot_mode):
  """Convert an is:ownerbouncing cond to SQL."""
  if snapshot_mode:
    return [], [], [cond]

  left_joins = [(
      'User AS {alias} '
      'ON (Issue.owner_id = {alias}.user_id OR '
      'Issue.derived_owner_id = {alias}.user_id)'.format(alias=alias),
      [])]
  if cond.op == ast_pb2.QueryOp.EQ:
    op = ast_pb2.QueryOp.IS_DEFINED
  else:
    op = ast_pb2.QueryOp.IS_NOT_DEFINED

  where = [_Compare(alias, op, tracker_pb2.FieldTypes.INT_TYPE,
                    'email_bounce_timestamp', [])]
  return left_joins, where, []


def _ProcessReporterCond(cond, alias, _spare_alias, snapshot_mode):
  """Convert a reporter:substring cond to SQL."""
  if snapshot_mode:
    left_joins = [(
        'User AS {alias} ON IssueSnapshot.reporter_id = {alias}.user_id'.format(
            alias=alias), [])]
  else:
    left_joins = [(
        'User AS {alias} ON Issue.reporter_id = {alias}.user_id'.format(
            alias=alias), [])]
  where = [_Compare(alias, cond.op, tracker_pb2.FieldTypes.STR_TYPE, 'email',
                    cond.str_values)]

  return left_joins, where, []


def _ProcessReporterIDCond(cond, _alias, _spare_alias, snapshot_mode):
  """Convert a reporter_ID=user_id cond to SQL."""
  field_type, field_values = _GetFieldTypeAndValues(cond)

  if snapshot_mode:
    where = [_Compare(
        'IssueSnapshot', cond.op, field_type, 'reporter_id', field_values)]
  else:
    where = [_Compare(
        'Issue', cond.op, field_type, 'reporter_id', field_values)]
  return [], where, []


def _ProcessCcCond(cond, alias, user_alias, snapshot_mode):
  """Convert a cc:substring cond to SQL."""
  email_cond_str, email_cond_args = _Compare(
      user_alias, ast_pb2.QueryOp.TEXT_HAS, tracker_pb2.FieldTypes.STR_TYPE,
      'email', cond.str_values)

  if snapshot_mode:
    left_joins = [(
        '(IssueSnapshot2Cc AS {alias} JOIN User AS {user_alias} '
        'ON {alias}.cc_id = {user_alias}.user_id AND {email_cond}) '
        'ON IssueSnapshot.id = {alias}.issuesnapshot_id'.format(
            alias=alias, user_alias=user_alias, email_cond=email_cond_str),
        email_cond_args)]
  else:
    # Note: email_cond_str will have parens, if needed.
    left_joins = [(
        '(Issue2Cc AS {alias} JOIN User AS {user_alias} '
        'ON {alias}.cc_id = {user_alias}.user_id AND {email_cond}) '
        'ON Issue.id = {alias}.issue_id AND '
        'Issue.shard = {alias}.issue_shard'.format(
            alias=alias, user_alias=user_alias, email_cond=email_cond_str),
        email_cond_args)]
  where = [_CompareAlreadyJoined(user_alias, cond.op, 'email')]

  return left_joins, where, []


def _ProcessCcIDCond(cond, alias, _spare_alias, snapshot_mode):
  """Convert a cc_id=user_id cond to SQL."""
  if snapshot_mode:
    join_str = (
        'IssueSnapshot2Cc AS {alias} '
        'ON IssueSnapshot.id = {alias}.issuesnapshot_id'.format(alias=alias))
  else:
    join_str = (
        'Issue2Cc AS {alias} ON Issue.id = {alias}.issue_id AND '
        'Issue.shard = {alias}.issue_shard'.format(
            alias=alias))
  if cond.op in (ast_pb2.QueryOp.IS_DEFINED, ast_pb2.QueryOp.IS_NOT_DEFINED):
    left_joins = [(join_str, [])]
  else:
    field_type, field_values = _GetFieldTypeAndValues(cond)
    cond_str, cond_args = _Compare(
        alias, ast_pb2.QueryOp.EQ, field_type, 'cc_id', field_values)
    left_joins = [(join_str + ' AND ' + cond_str, cond_args)]

  where = [_CompareAlreadyJoined(alias, cond.op, 'cc_id')]
  return left_joins, where, []


def _ProcessStarredByCond(cond, alias, user_alias, snapshot_mode):
  """Convert a starredby:substring cond to SQL."""
  if snapshot_mode:
    return [], [], [cond]

  email_cond_str, email_cond_args = _Compare(
      user_alias, cond.op, tracker_pb2.FieldTypes.STR_TYPE, 'email',
      cond.str_values)
  # Note: email_cond_str will have parens, if needed.
  left_joins = [(
      '(IssueStar AS {alias} JOIN User AS {user_alias} '
      'ON {alias}.user_id = {user_alias}.user_id AND {email_cond}) '
      'ON Issue.id = {alias}.issue_id'.format(
          alias=alias, user_alias=user_alias, email_cond=email_cond_str),
      email_cond_args)]
  where = [_CompareAlreadyJoined(user_alias, cond.op, 'email')]

  return left_joins, where, []


def _ProcessStarredByIDCond(cond, alias, _spare_alias, snapshot_mode):
  """Convert a starredby_id=user_id cond to SQL."""
  if snapshot_mode:
    return [], [], [cond]

  join_str = 'IssueStar AS {alias} ON Issue.id = {alias}.issue_id'.format(
      alias=alias)
  if cond.op in (ast_pb2.QueryOp.IS_DEFINED, ast_pb2.QueryOp.IS_NOT_DEFINED):
    left_joins = [(join_str, [])]
  else:
    field_type, field_values = _GetFieldTypeAndValues(cond)
    cond_str, cond_args = _Compare(
        alias, ast_pb2.QueryOp.EQ, field_type, 'user_id', field_values)
    left_joins = [(join_str + ' AND ' + cond_str, cond_args)]

  where = [_CompareAlreadyJoined(alias, cond.op, 'user_id')]
  return left_joins, where, []


def _ProcessCommentByCond(cond, alias, user_alias, snapshot_mode):
  """Convert a commentby:substring cond to SQL."""
  if snapshot_mode:
    return [], [], [cond]

  email_cond_str, email_cond_args = _Compare(
      user_alias, ast_pb2.QueryOp.TEXT_HAS, tracker_pb2.FieldTypes.STR_TYPE,
      'email', cond.str_values)
  # Note: email_cond_str will have parens, if needed.
  left_joins = [(
      '(Comment AS {alias} JOIN User AS {user_alias} '
      'ON {alias}.commenter_id = {user_alias}.user_id AND {email_cond}) '
      'ON Issue.id = {alias}.issue_id AND '
      '{alias}.deleted_by IS NULL'.format(
          alias=alias, user_alias=user_alias, email_cond=email_cond_str),
      email_cond_args)]
  where = [_CompareAlreadyJoined(user_alias, cond.op, 'email')]

  return left_joins, where, []


def _ProcessCommentByIDCond(cond, alias, _spare_alias, snapshot_mode):
  """Convert a commentby_id=user_id cond to SQL."""
  if snapshot_mode:
    return [], [], [cond]

  field_type, field_values = _GetFieldTypeAndValues(cond)
  commenter_cond_str, commenter_cond_args = _Compare(
      alias, ast_pb2.QueryOp.EQ, field_type, 'commenter_id', field_values)
  left_joins = [(
      'Comment AS {alias} ON Issue.id = {alias}.issue_id AND '
      '{commenter_cond} AND '
      '{alias}.deleted_by IS NULL'.format(
          alias=alias, commenter_cond=commenter_cond_str),
      commenter_cond_args)]
  where = [_CompareAlreadyJoined(alias, cond.op, 'commenter_id')]

  return left_joins, where, []


def _ProcessStatusIDCond(cond, _alias, _spare_alias, snapshot_mode):
  """Convert a status_id=ID cond to SQL."""
  field_type, field_values = _GetFieldTypeAndValues(cond)
  if snapshot_mode:
    explicit_str, explicit_args = _Compare(
        'IssueSnapshot', cond.op, field_type, 'status_id', field_values)
    where = [(explicit_str, explicit_args)]
  else:
    explicit_str, explicit_args = _Compare(
        'Issue', cond.op, field_type, 'status_id', field_values)
    derived_str, derived_args = _Compare(
        'Issue', cond.op, field_type, 'derived_status_id', field_values)
    if cond.op in (ast_pb2.QueryOp.IS_NOT_DEFINED, ast_pb2.QueryOp.NE):
      where = [(explicit_str, explicit_args), (derived_str, derived_args)]
    else:
      where = [
          ('(' + explicit_str + ' OR ' + derived_str + ')',
           explicit_args + derived_args)]

  return [], where, []


def _ProcessLabelIDCond(cond, alias, _spare_alias, snapshot_mode):
  """Convert a label_id=ID cond to SQL."""
  if snapshot_mode:
    join_str = (
        'IssueSnapshot2Label AS {alias} '
        'ON IssueSnapshot.id = {alias}.issuesnapshot_id'.format(alias=alias))
  else:
    join_str = (
        'Issue2Label AS {alias} ON Issue.id = {alias}.issue_id AND '
        'Issue.shard = {alias}.issue_shard'.format(alias=alias))

  field_type, field_values = _GetFieldTypeAndValues(cond)
  if not field_values and cond.op == ast_pb2.QueryOp.NE:
    return [], [], []
  cond_str, cond_args = _Compare(
      alias, ast_pb2.QueryOp.EQ, field_type, 'label_id', field_values)
  left_joins = [(join_str + ' AND ' + cond_str, cond_args)]
  where = [_CompareAlreadyJoined(alias, cond.op, 'label_id')]
  return left_joins, where, []


def _ProcessComponentIDCond(cond, alias, _spare_alias, snapshot_mode):
  """Convert a component_id=ID cond to SQL."""
  # This is a built-in field, so it shadows any other fields w/ the same name.
  if snapshot_mode:
    join_str = (
        'IssueSnapshot2Component AS {alias} '
        'ON IssueSnapshot.id = {alias}.issuesnapshot_id'.format(alias=alias))
  else:
    join_str = (
        'Issue2Component AS {alias} ON Issue.id = {alias}.issue_id AND '
        'Issue.shard = {alias}.issue_shard'.format(alias=alias))
  if cond.op in (ast_pb2.QueryOp.IS_DEFINED, ast_pb2.QueryOp.IS_NOT_DEFINED):
    left_joins = [(join_str, [])]
  else:
    field_type, field_values = _GetFieldTypeAndValues(cond)
    cond_str, cond_args = _Compare(
        alias, ast_pb2.QueryOp.EQ, field_type, 'component_id', field_values)
    left_joins = [(join_str + ' AND ' + cond_str, cond_args)]

  where = [_CompareAlreadyJoined(alias, cond.op, 'component_id')]
  return left_joins, where, []


# TODO(jojang): monorail:3819, check for cond.phase_name and process
# appropriately so users can search 'Canary.UXReview-status:Approved'
def _ProcessApprovalFieldCond(cond, alias, user_alias, snapshot_mode):
  """Convert a custom approval field cond to SQL."""
  if snapshot_mode:
    return [], [], [cond]

  approval_fd = cond.field_defs[0]
  left_joins = []

  join_str_tmpl = (
    '{tbl_name} AS {alias} ON Issue.id = {alias}.issue_id AND '
    '{alias}.approval_id = %s')

  join_args = [approval_fd.field_id]

  val_type, values = _GetFieldTypeAndValues(cond)
  if val_type is tracker_pb2.FieldTypes.STR_TYPE:
    values = [val.lower() for val in values]
  # TODO(jojwang):monorail:3809, check if there is a cond.key_suffx.
  # status, approver should always have a value, so 'has:UXReview-approver'
  # should return the same issues as 'has:UXReview'.
  # There will not always be values approval.setter_id and approval.set_on
  # and the current code would not process 'has:UXReview-by' correctly.
  if cond.op in (
      ast_pb2.QueryOp.IS_DEFINED, ast_pb2.QueryOp.IS_NOT_DEFINED):
    join_str = join_str_tmpl.format(
        tbl_name='Issue2ApprovalValue', alias=alias)
    left_joins = [(join_str, join_args)]
  else:
    op = cond.op
    if op == ast_pb2.QueryOp.NE:
      op = ast_pb2.QueryOp.EQ  # Negation is done in WHERE clause.

    if (not cond.key_suffix) or cond.key_suffix == query2ast.STATUS_SUFFIX:
      tbl_str = 'Issue2ApprovalValue'
      cond_str, cond_args = _Compare(
          alias, op, val_type, 'status', values)
    elif cond.key_suffix == query2ast.SET_ON_SUFFIX:
      tbl_str = 'Issue2ApprovalValue'
      cond_str, cond_args = _Compare(
          alias, op, val_type, 'set_on', values)
    elif cond.key_suffix in [query2ast.APPROVER_SUFFIX, query2ast.SET_BY_SUFFIX]:
      if cond.key_suffix == query2ast.SET_BY_SUFFIX:
        tbl_str = 'Issue2ApprovalValue'
        col_name = 'setter_id'
      else:
        tbl_str = 'IssueApproval2Approver'
        col_name = 'approver_id'

      if val_type == tracker_pb2.FieldTypes.INT_TYPE:
        cond_str, cond_args = _Compare(
            alias, op, val_type, col_name, values)
      else:
        email_cond_str, email_cond_args = _Compare(
            user_alias, op, val_type, 'email', values)
        left_joins.append((
          'User AS {user_alias} ON {email_cond}'.format(
              user_alias=user_alias, email_cond=email_cond_str),
          email_cond_args))

        cond_str = '{alias}.setter_id = {user_alias}.user_id'.format(
            alias=alias, user_alias=user_alias)
        cond_args = []
    if cond_str or cond_args:
      join_str = join_str_tmpl.format(tbl_name=tbl_str, alias=alias)
      join_str += ' AND ' + cond_str
      join_args.extend(cond_args)
    left_joins.append((join_str, join_args))

  where = [_CompareAlreadyJoined(alias, cond.op, 'approval_id')]
  return left_joins, where, []


def _ProcessCustomFieldCond(cond, alias, user_alias, snapshot_mode):
  """Convert a custom field cond to SQL."""
  if snapshot_mode:
    return [], [], [cond]

  # TODO(jrobbins): handle ambiguous field names that map to multiple
  # field definitions, especially for cross-project search.
  field_def = cond.field_defs[0]
  field_type = field_def.field_type
  left_joins = []

  join_str = (
      'Issue2FieldValue AS {alias} ON Issue.id = {alias}.issue_id AND '
      'Issue.shard = {alias}.issue_shard AND '
      '{alias}.field_id = %s'.format(alias=alias))
  join_args = [field_def.field_id]

  if cond.op not in (
      ast_pb2.QueryOp.IS_DEFINED, ast_pb2.QueryOp.IS_NOT_DEFINED):
    op = cond.op
    if op == ast_pb2.QueryOp.NE:
      op = ast_pb2.QueryOp.EQ  # Negation is done in WHERE clause.
    if field_type == tracker_pb2.FieldTypes.INT_TYPE:
      cond_str, cond_args = _Compare(
          alias, op, field_type, 'int_value', cond.int_values)
    elif field_type == tracker_pb2.FieldTypes.STR_TYPE:
      cond_str, cond_args = _Compare(
          alias, op, field_type, 'str_value', cond.str_values)
    elif field_type == tracker_pb2.FieldTypes.USER_TYPE:
      if cond.int_values:
        cond_str, cond_args = _Compare(
            alias, op, field_type, 'user_id', cond.int_values)
      else:
        email_cond_str, email_cond_args = _Compare(
            user_alias, op, field_type, 'email', cond.str_values)
        left_joins.append((
            'User AS {user_alias} ON {email_cond}'.format(
                user_alias=user_alias, email_cond=email_cond_str),
            email_cond_args))
        cond_str = '{alias}.user_id = {user_alias}.user_id'.format(
            alias=alias, user_alias=user_alias)
        cond_args = []
    elif field_type == tracker_pb2.FieldTypes.URL_TYPE:
      cond_str, cond_args = _Compare(
          alias, op, field_type, 'url_value', cond.str_values)
    if field_type == tracker_pb2.FieldTypes.DATE_TYPE:
      cond_str, cond_args = _Compare(
          alias, op, field_type, 'date_value', cond.int_values)
    if cond_str or cond_args:
      join_str += ' AND ' + cond_str
      join_args.extend(cond_args)

  left_joins.append((join_str, join_args))
  where = [_CompareAlreadyJoined(alias, cond.op, 'field_id')]
  return left_joins, where, []


def _ProcessAttachmentCond(cond, alias, _spare_alias, snapshot_mode):
  """Convert has:attachment and -has:attachment cond to SQL."""
  if snapshot_mode:
    return [], [], [cond]

  if cond.op in (ast_pb2.QueryOp.IS_DEFINED, ast_pb2.QueryOp.IS_NOT_DEFINED):
    left_joins = []
    where = [_Compare('Issue', cond.op, tracker_pb2.FieldTypes.INT_TYPE,
                      'attachment_count', cond.int_values)]
  else:
    field_def = cond.field_defs[0]
    field_type = field_def.field_type
    left_joins = [
      ('Attachment AS {alias} ON Issue.id = {alias}.issue_id AND '
       '{alias}.deleted = %s'.format(alias=alias),
       [False])]
    where = [_Compare(alias, cond.op, field_type, 'filename', cond.str_values)]

  return left_joins, where, []


def _ProcessHotlistIDCond(cond, alias, _spare_alias, snapshot_mode):
  """Convert hotlist_id=IDS cond to SQL."""
  if snapshot_mode:
    join_str = (
      'IssueSnapshot2Hotlist AS {alias} '
      'ON IssueSnapshot.id = {alias}.issuesnapshot_id'.format(alias=alias))
  else:
    join_str = (
      'Hotlist2Issue AS {alias} ON Issue.id = {alias}.issue_id'.format(
          alias=alias))

  field_type, field_values = _GetFieldTypeAndValues(cond)
  if not field_values and cond.op == ast_pb2.QueryOp.NE:
    return [], [], []
  cond_str, cond_args = _Compare(
      alias, ast_pb2.QueryOp.EQ, field_type, 'hotlist_id', field_values)
  left_joins = [(join_str + ' AND ' + cond_str, cond_args)]
  where = [_CompareAlreadyJoined(alias, cond.op, 'hotlist_id')]

  return left_joins, where, []


def _ProcessHotlistCond(cond, alias, _spare_alias, snapshot_mode):
  """Convert hotlist=user:hotlist-name to SQL"""
  # hotlist conditions that reach this function definitely have invalid
  # user_name/id/email. This validity was determined in
  # ast2ast._PreprocessHotlistCond. Any possible user identification is ignored.
  hotlist_substrings = []
  for val in cond.str_values:
    substring = val.split(':')[-1]
    if substring:
      hotlist_substrings.append(substring)
  hotlist_cond_str, hotlist_cond_args = _Compare(
      alias, ast_pb2.QueryOp.TEXT_HAS, tracker_pb2.FieldTypes.STR_TYPE,
      'name', hotlist_substrings)
  if snapshot_mode:
    left_joins = [(
        '(IssueSnapshot2Hotlist JOIN Hotlist AS {alias} '
        'ON IssueSnapshot2Hotlist.hotlist_id = {alias}.id AND {hotlist_cond}) '
        'ON IssueSnapshot.id = IssueSnapshot2Hotlist.issuesnapshot_id'.format(
            alias=alias, hotlist_cond=hotlist_cond_str), hotlist_cond_args)]
  else:
    left_joins = [(
        '(Hotlist2Issue JOIN Hotlist AS {alias} '
        'ON Hotlist2Issue.hotlist_id = {alias}.id AND {hotlist_cond}) '
        'ON Issue.id = Hotlist2Issue.issue_id'.format(
            alias=alias, hotlist_cond=hotlist_cond_str), hotlist_cond_args)]
  where = [_CompareAlreadyJoined(alias, cond.op, 'name')]

  return left_joins, where, []


def _ProcessPhaseCond(cond, alias, phase_alias, _snapshot_mode):
  """Convert gate:<phase_name> to SQL."""

  cond_str, cond_args = _Compare(
      phase_alias, cond.op, tracker_pb2.FieldTypes.STR_TYPE, 'name', cond.str_values)
  left_joins = [(
      '(Issue2ApprovalValue AS {alias} JOIN IssuePhaseDef AS {phase_alias} '
      'ON {alias}.phase_id = {phase_alias}.id AND {name_cond}) '
      'ON Issue.id = {alias}.issue_id'.format(
          alias=alias, phase_alias=phase_alias, name_cond=cond_str),
      cond_args)]
  where = [_CompareAlreadyJoined(phase_alias, cond.op, 'name')]

  return left_joins, where, []


_PROCESSORS = {
    'owner': _ProcessOwnerCond,
    'owner_id': _ProcessOwnerIDCond,
    'ownerlastvisit': _ProcessOwnerLastVisitCond,
    'ownerbouncing': _ProcessIsOwnerBouncing,
    'reporter': _ProcessReporterCond,
    'reporter_id': _ProcessReporterIDCond,
    'cc': _ProcessCcCond,
    'cc_id': _ProcessCcIDCond,
    'starredby': _ProcessStarredByCond,
    'starredby_id': _ProcessStarredByIDCond,
    'commentby': _ProcessCommentByCond,
    'commentby_id': _ProcessCommentByIDCond,
    'status_id': _ProcessStatusIDCond,
    'label_id': _ProcessLabelIDCond,
    'component_id': _ProcessComponentIDCond,
    'blockedon_id': _ProcessBlockedOnIDCond,
    'blocking_id': _ProcessBlockingIDCond,
    'mergedinto_id': _ProcessMergedIntoIDCond,
    'attachment': _ProcessAttachmentCond,
    'hotlist_id': _ProcessHotlistIDCond,
    'hotlist': _ProcessHotlistCond,
    'gate': _ProcessPhaseCond,
    }


def _ProcessCond(cond_num, cond, snapshot_mode):
  """Translate one term of the user's search into an SQL query.

  Args:
    cond_num: integer cond number used to make distinct local variable names.
    cond: user query cond parsed by query2ast.py.

  Returns:
    A pair of lists (left_joins, where) to use when building the SQL SELECT
    statement.  Each of them is a list of (str, [val, ...]) pairs.
  """
  alias = 'Cond%d' % cond_num
  spare_alias = 'Spare%d' % cond_num
  # Note: a condition like [x=y] has field_name "x", there may be multiple
  # field definitions that match "x", but they will all have field_name "x".
  field_def = cond.field_defs[0]
  assert all(field_def.field_name == fd.field_name for fd in cond.field_defs)

  if field_def.field_name in NATIVE_SEARCHABLE_FIELDS:
    # TODO(jeffcarp): Support local_id search here.
    if snapshot_mode:
      return [], [], [cond]
    else:
      col = NATIVE_SEARCHABLE_FIELDS[field_def.field_name]
      where = [_Compare(
          'Issue', cond.op, field_def.field_type, col,
          cond.str_values or cond.int_values)]
      return [], where, []

  elif field_def.field_name in _PROCESSORS:
    proc = _PROCESSORS[field_def.field_name]
    return proc(cond, alias, spare_alias, snapshot_mode)

  elif field_def.field_id:  # it is a search on a custom field
    if field_def.field_type == tracker_pb2.FieldTypes.APPROVAL_TYPE:
      return _ProcessCustomApprovalCond(cond, alias, user_alias, snapshot_mode)
    return _ProcessCustomFieldCond(cond, alias, user_alias, snapshot_mode)

  elif (field_def.field_name in tracker_fulltext.ISSUE_FULLTEXT_FIELDS or
        field_def.field_name == 'any_field'):
    if snapshot_mode:
      return [], [], [cond]
    # This case handled by full-text search.

  else:
    logging.error('untranslated search cond %r', cond)

  return [], [], []


def _Compare(alias, op, val_type, col, vals):
  """Return an SQL comparison for the given values. For use in WHERE or ON.

  Args:
    alias: String name of the table or alias defined in a JOIN clause.
    op: One of the operators defined in ast_pb2.py.
    val_type: One of the value types defined in ast_pb2.py.
    col: string column name to compare to vals.
    vals: list of values that the user is searching for.

  Returns:
    (cond_str, cond_args) where cond_str is a SQL condition that may contain
    some %s placeholders, and cond_args is the list of values that fill those
    placeholders.  If the condition string contains any AND or OR operators,
    the whole expression is put inside parens.

  Raises:
    NoPossibleResults: The user's query is impossible to ever satisfy, e.g.,
        it requires matching an empty set of labels.
  """
  vals_ph = sql.PlaceHolders(vals)
  if col in ['label', 'status', 'email', 'name']:
    alias_col = 'LOWER(%s.%s)' % (alias, col)
  else:
    alias_col = '%s.%s' % (alias, col)

  def Fmt(cond_str):
    return cond_str.format(alias_col=alias_col, vals_ph=vals_ph)

  no_value = (0 if val_type in [tracker_pb2.FieldTypes.DATE_TYPE,
                                tracker_pb2.FieldTypes.INT_TYPE] else '')
  if op == ast_pb2.QueryOp.IS_DEFINED:
    return Fmt('({alias_col} IS NOT NULL AND {alias_col} != %s)'), [no_value]
  if op == ast_pb2.QueryOp.IS_NOT_DEFINED:
    return Fmt('({alias_col} IS NULL OR {alias_col} = %s)'), [no_value]

  if val_type in [tracker_pb2.FieldTypes.DATE_TYPE,
                  tracker_pb2.FieldTypes.INT_TYPE]:
    if op == ast_pb2.QueryOp.TEXT_HAS:
      op = ast_pb2.QueryOp.EQ
    if op == ast_pb2.QueryOp.NOT_TEXT_HAS:
      op = ast_pb2.QueryOp.NE

  if op == ast_pb2.QueryOp.EQ:
    if not vals:
      raise NoPossibleResults('Column %s has no possible value' % alias_col)
    elif len(vals) == 1:
      cond_str = Fmt('{alias_col} = %s')
    else:
      cond_str = Fmt('{alias_col} IN ({vals_ph})')
    return cond_str, vals

  if op == ast_pb2.QueryOp.NE:
    if not vals:
      return 'TRUE', []  # a no-op that matches every row.
    elif len(vals) == 1:
      comp = Fmt('{alias_col} != %s')
    else:
      comp = Fmt('{alias_col} NOT IN ({vals_ph})')
    return '(%s IS NULL OR %s)' % (alias_col, comp), vals

  wild_vals = ['%%%s%%' % val for val in vals]
  if op == ast_pb2.QueryOp.TEXT_HAS:
    cond_str = ' OR '.join(Fmt('{alias_col} LIKE %s') for v in vals)
    return ('(%s)' % cond_str), wild_vals
  if op == ast_pb2.QueryOp.NOT_TEXT_HAS:
    cond_str = (Fmt('{alias_col} IS NULL OR ') +
                ' AND '.join(Fmt('{alias_col} NOT LIKE %s') for v in vals))
    return ('(%s)' % cond_str), wild_vals


  # Note: These operators do not support quick-OR
  val = vals[0]

  if op == ast_pb2.QueryOp.GT:
    return Fmt('{alias_col} > %s'), [val]
  if op == ast_pb2.QueryOp.LT:
    return Fmt('{alias_col} < %s'), [val]
  if op == ast_pb2.QueryOp.GE:
    return Fmt('{alias_col} >= %s'), [val]
  if op == ast_pb2.QueryOp.LE:
    return Fmt('{alias_col} <= %s'), [val]

  logging.error('unknown op: %r', op)


def _CompareAlreadyJoined(alias, op, col):
  """Return a WHERE clause comparison that checks that a join succeeded."""
  def Fmt(cond_str):
    return cond_str.format(alias_col='%s.%s' % (alias, col))

  if op in (ast_pb2.QueryOp.NE, ast_pb2.QueryOp.NOT_TEXT_HAS,
            ast_pb2.QueryOp.IS_NOT_DEFINED):
    return Fmt('{alias_col} IS NULL'), []
  else:
    return Fmt('{alias_col} IS NOT NULL'), []


class Error(Exception):
  """Base class for errors from this module."""


class NoPossibleResults(Error):
  """The query could never match any rows from the database, so don't try.."""
