// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import {
  AssetInstanceService,
  CreateAssetInstanceRequest,
  IAssetInstanceService,
} from '../../api/asset_instance_service';
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
  GetDefaultResourcesRequest,
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
  assetSpinRecord: string;
  defaultAssetResources: AssetResourceModel[];
  fetchResourceStatus: string;
  recordValidation: AssetRecordValidation;
  showDefaultMachines: boolean;
}

export interface AssetRecordValidation {
  nameValid: boolean;
  descriptionValid: boolean;
  resourceIdValid: boolean[];
  aliasNameValid: boolean[];
  aliasNameUnique: boolean[];
}

export const AssetRecordValidation = {
  defaultEntity(): AssetRecordValidation {
    return {
      nameValid: true,
      descriptionValid: true,
      resourceIdValid: [true],
      aliasNameValid: [true],
      aliasNameUnique: [true],
    };
  },
};

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
  assetSpinRecord: '',
  defaultAssetResources: [],
  fetchResourceStatus: 'idle',
  recordValidation: AssetRecordValidation.defaultEntity(),
  showDefaultMachines: false,
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
      assetId: assetId,
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
      assetId: assetId,
    };
    const service: IAssetService = new AssetService();
    const response = await service.delete(request);
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

export const createAssetInstanceAsync = createAsyncThunk(
  'asset/createAssetInstance',
  async (assetId: string) => {
    const request: CreateAssetInstanceRequest = {
      assetId: assetId,
      status: 'STATUS_PENDING',
    };
    const service: IAssetInstanceService = new AssetInstanceService();
    service.create(request);
  }
);

