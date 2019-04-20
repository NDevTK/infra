// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {combineReducers} from 'redux';
import {createSelector} from 'reselect';
import {autolink} from '../../autolink.js';
import {fieldTypes} from '../shared/field-types.js';
import {removePrefix} from '../shared/helpers.js';
import {issueRefToString} from '../shared/converters.js';
import {createReducer, createRequestReducer} from './redux-helpers.js';
import * as project from './project.js';

// Actions
const SET_ISSUE_REF = 'SET_ISSUE_REF';

export const FETCH_START = 'FETCH_START';
export const FETCH_SUCCESS = 'FETCH_SUCCESS';
export const FETCH_FAILURE = 'FETCH_FAILURE';

const FETCH_HOTLISTS_START = 'FETCH_HOTLISTS_START';
const FETCH_HOTLISTS_SUCCESS = 'FETCH_HOTLISTS_SUCCESS';
const FETCH_HOTLISTS_FAILURE = 'FETCH_HOTLISTS_FAILURE';

const FETCH_PERMISSIONS_START = 'FETCH_PERMISSIONS_START';
const FETCH_PERMISSIONS_SUCCESS = 'FETCH_PERMISSIONS_SUCCESS';
const FETCH_PERMISSIONS_FAILURE = 'FETCH_PERMISSIONS_FAILURE';

const STAR_START = 'STAR_START';
const STAR_SUCCESS = 'STAR_SUCCESS';
const STAR_FAILURE = 'STAR_FAILURE';

const PRESUBMIT_START = 'PRESUBMIT_START';
const PRESUBMIT_SUCCESS = 'PRESUBMIT_SUCCESS';
const PRESUBMIT_FAILURE = 'PRESUBMIT_FAILURE';

const FETCH_IS_STARRED_START = 'FETCH_IS_STARRED_START';
const FETCH_IS_STARRED_SUCCESS = 'FETCH_IS_STARRED_SUCCESS';
const FETCH_IS_STARRED_FAILURE = 'FETCH_IS_STARRED_FAILURE';

const FETCH_COMMENTS_START = 'FETCH_COMMENTS_START';
export const FETCH_COMMENTS_SUCCESS = 'FETCH_COMMENTS_SUCCESS';
const FETCH_COMMENTS_FAILURE = 'FETCH_COMMENTS_FAILURE';

const FETCH_COMMENT_REFERENCES_START = 'FETCH_COMMENT_REFERENCES_START';
const FETCH_COMMENT_REFERENCES_SUCCESS = 'FETCH_COMMENT_REFERENCES_SUCCESS';
const FETCH_COMMENT_REFERENCES_FAILURE = 'FETCH_COMMENT_REFERENCES_FAILURE';

const FETCH_REFERENCED_USERS_START = 'FETCH_REFERENCED_USERS_START';
const FETCH_REFERENCED_USERS_SUCCESS = 'FETCH_REFERENCED_USERS_SUCCESS';
const FETCH_REFERENCED_USERS_FAILURE = 'FETCH_REFERENCED_USERS_FAILURE';

const FETCH_RELATED_ISSUES_START = 'FETCH_RELATED_ISSUES_START';
const FETCH_RELATED_ISSUES_SUCCESS = 'FETCH_RELATED_ISSUES_SUCCESS';
const FETCH_RELATED_ISSUES_FAILURE = 'FETCH_RELATED_ISSUES_FAILURE';

const CONVERT_START = 'CONVERT_START';
const CONVERT_SUCCESS = 'CONVERT_SUCCESS';
const CONVERT_FAILURE = 'CONVERT_FAILURE';

const UPDATE_START = 'UPDATE_START';
const UPDATE_SUCCESS = 'UPDATE_SUCCESS';
const UPDATE_FAILURE = 'UPDATE_FAILURE';

const UPDATE_APPROVAL_START = 'UPDATE_APPROVAL_START';
const UPDATE_APPROVAL_SUCCESS = 'UPDATE_APPROVAL_SUCCESS';
const UPDATE_APPROVAL_FAILURE = 'UPDATE_APPROVAL_FAILURE';

