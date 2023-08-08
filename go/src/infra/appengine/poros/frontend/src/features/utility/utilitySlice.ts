// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import { RootState } from '../../app/store';
import { IUtilityService, UtilityService } from '../../api/utility_service';
import {
  AssetInstanceService,
  FetchLogsRequest,
  IAssetInstanceService,
} from '../../api/asset_instance_service';

export interface UtilityState {
  userEmail: string;
  userPicture: string;
  rightSideDrawerOpen: boolean;
  activeEntity: string;
  spinDialogOpen: boolean;
  deleteResourceDialogOpen: boolean;
  deleteAssetDialogOpen: boolean;
  showLogs: boolean;
  logs: string;
  activeLogsAssetInstanceId: string;
  activeLogsAssetDetails: object;
}

const initialState: UtilityState = {
  userEmail: '',
  userPicture: '',
  rightSideDrawerOpen: false,
  activeEntity: '',
  spinDialogOpen: false,
  deleteResourceDialogOpen: false,
  deleteAssetDialogOpen: false,
  showLogs: false,
  logs: '',
  activeLogsAssetInstanceId: '',
  activeLogsAssetDetails: { name: '', createdAt: '', status: '' },
};

export const fetchUserPictureAsync = createAsyncThunk(
  'asset/UserPicture',
  async () => {
    const service: IUtilityService = new UtilityService();
    const response = await service.getUserPicture();
    return response;
  }
);

export const fetchLogsAsync = createAsyncThunk(
  'asset/logs',
  async ({ assetInstanceId }: { assetInstanceId: string }) => {
    const service: IAssetInstanceService = new AssetInstanceService();
    const fetchLogsReq: FetchLogsRequest = {
      assetInstanceId,
    };
    const response = await service.fetchLogs(fetchLogsReq);
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
    setDeleteAssetDialogClose: (state) => {
      state.deleteAssetDialogOpen = false;
    },
    setDeleteAssetDialogOpen: (state) => {
      state.deleteAssetDialogOpen = true;
    },
    hideLogs: (state) => {
      state.showLogs = false;
      state.logs = '';
      state.activeLogsAssetInstanceId = '';
    },
    setActiveLogAssetDetails: (state, action) => {
      state.activeLogsAssetDetails = { ...action.payload };
    },
  },
  extraReducers: (builder) => {
    builder.addCase(fetchUserPictureAsync.fulfilled, (state, action) => {
      state.userPicture = action.payload;
    });
    builder.addCase(logoutAsync.fulfilled, () => {
      window.location.href = window.logoutUrl;
    });
    builder.addCase(fetchLogsAsync.pending, (state, action) => {
      state.showLogs = true;
      state.activeLogsAssetInstanceId = action.meta.arg.assetInstanceId;
    });
    builder.addCase(fetchLogsAsync.fulfilled, (state, action) => {
      state.logs = action.payload.logs;
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
  setDeleteAssetDialogClose,
  setDeleteAssetDialogOpen,
  hideLogs,
  setActiveLogAssetDetails,
} = utilitySlice.actions;

export default utilitySlice.reducer;