export const getDefaultResources = createAsyncThunk(
  'asset/getDefaultResources',
  async (assetType: string) => {
    const request: GetDefaultResourcesRequest = {
      assetType: assetType,
    };
    const service: IAssetService = new AssetService();
    const response = await service.getDefaultResources(request);
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
      if (state.record.name === '') {
        state.recordValidation.nameValid = false;
      } else {
        state.recordValidation.nameValid = true;
      }
    },
    setDescription: (state, action) => {
      state.record.description = action.payload;
      if (state.record.description === '') {
        state.recordValidation.descriptionValid = false;
      } else {
        state.recordValidation.descriptionValid = true;
      }
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
      state.defaultAssetResources = [];
      state.assetResourcesToSave = [AssetResourceModel.defaultEntity()];
      state.assetResourcesToDelete = [];
      state.recordValidation = AssetRecordValidation.defaultEntity();
      state.showDefaultMachines = false;
    },
    addMachine: (state) => {
      state.assetResourcesToSave = [
        ...state.assetResourcesToSave,
        AssetResourceModel.defaultEntity(),
      ];
      state.recordValidation.resourceIdValid.push(true);
      state.recordValidation.aliasNameValid.push(true);
      state.recordValidation.aliasNameUnique.push(true);
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
        state.recordValidation.resourceIdValid = state.recordValidation.resourceIdValid.filter(
          (_, index) => index !== action.payload
        );
        state.recordValidation.aliasNameValid = state.recordValidation.aliasNameValid.filter(
          (_, index) => index !== action.payload
        );
        state.recordValidation.aliasNameUnique = state.recordValidation.aliasNameUnique.filter(
          (_, index) => index !== action.payload
        );
      }
    },
    setResourceId: (state, action) => {
      state.assetResourcesToSave[action.payload.id].resourceId =
        action.payload.value;
      if (state.assetResourcesToSave[action.payload.id].resourceId === '') {
        state.recordValidation.resourceIdValid[action.payload.id] = false;
      } else {
        state.recordValidation.resourceIdValid[action.payload.id] = true;
      }
    },
    setAlias: (state, action) => {
      if (
        state.assetResourcesToSave.some(
          (assetResource) => assetResource.aliasName === action.payload.value
        ) ||
        state.defaultAssetResources.some(
          (assetResource) => assetResource.aliasName === action.payload.value
        )
      ) {
        state.recordValidation.aliasNameUnique[action.payload.id] = false;
      } else {
        state.recordValidation.aliasNameUnique[action.payload.id] = true;
      }
      state.assetResourcesToSave[action.payload.id].aliasName =
        action.payload.value;
      if (state.assetResourcesToSave[action.payload.id].aliasName === '') {
        state.recordValidation.aliasNameValid[action.payload.id] = false;
      } else {
        state.recordValidation.aliasNameValid[action.payload.id] = true;
      }
    },
    setState: (state, action) => {
      return action.payload;
    },
    setAssetSpinRecord: (state, action) => {
      state.assetSpinRecord = action.payload;
    },
    setNameValidFalse: (state) => {
      state.recordValidation.nameValid = false;
    },
    setDescriptionValidFalse: (state) => {
      state.recordValidation.descriptionValid = false;
    },
    setResourceIdValidFalse: (state, action) => {
      state.recordValidation.resourceIdValid[action.payload.index] = false;
    },
    setAliasNameValidFalse: (state, action) => {
      state.recordValidation.aliasNameValid[action.payload.index] = false;
    },
    changeShowDefaultMachines: (state, action) => {
      state.showDefaultMachines = action.payload;
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
        state.assets = [action.payload.asset, ...state.assets];
        state.assetResourcesToSave = action.payload.assetResources.filter(
          (entity) => entity.default === false
        );
        state.defaultAssetResources = action.payload.assetResources.filter(
          (entity) => entity.default === true
        );
        state.assetResourcesToDelete = [];
      })
      .addCase(updateAssetAsync.pending, (state) => {
        state.savingStatus = 'loading';
      })
      .addCase(updateAssetAsync.fulfilled, (state, action) => {
        state.savingStatus = 'idle';
        state.record = action.payload.asset;
        state.assets[
          state.assets.findIndex(function (asset: AssetModel) {
            return asset.assetId === action.payload.asset.assetId;
          })
        ] = action.payload.asset;
        state.assetResourcesToSave = action.payload.assetResources.filter(
          (entity) => entity.default === false
        );
        state.defaultAssetResources = action.payload.assetResources.filter(
          (entity) => entity.default === true
        );
        state.assetResourcesToDelete = [];
      })
      .addCase(queryAssetAsync.pending, (state) => {
        state.fetchAssetStatus = 'loading';
      })
      .addCase(queryAssetAsync.fulfilled, (state, action) => {
        state.fetchAssetStatus = 'idle';
        state.assets = action.payload.assets.filter(
          (asset) => asset.deleted === false
        );
        state.pageToken = action.payload.nextPageToken;
      })
      .addCase(deleteAssetAsync.pending, (state) => {
        state.deletingStatus = 'loading';
      })
      .addCase(deleteAssetAsync.fulfilled, (state) => {
        state.deletingStatus = 'idle';
        state.assets = state.assets.filter(
          (asset: AssetModel) => asset.assetId !== state.record.assetId
        );
        state.record = AssetModel.defaultEntity();
        state.assetResourcesToDelete = [];
        state.assetResourcesToSave = [AssetResourceModel.defaultEntity()];
      })
      .addCase(queryAssetResourceAsync.pending, (state) => {
        state.fetchAssetResourceStatus = 'loading';
      })
      .addCase(queryAssetResourceAsync.fulfilled, (state, action) => {
        state.fetchAssetResourceStatus = 'idle';
        state.assetResourcesToSave = [
          ...action.payload.assetResources.filter(
            (entity) =>
              entity.assetId === state.record.assetId &&
              entity.default === false
          ),
        ];
        if (state.assetResourcesToSave.length == 0) {
          state.assetResourcesToSave = [AssetResourceModel.defaultEntity()];
        }
        state.defaultAssetResources = [
          ...action.payload.assetResources.filter(
            (entity) =>
              entity.assetId === state.record.assetId && entity.default === true
          ),
        ];
        state.recordValidation.resourceIdValid = new Array(
          state.assetResourcesToSave.length
        ).fill(true);
        state.recordValidation.aliasNameValid = new Array(
          state.assetResourcesToSave.length
        ).fill(true);
        state.recordValidation.aliasNameUnique = new Array(
          state.assetResourcesToSave.length
        ).fill(true);
      })
      .addCase(queryResourceAsync.fulfilled, (state, action) => {
        state.resources = action.payload.resources.filter(
          (resource) => resource.deleted === false
        );
      })
      .addCase(createAssetInstanceAsync.pending, (state) => {
        state.savingStatus = 'loading';
      })
      .addCase(createAssetInstanceAsync.fulfilled, (state) => {
        state.savingStatus = 'idle';
        state.assetSpinRecord = '';
      })
      .addCase(getDefaultResources.pending, (state) => {
        state.fetchResourceStatus = 'loading';

        state.defaultAssetResources.forEach(function (
          assetResource: AssetResourceModel
        ) {
          // Need to delete this default AssetResource if it is already created
          if (assetResource.assetResourceId !== '') {
            state.assetResourcesToDelete = [
              ...state.assetResourcesToDelete,
              assetResource,
            ];
          }
        });
        state.defaultAssetResources = [];
      })
      .addCase(getDefaultResources.fulfilled, (state, action) => {
        state.fetchResourceStatus = 'idle';
        action.payload.assetResources.forEach(function (
          assetRes: AssetResourceModel
        ) {
          const assetResource: AssetResourceModel = AssetResourceModel.defaultEntity();
          assetResource.resourceId = assetRes.resourceId;
          assetResource.aliasName = assetRes.aliasName;
          assetResource.default = assetRes.default;
          state.defaultAssetResources = [
            assetResource,
            ...state.defaultAssetResources,
          ];
        });
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
  setState,
  setAssetSpinRecord,
  setNameValidFalse,
  setDescriptionValidFalse,
  setResourceIdValidFalse,
  setAliasNameValidFalse,
  changeShowDefaultMachines,
} = assetSlice.actions;

export default assetSlice.reducer;