/* State Shape
{
  issueRef: {
    localId: Number,
    projectName: String,
  },

  currentIssue: Object,

  hotlists: Array,
  comments: Array,
  commentReferences: Map,
  relatedIssues: Map,
  referencedUsers: Array,
  isStarred: Boolean,
  permissions: Array,
  presubmitResponse: Object,

  requests: {
    fetch: Object,
    fetchHotlists: Object,
    fetchPermissions: Object,
    star: Object,
    presubmit: Object,
    fetchComments: Object,
    fetchCommentReferences: Object,
    fetchRelatedIssues: Object,
    fetchIsStarred: Object,
    convert: Object,
    update: Object,
    updateApproval: Object,
  },
}
*/

// Helpers for the reducers.
const updateApprovalValues = (issue, approval) => {
  if (!issue.approvalValues) return issue;
  const newApprovals = issue.approvalValues.map((item, i) => {
    if (item.fieldRef.fieldName === approval.fieldRef.fieldName) {
      // PhaseRef isn't populated on the response so we want to make sure
      // it doesn't overwrite the original phaseRef with {}.
      return {...approval, phaseRef: item.phaseRef};
    }
    return item;
  });
  return {...issue, approvalValues: newApprovals};
};

// Reducers
const localIdReducer = createReducer(0, {
  [SET_ISSUE_REF]: (state, action) => action.localId || state,
});

const projectNameReducer = createReducer('', {
  [SET_ISSUE_REF]: (state, action) => action.projectName || state,
});

const currentIssueReducer = createReducer({}, {
  [FETCH_SUCCESS]: (_state, action) => action.issue,
  [STAR_SUCCESS]: (state, action) => {
    return {...state, starCount: action.starCount};
  },
  [CONVERT_SUCCESS]: (_state, action) => action.issue,
  [UPDATE_SUCCESS]: (_state, action) => action.issue,
  [UPDATE_APPROVAL_SUCCESS]: (state, action) => {
    return updateApprovalValues(state, action.approval);
  },
});

const hotlistsReducer = createReducer([], {
  [FETCH_HOTLISTS_SUCCESS]: (_, action) => action.hotlists,
});

const commentsReducer = createReducer([], {
  [FETCH_COMMENTS_SUCCESS]: (_state, action) => action.comments,
});

const commentReferencesReducer = createReducer(new Map(), {
  [FETCH_COMMENTS_START]: (_state, _action) => new Map(),
  [FETCH_COMMENT_REFERENCES_SUCCESS]: (_state, action) => {
    return action.commentReferences;
  },
});

const relatedIssuesReducer = createReducer(new Map(), {
  [FETCH_RELATED_ISSUES_SUCCESS]: (_state, action) => action.relatedIssues,
});

const referencedUsersReducer = createReducer(new Map(), {
  [FETCH_REFERENCED_USERS_SUCCESS]: (_state, action) => action.referencedUsers,
});

const isStarredReducer = createReducer(false, {
  [STAR_SUCCESS]: (state, _action) => !state,
  [FETCH_IS_STARRED_SUCCESS]: (_state, action) => !!action.isStarred,
});

const presubmitResponseReducer = createReducer({}, {
  [PRESUBMIT_SUCCESS]: (_state, action) => action.presubmitResponse,
});

const permissionsReducer = createReducer([], {
  [FETCH_PERMISSIONS_SUCCESS]: (_state, action) => action.permissions,
});

