// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { ResourceModel } from '../../api/resource_service';
import resourceReducer, { ResourceState } from './resourceSlice';

describe('resource reducer', () => {
  const initialState: ResourceState = {
    resources: [],
    pageToken: undefined,
    pageNumber: 1,
    pageSize: 25,
    fetchStatus: 'idle',
    record: ResourceModel.defaultEntity(),
    savingStatus: 'idle',
    deletingStatus: 'idle',
  };
  it('should handle initial state', () => {
    expect(resourceReducer(undefined, { type: 'unknown' })).toEqual(
      initialState
    );
  });
});
