// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { AssetInstanceModel } from '../../api/asset_instance_service';
import assetInstanceReducer, { AssetInstanceState } from './assetInstanceSlice';

describe('asset instance reducer', () => {
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
  it('should handle initial state', () => {
    expect(assetInstanceReducer(undefined, { type: 'unknown' })).toEqual(
      initialState
    );
  });
});
