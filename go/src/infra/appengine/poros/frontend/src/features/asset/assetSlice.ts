// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import {
  AssetResourceModel,
  AssetResourceService,
  IAssetResourceService,
  ListAssetResourcesRequest,
} from '../../api/asset_resource_service';
import {
  CreateAssetRequest,
  DeleteAssetRequest,
  GetAssetRequest,
  IAssetService,
  AssetService,
  ListAssetsRequest,
  AssetModel,
  UpdateAssetRequest,
} from '../../api/asset_service';
import { ResourceModel } from '../../api/resource_service';
import { RootState } from '../../app/store';
import { queryResourceAsync } from '../resource/resourceSlice';

export interface AssetState {
  assets: AssetModel[];
  pageToken: string | undefined;
  record: AssetModel;
  fetchAssetStatus: string;
  savingStatus: string;
  deletingStatus: string;
  fetchAssetResourceStatus: string;
  pageNumber: number;
  pageSize: number;
  resources: ResourceModel[];
  assetResourcesToSave: AssetResourceModel[];
  assetResourcesToDelete: AssetResourceModel[];
}

const initialState: AssetState = {
  assets: [],
  pageToken: undefined,
  pageNumber: 1,
  pageSize: 25,
  fetchAssetStatus: 'idle',
  record: AssetModel.defaultEntity(),
  savingStatus: 'idle',
  deletingStatus: 'idle',
  fetchAssetResourceStatus: 'idle',
  resources: [],
  assetResourcesToSave: [AssetResourceModel.defaultEntity()],
  assetResourcesToDelete: [],
};

// The function below is called a thunk and allows us to perform async logic. It
// can be dispatched like a regular action: `dispatch(fetchAssetAsync(10))`. This
// will call the thunk with the `dispatch` function as the first argument. Async
// code can then be executed and other actions can be dispatched. Thunks are
// typically used to make async requests.
export const fetchAssetAsync = createAsyncThunk(
  'asset/fetchAsset',
  async (assetId: string) => {
    const request: GetAssetRequest = {
      id: assetId,
    };
    const service: IAssetService = new AssetService();
    const response = await service.get(request);
    // The value we return becomes the `fulfilled` action payload
    return response;
  }
);

export const createAssetAsync = createAsyncThunk(
  'asset/createAsset',
  async ({
    name,
    description,
    assetType,
    assetResourcesToSave,
  }: {
    name: string;
    description: string;
    assetType: string;
    assetResourcesToSave: AssetResourceModel[];
  }) => {
    const request: CreateAssetRequest = {
      name: name,
      description: description,
      assetType: assetType,
      assetResourcesToSave: assetResourcesToSave,
    };
    const service: IAssetService = new AssetService();
    const response = await service.create(request);
    return response;
  }
);

export const updateAssetAsync = createAsyncThunk(
  'asset/updateAsset',
  async ({
    asset,
    assetUpdateMask,
    assetResourceUpdateMask,
    assetResourcesToSave,
    assetResourcesToDelete,
  }: {
    asset: AssetModel;
    assetUpdateMask: string[];
    assetResourceUpdateMask: string[];
    assetResourcesToSave: AssetResourceModel[];
    assetResourcesToDelete: AssetResourceModel[];
  }) => {
    const request: UpdateAssetRequest = {
      asset: asset,
      assetUpdateMask: assetUpdateMask,
      assetResourceUpdateMask: assetResourceUpdateMask,
      assetResourcesToSave: assetResourcesToSave,
      assetResourcesToDelete: assetResourcesToDelete,
    };
    const service: IAssetService = new AssetService();
    const response = await service.update(request);
    return response;
  }
);

export const queryAssetAsync = createAsyncThunk(
  'asset/queryAsset',
  async ({ pageSize, pageToken }: { pageSize: number; pageToken: string }) => {
    const request: ListAssetsRequest = {
      pageSize: pageSize,
      pageToken: pageToken,
      readMask: undefined,
    };
    const service: IAssetService = new AssetService();
    const response = await service.list(request);
    return response;
  }
);

export const deleteAssetAsync = createAsyncThunk(
  'asset/deleteAsset',
  async (assetId: string) => {
    const request: DeleteAssetRequest = {
      id: assetId,
    };
    const service: IAssetService = new AssetService();
    const response = await service.get(request);
    return response;
  }
);

