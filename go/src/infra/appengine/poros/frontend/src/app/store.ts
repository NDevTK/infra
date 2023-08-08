// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { configureStore, ThunkAction, Action } from '@reduxjs/toolkit';
import { setupListeners } from '@reduxjs/toolkit/query';
import assetReducer from '../features/asset/assetSlice';
import resourceReducer from '../features/resource/resourceSlice';
import utilityReducer from '../features/utility/utilitySlice';
import assetInstanceReducer from '../features/asset_instance/assetInstanceSlice';

export const store = configureStore({
  reducer: {
    asset: assetReducer,
    resource: resourceReducer,
    utility: utilityReducer,
    assetInstance: assetInstanceReducer,
  },
  devTools: process.env.NODE_ENV !== 'production',
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: false,
    }),
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
