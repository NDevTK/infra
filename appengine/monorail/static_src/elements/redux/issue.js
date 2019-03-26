// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {combineReducers} from 'redux';
import {createSelector} from 'reselect';
import {fieldTypes} from '../shared/field-types.js';
import {removePrefix} from '../shared/helpers.js';
import {createReducer, createRequestReducer} from './redux-helpers.js';
import * as project from './project.js';

// Actions
const UPDATE_REF = 'issue/UPDATE_REF';

const FETCH_START = 'issue/FETCH_START';
const FETCH_SUCCESS = 'issue/FETCH_SUCCESS';
const FETCH_FAILURE = 'issue/FETCH_FAILURE';

const FETCH_HOTLISTS_START = 'issue/FETCH_HOTLISTS_START';
const FETCH_HOTLISTS_SUCCESS = 'issue/FETCH_HOTLISTS_SUCCESS';
const FETCH_HOTLISTS_FAILURE = 'issue/FETCH_HOTLISTS_FAILURE';

const FETCH_PERMISSIONS_START = 'issue/FETCH_PERMISSIONS_START';
const FETCH_PERMISSIONS_SUCCESS = 'issue/FETCH_PERMISSIONS_SUCCESS';
const FETCH_PERMISSIONS_FAILURE = 'issue/FETCH_PERMISSIONS_FAILURE';

const STAR_START = 'issue/STAR_START';
const STAR_SUCCESS = 'issue/STAR_SUCCESS';
const STAR_FAILURE = 'issue/STAR_FAILURE';

const FETCH_IS_STARRED_START = 'issue/FETCH_IS_STARRED_START';
const FETCH_IS_STARRED_SUCCESS = 'issue/FETCH_IS_STARRED_SUCCESS';
const FETCH_IS_STARRED_FAILURE = 'issue/FETCH_IS_STARRED_FAILURE';

const FETCH_COMMENTS_START = 'issue/FETCH_COMMENTS_START';
const FETCH_COMMENTS_SUCCESS = 'issue/FETCH_COMMENTS_SUCCESS';
const FETCH_COMMENTS_FAILURE = 'issue/FETCH_COMMENTS_FAILURE';

const FETCH_COMMENT_REFERENCES_START = 'issue/FETCH_COMMENT_REFERENCES_START';
const FETCH_COMMENT_REFERENCES_SUCCESS = 'issue/FETCH_COMMENT_REFERENCES_SUCCESS';
const FETCH_COMMENT_REFERENCES_FAILURE = 'issue/FETCH_COMMENT_REFERENCES_FAILURE';

const FETCH_BLOCKER_REFERENCES_START = 'issue/FETCH_BLOCKER_REFERENCES_START';
const FETCH_BLOCKER_REFERENCES_SUCCESS = 'issue/FETCH_BLOCKER_REFERENCES_SUCCESS';
const FETCH_BLOCKER_REFERENCES_FAILURE = 'issue/FETCH_BLOCKER_REFERENCES_FAILURE';

const CONVERT_START = 'issue/CONVERT_START';
const CONVERT_SUCCESS = 'issue/CONVERT_SUCCESS';
const CONVERT_FAILURE = 'issue/CONVERT_FAILURE';

const UPDATE_START = 'issue/UPDATE_START';
const UPDATE_SUCCESS = 'issue/UPDATE_SUCCESS';
const UPDATE_FAILURE = 'issue/UPDATE_FAILURE';

const UPDATE_APPROVAL_START = 'issue/UPDATE_APPROVAL_START';
const UPDATE_APPROVAL_SUCCESS = 'issue/UPDATE_APPROVAL_SUCCESS';
const UPDATE_APPROVAL_FAILURE = 'issue/UPDATE_APPROVAL_FAILURE';

/* State Shape
{
  issueRef: {
    issueId: Number,
    projectName: String,
  },
  issue: {
    ...issue: Object,
    approvalValues: Array,
    blockerReferences: Map,
    comments: Array,
    commentReferences: Map,
    hotlists: Array,
    isStarred: Boolean,
    loaded: Boolean,
    permissions: Array,
    starCount: Number,
  }
  requests: {
    fetch: Object,
    fetchHotlists: Object,
    fetchPrefs: Object,
  },
}
*/

// Reducers
const issueIdReducer = createReducer(0, {
  [UPDATE_REF]: (state, action) => action.issueId || state,
});

const projectNameReducer = createReducer('', {
  [UPDATE_REF]: (state, action) => action.projectName || state,
});

const approvalValuesReducer = createReducer([], {
  [UPDATE_APPROVAL_SUCCESS]: (state, action) => {
    return state.map((item, i) => {
      if (item.fieldRef.fieldName === action.approval.fieldRef.fieldName) {
        // PhaseRef isn't populated on the response so we want to make sure
        // it doesn't overwrite the original phaseRef with {}.
        return {...action.approval, phaseRef: item.phaseRef};
      }
      return item;
    });
  },
});