const requestsReducer = combineReducers({
  fetch: createRequestReducer(
    FETCH_START, FETCH_SUCCESS, FETCH_FAILURE),
  fetchHotlists: createRequestReducer(
    FETCH_HOTLISTS_START, FETCH_HOTLISTS_SUCCESS, FETCH_HOTLISTS_FAILURE),
  fetchPermissions: createRequestReducer(
    FETCH_PERMISSIONS_START,
    FETCH_PERMISSIONS_SUCCESS,
    FETCH_PERMISSIONS_FAILURE),
  star: createRequestReducer(
    STAR_START, STAR_SUCCESS, STAR_FAILURE),
  presubmit: createRequestReducer(
    PRESUBMIT_START, PRESUBMIT_SUCCESS, PRESUBMIT_FAILURE),
  fetchComments: createRequestReducer(
    FETCH_COMMENTS_START, FETCH_COMMENTS_SUCCESS, FETCH_COMMENTS_FAILURE),
  fetchCommentReferences: createRequestReducer(
    FETCH_COMMENT_REFERENCES_START,
    FETCH_COMMENT_REFERENCES_SUCCESS,
    FETCH_COMMENT_REFERENCES_FAILURE),
  fetchRelatedIssues: createRequestReducer(
    FETCH_RELATED_ISSUES_START,
    FETCH_RELATED_ISSUES_SUCCESS,
    FETCH_RELATED_ISSUES_FAILURE),
  fetchReferencedUsers: createRequestReducer(
    FETCH_REFERENCED_USERS_START,
    FETCH_REFERENCED_USERS_SUCCESS,
    FETCH_REFERENCED_USERS_FAILURE),
  fetchIsStarred: createRequestReducer(
    FETCH_IS_STARRED_START, FETCH_IS_STARRED_SUCCESS, FETCH_IS_STARRED_FAILURE),
  convert: createRequestReducer(
    CONVERT_START, CONVERT_SUCCESS, CONVERT_FAILURE),
  update: createRequestReducer(
    UPDATE_START, UPDATE_SUCCESS, UPDATE_FAILURE),
  // Assumption: It's okay to prevent the user from sending multiple
  // approval update requests at once, even for different approvals.
  updateApproval: createRequestReducer(
    UPDATE_APPROVAL_START, UPDATE_APPROVAL_SUCCESS, UPDATE_APPROVAL_FAILURE),
});

export const reducer = combineReducers({
  issueRef: combineReducers({
    localId: localIdReducer,
    projectName: projectNameReducer,
  }),

  currentIssue: currentIssueReducer,

  hotlists: hotlistsReducer,
  comments: commentsReducer,
  commentReferences: commentReferencesReducer,
  relatedIssues: relatedIssuesReducer,
  referencedUsers: referencedUsersReducer,
  isStarred: isStarredReducer,
  permissions: permissionsReducer,
  presubmitResponse: presubmitResponseReducer,

  requests: requestsReducer,
});

// Selectors
const RESTRICT_VIEW_PREFIX = 'restrict-view-';
const RESTRICT_EDIT_PREFIX = 'restrict-editissue-';
const RESTRICT_COMMENT_PREFIX = 'restrict-addissuecomment-';

export const issueRef = (state) => state.issue.issueRef;

// TODO(zhangtiff): Eventually Monorail's Redux state will store
// multiple issues, and this selector will have to find the viewed
// issue based on a viewed issue ref.
export const issue = (state) => state.issue.currentIssue;

export const comments = (state) => state.issue.comments;
export const commentsLoaded = (state) => state.issue.commentsLoaded;
export const commentReferences = (state) => state.issue.commentReferences;
export const hotlists = (state) => state.issue.hotlists;
export const issueLoaded = (state) => state.issue.issueLoaded;
export const permissions = (state) => state.issue.permissions;
export const presubmitResponse = (state) => state.issue.presubmitResponse;
export const relatedIssues = (state) => state.issue.relatedIssues;
export const referencedUsers = (state) => state.issue.referencedUsers;
export const isStarred = (state) => state.issue.isStarred;

export const requests = (state) => state.issue.requests;

export const fieldValues = createSelector(
  issue,
  (issue) => issue && issue.fieldValues
);

export const type = createSelector(
  fieldValues,
  (fieldValues) => {
    if (!fieldValues) return;
    const typeFieldValue = fieldValues.find(
      (f) => (f.fieldRef && f.fieldRef.fieldName === 'Type')
    );
    if (typeFieldValue) {
      return typeFieldValue.value;
    }
    return;
  }
);

