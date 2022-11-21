// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import {
  CreateResourceRequest,
  UpdateResourceRequest,
  DeleteResourceRequest,
  GetResourceRequest,
  IResourceService,
  ResourceService,
  ListResourcesRequest,
  ResourceModel,
} from '../../api/resource_service';
import { RootState } from '../../app/store';

export interface ResourceState {
  resources: ResourceModel[];
  pageToken: string | undefined;
  record: ResourceModel;
  fetchStatus: string;
  savingStatus: string;
  deletingStatus: string;
  pageNumber: number;
  pageSize: number;
  recordValidation: ResourceRecordValidation;
}

export interface ResourceRecordValidation {
  nameValid: boolean;
  nameUnique: boolean;
  descriptionValid: boolean;
  operatingSystemValid: boolean;
  imageProjectValid: boolean;
  imageFamilyValid: boolean;
  imageSourceValid: boolean;
  typeValid: boolean;
}

export const ResourceRecordValidation = {
  defaultEntity(): ResourceRecordValidation {
    return {
      nameValid: true,
      nameUnique: true,
      descriptionValid: true,
      operatingSystemValid: true,
      imageProjectValid: true,
      imageFamilyValid: true,
      imageSourceValid: true,
      typeValid: true,
    };
  },
};

const initialState: ResourceState = {
  resources: [],
  pageToken: undefined,
  pageNumber: 1,
  pageSize: 25,
  fetchStatus: 'idle',
  record: ResourceModel.defaultEntity(),
  savingStatus: 'idle',
  deletingStatus: 'idle',
  recordValidation: ResourceRecordValidation.defaultEntity(),
};

// The function below is called a thunk and allows us to perform async logic. It
// can be dispatched like a regular action: `dispatch(fetchResourceAsync(10))`. This
// will call the thunk with the `dispatch` function as the first argument. Async
// code can then be executed and other actions can be dispatched. Thunks are
// typically used to make async requests.
export const fetchResourceAsync = createAsyncThunk(
  'resource/fetchResource',
  async (resourceId: string) => {
    const request: GetResourceRequest = {
      resourceId: resourceId,
    };
    const service: IResourceService = new ResourceService();
    const response = await service.get(request);
    // The value we return becomes the `fulfilled` action payload
    return response;
  }
);

export const createResourceAsync = createAsyncThunk(
  'resource/createResource',
  async ({
    name,
    type,
    operatingSystem,
    description,
    imageProject,
    imageFamily,
    imageSource,
  }: {
    name: string;
    type: string;
    operatingSystem: string;
    description: string;
    imageProject: string;
    imageFamily: string;
    imageSource: string;
  }) => {
    const request: CreateResourceRequest = {
      name,
      type,
      operatingSystem,
      description,
      imageProject,
      imageFamily,
      imageSource,
    };
    const service: IResourceService = new ResourceService();
    const response = await service.create(request);
    return response;
  }
);

export const updateResourceAsync = createAsyncThunk(
  'asset/updateResource',
  async ({
    resource,
    updateMask,
  }: {
    resource: ResourceModel;
    updateMask: string[];
  }) => {
    const request: UpdateResourceRequest = {
      resource: resource,
      updateMask: updateMask,
    };
    const service: IResourceService = new ResourceService();
    const response = await service.update(request);
    return response;
  }
);

export const queryResourceAsync = createAsyncThunk(
  'resource/queryResource',
  async ({ pageSize, pageToken }: { pageSize: number; pageToken: string }) => {
    const request: ListResourcesRequest = {
      pageSize: pageSize,
      pageToken: pageToken,
      readMask: undefined,
    };
    const service: IResourceService = new ResourceService();
    const response = await service.list(request);
    return response;
  }
);

export const deleteResourceAsync = createAsyncThunk(
  'resource/deleteResource',
  async (resourceId: string) => {
    const request: DeleteResourceRequest = {
      resourceId: resourceId,
    };
    const service: IResourceService = new ResourceService();
    const response = await service.delete(request);
    return response;
  }
);

