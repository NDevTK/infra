// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Empty } from './common/empty';
import { FieldMask } from './common/field_mask';
import { rpcClient } from './common/rpc_client';
import { fromJsonTimestamp, isSet } from './common/utils';

/** Performs operations on AssetResourceResources. */
export interface IAssetResourceService {
  /** Creates the given AssetResource. */
  create(request: CreateAssetResourceRequest): Promise<AssetResourceModel>;
  /** Retrieves a AssetResourceResource for a given unique value. */
  get(request: GetAssetResourceRequest): Promise<AssetResourceModel>;
  /** Update a single AssetResource in poros. */
  update(request: UpdateAssetResourceRequest): Promise<AssetResourceModel>;
  /** Deletes the given AssetResource. */
  delete(request: DeleteAssetResourceRequest): Promise<Empty>;
  /** Lists all AssetResources. */
  list(request: ListAssetResourcesRequest): Promise<ListAssetResourcesResponse>;
}

export class AssetResourceService implements IAssetResourceService {
  constructor() {
    this.create = this.create.bind(this);
    this.get = this.get.bind(this);
    this.update = this.update.bind(this);
    this.delete = this.delete.bind(this);
    this.list = this.list.bind(this);
  }

  create = (
    request: CreateAssetResourceRequest
  ): Promise<AssetResourceModel> => {
    const data = CreateAssetResourceRequest.toJSON(request);
    const promise = rpcClient.request('poros.AssetResource', 'Create', data);
    return promise.then((data) => AssetResourceModel.fromJSON(data));
  };

  get = (request: GetAssetResourceRequest): Promise<AssetResourceModel> => {
    const data = GetAssetResourceRequest.toJSON(request);
    const promise = rpcClient.request('poros.AssetResource', 'Get', data);
    return promise.then((data) => AssetResourceModel.fromJSON(data));
  };

  update = (
    request: UpdateAssetResourceRequest
  ): Promise<AssetResourceModel> => {
    const data = UpdateAssetResourceRequest.toJSON(request);
    const promise = rpcClient.request('poros.AssetResource', 'Update', data);
    return promise.then((data) => AssetResourceModel.fromJSON(data));
  };

  delete = (request: DeleteAssetResourceRequest): Promise<Empty> => {
    const data = DeleteAssetResourceRequest.toJSON(request);
    const promise = rpcClient.request('poros.AssetResource', 'Delete', data);
    return promise.then((data) => Empty.fromJSON(data));
  };

  list = (
    request: ListAssetResourcesRequest
  ): Promise<ListAssetResourcesResponse> => {
    const data = ListAssetResourcesRequest.toJSON(request);
    const promise = rpcClient.request('poros.AssetResource', 'List', data);
    return promise.then((data) => ListAssetResourcesResponse.fromJSON(data));
  };
}

export interface AssetResourceModel {
  // Unique identifier of the entity
  assetResourceId: string;
  // Identifier of the asset associated with the entity
  assetId: string;
  // Identifier of the resource associated with the entity
  resourceId: string;
  // Alias name of the entity
  aliasName: string;
  // User who created the record.
  createdBy: string;
  // Timestamp for the creation of the record.
  createdAt: Date | undefined;
  // Timestamp for the last update of the record.
  modifiedAt: Date | undefined;
  // User who modified the record.
  modifiedBy: string;
  // Flag to denote whether a given AssetResource is default
  default: boolean;
}

export const AssetResourceModel = {
  defaultEntity(): AssetResourceModel {
    return {
      assetResourceId: '',
      assetId: '',
      resourceId: '',
      aliasName: '',
      createdBy: '',
      createdAt: undefined,
      modifiedBy: '',
      modifiedAt: undefined,
      default: false,
    };
  },
  fromJSON(object: any): AssetResourceModel {
    return {
      assetResourceId: isSet(object.assetResourceId)
        ? String(object.assetResourceId)
        : '',
      assetId: isSet(object.assetId) ? String(object.assetId) : '',
      resourceId: isSet(object.resourceId) ? String(object.resourceId) : '',
      aliasName: isSet(object.aliasName) ? String(object.aliasName) : '',
      createdBy: isSet(object.createdBy) ? String(object.createdBy) : '',
      createdAt: isSet(object.createdAt)
        ? fromJsonTimestamp(object.createdAt)
        : undefined,
      modifiedBy: isSet(object.modifiedBy) ? String(object.modifiedBy) : '',
      modifiedAt: isSet(object.modifiedAt)
        ? fromJsonTimestamp(object.modifiedAt)
        : undefined,
      default: isSet(object.default) ? Boolean(object.default) : false,
    };
  },

  toJSON(message: AssetResourceModel): unknown {
    const obj: any = {};
    message.assetResourceId !== undefined &&
      (obj.assetResourceId = message.assetResourceId);
    message.assetId !== undefined && (obj.assetId = message.assetId);
    message.resourceId !== undefined && (obj.resourceId = message.resourceId);
    message.aliasName !== undefined && (obj.aliasName = message.aliasName);
    message.createdBy !== undefined && (obj.createdBy = message.createdBy);
    message.createdAt !== undefined &&
      (obj.createdAt = message.createdAt.toISOString());
    message.modifiedBy !== undefined && (obj.modifiedBy = message.modifiedBy);
    message.modifiedAt !== undefined &&
      (obj.modifiedAt = message.modifiedAt.toISOString());
    message.default !== undefined && (obj.default = message.default);
    return obj;
  },
};