export const restrictions = createSelector(
  issue,
  (issue) => {
    if (!issue || !issue.labelRefs) return {};

    const restrictions = {};

    issue.labelRefs.forEach((labelRef) => {
      const label = labelRef.label;
      const lowerCaseLabel = label.toLowerCase();

      if (lowerCaseLabel.startsWith(RESTRICT_VIEW_PREFIX)) {
        const permissionType = removePrefix(label, RESTRICT_VIEW_PREFIX);
        if (!('view' in restrictions)) {
          restrictions['view'] = [permissionType];
        } else {
          restrictions['view'].push(permissionType);
        }
      } else if (lowerCaseLabel.startsWith(RESTRICT_EDIT_PREFIX)) {
        const permissionType = removePrefix(label, RESTRICT_EDIT_PREFIX);
        if (!('edit' in restrictions)) {
          restrictions['edit'] = [permissionType];
        } else {
          restrictions['edit'].push(permissionType);
        }
      } else if (lowerCaseLabel.startsWith(RESTRICT_COMMENT_PREFIX)) {
        const permissionType = removePrefix(label, RESTRICT_COMMENT_PREFIX);
        if (!('comment' in restrictions)) {
          restrictions['comment'] = [permissionType];
        } else {
          restrictions['comment'].push(permissionType);
        }
      }
    });

    return restrictions;
  }
);

export const isRestricted = createSelector(
  restrictions,
  (restrictions) => {
    if (!restrictions) return false;
    return ('view' in restrictions && !!restrictions['view'].length) ||
      ('edit' in restrictions && !!restrictions['edit'].length) ||
      ('comment' in restrictions && !!restrictions['comment'].length);
  }
);

export const isOpen = createSelector(
  issue,
  (issue) => issue && issue.statusRef && issue.statusRef.meansOpen || false);

const blockingIssueRefs = createSelector(
  issue, (issue) => issue && issue.blockingIssueRefs || []);

const blockedOnIssueRefs = createSelector(
  issue, (issue) => issue && issue.blockedOnIssueRefs || []);

export const blockingIssues = createSelector(
  blockingIssueRefs, relatedIssues,
  (blockingRefs, relatedIssues) => blockingRefs.map((ref) => {
    const key = issueRefToString(ref);
    if (relatedIssues.has(key)) {
      return relatedIssues.get(key);
    }
    return ref;
  })
);

export const blockedOnIssues = createSelector(
  blockedOnIssueRefs, relatedIssues,
  (blockedOnRefs, relatedIssues) => blockedOnRefs.map((ref) => {
    const key = issueRefToString(ref);
    if (relatedIssues.has(key)) {
      return relatedIssues.get(key);
    }
    return ref;
  })
);

export const mergedInto = createSelector(
  issue, relatedIssues,
  (issue, relatedIssues) => issue && issue.mergedIntoRef
    && relatedIssues.get(issueRefToString(issue.mergedIntoRef))
);

export const sortedBlockedOn = createSelector(
  blockedOnIssues,
  (blockedOn) => blockedOn.sort((a, b) => {
    const aIsOpen = a.statusRef && a.statusRef.meansOpen ? 1 : 0;
    const bIsOpen = b.statusRef && b.statusRef.meansOpen ? 1 : 0;
    return bIsOpen - aIsOpen;
  })
);

// values (from issue.fieldValues) is an array with one entry per value.
// We want to turn this into a map of fieldNames -> values.
export const fieldValueMap = createSelector(
  fieldValues,
  (fieldValues) => {
    if (!fieldValues) return new Map();
    const acc = new Map();
    for (const v of fieldValues) {
      if (!v || !v.fieldRef || !v.fieldRef.fieldName || !v.value) continue;
      let key = [v.fieldRef.fieldName];
      if (v.phaseRef && v.phaseRef.phaseName) {
        key.push(v.phaseRef.phaseName);
      }
      key = key.join(' ');
      if (acc.has(key)) {
        acc.get(key).push(v.value);
      } else {
        acc.set(key, [v.value]);
      }
    }
    return acc;
  }
);

// Get the list of full componentDefs for the viewed issue.
export const components = createSelector(
  issue,
  project.componentsMap,
  (issue, components) => {
    if (!issue || !issue.componentRefs) return [];
    return issue.componentRefs.map((comp) => components.get(comp.path));
  }
);

export const fieldDefs = createSelector(
  project.fieldDefs,
  type,
  (fieldDefs, type) => {
    if (!fieldDefs) return [];
    type = type || '';
    return fieldDefs.filter((f) => {
      // Skip approval type and phase fields here.
      if (f.fieldRef.approvalName
          || f.fieldRef.type === fieldTypes.APPROVAL_TYPE
          || f.isPhaseField) {
        return false;
      }

      // If this fieldDef belongs to only one type, filter out the field if
      // that type isn't the specified type.
      if (f.applicableType && type.toLowerCase()
          !== f.applicableType.toLowerCase()) {
        return false;
      }

      return true;
    });
  }
);

