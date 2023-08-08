// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { AssetResourceModel } from './asset_resource_service';
import { Empty } from './common/empty';
import { FieldMask } from './common/field_mask';
import { rpcClient } from './common/rpc_client';
import { fromJsonTimestamp, isSet } from './common/utils';
import { ResourceModel } from './resource_service';

/** Performs operations on Assets. */
export interface IAssetService {
  /** Creates the given Asset. */
  create(request: CreateAssetRequest): Promise<CreateAssetResponse>;
  /** Retrieves a Asset for a given unique value. */
  get(request: GetAssetRequest): Promise<AssetModel>;
  /** Update a single asset in poros. */
  update(request: UpdateAssetRequest): Promise<UpdateAssetResponse>;
  /** Deletes the given Asset. */
  delete(request: DeleteAssetRequest): Promise<Empty>;
  /** Lists all Assets. */
  list(request: ListAssetsRequest): Promise<ListAssetsResponse>;
  /** Get default resourced given asset type. */
  getDefaultResources(
    request: GetDefaultResourcesRequest
  ): Promise<GetDefaultResourcesResponse>;
}

export class AssetService implements IAssetService {
  constructor() {
    this.create = this.create.bind(this);
    this.get = this.get.bind(this);
    this.update = this.update.bind(this);
    this.delete = this.delete.bind(this);
    this.list = this.list.bind(this);
    this.getDefaultResources = this.getDefaultResources.bind(this);
  }

  create = (request: CreateAssetRequest): Promise<CreateAssetResponse> => {
    const data = CreateAssetRequest.toJSON(request);
    const promise = rpcClient.request('poros.Asset', 'Create', data);
    return promise.then((data) => CreateAssetResponse.fromJSON(data));
  };

  get = (request: GetAssetRequest): Promise<AssetModel> => {
    const data = GetAssetRequest.toJSON(request);
    const promise = rpcClient.request('poros.Asset', 'Get', data);
    return promise.then((data) => AssetModel.fromJSON(data));
  };

  update = (request: UpdateAssetRequest): Promise<UpdateAssetResponse> => {
    const data = UpdateAssetRequest.toJSON(request);
    const promise = rpcClient.request('poros.Asset', 'Update', data);
    return promise.then((data) => UpdateAssetResponse.fromJSON(data));
  };

  delete = (request: DeleteAssetRequest): Promise<Empty> => {
    const data = DeleteAssetRequest.toJSON(request);
    const promise = rpcClient.request('poros.Asset', 'Delete', data);
    return promise.then((data) => Empty.fromJSON(data));
  };

  list = (request: ListAssetsRequest): Promise<ListAssetsResponse> => {
    const data = ListAssetsRequest.toJSON(request);
    const promise = rpcClient.request('poros.Asset', 'List', data);
    return promise.then((data) => ListAssetsResponse.fromJSON(data));
  };

  getDefaultResources = (
    request: GetDefaultResourcesRequest
  ): Promise<GetDefaultResourcesResponse> => {
    const data = GetDefaultResourcesRequest.toJSON(request);
    const promise = rpcClient.request(
      'poros.Asset',
      'GetDefaultResources',
      data
    );
    return promise.then((data) => GetDefaultResourcesResponse.fromJSON(data));
  };
}

export interface AssetModel {
  /** Unique identifier of the asset */
  assetId: string;
  /** Name of the asset */
  name: string;
  /** Description of the asset */
  description: string;
  /** Type of the asset (active_directory, etc) */
  assetType: string;
  /** User who created the record */
  createdBy: string;
  /** Timestamp for the creation of the record */
  createdAt: Date | undefined;
  /** User who modified the record */
  modifiedBy: string;
  /** Timestamp for the last update of the record */
  modifiedAt: Date | undefined;
  /** Flag to denote whether this Asset is deleted */
  deleted: boolean;
}

