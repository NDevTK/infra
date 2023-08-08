// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {combineReducers, Reducer} from 'redux';

import {messageReducer, MessageStateType} from './message';
import {queryReducer, QueryStoreStateType} from './query';
import {repairRecordReducer, RepairRecordStateType} from './repair-record';
import {userReducer, UserStateType} from './user';

export interface ApplicationState {
  record: RepairRecordStateType;
  user: UserStateType;
  message: MessageStateType;
  queryStore: QueryStoreStateType;
}

export const reducers: Reducer<ApplicationState> =
    combineReducers<ApplicationState>({
      record: repairRecordReducer,
      user: userReducer,
      message: messageReducer,
      queryStore: queryReducer,
    });
