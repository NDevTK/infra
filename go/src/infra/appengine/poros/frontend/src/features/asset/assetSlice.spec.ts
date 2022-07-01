// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { AssetResourceModel } from '../../api/asset_resource_service';
import { AssetModel } from '../../api/asset_service';
import assetReducer, { AssetState } from './assetSlice';

describe('asset reducer', () => {
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
  };
  it('should handle initial state', () => {
    expect(assetReducer(undefined, { type: 'unknown' })).toEqual(initialState);
  });
});