export const AssetModel = {
  defaultEntity(): AssetModel {
    return {
      assetId: '',
      name: '',
      description: '',
      assetType: 'active_directory',
      createdBy: '',
      createdAt: undefined,
      modifiedBy: '',
      modifiedAt: undefined,
      deleted: false,
    };
  },
  fromJSON(object: any): AssetModel {
    return {
      assetId: isSet(object.assetId) ? String(object.assetId) : '',
      name: isSet(object.name) ? String(object.name) : '',
      description: isSet(object.description) ? String(object.description) : '',
      assetType: isSet(object.assetType) ? String(object.assetType) : '',
      createdBy: isSet(object.createdBy) ? String(object.createdBy) : '',
      createdAt: isSet(object.createdAt)
        ? fromJsonTimestamp(object.createdAt)
        : undefined,
      modifiedBy: isSet(object.modifiedBy) ? String(object.modifiedBy) : '',
      modifiedAt: isSet(object.modifiedAt)
        ? fromJsonTimestamp(object.modifiedAt)
        : undefined,
      deleted: isSet(object.deleted) ? Boolean(object.deleted) : false,
    };
  },

  toJSON(message: AssetModel): unknown {
    const obj: any = {};
    message.assetId !== undefined && (obj.assetId = message.assetId);
    message.name !== undefined && (obj.name = message.name);
    message.description !== undefined &&
      (obj.description = message.description);
    message.assetType !== undefined && (obj.assetType = message.assetType);
    message.createdBy !== undefined && (obj.createdBy = message.createdBy);
    message.createdAt !== undefined &&
      (obj.createdAt = message.createdAt.toISOString());
    message.modifiedBy !== undefined && (obj.modifiedBy = message.modifiedBy);
    message.modifiedAt !== undefined &&
      (obj.modifiedAt = message.modifiedAt.toISOString());
    message.deleted !== undefined && (obj.deleted = message.deleted);
    return obj;
  },
};

/** Request to create a single asset in AssetServ */
export interface CreateAssetRequest {
  /** Name of the asset */
  name: string;
  /** Description of the asset */
  description: string;
  /** Type of the asset (active_directory, etc) */
  assetType: string;
  /** List of asset resource to create/update */
  assetResourcesToSave: AssetResourceModel[];
}

export interface CreateAssetResponse {
  /** Asset created */
  asset: AssetModel;
  /** List of AssetResources saved */
  assetResources: AssetResourceModel[];
}

export const CreateAssetRequest = {
  toJSON(message: CreateAssetRequest): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    message.description !== undefined &&
      (obj.description = message.description);
    message.assetType !== undefined && (obj.assetType = message.assetType);
    message.assetResourcesToSave !== undefined &&
      (obj.assetResourcesToSave = message.assetResourcesToSave);
    return obj;
  },
};

export const CreateAssetResponse = {
  fromJSON(object: any): CreateAssetResponse {
    return {
      asset: AssetModel.fromJSON(object.asset),
      assetResources: Array.isArray(object?.assetResources)
        ? object.assetResources.map((e: any) => AssetResourceModel.fromJSON(e))
        : [],
    };
  },
};

// Request to delete a single asset from poros.
export interface DeleteAssetRequest {
  /** Unique identifier of the asset */
  assetId: string;
}

export const DeleteAssetRequest = {
  toJSON(message: DeleteAssetRequest): unknown {
    const obj: any = {};
    message.assetId !== undefined && (obj.assetId = message.assetId);
    return obj;
  },
};

/** Gets a Asset resource. */
export interface GetAssetRequest {
  /**
   * The name of the asset to retrieve.
   * Format: publishers/{publisher}/assets/{asset}
   */
  assetId: string;
}

export const GetAssetRequest = {
  toJSON(message: GetAssetRequest): unknown {
    const obj: any = {};
    message.assetId !== undefined && (obj.assetId = message.assetId);
    return obj;
  },
};

/** Request to list all assets in poros. */
export interface ListAssetsRequest {
  /** Fields to include on each result */
  readMask: string[] | undefined;
  /** Number of results per page */
  pageSize: number;
  /** Page token from a previous page's ListAssetsResponse */
  pageToken: string;
}

