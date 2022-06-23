// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { FieldMask } from './common/field_mask';
import { rpcClient } from './common/rpc_client';
import { fromJsonTimestamp, isSet } from './common/utils';

/** Performs operations on AssetInstance. */
export interface IAssetInstanceService {
  /** Lists all AssetInstances. */
  list(request: ListAssetInstancesRequest): Promise<ListAssetInstancesResponse>;
}

export class AssetInstanceService implements IAssetInstanceService {
  constructor() {
    this.list = this.list.bind(this);
  }

  list = (
    request: ListAssetInstancesRequest
  ): Promise<ListAssetInstancesResponse> => {
    const data = ListAssetInstancesRequest.toJSON(request);
    const promise = rpcClient.request('poros.AssetInstance', 'List', data);
    return promise.then((data) => ListAssetInstancesResponse.fromJSON(data));
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
  /** Timestamp for the creation of the record */
  createdAt: Date | undefined;
  /** User who modified the record */
  modifiedBy: string;
  /** Timestamp for the last update of the record */
  modifiedAt: Date | undefined;
}

export const AssetInstanceModel = {
  defaultEntity(): AssetInstanceModel {
    return {
      assetInstanceId: '',
      assetId: '',
      status: '',
      createdBy: '',
      createdAt: undefined,
      modifiedBy: '',
      modifiedAt: undefined,
    };
  },
  fromJSON(object: any): AssetInstanceModel {
    return {
      assetInstanceId: isSet(object.assetInstanceId)
        ? String(object.assetInstanceId)
        : '',
      assetId: isSet(object.assetId) ? String(object.assetId) : '',
      status: isSet(object.status)
        ? String(object.status)
        : '',
      createdBy: isSet(object.createdBy) ? String(object.createdBy) : '',
      createdAt: isSet(object.createdAt)
        ? fromJsonTimestamp(object.createdAt)
        : undefined,
      modifiedBy: isSet(object.modifiedBy) ? String(object.modifiedBy) : '',
      modifiedAt: isSet(object.modifiedAt)
        ? fromJsonTimestamp(object.modifiedAt)
        : undefined,
    };
  },

  toJSON(message: AssetInstanceModel): unknown {
    const obj: any = {};
    message.assetInstanceId !== undefined &&
      (obj.assetInstanceId = message.assetInstanceId);
    message.assetId !== undefined && (obj.assetId = message.assetId);
    message.status !== undefined &&
      (obj.status = message.status);
    message.createdBy !== undefined && (obj.createdBy = message.createdBy);
    message.createdAt !== undefined &&
      (obj.createdAt = message.createdAt.toISOString());
    message.modifiedBy !== undefined && (obj.modifiedBy = message.modifiedBy);
    message.modifiedAt !== undefined &&
      (obj.modifiedAt = message.modifiedAt.toISOString());
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