/** Request to create a single AssetResource in AssetResourceServ */
export interface CreateAssetResourceRequest {
  // Identifier of the asset associated with the entity
  assetId: string;
  // Identifier of the resource associated with the entity
  resourceId: string;
  // Alias name of the entity
  aliasName: string;
  // Flag to denote whether a given AssetResource is default
  default: boolean;
}

export const CreateAssetResourceRequest = {
  toJSON(message: CreateAssetResourceRequest): unknown {
    const obj: any = {};
    message.assetId !== undefined && (obj.assetId = message.assetId);
    message.resourceId !== undefined && (obj.resourceId = message.resourceId);
    message.aliasName !== undefined && (obj.aliasName = message.aliasName);
    message.default !== undefined && (obj.default = message.default);
    return obj;
  },
};

// Request to delete a single AssetResource from poros.
export interface DeleteAssetResourceRequest {
  // Unique identifier for the asset resource entity
  assetResourceId: string;
}

export const DeleteAssetResourceRequest = {
  toJSON(message: DeleteAssetResourceRequest): unknown {
    const obj: any = {};
    message.assetResourceId !== undefined &&
      (obj.assetResourceId = message.assetResourceId);
    return obj;
  },
};

/** Gets a AssetResource resource. */
export interface GetAssetResourceRequest {
  // The id of the AssetResource to retrieve.
  assetResourceId: string;
}

export const GetAssetResourceRequest = {
  toJSON(message: GetAssetResourceRequest): unknown {
    const obj: any = {};
    message.assetResourceId !== undefined && (obj.id = message.assetResourceId);
    return obj;
  },
};

/** Request to list all AssetResources in poros. */
export interface ListAssetResourcesRequest {
  /** Fields to include on each result */
  readMask: string[] | undefined;
  /** Number of results per page */
  pageSize: number;
  /** Page token from a previous page's ListAssetResourcesResponse */
  pageToken: string;
}

/** Response to ListAssetResourcesRequest. */
export interface ListAssetResourcesResponse {
  /** The result set. */
  assetResources: AssetResourceModel[];
  /**
   * A page token that can be passed into future requests to get the next page.
   * Empty if there is no next page.
   */
  nextPageToken: string;
}

export const ListAssetResourcesRequest = {
  toJSON(message: ListAssetResourcesRequest): unknown {
    const obj: any = {};
    message.readMask !== undefined &&
      (obj.readMask = FieldMask.toJSON(FieldMask.wrap(message.readMask)));
    message.pageSize !== undefined &&
      (obj.pageSize = Math.round(message.pageSize));
    message.pageToken !== undefined && (obj.pageToken = message.pageToken);
    return obj;
  },
};

export const ListAssetResourcesResponse = {
  fromJSON(object: any): ListAssetResourcesResponse {
    return {
      assetResources: Array.isArray(object?.assetResources)
        ? object.assetResources.map((e: any) => AssetResourceModel.fromJSON(e))
        : [],
      nextPageToken: isSet(object.nextPageToken)
        ? String(object.nextPageToken)
        : '',
    };
  },
};

/** Request to update a single AssetResource in poros. */
export interface UpdateAssetResourceRequest {
  /** The existing AssetResource to update in the database. */
  assetResource: AssetResourceModel | undefined;
  /** The list of fields to update. */
  updateMask: string[] | undefined;
}

export const UpdateAssetResourceRequest = {
  toJSON(message: UpdateAssetResourceRequest): unknown {
    const obj: any = {};
    message.assetResource !== undefined &&
      (obj.assetResource = message.assetResource
        ? AssetResourceModel.toJSON(message.assetResource)
        : undefined);
    message.updateMask !== undefined &&
      (obj.updateMask = FieldMask.toJSON(FieldMask.wrap(message.updateMask)));
    return obj;
  },
};