/** Response to ListAssetsRequest. */
export interface ListAssetsResponse {
  /** The result set. */
  assets: AssetModel[];
  /**
   * A page token that can be passed into future requests to get the next page.
   * Empty if there is no next page.
   */
  nextPageToken: string;
}

export const ListAssetsRequest = {
  toJSON(message: ListAssetsRequest): unknown {
    const obj: any = {};
    message.readMask !== undefined &&
      (obj.readMask = FieldMask.toJSON(FieldMask.wrap(message.readMask)));
    message.pageSize !== undefined &&
      (obj.pageSize = Math.round(message.pageSize));
    message.pageToken !== undefined && (obj.pageToken = message.pageToken);
    return obj;
  },
};

export const ListAssetsResponse = {
  fromJSON(object: any): ListAssetsResponse {
    return {
      assets: Array.isArray(object?.assets)
        ? object.assets.map((e: any) => AssetModel.fromJSON(e))
        : [],
      nextPageToken: isSet(object.nextPageToken)
        ? String(object.nextPageToken)
        : '',
    };
  },
};

/** Request to update a single asset in poros. */
export interface UpdateAssetRequest {
  /** The existing asset to update in the database. */
  asset: AssetModel | undefined;
  /** The list of fields to update Asset. */
  assetUpdateMask: string[] | undefined;
  /** The list of fields to update the asset_resource. */
  assetResourceUpdateMask: string[] | undefined;
  /** The list of fields to update the asset_resource. */
  assetResourcesToSave: AssetResourceModel[];
  /** The list of fields to update the asset_resource. */
  assetResourcesToDelete: AssetResourceModel[];
}

export interface UpdateAssetResponse {
  /** Asset updated */
  asset: AssetModel;
  /** List of AssetResources saved/updated */
  assetResources: AssetResourceModel[];
}

export const UpdateAssetRequest = {
  toJSON(message: UpdateAssetRequest): unknown {
    const obj: any = {};
    message.asset !== undefined &&
      (obj.asset = message.asset
        ? AssetModel.toJSON(message.asset)
        : undefined);
    message.assetUpdateMask !== undefined &&
      (obj.assetUpdateMask = FieldMask.toJSON(
        FieldMask.wrap(message.assetUpdateMask)
      ));
    message.assetResourceUpdateMask !== undefined &&
      (obj.assetResourceUpdateMask = FieldMask.toJSON(
        FieldMask.wrap(message.assetResourceUpdateMask)
      ));
    obj.assetResourcesToSave = message.assetResourcesToSave;
    message.assetResourcesToDelete !== undefined &&
      (obj.assetResourcesToDelete = message.assetResourcesToDelete);
    return obj;
  },
};

export const UpdateAssetResponse = {
  fromJSON(object: any): CreateAssetResponse {
    return {
      asset: AssetModel.fromJSON(object.asset),
      assetResources: Array.isArray(object?.assetResources)
        ? object.assetResources.map((e: any) => AssetResourceModel.fromJSON(e))
        : [],
    };
  },
};

/** Request to get the default Resources given Asset type */
export interface GetDefaultResourcesRequest {
  /** The type of the given Asset. */
  assetType: string;
}

export interface GetDefaultResourcesResponse {
  /** List of default Resources*/
  assetResources: AssetResourceModel[];
}

export const GetDefaultResourcesRequest = {
  toJSON(message: GetDefaultResourcesRequest): unknown {
    const obj: any = {};
    message.assetType !== undefined && (obj.assetType = message.assetType);
    return obj;
  },
};

export const GetDefaultResourcesResponse = {
  fromJSON(object: any): GetDefaultResourcesResponse {
    return {
      assetResources: Array.isArray(object?.assetResources)
        ? object.assetResources.map((e: any) => AssetResourceModel.fromJSON(e))
        : [],
    };
  },
};
