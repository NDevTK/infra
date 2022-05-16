// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import { RootState } from '../../app/store';
import { IUtilityService, UtilityService } from '../../api/utility_service';

export interface UtilityState {
  userEmail: string;
  userPicture: string;
}

const initialState: UtilityState = {
  userEmail: '',
  userPicture: '',
};

export const fetchUserPictureAsync = createAsyncThunk(
  'asset/UserPicture',
  async () => {
    const service: IUtilityService = new UtilityService();
    const response = await service.getUserPicture();
    return response;
  }
);

export const logoutAsync = createAsyncThunk('asset/logout', async () => {
  const service: IUtilityService = new UtilityService();
  await service.logout();
});

export const utilitySlice = createSlice({
  name: 'utility',
  initialState,
  reducers: {
    setUserPicture: (state, action) => {
      state.userPicture = action.payload;
    },
  },
  extraReducers: (builder) => {
    builder.addCase(fetchUserPictureAsync.fulfilled, (state, action) => {
      state.userPicture = action.payload;
    });
    builder.addCase(logoutAsync.fulfilled, () => {
      window.location.href = 'www.google.com';
    });
  },
});

export const selectUtilityState = (state: RootState) => state.utility;

export const { setUserPicture } = utilitySlice.actions;

export default utilitySlice.reducer;
