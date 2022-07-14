// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import { RootState } from '../../app/store';
import { IUtilityService, UtilityService } from '../../api/utility_service';

export interface UtilityState {
  userEmail: string;
  userPicture: string;
  rightSideDrawerOpen: boolean;
  activeEntity: string;
  spinDialogOpen: boolean;
  deleteResourceDialogOpen: boolean;
}

const initialState: UtilityState = {
  userEmail: '',
  userPicture: '',
  rightSideDrawerOpen: false,
  activeEntity: '',
  spinDialogOpen: false,
  deleteResourceDialogOpen: false,
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
    setRightSideDrawerClose: (state) => {
      state.rightSideDrawerOpen = false;
    },
    setRightSideDrawerOpen: (state) => {
      state.rightSideDrawerOpen = true;
    },
    setActiveEntity: (state, action) => {
      state.activeEntity = action.payload;
    },
    setSpinDialogClose: (state) => {
      state.spinDialogOpen = false;
    },
    setSpinDialogOpen: (state) => {
      state.spinDialogOpen = true;
    },
    setDeleteResourceDialogClose: (state) => {
      state.deleteResourceDialogOpen = false;
    },
    setDeleteResourceDialogOpen: (state) => {
      state.deleteResourceDialogOpen = true;
    },
  },
  extraReducers: (builder) => {
    builder.addCase(fetchUserPictureAsync.fulfilled, (state, action) => {
      state.userPicture = action.payload;
    });
    builder.addCase(logoutAsync.fulfilled, () => {
      window.location.href = window.logoutUrl;
    });
  },
});

export const selectUtilityState = (state: RootState) => state.utility;

export const {
  setUserPicture,
  setRightSideDrawerClose,
  setRightSideDrawerOpen,
  setActiveEntity,
  setSpinDialogClose,
  setSpinDialogOpen,
  setDeleteResourceDialogClose,
  setDeleteResourceDialogOpen,
} = utilitySlice.actions;

export default utilitySlice.reducer;