// Action Creators
export const setIssueRef = (localId, projectName) => {
  return {type: SET_ISSUE_REF, localId, projectName};
};

export const fetchCommentReferences = (comments, projectName) => {
  return async (dispatch) => {
    dispatch({type: FETCH_COMMENT_REFERENCES_START});

    try {
      const refs = await autolink.getReferencedArtifacts(comments, projectName);
      const commentRefs = new Map();
      refs.forEach(({componentName, existingRefs}) => {
        commentRefs.set(componentName, existingRefs);
      });
      dispatch({
        type: FETCH_COMMENT_REFERENCES_SUCCESS,
        commentReferences: commentRefs,
      });
    } catch (error) {
      dispatch({type: FETCH_COMMENT_REFERENCES_FAILURE, error});
    }
  };
};

export const fetchReferencedUsers = (issue) => async (dispatch) => {
  if (!issue) return;
  dispatch({type: FETCH_REFERENCED_USERS_START});

  const userRefs = [...(issue.ccRefs || [])];
  if (issue.ownerRef) {
    userRefs.push(issue.ownerRef);
  }
  (issue.approvalValues || []).forEach((approval) => {
    userRefs.push(...(approval.approverRefs || []));
    if (approval.setterRef) {
      userRefs.push(approval.setterRef);
    }
  });

  try {
    const resp = await window.prpcClient.call(
      'monorail.Users', 'ListReferencedUsers', {userRefs});

    const referencedUsers = new Map();
    (resp.users || []).forEach((user) => {
      referencedUsers.set(user.email, user);
    });
    dispatch({type: FETCH_REFERENCED_USERS_SUCCESS, referencedUsers});
  } catch (error) {
    dispatch({type: FETCH_REFERENCED_USERS_FAILURE, error});
  }
};

// TODO(zhangtiff): Figure out if we can reduce request/response sizes by
// diffing issues to fetch against issues we already know about to avoid
// fetching duplicate info.
export const fetchRelatedIssues = (issue) => async (dispatch) => {
  if (!issue) return;
  dispatch({type: FETCH_RELATED_ISSUES_START});

  const refsToFetch = (issue.blockedOnIssueRefs || []).concat(
    issue.blockingIssueRefs || []);
  if (issue.mergedIntoIssueRef) {
    refsToFetch.push(issue.mergedIntoIssueRef);
  }

  const message = {
    issueRefs: refsToFetch,
  };
  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'ListReferencedIssues', message);

    const relatedIssues = new Map();

    const openIssues = resp.openRefs || [];
    const closedIssues = resp.closedRefs || [];
    openIssues.forEach((issue) => {
      issue.statusRef.meansOpen = true;
      relatedIssues.set(issueRefToString(issue), issue);
    });
    closedIssues.forEach((issue) => {
      issue.statusRef.meansOpen = false;
      relatedIssues.set(issueRefToString(issue), issue);
    });
    dispatch({
      type: FETCH_RELATED_ISSUES_SUCCESS,
      relatedIssues: relatedIssues,
    });
  } catch (error) {
    dispatch({type: FETCH_RELATED_ISSUES_FAILURE, error});
  };
};

export const fetchIssuePageData = (message) => async (dispatch) => {
  dispatch(fetchComments(message));
  dispatch(fetch(message));
  dispatch(fetchPermissions(message));
  dispatch(fetchIsStarred(message));
};

export const fetch = (message) => async (dispatch) => {
  dispatch({type: FETCH_START});

  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'GetIssue', message
    );
    const movedToRef = resp.movedToRef;
    const issue = {...resp.issue};
    if (movedToRef) {
      issue.movedToRef = movedToRef;
    }

    dispatch({type: FETCH_SUCCESS, issue});

    dispatch(fetchPermissions(message));
    if (!issue.isDeleted && !movedToRef) {
      dispatch(fetchRelatedIssues(issue));
      dispatch(fetchHotlists(issueRef));
      dispatch(fetchReferencedUsers(issue));
    }
  } catch (error) {
    dispatch({type: FETCH_FAILURE, error});
  }
};

