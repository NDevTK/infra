// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { FieldMask } from './common/field_mask';
import { rpcClient } from './common/rpc_client';
import { fromJsonTimestamp, isSet } from './common/utils';

/** Performs operations on AssetInstance. */
export interface IAssetInstanceService {
  /** Creates the given AssetInstance. */
  create(request: CreateAssetInstanceRequest): Promise<AssetInstanceModel>;
  /** Update the requested AssetInstanceModel. */
  update(request: UpdateAssetInstanceRequest): Promise<AssetInstanceModel>;
  /** Lists all AssetInstances. */
  list(request: ListAssetInstancesRequest): Promise<ListAssetInstancesResponse>;
  /** Fetch the latest deployment logs associated with asset instance id */
  fetchLogs(request: FetchLogsRequest): Promise<FetchLogsResponse>;
}

export class AssetInstanceService implements IAssetInstanceService {
  constructor() {
    this.create = this.create.bind(this);
    this.update = this.update.bind(this);
    this.list = this.list.bind(this);
    this.fetchLogs = this.fetchLogs.bind(this);
  }

  create = (
    request: CreateAssetInstanceRequest
  ): Promise<AssetInstanceModel> => {
    const data = CreateAssetInstanceRequest.toJSON(request);
    const promise = rpcClient.request('poros.AssetInstance', 'Create', data);
    return promise.then((data) => AssetInstanceModel.fromJSON(data));
  };

  update = (
    request: UpdateAssetInstanceRequest
  ): Promise<AssetInstanceModel> => {
    const data = UpdateAssetInstanceRequest.toJSON(request);
    const promise = rpcClient.request('poros.AssetInstance', 'Update', data);
    return promise.then((data) => AssetInstanceModel.fromJSON(data));
  };

  list = (
    request: ListAssetInstancesRequest
  ): Promise<ListAssetInstancesResponse> => {
    const data = ListAssetInstancesRequest.toJSON(request);
    const promise = rpcClient.request('poros.AssetInstance', 'List', data);
    return promise.then((data) => ListAssetInstancesResponse.fromJSON(data));
  };

  fetchLogs = (request: FetchLogsRequest): Promise<FetchLogsResponse> => {
    const data = FetchLogsRequest.toJSON(request);
    const promise = rpcClient.request('poros.AssetInstance', 'FetchLogs', data);
    return promise.then((data) => FetchLogsResponse.fromJSON(data));
  };
}

export interface AssetInstanceModel {
  /** Unique identifier of the AssetInstance */
  assetInstanceId: string;
  /** Name of the AssetInstance */
  assetId: string;
  /** Type of the AssetInstance */
  status: string;
  /** User who created the record */
  createdBy: string;
  /** GCP Project on which project will be deployed */
  projectId: string;
  /** Timestamp for the creation of the record */
  createdAt: Date | undefined;
  /** User who modified the record */
  modifiedBy: string;
  /** Timestamp for the last update of the record */
  modifiedAt: Date | undefined;
  /** Timestamp to delete the machines */
  deleteAt: Date | undefined;
}

export const AssetInstanceModel = {
  defaultEntity(): AssetInstanceModel {
    return {
      assetInstanceId: '',
      assetId: '',
      status: '',
      createdBy: '',
      projectId: '',
      createdAt: undefined,
      modifiedBy: '',
      modifiedAt: undefined,
      deleteAt: undefined,
    };
  },
  fromJSON(object: any): AssetInstanceModel {
    return {
      assetInstanceId: isSet(object.assetInstanceId)
        ? String(object.assetInstanceId)
        : '',
      assetId: isSet(object.assetId) ? String(object.assetId) : '',
      status: isSet(object.status) ? String(object.status) : '',
      createdBy: isSet(object.createdBy) ? String(object.createdBy) : '',
      createdAt: isSet(object.createdAt)
        ? fromJsonTimestamp(object.createdAt)
        : undefined,
      projectId: isSet(object.projectId) ? String(object.projectId) : '',
      modifiedBy: isSet(object.modifiedBy) ? String(object.modifiedBy) : '',
      modifiedAt: isSet(object.modifiedAt)
        ? fromJsonTimestamp(object.modifiedAt)
        : undefined,
      deleteAt: isSet(object.deleteAt)
        ? fromJsonTimestamp(object.deleteAt)
        : undefined,
    };
  },

  toJSON(message: AssetInstanceModel): unknown {
    const obj: any = {};
    message.assetInstanceId !== undefined &&
      (obj.assetInstanceId = message.assetInstanceId);
    message.assetId !== undefined && (obj.assetId = message.assetId);
    message.status !== undefined && (obj.status = message.status);
    message.createdBy !== undefined && (obj.createdBy = message.createdBy);
    message.createdAt !== undefined &&
      (obj.createdAt = message.createdAt.toISOString());
    message.modifiedBy !== undefined && (obj.modifiedBy = message.modifiedBy);
    message.modifiedAt !== undefined &&
      (obj.modifiedAt = message.modifiedAt.toISOString());
    message.deleteAt !== undefined &&
      (obj.deleteAt = message.deleteAt.toISOString());
    return obj;
  },
};

