// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { ResourceModel } from '../../api/resource_service';
import { store } from '../../app/store';
import {
  clearSelectedRecord,
  ResourceRecordValidation,
  ResourceState,
  setDefaultState,
  setDescription,
  setImageFamily,
  setImageProject,
  setName,
  setOperatingSystem,
  setPageSize,
  setType,
} from './resourceSlice';

const initialState: ResourceState = {
  resources: [],
  pageToken: undefined,
  pageNumber: 1,
  pageSize: 25,
  fetchStatus: 'idle',
  record: ResourceModel.defaultEntity(),
  savingStatus: 'idle',
  deletingStatus: 'idle',
  recordValidation: ResourceRecordValidation.defaultEntity(),
};

test('should return the initial state', () => {
  expect(store.getState().resource).toEqual(initialState);
});

test('should change page size', () => {
  store.dispatch(setPageSize({ pageSize: 100 }));
  expect(store.getState().resource.pageSize).toEqual(100);
});

test('should change name of selected record', () => {
  store.dispatch(setName('Resource 1'));
  expect(store.getState().resource.record.name).toEqual('Resource 1');
});

test('should change type of selected record', () => {
  store.dispatch(setType('network'));
  expect(store.getState().resource.record.type).toEqual('network');
});

test('should change description of selected record', () => {
  store.dispatch(setDescription('Resource 1 description'));
  expect(store.getState().resource.record.description).toEqual(
    'Resource 1 description'
  );
});

test('should change image project of selected record', () => {
  store.dispatch(setImageProject('test-project'));
  expect(store.getState().resource.record.imageProject).toEqual('test-project');
});

test('should change image family of selected record', () => {
  store.dispatch(setImageFamily('test-family'));
  expect(store.getState().resource.record.imageFamily).toEqual('test-family');
});

test('should clear selected record', () => {
  store.dispatch(clearSelectedRecord());
  expect(store.getState().resource.record).toEqual(
    ResourceModel.defaultEntity()
  );
});

test('should set default state', () => {
  store.dispatch(setDefaultState());
  expect(store.getState().resource).toEqual(initialState);
});
