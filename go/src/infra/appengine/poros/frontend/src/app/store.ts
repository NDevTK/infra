// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { configureStore, ThunkAction, Action } from '@reduxjs/toolkit';
import { setupListeners } from '@reduxjs/toolkit/query';
import assetReducer from '../features/asset/assetSlice';
import resourceReducer from '../features/resource/resourceSlice';
import utilityReducer from '../features/utility/utilitySlice';

export const store = configureStore({
  reducer: {
    asset: assetReducer,
    resource: resourceReducer,
    utility: utilityReducer,
  },
  devTools: process.env.NODE_ENV !== 'production',
});

setupListeners(store.dispatch);

export type AppDispatch = typeof store.dispatch;
export type RootState = ReturnType<typeof store.getState>;
export type AppThunk<ReturnType = void> = ThunkAction<
  ReturnType,
  RootState,
  unknown,
  Action<string>
>;
