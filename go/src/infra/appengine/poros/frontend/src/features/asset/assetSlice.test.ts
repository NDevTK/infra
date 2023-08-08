// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { AssetResourceModel } from '../../api/asset_resource_service';
import { AssetModel } from '../../api/asset_service';
import { AssetState, AssetRecordValidation } from './assetSlice';
import { store } from '../../app/store';
import {
  setPageSize,
  onSelectRecord,
  clearSelectedRecord,
  setName,
  setDescription,
  setAssetType,
  addMachine,
  removeMachine,
  setResourceId,
  setAlias,
  setState,
  setAssetSpinRecord,
} from './assetSlice';
import { ResourceModel } from '../../api/resource_service';

const initialState: AssetState = {
  assets: [],
  pageToken: undefined,
  pageNumber: 1,
  pageSize: 25,
  fetchAssetStatus: 'idle',
  fetchAssetResourceStatus: 'idle',
  record: AssetModel.defaultEntity(),
  savingStatus: 'idle',
  deletingStatus: 'idle',
  resources: [],
  assetResourcesToSave: [AssetResourceModel.defaultEntity()],
  assetResourcesToDelete: [],
  assetSpinRecord: '',
  fetchResourceStatus: 'idle',
  defaultAssetResources: [],
  recordValidation: AssetRecordValidation.defaultEntity(),
  showDefaultMachines: false,
};

const testAsset: AssetModel = {
  assetId: 'test asset id',
  name: 'test name',
  description: 'test description',
  assetType: 'test type',
  createdBy: '',
  createdAt: undefined,
  modifiedBy: '',
  modifiedAt: undefined,
  deleted: false,
};

const testResource: ResourceModel = {
  resourceId: 'test resource id',
  name: 'test name',
  type: 'test type',
  imageFamily: 'test image family',
  imageProject: 'test image project',
  imageSource: 'test image source',
  description: 'test description',
  operatingSystem: 'test os',
  createdAt: undefined,
  createdBy: '',
  modifiedAt: undefined,
  modifiedBy: '',
  deleted: false,
};

const testAssetResource: AssetResourceModel = {
  assetResourceId: 'test asset resource id',
  assetId: 'test asset id',
  resourceId: 'test resource id',
  aliasName: 'test alias',
  createdAt: undefined,
  createdBy: '',
  modifiedAt: undefined,
  modifiedBy: '',
  default: false,
};

const testState: AssetState = {
  assets: [testAsset],
  pageToken: undefined,
  pageNumber: 1,
  pageSize: 25,
  fetchAssetStatus: 'idle',
  record: AssetModel.defaultEntity(),
  savingStatus: 'idle',
  deletingStatus: 'idle',
  fetchAssetResourceStatus: 'idle',
  resources: [testResource],
  assetResourcesToSave: [testAssetResource, AssetResourceModel.defaultEntity()],
  assetResourcesToDelete: [],
  assetSpinRecord: '',
  fetchResourceStatus: 'idle',
  defaultAssetResources: [],
  recordValidation: AssetRecordValidation.defaultEntity(),
  showDefaultMachines: false,
};
test('should return the initial state', () => {
  expect(store.getState().asset).toEqual(initialState);
});

test('should change page size', () => {
  store.dispatch(setPageSize({ pageSize: 100 }));
  expect(store.getState().asset.pageSize).toEqual(100);
});

test('should change name of selected record', () => {
  store.dispatch(setName('test name'));
  expect(store.getState().asset.record.name).toEqual('test name');
});

test('should change asset type of selected record', () => {
  store.dispatch(setAssetType('test type'));
  expect(store.getState().asset.record.assetType).toEqual('test type');
});

test('should change description of selected record', () => {
  store.dispatch(setDescription('test description'));
  expect(store.getState().asset.record.description).toEqual('test description');
});

test('should change record on which asset to spin', () => {
  store.dispatch(setAssetSpinRecord('test spin record'));
  expect(store.getState().asset.assetSpinRecord).toEqual('test spin record');
});

test('should clear selected record', () => {
  store.dispatch(clearSelectedRecord());
  expect(store.getState().asset.record).toEqual(AssetModel.defaultEntity());
});

test('should select the requested record', () => {
  store.dispatch(setState(testState));
  store.dispatch(onSelectRecord({ assetId: 'test asset id' }));
  expect(store.getState().asset.record).toEqual(testAsset);
});

test('should add an associated machine entry', () => {
  store.dispatch(setState(initialState));
  expect(store.getState().asset.assetResourcesToSave).toEqual([
    AssetResourceModel.defaultEntity(),
  ]);
  store.dispatch(addMachine());
  expect(store.getState().asset.assetResourcesToSave).toEqual([
    AssetResourceModel.defaultEntity(),
    AssetResourceModel.defaultEntity(),
  ]);
});

test('should remove the requested associated machine entry', () => {
  store.dispatch(setState(testState));
  expect(store.getState().asset.assetResourcesToSave).toEqual([
    testAssetResource,
    AssetResourceModel.defaultEntity(),
  ]);
  store.dispatch(removeMachine(0));
  expect(store.getState().asset.assetResourcesToSave).toEqual([
    AssetResourceModel.defaultEntity(),
  ]);
});

test('should change alias name for the requested associated machine', () => {
  store.dispatch(setState(initialState));
  expect(store.getState().asset.assetResourcesToSave).toEqual([
    AssetResourceModel.defaultEntity(),
  ]);
  store.dispatch(setAlias({ id: 0, value: 'test alias updated' }));
  expect(store.getState().asset.assetResourcesToSave[0].aliasName).toEqual(
    'test alias updated'
  );
});

test('should change resource id for the requested associated machine', () => {
  store.dispatch(setState(initialState));
  expect(store.getState().asset.assetResourcesToSave).toEqual([
    AssetResourceModel.defaultEntity(),
  ]);
  store.dispatch(setResourceId({ id: 0, value: 'resource id updated' }));
  expect(store.getState().asset.assetResourcesToSave[0].resourceId).toEqual(
    'resource id updated'
  );
});