export const queryAssetResourceAsync = createAsyncThunk(
  'asset/queryAssetResource',
  async () => {
    const request: ListAssetResourcesRequest = {
      readMask: undefined,
      pageSize: 100,
      pageToken: '',
    };
    const service: IAssetResourceService = new AssetResourceService();
    const response = await service.list(request);
    return response;
  }
);

export const assetSlice = createSlice({
  name: 'asset',
  initialState,
  reducers: {
    setPageSize: (state, action) => {
      state.pageSize = action.payload.pageSize;
    },
    setName: (state, action) => {
      state.record.name = action.payload;
    },
    setDescription: (state, action) => {
      state.record.description = action.payload;
    },
    setAssetType: (state, action) => {
      state.record.assetType = action.payload;
    },
    onSelectRecord: (state, action) => {
      state.record = state.assets.filter(
        (s) => s.assetId == action.payload.assetId
      )[0];
    },
    clearSelectedRecord: (state) => {
      state.record = AssetModel.defaultEntity();
      state.assetResourcesToSave = [AssetResourceModel.defaultEntity()];
      state.assetResourcesToDelete = [];
    },
    addMachine: (state) => {
      state.assetResourcesToSave = [
        ...state.assetResourcesToSave,
        AssetResourceModel.defaultEntity(),
      ];
    },
    removeMachine: (state, action) => {
      if (state.assetResourcesToSave.length > 1) {
        if (state.assetResourcesToSave[action.payload].assetResourceId !== '') {
          state.assetResourcesToDelete.push(
            state.assetResourcesToSave[action.payload]
          );
        }
        state.assetResourcesToSave = state.assetResourcesToSave.filter(
          (_, index) => index !== action.payload
        );
      }
    },
    setResourceId: (state, action) => {
      state.assetResourcesToSave[action.payload.id].resourceId =
        action.payload.value;
    },
    setAlias: (state, action) => {
      state.assetResourcesToSave[action.payload.id].aliasName =
        action.payload.value;
    },
  },

  // The `extraReducers` field lets the slice handle actions generated by
  // createAsyncThunk or in other slices.
  extraReducers: (builder) => {
    builder
      .addCase(fetchAssetAsync.pending, (state) => {
        state.fetchAssetStatus = 'loading';
      })
      .addCase(fetchAssetAsync.fulfilled, (state, action) => {
        state.fetchAssetStatus = 'idle';
        state.record = action.payload;
      })
      .addCase(createAssetAsync.pending, (state) => {
        state.savingStatus = 'loading';
      })
      .addCase(createAssetAsync.fulfilled, (state, action) => {
        state.savingStatus = 'idle';
        state.record = action.payload.asset;
        state.assetResourcesToSave = action.payload.assetResources;
      })
      .addCase(updateAssetAsync.pending, (state) => {
        state.savingStatus = 'loading';
      })
      .addCase(updateAssetAsync.fulfilled, (state, action) => {
        state.savingStatus = 'idle';
        state.record = action.payload.asset;
        state.assetResourcesToSave = action.payload.assetResources;
      })
      .addCase(queryAssetAsync.pending, (state) => {
        state.fetchAssetStatus = 'loading';
      })
      .addCase(queryAssetAsync.fulfilled, (state, action) => {
        state.fetchAssetStatus = 'idle';
        state.assets = action.payload.assets;
        state.pageToken = action.payload.nextPageToken;
      })
      .addCase(deleteAssetAsync.pending, (state) => {
        state.deletingStatus = 'loading';
      })
      .addCase(deleteAssetAsync.fulfilled, (state) => {
        state.deletingStatus = 'idle';
        state.record = AssetModel.defaultEntity();
      })
      .addCase(queryAssetResourceAsync.pending, (state) => {
        state.fetchAssetResourceStatus = 'loading';
      })
      .addCase(queryAssetResourceAsync.fulfilled, (state, action) => {
        state.fetchAssetResourceStatus = 'idle';
        state.assetResourcesToSave = [
          ...action.payload.assetResources.filter(
            (entity) => entity.assetId === state.record.assetId
          ),
          AssetResourceModel.defaultEntity(),
        ];
      })
      .addCase(queryResourceAsync.fulfilled, (state, action) => {
        state.resources = action.payload.resources;
      });
  },
});

// The function below is called a selector and allows us to select a value from
// the state. Selectors can also be defined inline where they're used instead of
// in the slice file. For example: `useSelector((state: RootState) => state.asset)`
export const selectAssetState = (state: RootState) => state.asset;

export const {
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
} = assetSlice.actions;

export default assetSlice.reducer;