export const fetchHotlists = (issue) => async (dispatch) => {
  dispatch({type: FETCH_HOTLISTS_START});

  try {
    const resp = await window.prpcClient.call(
      'monorail.Features', 'ListHotlistsByIssue', {issue});

    const hotlists = (resp.hotlists || []);
    hotlists.sort((hotlistA, hotlistB) => {
      return hotlistA.name.localeCompare(hotlistB.name);
    });
    dispatch({type: FETCH_HOTLISTS_SUCCESS, hotlists});
  } catch (error) {
    dispatch({type: FETCH_HOTLISTS_FAILURE, error});
  };
};

export const fetchPermissions = (message) => async (dispatch) => {
  dispatch({type: FETCH_PERMISSIONS_START});

  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'ListIssuePermissions', message
    );

    dispatch({type: FETCH_PERMISSIONS_SUCCESS, permissions: resp.permissions});
  } catch (error) {
    dispatch({type: FETCH_PERMISSIONS_FAILURE, error});
  };
};

export const fetchComments = (message) => async (dispatch) => {
  dispatch({type: FETCH_COMMENTS_START});

  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'ListComments', message
    );

    dispatch({type: FETCH_COMMENTS_SUCCESS, comments: resp.comments});
    dispatch(fetchCommentReferences(
      resp.comments, message.issueRef.projectName));
  } catch (error) {
    dispatch({type: FETCH_COMMENTS_FAILURE, error});
  };
};

export const fetchIsStarred = (message) => async (dispatch) => {
  dispatch({type: FETCH_IS_STARRED_START});

  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'IsIssueStarred', message
    );

    dispatch({type: FETCH_IS_STARRED_SUCCESS, isStarred: resp.isStarred});
  } catch (error) {
    dispatch({type: FETCH_IS_STARRED_FAILURE, error});
  };
};

export const star = (issueRef, starred) => async (dispatch) => {
  dispatch({type: STAR_START});

  const message = {issueRef, starred};

  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'StarIssue', message
    );

    dispatch({
      type: STAR_SUCCESS,
      starCount: resp.starCount,
      isStarred: starred,
    });
  } catch (error) {
    dispatch({type: STAR_FAILURE, error});
  }
};

export const presubmit = (message) => async (dispatch) => {
  dispatch({type: PRESUBMIT_START});

  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'PresubmitIssue', message);

    dispatch({type: PRESUBMIT_SUCCESS, presubmitResponse: resp});
  } catch (error) {
    dispatch({type: PRESUBMIT_FAILURE, error: error});
  }
};

export const updateApproval = (message) => async (dispatch) => {
  dispatch({type: UPDATE_APPROVAL_START});

  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'UpdateApproval', message);

    dispatch({type: UPDATE_APPROVAL_SUCCESS, approval: resp.approval});
    const baseMessage = {issueRef: message.issueRef};
    dispatch(fetch(baseMessage));
    dispatch(fetchComments(baseMessage));
  } catch (error) {
    dispatch({type: UPDATE_APPROVAL_FAILURE, error: error});
  };
};

export const update = (message) => async (dispatch) => {
  dispatch({type: UPDATE_START});

  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'UpdateIssue', message);

    dispatch({type: UPDATE_SUCCESS, issue: resp.issue});
    const fetchCommentsMessage = {issueRef: message.issueRef};
    dispatch(fetchComments(fetchCommentsMessage));
    dispatch(fetchRelatedIssues(resp.issue));
    dispatch(fetchReferencedUsers(resp.issue));
  } catch (error) {
    dispatch({type: UPDATE_FAILURE, error: error});
  };
};

export const convert = (message) => async (dispatch) => {
  dispatch({type: CONVERT_START});

  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'ConvertIssueApprovalsTemplate', message);

    dispatch({type: CONVERT_SUCCESS, issue: resp.issue});
    const fetchCommentsMessage = {issueRef: message.issueRef};
    dispatch(fetchComments(fetchCommentsMessage));
  } catch (error) {
    dispatch({type: CONVERT_FAILURE, error: error});
  };
};