const blockerReferencesReducer = createReducer(new Map(), {
  [FETCH_BLOCKER_REFERENCES_SUCCESS]: (_state, action) => {
    return action.blockerReferences;
  },
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

const hotlistsReducer = createReducer([], {
  [FETCH_HOTLISTS_SUCCESS]: (_, action) => action.hotlists,
});

const isStarredReducer = createReducer(false, {
  [STAR_SUCCESS]: (state, _action) => !state,
  [FETCH_IS_STARRED_SUCCESS]: (_state, action) => !!action.isStarred,
});

const loadedReducer = createReducer(false, {
  [FETCH_SUCCESS]: (_state, _action) => true,
});

const permissionsReducer = createReducer([], {
  [FETCH_PERMISSIONS_SUCCESS]: (_state, action) => action.permissions,
});

const starCountReducer = createReducer(0, {
  [STAR_SUCCESS]: (_state, action) => action.starCount,
});

const issueRefReducer = combineReducers({
  issueId: issueIdReducer,
  projectName: projectNameReducer,
});

const issueExtraFieldsReducer = combineReducers({
  approvalValues: approvalValuesReducer,
  blockerReferences: blockerReferencesReducer,
  comments: commentsReducer,
  commentReferences: commentReferencesReducer,
  hotlists: hotlistsReducer,
  isStarred: isStarredReducer,
  loaded: loadedReducer,
  permissions: permissionsReducer,
  starCount: starCountReducer,
});

const issueReducer = (state, action) => {
  switch (action.type) {
    case FETCH_SUCCESS:
    case CONVERT_SUCCESS:
    case UPDATE_SUCCESS:
      // We want the newly fetched issue to override any of the
      // defaults from issueExtraFieldsReducer(), so reverse the order.
      return {...issueExtraFieldsReducer(), ...action.issue};
    default:
      return {...state, ...issueExtraFieldsReducer(state, action)};
  }
};

const requestsReducer = combineReducers({
  fetchIssue: createRequestReducer(
    FETCH_START, FETCH_SUCCESS, FETCH_FAILURE),
  fetchIssueHotlists: createRequestReducer(
    FETCH_HOTLISTS_START, FETCH_HOTLISTS_SUCCESS, FETCH_HOTLISTS_FAILURE),
  fetchIssuePermissions: createRequestReducer(
    FETCH_PERMISSIONS_START,
    FETCH_PERMISSIONS_SUCCESS,
    FETCH_PERMISSIONS_FAILURE),
  starIssue: createRequestReducer(
    STAR_START, STAR_SUCCESS, STAR_FAILURE),
  fetchComments: createRequestReducer(
    FETCH_COMMENTS_START, FETCH_COMMENTS_SUCCESS, FETCH_COMMENTS_FAILURE),
  fetchCommentReferences: createRequestReducer(
    FETCH_COMMENT_REFERENCES_START,
    FETCH_COMMENT_REFERENCES_SUCCESS,
    FETCH_COMMENT_REFERENCES_FAILURE),
  fetchBlockerReferences: createRequestReducer(
    FETCH_BLOCKER_REFERENCES_START,
    FETCH_BLOCKER_REFERENCES_SUCCESS,
    FETCH_BLOCKER_REFERENCES_FAILURE),
  fetchIsStarred: createRequestReducer(
    FETCH_IS_STARRED_START, FETCH_IS_STARRED_SUCCESS, FETCH_IS_STARRED_FAILURE),
  convertIssue: createRequestReducer(
    CONVERT_START, CONVERT_SUCCESS, CONVERT_FAILURE),
  updateIssue: createRequestReducer(
    UPDATE_START, UPDATE_SUCCESS, UPDATE_FAILURE),
  // Assumption: It's okay to prevent the user from sending multiple
  // approval update requests at once, even for different approvals.
  updateApproval: createRequestReducer(
    UPDATE_APPROVAL_START, UPDATE_APPROVAL_SUCCESS, UPDATE_APPROVAL_FAILURE),
});

export const reducer = combineReducers({
  issueRef: issueRefReducer,
  issue: issueReducer,
  requests: requestsReducer,
});

// Selectors
const RESTRICT_VIEW_PREFIX = 'restrict-view-';
const RESTRICT_EDIT_PREFIX = 'restrict-editissue-';
const RESTRICT_COMMENT_PREFIX = 'restrict-addissuecomment-';

// TODO(zhangtiff): Eventually Monorail's Redux state will store
// multiple issues, and this selector will have to find the viewed
// issue based on a viewed issue ref.
export const issue = (state) => state.issue.issue;

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
  (issue) => issue && issue.statusRef && issue.statusRef.meansOpen
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
export const fetchCommentReferences = (comments, projectName) => async (dispatch) => {
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

// TODO(zhangtiff): Figure out if we can reduce request/response sizes by
// diffing issues to fetch against issues we already know about to avoid
// fetching duplicate info.
export const fetchBlockerReferences = (issue) => async (dispatch) => {
  if (!issue) return;
  dispatch({type: FETCH_BLOCKER_REFERENCES_START});

  const refsToFetch = (issue.blockedOnIssueRefs || []).concat(
      issue.blockingIssueRefs || []);
  if (issue.mergedIntoIssueRef) {
    refsToFetch.push(issue.mergedIntoIssueRef);
  }

  const message = {issueRefs: refsToFetch};
  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'ListReferencedIssues', message);

    let blockerReferences = new Map();

    const openIssues = resp.openRefs || [];
    const closedIssues = resp.closedRefs || [];
    openIssues.forEach((issue) => {
      blockerReferences.set(
        `${issue.projectName}:${issue.localId}`, {
          issue: issue,
          isClosed: false,
        });
    });
    closedIssues.forEach((issue) => {
      blockerReferences.set(
        `${issue.projectName}:${issue.localId}`, {
          issue: issue,
          isClosed: true,
        });
    });
    dispatch({
      type: FETCH_BLOCKER_REFERENCES_SUCCESS,
      blockerReferences: blockerReferences,
    });
  } catch (error) {
    dispatch({type: FETCH_BLOCKER_REFERENCES_FAILURE, error});
  };
};

export const fetchIssue = (message) => async (dispatch) => {
  dispatch({type: FETCH_START});

  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'GetIssue', message);

    dispatch({type: FETCH_SUCCESS, issue: resp.issue});

    dispatch(fetchIssuePermissions(message));
    if (!resp.issue.isDeleted) {
      dispatch(fetchBlockerReferences(resp.issue));
      dispatch(fetchIssueHotlists(message.issueRef));
    }
  } catch (error) {
    dispatch({type: FETCH_FAILURE, error});
  };
};

