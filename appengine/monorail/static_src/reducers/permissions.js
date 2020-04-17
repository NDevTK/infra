// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

/**
 * @fileoverview Permissions actions, selectors, and reducers organized into
 * a single Redux "Duck" that manages updating and retrieving permissions state
 * on the frontend.
 *
 * The Permissions data is stored in a normalized format.
 * `permissions` stores all PermissionSets[] indexed by resource name.
 *
 * Reference: https://github.com/erikras/ducks-modular-redux
 */

import {combineReducers} from 'redux';
import {createReducer, createRequestReducer} from './redux-helpers.js';

import {prpcClient} from 'prpc-client-instance.js';

import 'shared/typedef.js';

/** @typedef {import('redux').AnyAction} AnyAction */

// Actions
export const BATCH_GET_START = 'permissions/BATCH_GET_START';
export const BATCH_GET_SUCCESS = 'permissions/BATCH_GET_SUCCESS';
export const BATCH_GET_FAILURE = 'permissions/BATCH_GET_FAILURE';

/* State Shape
{
  permissionSets: Object<string, PermissionSet>,

  requests: {
    batchGet: ReduxRequestState,
  },
}
*/

// Reducers

/**
 * All PermissionSets indexed by resource name.
 * @param {Object<string, PermissionSet>} state The existing items.
 * @param {AnyAction} action
 * @param {Array<PermissionSet>} action.permissionSets
 * @return {Object<string, PermissionSet>}
 */
export const permissionSetsReducer = createReducer({}, {
  [BATCH_GET_SUCCESS]: (state, {permissionSets}) => {
    const newState = {...state};
    for (const permissionSet of permissionSets) {
      newState[permissionSet.resource] = permissionSet;
    }
    return newState;
  },
});

const requestsReducer = combineReducers({
  batchGet: createRequestReducer(
      BATCH_GET_START, BATCH_GET_SUCCESS, BATCH_GET_FAILURE),
});

export const reducer = combineReducers({
  permissionSets: permissionSetsReducer,

  requests: requestsReducer,
});

// Selectors

/**
 * Returns all the PermissionSets in the store as a mapping.
 * @param {any} state
 * @return {Object<string, PermissionSet>}
 */
export const permissionSets = (state) => state.permissions.permissionSets;

/**
 * Returns the Permissions requests.
 * @param {any} state
 * @return {Object.<string, ReduxRequestState>}
 */
export const requests = (state) => state.permissions.requests;

// Action Creators

/**
 * Action creator to fetch PermissionSets.
 * @param {Array<string>} resourceNames The resource names to get.
 * @return {function(function): Promise<Array<PermissionSet>>}
 */
export const batchGet = (resourceNames) => async (dispatch) => {
  dispatch({type: BATCH_GET_START});

  try {
    const args = {names: resourceNames};
    /** @type {{permissionSets: Array<PermissionSet>}} */
    const {permissionSets} = await prpcClient.call(
        'monorail.v1.Permissions', 'BatchGetPermissionSets', args);

    dispatch({type: BATCH_GET_SUCCESS, permissionSets});

    return permissionSets;
  } catch (error) {
    dispatch({type: BATCH_GET_FAILURE, error});
  };
};
