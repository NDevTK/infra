// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { PrpcClient } from '@chopsui/prpc-client';
import { obtainAuthState } from '../libs/auth_state';


export class AuthorizedPrpcClient {
    client: PrpcClient;

    constructor() {
        // Only allow insecure connections in local development, where risk of
        // man-in-the-middle attack to server is negligable.
        const insecure = document.location.hostname === "localhost" && document.location.protocol === "http:";
        this.client = new PrpcClient({
            insecure: insecure
        });
    }

    async call(service: string, method: string, message: object, additionalHeaders?: {
        [key: string]: string;
    } | undefined): Promise<any> {
        // Although PrpcClient allows us to pass a token to the constructor,
        // we prefer to inject it at request time to ensure the most recent
        // token is used.
        const authState = await obtainAuthState();
        additionalHeaders = {
            Authorization: 'Bearer ' + authState.accessToken,
            ...additionalHeaders,
        };
        return this.client.call(service, method, message, additionalHeaders);
    }
}