export const resourceSlice = createSlice({
  name: 'resource',
  initialState,
  reducers: {
    setPageSize: (state, action) => {
      state.pageSize = action.payload.pageSize;
    },
    setName: (state, action) => {
      if (
        state.resources.some((resource) => resource.name === action.payload)
      ) {
        state.recordValidation.nameUnique = false;
      } else {
        state.recordValidation.nameUnique = true;
      }
      state.record.name = action.payload;
      if (state.record.name === '') {
        state.recordValidation.nameValid = false;
      } else {
        state.recordValidation.nameValid = true;
      }
    },
    setType: (state, action) => {
      state.record.type = action.payload;
    },
    setOperatingSystem: (state, action) => {
      state.record.operatingSystem = action.payload;
      if (
        (state.record.type === 'ad_joined_machine' || state.record.type === 'machine') &&
        state.record.operatingSystem === ''
      ) {
        state.recordValidation.operatingSystemValid = false;
      } else {
        state.recordValidation.operatingSystemValid = true;
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
    setImageProject: (state, action) => {
      state.record.imageProject = action.payload;
      if ((state.record.type === 'ad_joined_machine' || state.record.type === 'machine') && state.record.imageProject === '') {
        state.recordValidation.imageProjectValid = false;
      } else {
        state.recordValidation.imageProjectValid = true;
      }
    },
    setImageFamily: (state, action) => {
      state.record.imageFamily = action.payload;
      if ((state.record.type === 'ad_joined_machine' || state.record.type === 'machine') && state.record.imageFamily === '') {
        state.recordValidation.imageFamilyValid = false;
      } else {
        state.recordValidation.imageFamilyValid = true;
      }
    },
    setImageSource: (state, action) => {
      state.record.imageSource = action.payload;
      if (state.record.type === 'custom_image_machine' && state.record.imageSource === '') {
        state.recordValidation.imageSourceValid = false;
      } else {
        state.recordValidation.imageSourceValid = true;
      }
    },
    onSelectRecord: (state, action) => {
      state.record = state.resources.filter(
        (s) => s.resourceId == action.payload.resourceId
      )[0];
    },
    clearSelectedRecord: (state) => {
      state.record = ResourceModel.defaultEntity();
      state.recordValidation = ResourceRecordValidation.defaultEntity();
    },
    setState: (state, action) => {
      return action.payload;
    },
    setDefaultState: () => {
      return initialState;
    },
    setNameValidFalse: (state) => {
      state.recordValidation.nameValid = false;
    },
    setDescriptionValidFalse: (state) => {
      state.recordValidation.descriptionValid = false;
    },
    setOperatingSystemValidFalse: (state) => {
      state.recordValidation.operatingSystemValid = false;
    },
    setImageProjectValidFalse: (state) => {
      state.recordValidation.imageProjectValid = false;
    },
    setImageFamilyValidFalse: (state) => {
      state.recordValidation.imageFamilyValid = false;
    },
    setImageSourceValidFalse: (state) => {
      state.recordValidation.imageSourceValid = false;
    },
    setTypeValidFalse: (state) => {
      state.recordValidation.typeValid = false;
    },
  },

  // The `extraReducers` field lets the slice handle actions generated by
  // createAsyncThunk or in other slices.
  extraReducers: (builder) => {
    builder
      .addCase(fetchResourceAsync.pending, (state) => {
        state.fetchStatus = 'loading';
      })
      .addCase(fetchResourceAsync.fulfilled, (state, action) => {
        state.fetchStatus = 'idle';
        state.record = action.payload;
      })
      .addCase(createResourceAsync.pending, (state) => {
        state.savingStatus = 'loading';
      })
      .addCase(createResourceAsync.fulfilled, (state, action) => {
        state.savingStatus = 'idle';
        state.record = action.payload;
        state.resources = [action.payload, ...state.resources];
      })
      .addCase(updateResourceAsync.pending, (state) => {
        state.savingStatus = 'loading';
      })
      .addCase(updateResourceAsync.fulfilled, (state, action) => {
        state.savingStatus = 'idle';
        state.record = action.payload;
        state.resources[
          state.resources.findIndex(function (resource: ResourceModel) {
            return resource.resourceId === action.payload.resourceId;
          })
        ] = action.payload;
      })
      .addCase(queryResourceAsync.pending, (state) => {
        state.fetchStatus = 'loading';
      })
      .addCase(queryResourceAsync.fulfilled, (state, action) => {
        state.fetchStatus = 'idle';
        state.resources = action.payload.resources.filter(
          (resource) => resource.deleted === false
        );
        state.pageToken = action.payload.nextPageToken;
        if (
          state.record.name !== '' &&
          state.resources.some(
            (resource: ResourceModel) => resource.name === state.record.name
          )
        ) {
          state.recordValidation.nameUnique = false;
        }
      })
      .addCase(deleteResourceAsync.pending, (state) => {
        state.deletingStatus = 'loading';
      })
      .addCase(deleteResourceAsync.fulfilled, (state) => {
        state.deletingStatus = 'idle';
        state.resources = state.resources.filter(
          (resource: ResourceModel) =>
            resource.resourceId !== state.record.resourceId
        );
        state.record = ResourceModel.defaultEntity();
      });
  },
});

// The function below is called a selector and allows us to select a value from
// the state. Selectors can also be defined inline where they're used instead of
// in the slice file. For example: `useSelector((state: RootState) => state.resource)`
export const selectResourceState = (state: RootState) => state.resource;

export const {
  setPageSize,
  onSelectRecord,
  clearSelectedRecord,
  setName,
  setType,
  setOperatingSystem,
  setDescription,
  setImageFamily,
  setImageProject,
  setImageSource,
  setDefaultState,
  setState,
  setNameValidFalse,
  setDescriptionValidFalse,
  setOperatingSystemValidFalse,
  setImageFamilyValidFalse,
  setImageProjectValidFalse,
  setImageSourceValidFalse,
  setTypeValidFalse,
} = resourceSlice.actions;

export default resourceSlice.reducer;