export const fetchIssueHotlists = (issue) => async (dispatch) => {
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

export const fetchIssuePermissions = (message) => async (dispatch) => {
  dispatch({type: FETCH_PERMISSIONS_START});

  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'ListIssuePermissions', message);

    dispatch({type: FETCH_PERMISSIONS_SUCCESS, permissions: resp.permissions});
  } catch (error) {
    dispatch({type: FETCH_PERMISSIONS_FAILURE, error});
  };
};

export const fetchComments = (message) => async (dispatch) => {
  dispatch({type: FETCH_COMMENTS_START});

  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'ListComments', message);

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
      'monorail.Issues', 'IsIssueStarred', message);

    dispatch({type: FETCH_IS_STARRED_SUCCESS, isStarred: resp.isStarred});
  } catch (error) {
    dispatch({type: FETCH_IS_STARRED_FAILURE, error});
  };
};

export const starIssue = (issueRef, starred) => async (dispatch) => {
  dispatch({type: STAR_START});

  const message = {issueRef, starred};

  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'StarIssue', message);

    dispatch({
      type: STAR_SUCCESS,
      starCount: resp.starCount,
      isStarred: starred,
    });
  } catch (error) {
    dispatch({type: STAR_FAILURE, error});
  }
};

export const updateApproval = (message) => async (dispatch) => {
  dispatch({type: UPDATE_APPROVAL_START});

  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'UpdateApproval', message);

    dispatch({type: UPDATE_APPROVAL_SUCCESS, approval: resp.approval});
    const baseMessage = {issueRef: message.issueRef};
    dispatch(fetchIssue(baseMessage));
    dispatch(fetchComments(baseMessage));
  } catch (error) {
    dispatch({type: UPDATE_APPROVAL_FAILURE, error: error});
  };
};

export const updateIssue = (message) => async (dispatch) => {
  dispatch({type: UPDATE_START});

  try {
    const resp = await window.prpcClient.call(
      'monorail.Issues', 'UpdateIssue', message);

    dispatch({type: UPDATE_SUCCESS, issue: resp.issue});
    const fetchCommentsMessage = {issueRef: message.issueRef};
    dispatch(fetchComments(fetchCommentsMessage));
    dispatch(fetchBlockerReferences(resp.issue));
  } catch (error) {
    dispatch({type: UPDATE_FAILURE, error: error});
  };
};

export const convertIssue = (message) => async (dispatch) => {
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

export const updateRef = (issueId, projectName) => {
  return {type: UPDATE_REF, issueId, projectName};
};
