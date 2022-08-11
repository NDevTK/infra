// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { AssetInstanceModel } from '../../api/asset_instance_service';
import { store } from '../../app/store';
import {
  setDeleteTime,
  onSelectRecord,
  setState,
  AssetInstanceState,
} from './assetInstanceSlice';

const initialState: AssetInstanceState = {
  assetInstances: [],
  pageToken: undefined,
  pageNumber: 1,
  pageSize: 10,
  fetchStatus: 'idle',
  assets: [],
  savingStatus: 'idle',
  deleteAtBuffer: '',
  record: AssetInstanceModel.defaultEntity(),
};

const testAssetInstance: AssetInstanceModel = {
  assetInstanceId: 'test asset instance id',
  assetId: 'test asset id',
  status: 'test status',
  createdBy: '',
  projectId: 'testProjectId',
  createdAt: undefined,
  modifiedBy: '',
  modifiedAt: undefined,
  deleteAt: undefined,
};

const testState: AssetInstanceState = {
  assetInstances: [testAssetInstance],
  pageToken: undefined,
  pageNumber: 1,
  pageSize: 10,
  fetchStatus: 'idle',
  assets: [],
  savingStatus: 'idle',
  deleteAtBuffer: '',
  record: AssetInstanceModel.defaultEntity(),
};

test('should return the initial state', () => {
  expect(store.getState().assetInstance).toEqual(initialState);
});

test('should select asset instance', () => {
  store.dispatch(setState(testState));
  store.dispatch(onSelectRecord({ assetInstanceId: 'test asset instance id' }));
  expect(store.getState().assetInstance.record).toStrictEqual(
    testAssetInstance
  );
});

test('should set delete time', () => {
  store.dispatch(setState(testState));
  store.dispatch(onSelectRecord({ assetInstanceId: 'test asset instance id' }));
  expect(store.getState().assetInstance.deleteAtBuffer).toEqual('');
  store.dispatch(setDeleteTime('0000-00-00T00:00'));
  expect(store.getState().assetInstance.deleteAtBuffer).toEqual(
    '0000-00-00T00:00'
  );
  expect(store.getState().assetInstance.record.deleteAt).toEqual(undefined);
  store.dispatch(setDeleteTime('2022-07-01T00:00'));
  expect(store.getState().assetInstance.deleteAtBuffer).toEqual(
    '2022-07-01T00:00'
  );
  expect(store.getState().assetInstance.record.deleteAt).toEqual(
    new Date('2022-07-01T00:00')
  );
});
