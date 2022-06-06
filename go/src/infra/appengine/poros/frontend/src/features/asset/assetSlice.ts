// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import {
  AssetResourceModel,
  AssetResourceService,
  CreateAssetResourceRequest,
  IAssetResourceService,
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

export interface AssetState {
  assets: AssetModel[];
  pageToken: string | undefined;
  record: AssetModel;
  fetchStatus: string;
  savingStatus: string;
  deletingStatus: string;
  pageNumber: number;
  pageSize: number;
  resources: ResourceModel[];
  assetResources: AssetResourceModel[];
}

const initialState: AssetState = {
  assets: [],
  pageToken: undefined,
  pageNumber: 1,
  pageSize: 25,
  fetchStatus: 'idle',
  record: AssetModel.defaultEntity(),
  savingStatus: 'idle',
  deletingStatus: 'idle',
  resources: [],
  assetResources: [AssetResourceModel.defaultEntity()],
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
    assetResources,
  }: {
    name: string;
    description: string;
    assetResources: AssetResourceModel[];
  }) => {
    const request: CreateAssetRequest = {
      name: name,
      description: description,
    };
    const service: IAssetService = new AssetService();
    const response = await service.create(request);
    createAssetResourceAsync(response.assetId, assetResources);
    return response;
  }
);

export const updateAssetAsync = createAsyncThunk(
  'asset/updateAsset',
  async ({
    asset,
    updateMask,
  }: {
    asset: AssetModel;
    updateMask: string[];
  }) => {
    const request: UpdateAssetRequest = {
      asset: asset,
      updateMask: updateMask,
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

const createAssetResourceAsync = async (
  assetId: string,
  assetResource: AssetResourceModel[]
) => {
  assetResource.forEach(function (entity) {
    const request: CreateAssetResourceRequest = {
      assetId: assetId,
      resourceId: entity.resourceId,
      aliasName: entity.aliasName,
    };
    const service: IAssetResourceService = new AssetResourceService();
    service.create(request);
  });
};

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
    onSelectRecord: (state, action) => {
      state.record = state.assets.filter(
        (s) => s.assetId == action.payload.assetId
      )[0];
    },
    clearSelectedRecord: (state) => {
      state.record = AssetModel.defaultEntity();
      state.assetResources = [AssetResourceModel.defaultEntity()];
    },
    addMachine: (state) => {
      state.assetResources = [
        ...state.assetResources,
        AssetResourceModel.defaultEntity(),
      ];
    },
    removeMachine: (state, action) => {
      state.assetResources = state.assetResources.filter(
        (_, index) => index !== action.payload
      );
    },
    setResourceId: (state, action) => {
      state.assetResources[action.payload.id].resourceId = action.payload.value;
    },
    setAlias: (state, action) => {
      state.assetResources[action.payload.id].aliasName = action.payload.value;
    },
  },

  // The `extraReducers` field lets the slice handle actions generated by
  // createAsyncThunk or in other slices.
  extraReducers: (builder) => {
    builder
      .addCase(fetchAssetAsync.pending, (state) => {
        state.fetchStatus = 'loading';
      })
      .addCase(fetchAssetAsync.fulfilled, (state, action) => {
        state.fetchStatus = 'idle';
        state.record = action.payload;
      })
      .addCase(createAssetAsync.pending, (state) => {
        state.savingStatus = 'loading';
      })
      .addCase(createAssetAsync.fulfilled, (state, action) => {
        state.savingStatus = 'idle';
        state.record = action.payload;
      })
      .addCase(updateAssetAsync.pending, (state) => {
        state.savingStatus = 'loading';
      })
      .addCase(updateAssetAsync.fulfilled, (state, action) => {
        state.savingStatus = 'idle';
        state.record = action.payload;
      })
      .addCase(queryAssetAsync.pending, (state) => {
        state.fetchStatus = 'loading';
      })
      .addCase(queryAssetAsync.fulfilled, (state, action) => {
        state.fetchStatus = 'idle';
        console.log(action.payload.assets);
        state.assets = action.payload.assets;
        state.pageToken = action.payload.nextPageToken;
      })
      .addCase(deleteAssetAsync.pending, (state) => {
        state.deletingStatus = 'loading';
      })
      .addCase(deleteAssetAsync.fulfilled, (state) => {
        state.deletingStatus = 'idle';
        state.record = AssetModel.defaultEntity();
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
  addMachine,
  removeMachine,
  setResourceId,
  setAlias,
} = assetSlice.actions;

export default assetSlice.reducer;