/** Request to create a single AssetInstance in AssetInstanceServ */
export interface CreateAssetInstanceRequest {
  // Identifier of the asset associated with the entity
  assetId: string;
  // Status of the instance
  status: string;
}

export const CreateAssetInstanceRequest = {
  toJSON(message: CreateAssetInstanceRequest): unknown {
    const obj: any = {};
    message.assetId !== undefined && (obj.assetId = message.assetId);
    message.status !== undefined && (obj.status = message.status);
    return obj;
  },
};

/** Request to list all AssetInstances in poros. */
export interface ListAssetInstancesRequest {
  /** Fields to include on each result */
  readMask: string[] | undefined;
  /** Number of results per page */
  pageSize: number;
  /** Page token from a previous page's ListAssetInstancesResponse */
  pageToken: string;
}

/** Response to ListAssetInstancesRequest. */
export interface ListAssetInstancesResponse {
  /** The result set. */
  assetInstances: AssetInstanceModel[];
  /**
   * A page token that can be passed into future requests to get the next page.
   * Empty if there is no next page.
   */
  nextPageToken: string;
}

/** Request to fetch deployment logs for an asset instance */
export interface FetchLogsRequest {
  assetInstanceId: string;
}

/** Response to FetchLogsRequest. */
export interface FetchLogsResponse {
  logs: string;
}

export const ListAssetInstancesRequest = {
  toJSON(message: ListAssetInstancesRequest): unknown {
    const obj: any = {};
    message.readMask !== undefined &&
      (obj.readMask = FieldMask.toJSON(FieldMask.wrap(message.readMask)));
    message.pageSize !== undefined &&
      (obj.pageSize = Math.round(message.pageSize));
    message.pageToken !== undefined && (obj.pageToken = message.pageToken);
    return obj;
  },
};

export const ListAssetInstancesResponse = {
  fromJSON(object: any): ListAssetInstancesResponse {
    return {
      assetInstances: Array.isArray(object?.assetInstances)
        ? object.assetInstances.map((e: any) => AssetInstanceModel.fromJSON(e))
        : [],
      nextPageToken: isSet(object.nextPageToken)
        ? String(object.nextPageToken)
        : '',
    };
  },
};

export const FetchLogsRequest = {
  toJSON(message: FetchLogsRequest): unknown {
    const obj: any = {};
    message.assetInstanceId !== undefined &&
      (obj.assetInstanceId = message.assetInstanceId);
    return obj;
  },
};

export const FetchLogsResponse = {
  fromJSON(object: any): FetchLogsResponse {
    return {
      logs: isSet(object.logs) ? object.logs : '',
    };
  },
};

/** Request to update a single AssetInstance in poros. */
export interface UpdateAssetInstanceRequest {
  /** The existing AssetInstance to update in the database. */
  assetInstance: AssetInstanceModel | undefined;
  /** The list of fields to update. */
  updateMask: string[] | undefined;
}

export const UpdateAssetInstanceRequest = {
  toJSON(message: UpdateAssetInstanceRequest): unknown {
    const obj: any = {};
    message.assetInstance !== undefined &&
      (obj.assetInstance = message.assetInstance
        ? AssetInstanceModel.toJSON(message.assetInstance)
        : undefined);
    message.updateMask !== undefined &&
      (obj.updateMask = FieldMask.toJSON(FieldMask.wrap(message.updateMask)));
    return obj;
  },
};
